package zone

import (
	"errors"
	"fmt"
	"strings"

	"github.com/miekg/dns"

	bdns "github.com/davidseybold/beacondns/internal/dns"
	"github.com/davidseybold/beacondns/internal/model"
)

var (
	ErrCNAMESelfReference = errors.New("CNAME record cannot point to itself")
	ErrCNAMEConflict      = errors.New("cannot add CNAME record: record of different type already exists")
	ErrSOADeletion        = errors.New("cannot delete SOA record: it is required for zone")
	ErrSOAExists          = errors.New("cannot add SOA record: zone already has an SOA record")
	ErrSOAMultipleRecords = errors.New("SOA record can only have one resource record")
	ErrNSDeletion         = errors.New("cannot delete all NS records: at least one NS record is required")
	ErrNSRequired         = errors.New("zone must have at least one NS record")
	ErrInvalidApexRecord  = errors.New("invalid apex record type: only SOA, NS, and A records are allowed at zone apex")
	ErrUnsupportedRRType  = errors.New("invalid record type: not supported")
	ErrInvalidRecordValue = errors.New("invalid record value")
	ErrOutsideZone        = errors.New("invalid domain name: not within zone")
	ErrInvalidWildcard    = errors.New("invalid wildcard record: wildcard must be at the leftmost label")
)

type rule func(zone *model.Zone, changes []model.ChangeAction) error

var rules = []rule{
	supportedRRTypeRule,
	cnameRule,
	soaRule,
	nsRule,
	apexRecordRule,
	recordValueRule,
	domainNameRule,
}

func validateChanges(zone *model.Zone, change *model.Change) error {
	for _, rule := range rules {
		if err := rule(zone, change.Actions); err != nil {
			return err
		}
	}

	return nil
}

func cnameRule(zone *model.Zone, changes []model.ChangeAction) error {
	cnameChangeNames := make(map[string]struct{})
	for _, change := range changes {
		if change.ResourceRecordSet.Type != model.RRTypeCNAME {
			continue
		}

		cnameChangeNames[change.ResourceRecordSet.Name] = struct{}{}

		for _, rr := range change.ResourceRecordSet.ResourceRecords {
			if rr.Value == change.ResourceRecordSet.Name {
				return fmt.Errorf("%w: %s", ErrCNAMESelfReference, change.ResourceRecordSet.Name)
			}
		}
	}

	if len(cnameChangeNames) == 0 {
		return nil
	}

	for _, rrset := range zone.ResourceRecordSets {
		if _, ok := cnameChangeNames[rrset.Name]; ok && rrset.Type != model.RRTypeCNAME {
			return fmt.Errorf("%w at %s: record of type %s already exists", ErrCNAMEConflict, rrset.Name, rrset.Type)
		}
	}

	return nil
}

func soaRule(zone *model.Zone, changes []model.ChangeAction) error {
	zoneHasSOA := false
	for _, rrset := range zone.ResourceRecordSets {
		if rrset.Type == model.RRTypeSOA {
			zoneHasSOA = true
			break
		}
	}

	for _, change := range changes {
		if change.ResourceRecordSet.Type != model.RRTypeSOA {
			continue
		}

		if change.ActionType == model.ChangeActionTypeDelete {
			return fmt.Errorf("%w %s", ErrSOADeletion, zone.Name)
		}

		if zoneHasSOA {
			return fmt.Errorf("%w %s", ErrSOAExists, zone.Name)
		}
		if len(change.ResourceRecordSet.ResourceRecords) > 1 {
			return ErrSOAMultipleRecords
		}
	}

	return nil
}

func nsRule(zone *model.Zone, changes []model.ChangeAction) error {
	hasNS := false
	for _, rrset := range zone.ResourceRecordSets {
		if rrset.Type == model.RRTypeNS {
			hasNS = true
			break
		}
	}

	nsDeletions := 0
	for _, change := range changes {
		if change.ResourceRecordSet.Type == model.RRTypeNS && change.ActionType == model.ChangeActionTypeDelete {
			nsDeletions++
		}
	}

	if hasNS && nsDeletions == len(zone.ResourceRecordSets) {
		return fmt.Errorf("%w %s", ErrNSDeletion, zone.Name)
	}

	if !hasNS {
		addingNS := false
		for _, change := range changes {
			if change.ResourceRecordSet.Type == model.RRTypeNS && change.ActionType == model.ChangeActionTypeUpsert {
				addingNS = true
				break
			}
		}
		if !addingNS {
			return fmt.Errorf("%w %s", ErrNSRequired, zone.Name)
		}
	}

	return nil
}

func apexRecordRule(zone *model.Zone, changes []model.ChangeAction) error {
	apexName := dns.Fqdn(zone.Name)

	for _, change := range changes {
		if dns.Fqdn(change.ResourceRecordSet.Name) == apexName {
			if change.ResourceRecordSet.Type != model.RRTypeSOA &&
				change.ResourceRecordSet.Type != model.RRTypeNS &&
				change.ResourceRecordSet.Type != model.RRTypeA {
				return fmt.Errorf("%w: %s", ErrInvalidApexRecord, change.ResourceRecordSet.Type)
			}
		}
	}

	return nil
}

func supportedRRTypeRule(_ *model.Zone, changes []model.ChangeAction) error {
	for _, change := range changes {
		if _, ok := model.SupportedRRTypes[change.ResourceRecordSet.Type]; !ok {
			return fmt.Errorf("%w: %s (supported types are %v)",
				ErrUnsupportedRRType,
				change.ResourceRecordSet.Type,
				model.SupportedRRTypes,
			)
		}
	}

	return nil
}

func recordValueRule(_ *model.Zone, changes []model.ChangeAction) error {
	for _, change := range changes {
		_, err := bdns.ParseRRs(change.ResourceRecordSet)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidRecordValue, err)
		}
	}
	return nil
}

func domainNameRule(zone *model.Zone, changes []model.ChangeAction) error {
	for _, change := range changes {
		// Check if the record name is within the zone.
		if !dns.IsSubDomain(dns.Fqdn(zone.Name), dns.Fqdn(change.ResourceRecordSet.Name)) {
			return fmt.Errorf("%w: %s is not within zone %s", ErrOutsideZone, change.ResourceRecordSet.Name, zone.Name)
		}

		// Check for wildcard records.
		numAsterisks := strings.Count(change.ResourceRecordSet.Name, "*")
		if numAsterisks > 0 {
			labels := dns.SplitDomainName(change.ResourceRecordSet.Name)
			if labels[0] != "*" || numAsterisks > 1 {
				return fmt.Errorf("%w: %s", ErrInvalidWildcard, change.ResourceRecordSet.Name)
			}
		}
	}
	return nil
}
