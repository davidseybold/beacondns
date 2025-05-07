package api

import "github.com/davidseybold/beacondns/internal/controller/domain"

type CreateZoneRequest struct {
	Name string `json:"name" binding:"required"`
}

type ChangeInfo struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submittedAt"`
}

func NewChangeInfoFromDomain(changeInfo domain.ChangeInfo) ChangeInfo {
	return ChangeInfo{
		ID:          changeInfo.ID.String(),
		Status:      string(changeInfo.Status),
		SubmittedAt: changeInfo.SubmittedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type DelegationSet struct {
	ID          string   `json:"id"`
	NameServers []string `json:"nameServers"`
}

func NewDelegationSetFromDomain(ds *domain.DelegationSet) *DelegationSet {
	if ds == nil {
		return nil
	}

	nameServers := make([]string, len(ds.NameServers))
	for i, ns := range ds.NameServers {
		nameServers[i] = ns.Name
	}

	return &DelegationSet{
		ID:          ds.ID.String(),
		NameServers: nameServers,
	}
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewZoneFromDomain(zone domain.Zone) Zone {
	return Zone{
		ID:   zone.ID.String(),
		Name: zone.Name,
	}
}

type CreateZoneResponse struct {
	ChangeInfo    ChangeInfo     `json:"changeInfo"`
	DelegationSet *DelegationSet `json:"delegationSet,omitempty"`
	Zone          Zone           `json:"zone"`
}

func NewCreateZoneResponse(res domain.CreateZoneResult) CreateZoneResponse {
	ds := NewDelegationSetFromDomain(res.DelegationSet)
	changeInfo := NewChangeInfoFromDomain(res.ChangeInfo)
	zone := NewZoneFromDomain(res.Zone)

	return CreateZoneResponse{
		ChangeInfo:    changeInfo,
		DelegationSet: ds,
		Zone:          zone,
	}
}

type AddNameServerRequest struct {
	Name      string `json:"name" binding:"required"`
	RouteKey  string `json:"routeKey" binding:"required"`
	IPAddress string `json:"ipAddress" binding:"required"`
}

type AddNameServerResponse struct {
	NameServer NameServer `json:"nameServer"`
}

type NameServer struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RouteKey  string `json:"routeKey"`
	IPAddress string `json:"ipAddress"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ListNameServersResponse struct {
	NameServers []NameServer `json:"nameServers"`
}

type ListZonesResponse struct {
	Zones []Zone `json:"zones"`
}
