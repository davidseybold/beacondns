package client

type CreateZoneRequest struct {
	Name string `json:"name"`
}

type ChangeResourceRecordSetsRequest struct {
	Changes []Change `json:"changes"`
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

type Change struct {
	Action            string            `json:"action"`
	ResourceRecordSet ResourceRecordSet `json:"resourceRecordSet"`
}

type ChangeInfo struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submittedAt"`
}

type CreateZoneResponse struct {
	ChangeInfo ChangeInfo `json:"changeInfo"`
	Zone       Zone       `json:"zone"`
}

type ListZonesResponse struct {
	Zones []Zone `json:"zones"`
}

type ListResourceRecordSetsResponse struct {
	ResourceRecordSets []ResourceRecordSet `json:"resourceRecordSets"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
