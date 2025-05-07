package api

import (
	"net/http"

	"github.com/davidseybold/beacondns/internal/controller/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func NewHTTPHandler(zoneService usecase.ZoneService, nameServerService usecase.NameServerService) http.Handler {
	r := gin.Default()

	handler := &handler{
		zoneService:       zoneService,
		nameServerService: nameServerService,
	}

	r.GET("/health", handler.Health)
	r.POST("/nameservers", handler.AddNameServer)
	r.GET("/nameservers", handler.ListNameServers)

	r.POST("/zones", handler.CreateZone)

	return r
}

type handler struct {
	zoneService       usecase.ZoneService
	nameServerService usecase.NameServerService
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

func (h *handler) AddNameServer(c *gin.Context) {
	var body AddNameServerRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "BadRequest",
			Message: err.Error(),
		})
		return
	}

	ns, err := h.nameServerService.AddNameServer(c.Request.Context(), body.Name, body.RouteKey, body.IPAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, AddNameServerResponse{
		NameServer: NameServer{
			ID:        ns.ID.String(),
			Name:      ns.Name,
			RouteKey:  ns.RouteKey,
			IPAddress: ns.IPAddress,
		},
	})
}

func (h *handler) ListNameServers(c *gin.Context) {
	nameServers, err := h.nameServerService.ListNameServers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "InternalServerError",
			Message: err.Error(),
		})
		return
	}

	responseBody := ListNameServersResponse{
		NameServers: make([]NameServer, len(nameServers)),
	}

	for i, ns := range nameServers {
		responseBody.NameServers[i] = NameServer{
			ID:        ns.ID.String(),
			Name:      ns.Name,
			RouteKey:  ns.RouteKey,
			IPAddress: ns.IPAddress,
		}
	}

	c.JSON(http.StatusOK, responseBody)
}
