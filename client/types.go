package client

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
