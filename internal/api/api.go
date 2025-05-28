package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/responsepolicy"
	"github.com/davidseybold/beacondns/internal/zone"
)

func NewHTTPHandler(
	logger *slog.Logger,
	zoneService zone.Service,
	responsePolicyService responsepolicy.Service,
) (http.Handler, error) {
	r := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		for name, fn := range validators {
			err := v.RegisterValidation(name, fn)
			if err != nil {
				return nil, fmt.Errorf("error registering validator: %w", err)
			}
		}
	}

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
		g.GET("", handler.ListResponsePolicies)
		g.GET("/:policyID", handler.GetResponsePolicy)
		g.POST("/:policyID", handler.UpdateResponsePolicy)
		g.DELETE("/:policyID", handler.DeleteResponsePolicy)
		g.POST("/:policyID/toggle", handler.ToggleResponsePolicy)
		g.POST("/:policyID/rules", handler.CreateResponsePolicyRule)
		g.GET("/:policyID/rules", handler.ListResponsePolicyRules)
		g.GET("/:policyID/rules/:ruleID", handler.GetResponsePolicyRule)
		g.POST("/:policyID/rules/:ruleID", handler.UpdateResponsePolicyRule)
		g.DELETE("/:policyID/rules/:ruleID", handler.DeleteResponsePolicyRule)
	}

	return r, nil
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

func translateGinBindingError(err error) error {
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors
	if !errors.As(err, &errs) {
		return err
	}

	if len(errs) == 0 {
		return err
	}

	field := errs[0].Field()

	return beaconerr.ErrInvalidArgument(fmt.Sprintf("invalid %s", field), field)
}

func getUUIDParam(c *gin.Context, paramName string) (uuid.UUID, error) {
	param := c.Param(paramName)

	id, err := uuid.Parse(param)
	if err != nil {
		return uuid.Nil, beaconerr.ErrInvalidArgument("invalid "+paramName, paramName)
	}

	return id, nil
}
