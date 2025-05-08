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
	ChangeInfo ChangeInfo `json:"changeInfo"`
	Zone       Zone       `json:"zone"`
}

func NewCreateZoneResponse(res domain.CreateZoneResult) CreateZoneResponse {
	changeInfo := NewChangeInfoFromDomain(res.ChangeInfo)
	zone := NewZoneFromDomain(res.Zone)

	return CreateZoneResponse{
		ChangeInfo: changeInfo,
		Zone:       zone,
	}
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ListZonesResponse struct {
	Zones []Zone `json:"zones"`
}
