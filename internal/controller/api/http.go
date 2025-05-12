package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/controller/zone"
)

func NewHTTPHandler(zoneService zone.Service) http.Handler {
	r := gin.Default()

	handler := &handler{
		zoneService: zoneService,
	}

	r.GET("/health", handler.Health)

	r.POST("/zones", handler.CreateZone)
	r.GET("/zones", handler.ListZones)
	r.GET("/zones/:id", handler.GetZone)

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
			ID:   zone.ID.String(),
			Name: zone.Name,
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

	resp := NewCreateZoneResponse(*res)

	c.JSON(http.StatusCreated, resp)
}

func (h *handler) GetZone(c *gin.Context) {
	zoneID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "BadRequest",
			Message: "invalid zone ID",
		})
		return
	}

	zone, err := h.zoneService.GetZone(c.Request.Context(), zoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Zone{
		ID:   zone.ID.String(),
		Name: zone.Name,
	})
}
