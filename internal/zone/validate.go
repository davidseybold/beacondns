package zone

import (
	"fmt"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/model"
)

type rule func(zone *model.Zone, changes []model.ResourceRecordSetChange) error

var rules = []rule{
	cnameRule,
	soaRule,
	nsRule,
	duplicateRecordRule,
	apexRecordRule,
}

func validateChanges(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	for _, rule := range rules {
		if err := rule(zone, changes); err != nil {
			return err
		}
	}

	return nil
}

// cnameRule ensures that CNAME records don't coexist with other records at the same name.
func cnameRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	// Track all CNAME changes.
	cnameChangeNames := make(map[string]struct{})
	for _, change := range changes {
		if change.ResourceRecordSet.Type == model.RRTypeCNAME {
			cnameChangeNames[change.ResourceRecordSet.Name] = struct{}{}
		}
	}

	if len(cnameChangeNames) == 0 {
		return nil
	}

	// Check existing records for conflicts.
	for _, rrset := range zone.ResourceRecordSets {
		if _, ok := cnameChangeNames[rrset.Name]; ok && rrset.Type != model.RRTypeCNAME {
			return fmt.Errorf("cannot add CNAME record at %s: record of type %s already exists", rrset.Name, rrset.Type)
		}
	}

	// Check other changes for conflicts.
	for _, change := range changes {
		if _, ok := cnameChangeNames[change.ResourceRecordSet.Name]; ok &&
			change.ResourceRecordSet.Type != model.RRTypeCNAME {
			return fmt.Errorf(
				"cannot add %s record at %s: CNAME record is being added",
				change.ResourceRecordSet.Type,
				change.ResourceRecordSet.Name,
			)
		}
	}

	return nil
}

// soaRule ensures SOA record is present and valid.
func soaRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	// Check if SOA is being deleted.
	for _, change := range changes {
		if change.ResourceRecordSet.Type == model.RRTypeSOA && change.Action == model.RRSetChangeActionDelete {
			return fmt.Errorf("cannot delete SOA record: it is required for zone %s", zone.Name)
		}
	}

	return nil
}

// nsRule ensures NS records are present and valid.
func nsRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	hasNS := false
	for _, rrset := range zone.ResourceRecordSets {
		if rrset.Type == model.RRTypeNS {
			hasNS = true
			break
		}
	}

	// Check if all NS records are being deleted.
	nsDeletions := 0
	for _, change := range changes {
		if change.ResourceRecordSet.Type == model.RRTypeNS && change.Action == model.RRSetChangeActionDelete {
			nsDeletions++
		}
	}

	if hasNS && nsDeletions == len(zone.ResourceRecordSets) {
		return fmt.Errorf("cannot delete all NS records: at least one NS record is required for zone %s", zone.Name)
	}

	// If no NS exists, ensure at least one is being added.
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

// duplicateRecordRule ensures no duplicate records exist.
func duplicateRecordRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	// Track existing records.
	existingRecords := make(map[string]map[model.RRType]struct{})
	for _, rrset := range zone.ResourceRecordSets {
		if _, ok := existingRecords[rrset.Name]; !ok {
			existingRecords[rrset.Name] = make(map[model.RRType]struct{})
		}
		existingRecords[rrset.Name][rrset.Type] = struct{}{}
	}

	// Check changes for duplicates.
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

// apexRecordRule ensures apex records are valid.
func apexRecordRule(zone *model.Zone, changes []model.ResourceRecordSetChange) error {
	apexName := dns.Fqdn(zone.Name)

	for _, change := range changes {
		if change.ResourceRecordSet.Name == apexName {
			// Apex records can only be SOA, NS, or A records.
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
