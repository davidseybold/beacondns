package beaconauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/dnsstore"
)

//nolint:gochecknoglobals // used for logging
var blog = clog.NewWithPlugin("beacon")

type Config struct {
	EtcdEndpoints []string
}

type BeaconAuth struct {
	Next plugin.Handler

	config Config

	store    dnsstore.DNSStore
	zoneTrie *DNTrie
	close    func() error
}

var _ plugin.Handler = (*BeaconAuth)(nil)

func (b *BeaconAuth) Name() string { return "beaconauth" }

func (b *BeaconAuth) lookup(zoneName, rrName string, t dns.Type) ([]dns.RR, bool) {
	val, err := b.store.GetRRSet(context.Background(), zoneName, rrName, t.String())
	if err != nil && errors.Is(err, dnsstore.ErrRRSetNotFound) {
		return nil, false
	} else if err != nil {
		blog.Errorf("error looking up rrset %s for zone %s: %s", rrName, zoneName, err.Error())
		return nil, false
	}

	return val, true
}

func (b *BeaconAuth) listenForZoneChanges(ctx context.Context, ch <-chan dnsstore.ZoneEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			switch event.Type {
			case dnsstore.ZoneEventTypeCreate:
				b.zoneTrie.Insert(event.Zone)
			case dnsstore.ZoneEventTypeDelete:
				b.zoneTrie.Remove(event.Zone)
			}
		}
	}
}

func (b *BeaconAuth) loadZones() error {
	zoneNames, err := b.store.GetAllZoneNames(context.Background())
	if err != nil {
		return fmt.Errorf("error getting zones: %w", err)
	}

	for _, zoneName := range zoneNames {
		b.zoneTrie.Insert(zoneName)
	}

	return nil
}
