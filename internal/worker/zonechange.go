package worker

import (
	"context"
	"fmt"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/dnsstore"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/rrbuilder"
)

func (w *Worker) processZoneChange(ctx context.Context, change *model.Change) error {
	zoneChange := change.ZoneChange

	if zoneChange.Action == model.ZoneChangeActionDelete {
		return w.store.DeleteZone(ctx, zoneChange.ZoneName)
	}

	tx := w.store.ZoneTxn(ctx, zoneChange.ZoneName)

	tx.CreateZoneMarker()

	for _, ch := range zoneChange.Changes {
		if err := processResourceRecordSetChange(tx, ch); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func processResourceRecordSetChange(tx dnsstore.ZoneTransaction, change model.ResourceRecordSetChange) error {
	if change.Action == model.RRSetChangeActionDelete {
		tx.DeleteRRSet(change.ResourceRecordSet.Name, string(change.ResourceRecordSet.Type))
		return nil
	}

	rrset, err := parseDNSRecordSet(change.ResourceRecordSet)
	if err != nil {
		return err
	}

	tx.PutRRSet(change.ResourceRecordSet.Name, string(change.ResourceRecordSet.Type), rrset)

	return nil
}

func parseDNSRecordSet(rrset model.ResourceRecordSet) ([]dns.RR, error) {
	switch rrset.Type {
	case model.RRTypeA:
		return rrbuilder.A(rrset)
	case model.RRTypeAAAA:
		return rrbuilder.AAAA(rrset)
	case model.RRTypeCNAME:
		return rrbuilder.CNAME(rrset)
	case model.RRTypeSOA:
		return rrbuilder.SOA(rrset)
	case model.RRTypeNS:
		return rrbuilder.NS(rrset)
	case model.RRTypeSRV:
		return rrbuilder.SRV(rrset)
	case model.RRTypeSVCB:
		return rrbuilder.SVCB(rrset)
	case model.RRTypeHTTPS:
		return rrbuilder.HTTPS(rrset)
	case model.RRTypeNAPTR:
		return rrbuilder.NAPTR(rrset)
	case model.RRTypeSSHFP:
		return rrbuilder.SSHFP(rrset)
	case model.RRTypeTLSA:
		return rrbuilder.TLSA(rrset)
	case model.RRTypeTXT:
		return rrbuilder.TXT(rrset)
	case model.RRTypeCAA:
		return rrbuilder.CAA(rrset)
	case model.RRTypeDS:
		return rrbuilder.DS(rrset)
	case model.RRTypePTR:
		return rrbuilder.PTR(rrset)
	case model.RRTypeMX:
		return rrbuilder.MX(rrset)
	}

	return nil, fmt.Errorf("invalid record type: %s", rrset.Type)
}
