package beacon

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/dnsserializer"
)

//nolint:gochecknoglobals // used for logging
var log = clog.NewWithPlugin("beacon")

type BeaconConfig struct {
	EtcdEndpoints []string
}

type Beacon struct {
	Next plugin.Handler

	config BeaconConfig

	store    kvstore.KVStore
	zoneTrie *ZoneTrie
	close    func() error
	logger   *slog.Logger
}

var _ plugin.Handler = (*Beacon)(nil)

func (b *Beacon) lookup(zoneName, rrName string, t dns.Type) ([]dns.RR, bool) {
	log.Info(createRecordKey(zoneName, rrName, t.String()))
	val, err := b.store.Get(context.Background(), createRecordKey(zoneName, rrName, t.String()))
	if err != nil && err == kvstore.ErrNotFound {
		log.Info("record not found", " zone ", zoneName, " rrName ", rrName, " type ", t)
		return nil, false
	} else if err != nil {
		log.Error("error getting record ", " error ", err)
		return nil, false
	}

	rrset, err := dnsserializer.UnmarshalRRSet(val)
	if err != nil {
		log.Error("error unmarshalling dns response", "error", err)
		return nil, false
	}

	return rrset.RRs, true
}

func (b *Beacon) listenForZoneChanges(ctx context.Context, ch <-chan kvstore.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			log.Info("zone change detected")
			if event.Type == kvstore.EventTypePut {
				b.zoneTrie.AddZone(event.Key)
			} else if event.Type == kvstore.EventTypeDelete {
				b.zoneTrie.RemoveZone(event.Key)
			}
		}
	}
}

func (b *Beacon) loadZones() error {
	log.Info("loading zones")
	items, err := b.store.GetPrefix(context.Background(), "/zones")
	if err != nil {
		return fmt.Errorf("error getting zones: %w", err)
	}

	log.Info("loading zones ", " count ", len(items))

	for _, item := range items {
		log.Info("loading zone key ", string(item.Key))
		b.zoneTrie.AddZone(string(item.Value))
	}
	return nil
}

func createRecordKey(zoneName, rrName string, rrType string) string {
	return fmt.Sprintf("/zone/%s/recordset/%s/%s", zoneName, rrName, rrType)
}
