package sessionresolver_test

import (
	"context"
	"testing"

	"github.com/ooni/probe-cli/v3/internal/engine/internal/sessionresolver"
)

func TestSessionResolverGood(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	reso := &sessionresolver.Resolver{}
	defer reso.CloseIdleConnections()
	if reso.Network() != "sessionresolver" {
		t.Fatal("unexpected Network")
	}
	if reso.Address() != "" {
		t.Fatal("unexpected Address")
	}
	addrs, err := reso.LookupHost(context.Background(), "google.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(addrs) < 1 {
		t.Fatal("expected some addrs here")
	}
}
