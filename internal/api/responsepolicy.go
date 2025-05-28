package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/model"
)

func (h *handler) CreateResponsePolicy(c *gin.Context) {
	var req CreateResponsePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, translateGinBindingError(err))
		return
	}

	p, err := h.responsePolicyService.CreateResponsePolicy(c.Request.Context(), &model.ResponsePolicy{
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, CreateResponsePolicyResponse{
		ResponsePolicy: convertModelResponsePolicyToAPI(p),
	})
}

func (h *handler) ListResponsePolicies(c *gin.Context) {
	policies, err := h.responsePolicyService.ListResponsePolicies(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ListResponsePoliciesResponse{
		ResponsePolicies: convertModelResponsePoliciesToAPI(policies),
	})
}

func (h *handler) GetResponsePolicy(c *gin.Context) {
	policyID, err := getUUIDParam(c, "policyID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	policy, err := h.responsePolicyService.GetResponsePolicy(c.Request.Context(), policyID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, GetResponsePolicyResponse{
		ResponsePolicy: *convertModelResponsePolicyToAPI(policy),
	})
}

func (h *handler) UpdateResponsePolicy(c *gin.Context) {
	policyID, err := getUUIDParam(c, "policyID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req UpdateResponsePolicyRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, translateGinBindingError(err))
		return
	}

	policy, err := h.responsePolicyService.UpdateResponsePolicy(c.Request.Context(), &model.ResponsePolicy{
		ID:          policyID,
		Name:        req.Name,
		Description: req.Description,
		Priority:    req.Priority,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, UpdateResponsePolicyResponse{
		ResponsePolicy: *convertModelResponsePolicyToAPI(policy),
	})
}

func (h *handler) DeleteResponsePolicy(c *gin.Context) {
	policyID, err := getUUIDParam(c, "policyID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	err = h.responsePolicyService.DeleteResponsePolicy(c.Request.Context(), policyID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *handler) CreateResponsePolicyRule(c *gin.Context) {
	policyID, err := getUUIDParam(c, "policyID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req CreateResponsePolicyRuleRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, translateGinBindingError(err))
		return
	}

	rule, err := h.responsePolicyService.CreateResponsePolicyRule(
		c.Request.Context(),
		policyID,
		&model.ResponsePolicyRule{
			Name:         req.Name,
			TriggerType:  model.ResponsePolicyRuleTriggerType(req.TriggerType),
			TriggerValue: req.TriggerValue,
			ActionType:   model.ResponsePolicyRuleActionType(req.ActionType),
			LocalData:    convertAPIResourceRecordSetsToModel(req.LocalData),
		},
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, CreateResponsePolicyRuleResponse{
		ResponsePolicyRule: *convertModelResponsePolicyRuleToAPI(rule),
	})
}

func (h *handler) ListResponsePolicyRules(c *gin.Context) {
	policyID, err := getUUIDParam(c, "policyID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	rules, err := h.responsePolicyService.ListResponsePolicyRules(c.Request.Context(), policyID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ListResponsePolicyRulesResponse{
		ResponsePolicyRules: convertModelResponsePolicyRulesToAPI(rules),
	})
}

func (h *handler) GetResponsePolicyRule(c *gin.Context) {
	policyID, err := getUUIDParam(c, "policyID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	ruleID, err := getUUIDParam(c, "ruleID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	rule, err := h.responsePolicyService.GetResponsePolicyRule(c.Request.Context(), policyID, ruleID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, GetResponsePolicyRuleResponse{
		ResponsePolicyRule: *convertModelResponsePolicyRuleToAPI(rule),
	})
}

func (h *handler) UpdateResponsePolicyRule(c *gin.Context) {
	policyID, err := getUUIDParam(c, "policyID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	ruleID, err := getUUIDParam(c, "ruleID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req UpdateResponsePolicyRuleRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, translateGinBindingError(err))
		return
	}

	rule, err := h.responsePolicyService.UpdateResponsePolicyRule(
		c.Request.Context(),
		policyID,
		&model.ResponsePolicyRule{
			ID:           ruleID,
			Name:         req.Name,
			TriggerType:  model.ResponsePolicyRuleTriggerType(req.TriggerType),
			TriggerValue: req.TriggerValue,
			ActionType:   model.ResponsePolicyRuleActionType(req.ActionType),
			LocalData:    convertAPIResourceRecordSetsToModel(req.LocalData),
		},
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, UpdateResponsePolicyRuleResponse{
		ResponsePolicyRule: *convertModelResponsePolicyRuleToAPI(rule),
	})
}

func (h *handler) DeleteResponsePolicyRule(c *gin.Context) {
	policyIDParam := c.Param("policyID")

	policyID, err := uuid.Parse(policyIDParam)
	if err != nil {
		h.handleError(c, beaconerr.ErrInvalidArgument("invalid policy ID", "policyID"))
		return
	}

	ruleIDParam := c.Param("ruleID")

	ruleID, err := uuid.Parse(ruleIDParam)
	if err != nil {
		h.handleError(c, beaconerr.ErrInvalidArgument("invalid rule ID", "ruleID"))
		return
	}

	err = h.responsePolicyService.DeleteResponsePolicyRule(c.Request.Context(), policyID, ruleID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *handler) ToggleResponsePolicy(c *gin.Context) {
	policyID, err := getUUIDParam(c, "policyID")
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req ToggleResponsePolicyRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, translateGinBindingError(err))
		return
	}

	err = h.responsePolicyService.ToggleResponsePolicy(c.Request.Context(), policyID, req.Enabled)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
