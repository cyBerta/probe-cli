package main

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ooni/probe-cli/v3/internal/engine/ooapi/internal/openapi"
)

const (
	tagForJSON = "json"
	tagForPath = "path"
)

func (d *Descriptor) genSwaggerURLPath() string {
	up := d.URLPath
	if up.InSwagger != "" {
		return up.InSwagger
	}
	if up.IsTemplate {
		panic("we should always use InSwapper and IsTemplate together")
	}
	return up.Value
}

func (d *Descriptor) genSwaggerSchema(cur reflect.Type) *openapi.Schema {
	switch cur.Kind() {
	case reflect.String:
		return &openapi.Schema{Type: "string"}
	case reflect.Bool:
		return &openapi.Schema{Type: "boolean"}
	case reflect.Int64:
		return &openapi.Schema{Type: "integer"}
	case reflect.Slice:
		return &openapi.Schema{Type: "array", Items: d.genSwaggerSchema(cur.Elem())}
	case reflect.Map:
		return &openapi.Schema{Type: "object"}
	case reflect.Ptr:
		return d.genSwaggerSchema(cur.Elem())
	case reflect.Struct:
		if cur.String() == "time.Time" {
			// Implementation note: we don't want to dive into time.Time but
			// rather we want to pretend it's a string. The JSON parser for
			// time.Time can indeed reconstruct a time.Time from a string, and
			// it's much easier for us to let it do the parsing.
			return &openapi.Schema{Type: "string"}
		}
		sinfo := &openapi.Schema{Type: "object"}
		var once sync.Once
		initmap := func() {
			sinfo.Properties = make(map[string]*openapi.Schema)
		}
		for idx := 0; idx < cur.NumField(); idx++ {
			field := cur.Field(idx)
			if field.Tag.Get(tagForPath) != "" {
				continue // skipping because this is a path param
			}
			if field.Tag.Get(tagForQuery) != "" {
				continue // skipping because this is a query param
			}
			v := field.Name
			if j := field.Tag.Get(tagForJSON); j != "" {
				j = strings.Replace(j, ",omitempty", "", 1) // remove options
				if j == "-" {
					continue // not exported via JSON
				}
				v = j
			}
			once.Do(initmap)
			sinfo.Properties[v] = d.genSwaggerSchema(field.Type)
		}
		return sinfo
	case reflect.Interface:
		return &openapi.Schema{Type: "object"}
	default:
		panic("unsupported type")
	}
}

func (d *Descriptor) swaggerParamForType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int64:
		return "integer"
	default:
		panic("unsupported type")
	}
}

func (d *Descriptor) genSwaggerParams(cur reflect.Type) []*openapi.Parameter {
	// when we have params the input must be a pointer to struct
	if cur.Kind() != reflect.Ptr {
		panic("not a pointer")
	}
	cur = cur.Elem()
	if cur.Kind() != reflect.Struct {
		panic("not a pointer to struct")
	}
	// now that we're sure of the type, inspect the fields
	var out []*openapi.Parameter
	for idx := 0; idx < cur.NumField(); idx++ {
		f := cur.Field(idx)
		if q := f.Tag.Get(tagForQuery); q != "" {
			out = append(
				out, &openapi.Parameter{
					Name:     q,
					In:       "query",
					Required: f.Tag.Get(tagForRequired) == "true",
					Type:     d.swaggerParamForType(f.Type),
				})
			continue
		}
		if p := f.Tag.Get(tagForPath); p != "" {
			out = append(out, &openapi.Parameter{
				Name:     p,
				In:       "path",
				Required: true,
				Type:     d.swaggerParamForType(f.Type),
			})
			continue
		}
	}
	return out
}

func (d *Descriptor) genSwaggerPath() (string, *openapi.Path) {
	pathStr, pathInfo := d.genSwaggerURLPath(), &openapi.Path{}
	rtinfo := &openapi.RoundTrip{Produces: []string{"application/json"}}
	switch d.Method {
	case "GET":
		pathInfo.Get = rtinfo
	case "POST":
		rtinfo.Consumes = append(rtinfo.Consumes, "application/json")
		pathInfo.Post = rtinfo
	default:
		panic("unsupported method")
	}
	rtinfo.Parameters = d.genSwaggerParams(reflect.TypeOf(d.Request))
	if d.Method != "GET" {
		rtinfo.Parameters = append(rtinfo.Parameters, &openapi.Parameter{
			Name:     "body",
			In:       "body",
			Required: true,
			Schema:   d.genSwaggerSchema(reflect.TypeOf(d.Request)),
		})
	}
	rtinfo.Responses = &openapi.Responses{Successful: openapi.Body{
		Description: "all good",
		Schema:      d.genSwaggerSchema(reflect.TypeOf(d.Response)),
	}}
	return pathStr, pathInfo
}

func genSwaggerVersion() string {
	return time.Now().UTC().Format("0.20060102.1150405")
}

// GenSwaggerTestGo generates swagger_test.go
func GenSwaggerTestGo(file string) {
	swagger := openapi.Swagger{
		Swagger: "2.0",
		Info: openapi.API{
			Title:   "OONI API specification",
			Version: genSwaggerVersion(),
		},
		Host:     "api.ooni.io",
		BasePath: "/",
		Schemes:  []string{"https"},
		Paths:    make(map[string]*openapi.Path),
	}
	for _, desc := range Descriptors {
		pathStr, pathInfo := desc.genSwaggerPath()
		swagger.Paths[pathStr] = pathInfo
	}
	data, err := json.MarshalIndent(swagger, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	var sb strings.Builder
	fmt.Fprint(&sb, "// Code generated by go generate; DO NOT EDIT.\n")
	fmt.Fprintf(&sb, "// %s\n\n", time.Now())
	fmt.Fprint(&sb, "package ooapi\n\n")
	fmt.Fprintf(&sb, "//go:generate go run ./internal/generator -file %s\n\n", file)
	fmt.Fprintf(&sb, "const swagger = `%s`\n", string(data))
	writefile(file, &sb)
}
