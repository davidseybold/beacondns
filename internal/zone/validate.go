package zone

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/rrbuilder"
)

type rule func(zone *model.Zone, changes []model.ResourceRecordSetChange) error

var rules = []rule{
	supportedRRTypeRule,
	cnameRule,
	soaRule,
	nsRule,
	duplicateRecordRule,
	apexRecordRule,
	ttlRule,
	recordValueRule,
	domainNameRule,
}

func validateChanges(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	for _, rule := range rules {
		if err := rule(zone, changes); err != nil {
			return err
		}
	}

	return nil
}

func cnameRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	cnameChangeNames := make(map[string]struct{})
	for _, change := range changes {
		if change.ResourceRecordSet.Type != model.RRTypeCNAME {
			continue
		}

		cnameChangeNames[change.ResourceRecordSet.Name] = struct{}{}

		for _, rr := range change.ResourceRecordSet.ResourceRecords {
			if rr.Value == change.ResourceRecordSet.Name {
				return fmt.Errorf("CNAME record cannot point to itself: %s", change.ResourceRecordSet.Name)
			}
		}
	}

	if len(cnameChangeNames) == 0 {
		return nil
	}

	for _, rrset := range zone.ResourceRecordSets {
		if _, ok := cnameChangeNames[rrset.Name]; ok && rrset.Type != model.RRTypeCNAME {
			return fmt.Errorf("cannot add CNAME record at %s: record of type %s already exists", rrset.Name, rrset.Type)
		}
	}

	return nil
}

func soaRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	hasSOA := false
	for _, rrset := range zone.ResourceRecordSets {
		if rrset.Type == model.RRTypeSOA {
			hasSOA = true
			break
		}
	}

	for _, change := range changes {
		if change.ResourceRecordSet.Type != model.RRTypeSOA {
			continue
		}

		if change.Action == model.RRSetChangeActionDelete {
			return fmt.Errorf("cannot delete SOA record: it is required for zone %s", zone.Name)
		}

		if hasSOA {
			return fmt.Errorf("cannot add SOA record: zone %s already has an SOA record", zone.Name)
		}
		if len(change.ResourceRecordSet.ResourceRecords) > 1 {
			return errors.New("SOA record can only have one resource record")
		}
	}

	return nil
}

func nsRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	hasNS := false
	for _, rrset := range zone.ResourceRecordSets {
		if rrset.Type == model.RRTypeNS {
			hasNS = true
			break
		}
	}

	nsDeletions := 0
	for _, change := range changes {
		if change.ResourceRecordSet.Type == model.RRTypeNS && change.Action == model.RRSetChangeActionDelete {
			nsDeletions++
		}
	}

	if hasNS && nsDeletions == len(zone.ResourceRecordSets) {
		return fmt.Errorf("cannot delete all NS records: at least one NS record is required for zone %s", zone.Name)
	}

	if !hasNS {
		addingNS := false
		for _, change := range changes {
			if change.ResourceRecordSet.Type == model.RRTypeNS && change.Action == model.RRSetChangeActionCreate {
				addingNS = true
				break
			}
		}
		if !addingNS {
			return fmt.Errorf("zone %s must have at least one NS record", zone.Name)
		}
	}

	return nil
}

func duplicateRecordRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	existingRecords := make(map[string]map[model.RRType]struct{})
	for _, rrset := range zone.ResourceRecordSets {
		if _, ok := existingRecords[rrset.Name]; !ok {
			existingRecords[rrset.Name] = make(map[model.RRType]struct{})
		}
		existingRecords[rrset.Name][rrset.Type] = struct{}{}
	}

	for _, change := range changes {
		if change.Action == model.RRSetChangeActionCreate {
			if types, ok := existingRecords[change.ResourceRecordSet.Name]; ok {
				if _, exists := types[change.ResourceRecordSet.Type]; exists {
					return fmt.Errorf("duplicate record: %s record already exists at %s",
						change.ResourceRecordSet.Type, change.ResourceRecordSet.Name)
				}
			}
		}
	}

	return nil
}

func apexRecordRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	apexName := dns.Fqdn(zone.Name)

	for _, change := range changes {
		if change.ResourceRecordSet.Name == apexName {
			if change.ResourceRecordSet.Type != model.RRTypeSOA &&
				change.ResourceRecordSet.Type != model.RRTypeNS &&
				change.ResourceRecordSet.Type != model.RRTypeA {
				return fmt.Errorf("invalid apex record type %s: only SOA, NS, and A records are allowed at zone apex",
					change.ResourceRecordSet.Type)
			}
		}
	}

	return nil
}

func supportedRRTypeRule(_ *model.Zone, changes []model.ResourceRecordSetChange) error {
	for _, change := range changes {
		if _, ok := model.SupportedRRTypes[change.ResourceRecordSet.Type]; !ok {
			return fmt.Errorf("invalid record type %s: only supported types are %v",
				change.ResourceRecordSet.Type,
				model.SupportedRRTypes,
			)
		}
	}

	return nil
}

// ttlRule ensures TTL values are within valid range.
func ttlRule(_ *model.Zone, changes []model.ResourceRecordSetChange) error {
	const (
		minTTL = 0
		maxTTL = 2147483647 // 2^31 - 1
	)

	for _, change := range changes {
		if change.ResourceRecordSet.TTL < minTTL || change.ResourceRecordSet.TTL > maxTTL {
			return fmt.Errorf("invalid TTL value %d: must be between %d and %d",
				change.ResourceRecordSet.TTL, minTTL, maxTTL)
		}
	}
	return nil
}

func recordValueRule(_ *model.Zone, changes []model.ResourceRecordSetChange) error {
	for _, change := range changes {
		var err error
		switch change.ResourceRecordSet.Type {
		case model.RRTypeA:
			_, err = rrbuilder.A(change.ResourceRecordSet)
		case model.RRTypeAAAA:
			_, err = rrbuilder.AAAA(change.ResourceRecordSet)
		case model.RRTypeMX:
			_, err = rrbuilder.MX(change.ResourceRecordSet)
		case model.RRTypeSRV:
			_, err = rrbuilder.SRV(change.ResourceRecordSet)
		case model.RRTypeCNAME:
			_, err = rrbuilder.CNAME(change.ResourceRecordSet)
		case model.RRTypeTXT:
			_, err = rrbuilder.TXT(change.ResourceRecordSet)
		case model.RRTypeSOA:
			_, err = rrbuilder.SOA(change.ResourceRecordSet)
		case model.RRTypeNS:
			_, err = rrbuilder.NS(change.ResourceRecordSet)
		case model.RRTypePTR:
			_, err = rrbuilder.PTR(change.ResourceRecordSet)
		case model.RRTypeCAA:
			_, err = rrbuilder.CAA(change.ResourceRecordSet)
		case model.RRTypeDS:
			_, err = rrbuilder.DS(change.ResourceRecordSet)
		case model.RRTypeSSHFP:
			_, err = rrbuilder.SSHFP(change.ResourceRecordSet)
		case model.RRTypeTLSA:
			_, err = rrbuilder.TLSA(change.ResourceRecordSet)
		case model.RRTypeSVCB:
			_, err = rrbuilder.SVCB(change.ResourceRecordSet)
		case model.RRTypeHTTPS:
			_, err = rrbuilder.HTTPS(change.ResourceRecordSet)
		case model.RRTypeNAPTR:
			_, err = rrbuilder.NAPTR(change.ResourceRecordSet)
		}
		if err != nil {
			return fmt.Errorf("invalid record value: %w", err)
		}
	}
	return nil
}

// domainNameRule ensures domain names are valid.
func domainNameRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	for _, change := range changes {
		// Check if the record name is a valid domain name
		if !dns.IsFqdn(change.ResourceRecordSet.Name) {
			return fmt.Errorf(
				"invalid domain name: %s is not a fully qualified domain name",
				change.ResourceRecordSet.Name,
			)
		}

		// Check if the record name is within the zone.
		if !strings.HasSuffix(change.ResourceRecordSet.Name, dns.Fqdn(zone.Name)) {
			return fmt.Errorf("invalid domain name: %s is not within zone %s", change.ResourceRecordSet.Name, zone.Name)
		}

		// Check for wildcard records.
		if strings.Contains(change.ResourceRecordSet.Name, "*") {
			if !strings.HasPrefix(change.ResourceRecordSet.Name, "*. ") {
				return fmt.Errorf(
					"invalid wildcard record: %s (wildcard must be at the leftmost label)",
					change.ResourceRecordSet.Name,
				)
			}
		}
	}
	return nil
}
