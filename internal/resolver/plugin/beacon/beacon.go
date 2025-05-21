package beacon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/dnsstore"
)

//nolint:gochecknoglobals // used for logging
var log = clog.NewWithPlugin("beacon")

type Config struct {
	EtcdEndpoints []string
}

type Beacon struct {
	Next plugin.Handler

	config Config

	store    dnsstore.DNSStore
	zoneTrie *ZoneTrie
	close    func() error
	logger   *slog.Logger
}

var _ plugin.Handler = (*Beacon)(nil)

func (b *Beacon) lookup(zoneName, rrName string, t dns.Type) ([]dns.RR, bool) {
	val, err := b.store.GetRRSet(context.Background(), zoneName, rrName, t.String())
	if err != nil && errors.Is(err, dnsstore.ErrNotFound) {
		log.Info("record not found", " zone ", zoneName, " rrName ", rrName, " type ", t)
		return nil, false
	} else if err != nil {
		log.Error("error getting record ", " error ", err)
		return nil, false
	}

	return val, true
}

func (b *Beacon) listenForZoneChanges(ctx context.Context, ch <-chan kvstore.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			switch event.Type {
			case kvstore.EventTypePut:
				b.zoneTrie.AddZone(event.Key)
			case kvstore.EventTypeDelete:
				b.zoneTrie.RemoveZone(event.Key)
			}
		}
	}
}

func (b *Beacon) loadZones() error {
	zoneNames, err := b.store.GetZoneNames(context.Background())
	if err != nil {
		return fmt.Errorf("error getting zones: %w", err)
	}

	for _, zoneName := range zoneNames {
		b.zoneTrie.AddZone(zoneName)
	}

	return nil
}
