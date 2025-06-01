package api

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
	Action            string             `json:"action"`
	BlockResponseType *string            `json:"blockResponseType,omitempty"`
	BlockResponse     *ResourceRecordSet `json:"blockResponse,omitempty"`
	Priority          uint               `json:"priority"`
}
