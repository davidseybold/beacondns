package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/responsepolicy"
	"github.com/davidseybold/beacondns/internal/zone"
)

func NewHTTPHandler(
	logger *slog.Logger,
	zoneService zone.Service,
	responsePolicyService responsepolicy.Service,
) http.Handler {
	r := gin.Default()

	handler := &handler{
		logger:                logger,
		zoneService:           zoneService,
		responsePolicyService: responsePolicyService,
	}

	r.GET("/health", handler.Health)

	{
		g := r.Group("/v1/zones")
		g.POST("", handler.CreateZone)
		g.DELETE("/:zoneName", handler.DeleteZone)
		g.GET("", handler.ListZones)
		g.GET("/:zoneName", handler.GetZone)
		g.POST("/:zoneName/rrsets", handler.UpsertResourceRecordSet)
		g.GET("/:zoneName/rrsets", handler.ListResourceRecordSets)
		g.DELETE("/:zoneName/rrsets/:name/:type", handler.DeleteResourceRecordSet)
		g.GET("/:zoneName/rrsets/:name/:type", handler.GetResourceRecordSet)
	}

	{
		g := r.Group("/v1/response-policies")
		g.POST("", handler.CreateResponsePolicy)
		// g.GET("", handler.ListPolicies)
		// g.GET("/:policyID", handler.GetPolicy)
		// g.PUT("/:policyID", handler.UpdatePolicy)
		// g.DELETE("/:policyID", handler.DeletePolicy)
	}

	return r
}

type handler struct {
	zoneService           zone.Service
	responsePolicyService responsepolicy.Service

	logger *slog.Logger
}

func (h *handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

func (h *handler) handleError(c *gin.Context, err error) {
	var beaconErr *beaconerr.BeaconError
	if !errors.As(err, &beaconErr) {
		beaconErr = beaconerr.NewBeaconError(beaconerr.ErrorCodeInternalError, err.Error(), err)
	}

	h.logger.Error("api error", "err", err)

	switch {
	case beaconerr.IsNoSuchError(err):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    beaconErr.Code(),
			Message: beaconErr.Message(),
		})
	case beaconerr.IsConflictError(err):
		c.JSON(http.StatusConflict, ErrorResponse{
			Code:    beaconErr.Code(),
			Message: beaconErr.Message(),
		})
	case beaconerr.IsInternalError(err):
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    beaconErr.Code(),
			Message: beaconErr.Message(),
		})
	case beaconerr.IsBadRequestError(err):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    beaconErr.Code(),
			Message: beaconErr.Message(),
		})
	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    beaconerr.ErrorCodeInternalError.String(),
			Message: "An unexpected error occurred",
		})
	}
}
