package api

import "github.com/google/uuid"

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CreateZoneRequest struct {
	Name string `json:"name" binding:"required"`
}

type Zone struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	ResourceRecordSetCount int    `json:"resourceRecordSetCount"`
}

type ListZonesResponse struct {
	Zones []Zone `json:"zones"`
}

type ListResourceRecordSetsResponse struct {
	ResourceRecordSets []ResourceRecordSet `json:"resourceRecordSets"`
}

type UpsertResourceRecordSetRequest struct {
	ResourceRecordSet
}

type ResourceRecordSet struct {
	Name            string           `json:"name"            binding:"required"`
	Type            string           `json:"type"            binding:"required"`
	TTL             uint32           `json:"ttl"             binding:"required"`
	ResourceRecords []ResourceRecord `json:"resourceRecords" binding:"required,min=1"`
}

type ResourceRecord struct {
	Value string `json:"value" binding:"required"`
}

type FirewallRule struct {
	ID                string             `json:"id"`
	DomainListID      string             `json:"domainListId"`
	Action            string             `json:"action"                      binding:"firewallRuleAction"`
	BlockResponseType *string            `json:"blockResponseType,omitempty" binding:"firewallRuleBlockResponseType"`
	BlockResponse     *ResourceRecordSet `json:"blockResponse,omitempty"`
	Priority          uint               `json:"priority"`
}

type AddDomainsToDomainListRequest struct {
	Domains []string `json:"domains" binding:"required"`
}

type RemoveDomainsFromDomainListRequest struct {
	Domains []string `json:"domains" binding:"required"`
}

type ListDomainListDomainsResponse struct {
	Domains    []string `json:"domains"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

type ListDomainListsResponse struct {
	DomainLists []DomainList `json:"domainLists"`
}

type CreateDomainListRequest struct {
	Name      string   `json:"name"                binding:"required"`
	IsManaged bool     `json:"isManaged"`
	Domains   []string `json:"domains"`
	SourceURL *string  `json:"sourceUrl,omitempty"`
}

type FirewallRuleRequest struct {
	Name              string             `json:"name"                        binding:"required"`
	DomainListID      uuid.UUID          `json:"domainListId"                binding:"required"`
	Action            string             `json:"action"                      binding:"required,firewallRuleAction"`
	BlockResponseType *string            `json:"blockResponseType,omitempty" binding:"firewallRuleBlockResponseType"`
	BlockResponse     *ResourceRecordSet `json:"blockResponse,omitempty"`
	Priority          uint               `json:"priority"                    binding:"required"`
}

type DomainList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DomainCount int       `json:"domainCount"`
}
