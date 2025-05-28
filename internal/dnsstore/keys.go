package dnsstore

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	keyPrefixZones          = "/zones"
	keyPrefixResponsePolicy = "/response-policy"
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

func createResponsePolicyRuleKey(rule *ResponsePolicyRule) string {
	return fmt.Sprintf("%s/%s/rule/%s", keyPrefixResponsePolicy, rule.Meta.PolicyID, rule.ID)
}

func createResponsePolicyRulePrefix(policyID uuid.UUID) string {
	return fmt.Sprintf("%s/%s/rule", keyPrefixResponsePolicy, policyID)
}
