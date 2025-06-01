package client

import "github.com/google/uuid"

type createZoneRequest struct {
	Name string `json:"name"`
}

type Zone struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	ResourceRecordSetCount int    `json:"resourceRecordSetCount"`
}

type ResourceRecordSet struct {
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	TTL             uint32           `json:"ttl"`
	ResourceRecords []ResourceRecord `json:"resourceRecords"`
}

type ResourceRecord struct {
	Value string `json:"value"`
}

type listZonesResponse struct {
	Zones []Zone `json:"zones"`
}

type listResourceRecordSetsResponse struct {
	ResourceRecordSets []ResourceRecordSet `json:"resourceRecordSets"`
}

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type upsertResourceRecordSetRequest struct {
	ResourceRecordSet
}

type CreateFirewallRuleRequest struct {
	DomainListID      string             `json:"domainListId"`
	Action            string             `json:"action"`
	BlockResponseType *string            `json:"blockResponseType,omitempty"`
	BlockResponse     *ResourceRecordSet `json:"blockResponse,omitempty"`
	Priority          uint               `json:"priority"`
}

type FirewallRule struct {
	ID                uuid.UUID          `json:"id"`
	Name              string             `json:"name"`
	DomainListID      uuid.UUID          `json:"domainListId"`
	Action            string             `json:"action"`
	BlockResponseType *string            `json:"blockResponseType,omitempty"`
	BlockResponse     *ResourceRecordSet `json:"blockResponse,omitempty"`
	Priority          uint               `json:"priority"`
}

type getFirewallRulesResponse struct {
	Rules []FirewallRule `json:"rules"`
}

type UpdateFirewallRuleRequest struct {
	DomainListID      uuid.UUID          `json:"domainListId"`
	Action            string             `json:"action"`
	BlockResponseType *string            `json:"blockResponseType,omitempty"`
	BlockResponse     *ResourceRecordSet `json:"blockResponse,omitempty"`
	Priority          uint               `json:"priority"`
}

type CreateDomainListRequest struct {
	Name    string   `json:"name"`
	Domains []string `json:"domains"`
}

type DomainList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DomainCount int       `json:"domainCount"`
}

type listDomainListsResponse struct {
	DomainLists []DomainList `json:"domainLists"`
}

type getDomainListDomainsResponse struct {
	Domains []string `json:"domains"`
}

type addDomainsToDomainListRequest struct {
	Domains []string `json:"domains"`
}

type removeDomainsFromDomainListRequest struct {
	Domains []string `json:"domains"`
}
