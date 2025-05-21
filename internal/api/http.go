package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/zone"
)

func NewHTTPHandler(zoneService zone.Service) http.Handler {
	r := gin.Default()

	handler := &handler{
		zoneService: zoneService,
	}

	r.GET("/health", handler.Health)

	{
		g := r.Group("/v1/zones")
		g.POST("", handler.CreateZone)
		g.DELETE("/:zoneID", handler.DeleteZone)
		g.GET("", handler.ListZones)
		g.GET("/:zoneID", handler.GetZone)
		g.POST("/:zoneID/rrsets", handler.ChangeResourceRecordSets)
		g.GET("/:zoneID/rrsets", handler.ListResourceRecordSets)
	}

	return r
}

type handler struct {
	zoneService zone.Service
}

func (h *handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

func (h *handler) ListZones(c *gin.Context) {
	zones, err := h.zoneService.ListZones(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
		return
	}

	responseBody := ListZonesResponse{
		Zones: make([]Zone, len(zones)),
	}

	for i, zone := range zones {
		responseBody.Zones[i] = Zone{
			ID:                     zone.ID.String(),
			Name:                   zone.Name,
			ResourceRecordSetCount: zone.ResourceRecordSetCount,
		}
	}

	c.JSON(http.StatusOK, responseBody)
}

func (h *handler) CreateZone(c *gin.Context) {
	var body CreateZoneRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "BadRequest",
			Message: err.Error(),
		})
		return
	}

	res, err := h.zoneService.CreateZone(c.Request.Context(), body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
		return
	}

	resp := CreateZoneResponse{
		ChangeInfo: ChangeInfo{
			ID:          res.Change.ID.String(),
			Status:      string(res.Change.Status),
			SubmittedAt: res.Change.SubmittedAt.Format("2006-01-02T15:04:05Z"),
		},
		Zone: Zone{
			ID:                     res.Zone.ID.String(),
			Name:                   res.Zone.Name,
			ResourceRecordSetCount: len(res.Zone.ResourceRecordSets),
		},
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *handler) GetZone(c *gin.Context) {
	zoneID, err := uuid.Parse(c.Param("zoneID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "BadRequest",
			Message: "invalid zone ID",
		})
		return
	}

	zone, err := h.zoneService.GetZoneInfo(c.Request.Context(), zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Zone{
		ID:                     zone.ID.String(),
		Name:                   zone.Name,
		ResourceRecordSetCount: zone.ResourceRecordSetCount,
	})
}

func (h *handler) ChangeResourceRecordSets(c *gin.Context) {
	zoneID, err := uuid.Parse(c.Param("zoneID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "BadRequest",
			Message: "invalid zone ID",
		})
		return
	}

	var body ChangeResourceRecordSetsRequest
	if err = c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "BadRequest",
			Message: err.Error(),
		})
		return
	}

	changes := make([]model.ResourceRecordSetChange, len(body.Changes))
	for i, change := range body.Changes {
		changes[i] = convertAPIChangeToModel(change)
	}

	change, err := h.zoneService.ChangeResourceRecordSets(c.Request.Context(), zoneID, changes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChangeInfo{
		ID:          change.ID.String(),
		Status:      string(change.Status),
		SubmittedAt: change.SubmittedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *handler) ListResourceRecordSets(c *gin.Context) {
	zoneID, err := uuid.Parse(c.Param("zoneID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "BadRequest",
			Message: "invalid zone ID",
		})
		return
	}

	rrsets, err := h.zoneService.ListResourceRecordSets(c.Request.Context(), zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
	}

	responseBody := ListResourceRecordSetsResponse{
		ResourceRecordSets: make([]ResourceRecordSet, len(rrsets)),
	}

	for i, rrset := range rrsets {
		responseBody.ResourceRecordSets[i] = ResourceRecordSet{
			Name:            rrset.Name,
			Type:            string(rrset.Type),
			TTL:             rrset.TTL,
			ResourceRecords: make([]ResourceRecord, len(rrset.ResourceRecords)),
		}

		for j, record := range rrset.ResourceRecords {
			responseBody.ResourceRecordSets[i].ResourceRecords[j] = ResourceRecord{
				Value: record.Value,
			}
		}
	}

	c.JSON(http.StatusOK, responseBody)
}

func (h *handler) DeleteZone(c *gin.Context) {
	zoneID, err := uuid.Parse(c.Param("zoneID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "BadRequest",
			Message: "invalid zone ID",
		})
		return
	}

	change, err := h.zoneService.DeleteZone(c.Request.Context(), zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChangeInfo{
		ID:          change.ID.String(),
		Status:      string(change.Status),
		SubmittedAt: change.SubmittedAt.Format("2006-01-02T15:04:05Z"),
	})
}
