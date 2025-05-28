package beacon

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

type Beacon struct {
	Next plugin.Handler

	config Config

	store    dnsstore.DNSStore
	zoneTrie *ZoneTrie
	close    func() error
}

var _ plugin.Handler = (*Beacon)(nil)

func (b *Beacon) Name() string { return "beacon" }

func (b *Beacon) lookup(zoneName, rrName string, t dns.Type) ([]dns.RR, bool) {
	val, err := b.store.GetRRSet(context.Background(), zoneName, rrName, t.String())
	if err != nil && errors.Is(err, dnsstore.ErrRRSetNotFound) {
		return nil, false
	} else if err != nil {
		blog.Errorf("error looking up rrset %s for zone %s: %s", rrName, zoneName, err.Error())
		return nil, false
	}

	return val, true
}

func (b *Beacon) listenForZoneChanges(ctx context.Context, ch <-chan dnsstore.ZoneEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			switch event.Type {
			case dnsstore.ZoneEventTypeCreate:
				b.zoneTrie.AddZone(event.Zone)
			case dnsstore.ZoneEventTypeDelete:
				b.zoneTrie.RemoveZone(event.Zone)
			}
		}
	}
}

func (b *Beacon) loadZones() error {
	zoneNames, err := b.store.GetAllZoneNames(context.Background())
	if err != nil {
		return fmt.Errorf("error getting zones: %w", err)
	}

	for _, zoneName := range zoneNames {
		b.zoneTrie.AddZone(zoneName)
	}

	return nil
}
