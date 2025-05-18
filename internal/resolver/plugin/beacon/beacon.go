package beacon

import (
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/miekg/dns"
	"google.golang.org/protobuf/proto"

	"github.com/davidseybold/beacondns/internal/convert"
	"github.com/davidseybold/beacondns/internal/db/kvstore"
	beacondnspb "github.com/davidseybold/beacondns/internal/gen/proto/beacondns/v1"
	"github.com/davidseybold/beacondns/internal/model"
)

//nolint:gochecknoglobals // used for logging
var log = clog.NewWithPlugin("beacon")

type Beacon struct {
	Next plugin.Handler

	hostName           string
	dbPath             string
	rabbitmqConnString string
	rabbitmqExchange   string
	changeQueue        string

	store    kvstore.KVStore
	zoneTrie *ZoneTrie
	close    func() error
	logger   *slog.Logger
}

var _ plugin.Handler = (*Beacon)(nil)

func (b *Beacon) A(rrset *model.ResourceRecordSet) ([]dns.RR, []dns.RR) {
	if rrset == nil {
		return nil, nil
	}

	answers := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rrset.Name),
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rrset.TTL,
		}
		ip := net.ParseIP(rr.Value)
		if ip == nil || ip.To4() == nil {
			continue
		}

		r.A = ip
		answers = append(answers, r)
	}
	return answers, nil
}

func (b *Beacon) CNAME(rrset *model.ResourceRecordSet) ([]dns.RR, []dns.RR) {
	if rrset == nil {
		return nil, nil
	}

	answers := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rrset.Name),
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    rrset.TTL,
		}
		r.Target = dns.Fqdn(rr.Value)
		answers = append(answers, r)
	}
	return answers, nil
}

func (b *Beacon) MX(_ string) ([]dns.RR, []dns.RR) {
	return nil, nil
}

func (b *Beacon) NS(rrset *model.ResourceRecordSet) ([]dns.RR, []dns.RR) {
	if rrset == nil {
		return nil, nil
	}

	answers := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.NS)
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rrset.Name),
			Rrtype: dns.TypeNS,
			Class:  dns.ClassINET,
			Ttl:    rrset.TTL,
		}
		r.Ns = dns.Fqdn(rr.Value)
		answers = append(answers, r)
	}
	return answers, nil
}

func (b *Beacon) TXT(_ *model.ResourceRecordSet) ([]dns.RR, []dns.RR) {
	return nil, nil
}

func (b *Beacon) PTR(_ *model.ResourceRecordSet) ([]dns.RR, []dns.RR) {
	return nil, nil
}

func (b *Beacon) SRV(_ *model.ResourceRecordSet) ([]dns.RR, []dns.RR) {
	return nil, nil
}

func (b *Beacon) SOA(rrset *model.ResourceRecordSet) ([]dns.RR, []dns.RR) {
	if rrset == nil {
		return nil, nil
	}

	answers := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.SOA)
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rrset.Name),
			Rrtype: dns.TypeSOA,
			Class:  dns.ClassINET,
			Ttl:    rrset.TTL,
		}

		parts := strings.Fields(rr.Value)
		if len(parts) != 7 {
			continue
		}

		r.Ns = dns.Fqdn(parts[0])
		r.Mbox = dns.Fqdn(parts[1])
		serial, _ := strconv.ParseUint(parts[2], 10, 32)
		r.Serial = uint32(serial)
		refresh, _ := strconv.ParseInt(parts[3], 10, 32)
		r.Refresh = uint32(refresh)
		retry, _ := strconv.ParseInt(parts[4], 10, 32)
		r.Retry = uint32(retry)
		expire, _ := strconv.ParseInt(parts[5], 10, 32)
		r.Expire = uint32(expire)
		minttl, _ := strconv.ParseUint(parts[6], 10, 32)
		r.Minttl = uint32(minttl)

		answers = append(answers, r)
	}
	return answers, nil
}

func (b *Beacon) AAAA(rrset *model.ResourceRecordSet) ([]dns.RR, []dns.RR) {
	if rrset == nil {
		return nil, nil
	}

	answers := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rrset.Name),
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rrset.TTL,
		}
		ip := net.ParseIP(rr.Value)
		if ip == nil || ip.To16() == nil {
			continue
		}
		r.AAAA = ip
		answers = append(answers, r)
	}
	return answers, nil
}

func (b *Beacon) handleZoneChange(ch *model.ZoneChange) error {
	switch ch.Action {
	case model.ZoneChangeActionDelete:
		return b.deleteZone(ch.ZoneName)
	case model.ZoneChangeActionCreate:
		return b.createZone(ch.ZoneName, ch.Changes)
	case model.ZoneChangeActionUpdate:
		return b.updateZone(ch.ZoneName, ch.Changes)
	}
	return nil
}

func (b *Beacon) deleteZone(zoneName string) error {
	err := b.store.DeletePrefix([]byte(createZonePrefix(zoneName)))
	if err != nil {
		return fmt.Errorf("error deleting zone: %w", err)
	}

	b.zoneTrie.RemoveZone(zoneName)

	return nil
}

func (b *Beacon) createZone(zoneName string, rrChanges []model.ResourceRecordSetChange) error {
	changes, err := b.makeRecordChanges(zoneName, rrChanges)
	if err != nil {
		return fmt.Errorf("error making record changes: %w", err)
	}

	fmt.Printf("changes: %+v\n", changes)
	zonesKey := []byte(createZonesKey(zoneName))
	changes = append(changes, kvstore.Change{
		Action: kvstore.ActionPut,
		Key:    zonesKey,
		Value:  []byte(zoneName),
	})

	err = b.store.BatchChange(changes)
	if err != nil {
		return fmt.Errorf("error applying changes: %w", err)
	}

	b.zoneTrie.AddZone(zoneName)

	return nil
}

func (b *Beacon) updateZone(zoneName string, rrChanges []model.ResourceRecordSetChange) error {
	changes, err := b.makeRecordChanges(zoneName, rrChanges)
	if err != nil {
		return fmt.Errorf("error making record changes: %w", err)
	}

	err = b.store.BatchChange(changes)
	if err != nil {
		return fmt.Errorf("error applying changes: %w", err)
	}

	return nil
}

func (b *Beacon) makeRecordChanges(zoneName string, changes []model.ResourceRecordSetChange) ([]kvstore.Change, error) {
	dbChanges := make([]kvstore.Change, 0, len(changes))
	for _, change := range changes {
		switch change.Action {
		case model.RRSetChangeActionCreate, model.RRSetChangeActionUpsert:
			pbRRSet := convert.ResourceRecordSetToProto(&change.ResourceRecordSet)
			buf, err := proto.Marshal(pbRRSet)
			if err != nil {
				return nil, fmt.Errorf("error marshalling resource record set: %w", err)
			}

			dbChanges = append(dbChanges, kvstore.Change{
				Action: kvstore.ActionPut,
				Key:    []byte(createRecordKey(zoneName, change.ResourceRecordSet.Name, change.ResourceRecordSet.Type)),
				Value:  buf,
			})
		case model.RRSetChangeActionDelete:
			dbChanges = append(dbChanges, kvstore.Change{
				Action: kvstore.ActionDelete,
				Key:    []byte(createRecordKey(zoneName, change.ResourceRecordSet.Name, change.ResourceRecordSet.Type)),
			})
		}
	}

	return dbChanges, nil
}
func (b *Beacon) lookup(zoneName, rrName string, rrType model.RRType) (*model.ResourceRecordSet, bool) {
	buf, err := b.store.Get([]byte(createRecordKey(zoneName, rrName, rrType)))
	if err != nil {
		return nil, false
	}

	if buf == nil {
		return nil, false
	}

	var pbRRSet beacondnspb.ResourceRecordSet
	err = proto.Unmarshal(buf, &pbRRSet)
	if err != nil {
		return nil, false
	}

	return convert.ResourceRecordSetFromProto(&pbRRSet), true
}

func createRecordKey(zoneName, rrName string, rrType model.RRType) string {
	return fmt.Sprintf("/zone/%s/records/%s/%s", zoneName, rrName, rrType)
}

func createZonePrefix(zoneName string) string {
	return fmt.Sprintf("/zone/%s", zoneName)
}

func createZonesKey(zoneName string) string {
	return fmt.Sprintf("/zones/%s", zoneName)
}
