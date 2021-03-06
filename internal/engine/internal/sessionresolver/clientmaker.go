package sessionresolver

import "github.com/ooni/probe-cli/v3/internal/engine/netx"

// dnsclientmaker makes a new resolver.
type dnsclientmaker interface {
	// Make makes a new resolver.
	Make(config netx.Config, URL string) (childResolver, error)
}

// clientmaker returns a valid dnsclientmaker
func (r *Resolver) clientmaker() dnsclientmaker {
	if r.dnsClientMaker != nil {
		return r.dnsClientMaker
	}
	return &defaultDNSClientMaker{}
}

// defaultDNSClientMaker is the default dnsclientmaker
type defaultDNSClientMaker struct{}

// Make implements dnsclientmaker.Make.
func (*defaultDNSClientMaker) Make(config netx.Config, URL string) (childResolver, error) {
	return netx.NewDNSClient(config, URL)
}
