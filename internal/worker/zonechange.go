package worker

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/dnsserializer"
	"github.com/davidseybold/beacondns/internal/model"
)

func (w *Worker) processZoneChange(ctx context.Context, change *model.Change) error {
	ops := make([]kvstore.Op, 0)

	zoneChange := change.ZoneChange

	if zoneChange.Action == model.ZoneChangeActionCreate {
		ops = append(ops, kvstore.Op{
			Action: kvstore.ActionPut,
			Key:    createZoneKey(zoneChange.ZoneName),
			Value:  []byte(zoneChange.ZoneName),
		})
	} else if zoneChange.Action == model.ZoneChangeActionDelete {
		ops = append(ops, kvstore.Op{
			Action: kvstore.ActionDelete,
			Key:    createZoneKey(zoneChange.ZoneName),
		})
	}

	for _, ch := range zoneChange.Changes {
		op, err := processResourceRecordSetChange(zoneChange.ZoneName, ch)
		if err != nil {
			return err
		}

		if op != nil {
			ops = append(ops, *op)
		}
	}

	err := w.kv.Txn(ctx, ops)
	if err != nil {
		return err
	}

	return nil
}

func processResourceRecordSetChange(zoneName string, change model.ResourceRecordSetChange) (*kvstore.Op, error) {
	if change.Action == model.RRSetChangeActionDelete {
		return &kvstore.Op{
			Action: kvstore.ActionDelete,
			Key:    createResourceRecordKey(zoneName, change.ResourceRecordSet.Name, change.ResourceRecordSet.Type),
		}, nil
	}

	rrset := newDNSRecordSet(change.ResourceRecordSet)

	dnsResp := dnsserializer.RRSet{
		RRs: rrset,
	}

	packed, err := dnsserializer.MarshalRRSet(&dnsResp)
	if err != nil {
		return nil, err
	}

	if change.Action == model.RRSetChangeActionCreate || change.Action == model.RRSetChangeActionUpsert {
		return &kvstore.Op{
			Action: kvstore.ActionPut,
			Key:    createResourceRecordKey(zoneName, change.ResourceRecordSet.Name, change.ResourceRecordSet.Type),
			Value:  packed,
		}, nil
	}

	return nil, nil
}

func createResourceRecordKey(zoneName string, rrName string, recordType model.RRType) string {
	dnsRType := dns.StringToType[string(recordType)]
	dnsRTypeStr := dns.TypeToString[uint16(dnsRType)]
	return fmt.Sprintf("/zone/%s/recordset/%s/%s", zoneName, rrName, dnsRTypeStr)
}

func createZoneKey(zone string) string {
	return fmt.Sprintf("/zones/%s", zone)
}

func newDNSRecordSet(rrset model.ResourceRecordSet) []dns.RR {
	switch rrset.Type {
	case model.RRTypeA:
		return newARecordSet(rrset)
	case model.RRTypeAAAA:
		return newAAAARecordSet(rrset)
	case model.RRTypeCNAME:
		return newCNAMERecordSet(rrset)
	case model.RRTypeSOA:
		return newSOARecordSet(rrset)
	case model.RRTypeNS:
		return newNSRecordSet(rrset)
	}

	return nil
}

func newARecordSet(rrset model.ResourceRecordSet) []dns.RR {
	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
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
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs
}

func newNSRecordSet(rrset model.ResourceRecordSet) []dns.RR {
	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.NS)
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rrset.Name),
			Rrtype: dns.TypeNS,
			Class:  dns.ClassINET,
			Ttl:    rrset.TTL,
		}
		r.Ns = dns.Fqdn(rr.Value)
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs
}

func newCNAMERecordSet(rrset model.ResourceRecordSet) []dns.RR {
	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
	for _, rr := range rrset.ResourceRecords {
		r := new(dns.CNAME)
		r.Hdr = dns.RR_Header{
			Name:   dns.Fqdn(rrset.Name),
			Rrtype: dns.TypeCNAME,
			Class:  dns.ClassINET,
			Ttl:    rrset.TTL,
		}
		r.Target = dns.Fqdn(rr.Value)
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs
}

func newSOARecordSet(rrset model.ResourceRecordSet) []dns.RR {
	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
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

		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs
}

func newAAAARecordSet(rrset model.ResourceRecordSet) []dns.RR {
	dnsRRs := make([]dns.RR, 0, len(rrset.ResourceRecords))
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
		dnsRRs = append(dnsRRs, r)
	}

	return dnsRRs
}
