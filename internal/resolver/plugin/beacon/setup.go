package beacon

import (
	"context"
	"fmt"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/dnsstore"
)

//nolint:gochecknoinits // used for plugin registration
func init() {
	plugin.Register("beacon", setup)
}

func setup(c *caddy.Controller) error {
	beacon, err := beaconParse(c)
	if err != nil {
		return plugin.Error("beacon", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		beacon.Next = next
		return beacon
	})

	c.OnStartup(beacon.OnStartup)
	c.OnFinalShutdown(beacon.OnFinalShutdown)

	return nil
}

func (b *Beacon) OnStartup() error {
	blog.Info("starting beacon")

	etcdClient, err := kvstore.NewEtcdClient(b.config.EtcdEndpoints, kvstore.Scope{
		Namespace: "beacon",
	})
	if err != nil {
		return fmt.Errorf("error creating etcd client: %w", err)
	}

	b.store = dnsstore.New(etcdClient)

	watchCtx, watchCancel := context.WithCancel(context.Background())

	zoneEventsCh, err := b.store.SubscribeToZoneEvents(watchCtx)
	if err != nil {
		watchCancel()
		return fmt.Errorf("error watching zones: %w", err)
	}

	go b.listenForZoneChanges(watchCtx, zoneEventsCh)

	b.close = func() error {
		watchCancel()

		if storeErr := etcdClient.Close(); storeErr != nil {
			return fmt.Errorf("error closing etcd client: %w", storeErr)
		}

		return nil
	}

	err = b.loadZones()
	if err != nil {
		return fmt.Errorf("error loading zones: %w", err)
	}

	return nil
}

func (b *Beacon) OnFinalShutdown() error {
	blog.Info("shutting down beacon")
	return b.close()
}

func beaconParse(c *caddy.Controller) (*Beacon, error) {
	if c.Next() {
		if c.NextBlock() {
			return parseAttributes(c)
		}
	}
	return &Beacon{}, nil
}

func parseAttributes(c *caddy.Controller) (*Beacon, error) {
	beacon := &Beacon{
		zoneTrie: NewZoneTrie(),
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
