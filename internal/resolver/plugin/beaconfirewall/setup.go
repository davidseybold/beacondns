package beaconfirewall

import (
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/dnsstore"
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

func (b *BeaconFirewall) OnStartup() error {
	blog.Info("starting beaconfirewall plugin")

	etcdClient, err := kvstore.NewEtcdClient(b.config.EtcdEndpoints, kvstore.Scope{
		Namespace: "beacon",
	})
	if err != nil {
		return fmt.Errorf("error creating etcd client: %w", err)
	}

	b.store = dnsstore.New(etcdClient)

	b.close = func() error {
		if storeErr := etcdClient.Close(); storeErr != nil {
			return fmt.Errorf("error closing etcd client: %w", storeErr)
		}

		return nil
	}

	return nil
}

func (b *BeaconFirewall) OnFinalShutdown() error {
	blog.Info("shutting down beaconfirewall plugin")
	return b.close()
}

func beaconParse(c *caddy.Controller) (*BeaconFirewall, error) {
	if c.Next() {
		if c.NextBlock() {
			return parseAttributes(c)
		}
	}
	return &BeaconFirewall{}, nil
}

func parseAttributes(c *caddy.Controller) (*BeaconFirewall, error) {
	beacon := &BeaconFirewall{
		ruleLookup: NewDNTrie[ruleMeta](),
	}

	config := Config{}

	for {
		switch c.Val() {
		case "etcd_endpoints":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			endpoints := append([]string{c.Val()}, c.RemainingArgs()...)
			if len(endpoints) == 0 {
				return nil, c.Errf("etcd_endpoints requires at least one endpoint")
			}
			config.EtcdEndpoints = endpoints
		default:
			if c.Val() != "}" {
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
		if !c.Next() {
			break
		}
	}

	beacon.config = config

	return beacon, nil
}
