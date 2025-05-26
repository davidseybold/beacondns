package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/model"
)

func (h *handler) CreateResponsePolicy(c *gin.Context) {
	var req CreateResponsePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, beaconerr.ErrInvalidArgument("invalid request body", ""))
		return
	}

	p, err := h.responsePolicyService.CreateResponsePolicy(c.Request.Context(), &model.ResponsePolicy{
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
		Enabled:     req.Enabled,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateResponsePolicyResponse{
		ResponsePolicy: convertModelResponsePolicyToAPI(*p),
	})
}
