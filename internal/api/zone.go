package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/model"
)

func (h *handler) ListZones(c *gin.Context) {
	zones, err := h.zoneService.ListZones(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
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
		h.handleGinBindingError(c, err)
		return
	}

	res, err := h.zoneService.CreateZone(c.Request.Context(), body.Name)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, Zone{
		ID:                     res.ID.String(),
		Name:                   res.Name,
		ResourceRecordSetCount: res.ResourceRecordSetCount,
	})
}

func (h *handler) GetZone(c *gin.Context) {
	zoneName := c.Param("zoneName")

	zone, err := h.zoneService.GetZoneInfo(c.Request.Context(), zoneName)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, Zone{
		ID:                     zone.ID.String(),
		Name:                   zone.Name,
		ResourceRecordSetCount: zone.ResourceRecordSetCount,
	})
}

func (h *handler) UpsertResourceRecordSet(c *gin.Context) {
	zoneName := c.Param("zoneName")

	if zoneName == "" {
		h.handleError(c, beaconerr.ErrInvalidArgument("zone name is required", "zoneName"))
		return
	}

	var body UpsertResourceRecordSetRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleGinBindingError(c, err)
		return
	}

	rrSet := convertAPIResourceRecordSetToModel(&body.ResourceRecordSet)

	newRRSet, err := h.zoneService.UpsertResourceRecordSet(c.Request.Context(), zoneName, rrSet)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertModelResourceRecordSetToAPI(newRRSet))
}

func (h *handler) ListResourceRecordSets(c *gin.Context) {
	zoneName := c.Param("zoneName")

	if zoneName == "" {
		h.handleError(c, beaconerr.ErrInvalidArgument("zone name is required", "zoneName"))
		return
	}

	rrsets, err := h.zoneService.ListResourceRecordSets(c.Request.Context(), zoneName)
	if err != nil {
		h.handleError(c, err)
		return
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
	zoneName := c.Param("zoneName")

	if zoneName == "" {
		h.handleError(c, beaconerr.ErrInvalidArgument("zone name is required", "zoneName"))
		return
	}

	err := h.zoneService.DeleteZone(c.Request.Context(), zoneName)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *handler) DeleteResourceRecordSet(c *gin.Context) {
	zoneName := c.Param("zoneName")

	if zoneName == "" {
		h.handleError(c, beaconerr.ErrInvalidArgument("zone name is required", "zoneName"))
		return
	}

	name := c.Param("name")
	rrType := model.RRType(strings.ToUpper(c.Param("type")))

	_, ok := model.SupportedRRTypes[rrType]
	if !ok {
		h.handleError(c, beaconerr.ErrInvalidArgument("invalid record type", "type"))
		return
	}

	if name == "" {
		h.handleError(c, beaconerr.ErrInvalidArgument("name is required", "name"))
		return
	}

	err := h.zoneService.DeleteResourceRecordSet(c.Request.Context(), zoneName, name, rrType)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *handler) GetResourceRecordSet(c *gin.Context) {
	zoneName := c.Param("zoneName")

	if zoneName == "" {
		h.handleError(c, beaconerr.ErrInvalidArgument("zone name is required", "zoneName"))
		return
	}

	name := c.Param("name")
	rrType := model.RRType(strings.ToUpper(c.Param("type")))

	_, ok := model.SupportedRRTypes[rrType]
	if !ok {
		h.handleError(c, beaconerr.ErrInvalidArgument("invalid record type", "type"))
		return
	}

	rrSet, err := h.zoneService.GetResourceRecordSet(c.Request.Context(), zoneName, name, rrType)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertModelResourceRecordSetToAPI(rrSet))
}
