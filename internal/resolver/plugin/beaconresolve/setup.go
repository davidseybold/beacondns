package beaconresolve

import (
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/davidseybold/beacondns/internal/recursive"
)

//nolint:gochecknoinits // used for plugin registration
func init() {
	plugin.Register(pluginName, setup)
}

func setup(c *caddy.Controller) error {
	beacon, err := beaconParse(c)
	if err != nil {
		return plugin.Error(pluginName, err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		beacon.Next = next
		return beacon
	})

	c.OnStartup(beacon.OnStartup)
	c.OnFinalShutdown(beacon.OnFinalShutdown)

	return nil
}

func (b *BeaconResolve) OnStartup() error {
	blog.Info("starting beaconresolve plugin")

	resolver, err := recursive.NewResolver(nil)
	if err != nil {
		return fmt.Errorf("error creating resolver: %w", err)
	}
	b.resolver = resolver

	return nil
}

func (b *BeaconResolve) OnFinalShutdown() error {
	blog.Info("shutting down beaconresolve plugin")
	return nil
}

func beaconParse(c *caddy.Controller) (*BeaconResolve, error) {
	if c.Next() {
		if c.NextBlock() {
			return parseAttributes(c)
		}
	}
	return &BeaconResolve{}, nil
}

func parseAttributes(_ *caddy.Controller) (*BeaconResolve, error) {
	beacon := &BeaconResolve{}

	beacon.config = Config{}

	return beacon, nil
}
