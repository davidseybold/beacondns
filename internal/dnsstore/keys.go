package dnsstore

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	keyPrefixZones         = "/zones"
	keyPrefixFirewallRules = "/firewall/rule"
)

func createRecordKey(zoneName, rrName string, rrType string) string {
	return fmt.Sprintf("/zone/%s/recordset/%s/%s", zoneName, rrName, rrType)
}

func createZonesKey(zoneName string) string {
	return fmt.Sprintf("%s/%s", keyPrefixZones, zoneName)
}

func createZoneRecordSetPrefix(zoneName string) string {
	return fmt.Sprintf("/zone/%s/recordset", zoneName)
}

func createFirewallRuleMappingPrefix(ruleID uuid.UUID) string {
	return fmt.Sprintf("%s/%s", keyPrefixFirewallRules, ruleID)
}

func createFirewallRuleMappingKey(ruleID uuid.UUID, domain string) string {
	return fmt.Sprintf("%s/domain/%s", createFirewallRuleMappingPrefix(ruleID), domain)
}

func createFirewallRuleKey(ruleID uuid.UUID) string {
	return fmt.Sprintf("/firewall/rules/%s", ruleID)
}
