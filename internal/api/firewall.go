package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/model"
)

type CreateDomainListRequest struct {
	Name    string   `json:"name"    binding:"required"`
	Domains []string `json:"domains" binding:"required"`
}

type DomainList struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DomainCount int       `json:"domainCount"`
}

func (h *handler) CreateDomainList(c *gin.Context) {
	var req CreateDomainListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err)
		return
	}

	info, err := h.firewallService.CreateDomainList(c, req.Name, req.Domains)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, DomainList{
		ID:          info.ID,
		Name:        info.Name,
		DomainCount: info.DomainCount,
	})
}

func (h *handler) DeleteDomainList(c *gin.Context) {
	id, err := getIDParam(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	err = h.firewallService.DeleteDomainList(c, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type AddDomainsToDomainListRequest struct {
	Domains []string `json:"domains" binding:"required"`
}

func (h *handler) AddDomainsToDomainList(c *gin.Context) {
	id, err := getIDParam(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req AddDomainsToDomainListRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err)
		return
	}

	err = h.firewallService.AddDomainsToDomainList(c, id, req.Domains)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type RemoveDomainsFromDomainListRequest struct {
	Domains []string `json:"domains" binding:"required"`
}

func (h *handler) RemoveDomainsFromDomainList(c *gin.Context) {
	id, err := getIDParam(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req RemoveDomainsFromDomainListRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err)
		return
	}

	err = h.firewallService.RemoveDomainsFromDomainList(c, id, req.Domains)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *handler) GetDomainList(c *gin.Context) {
	id, err := getIDParam(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	dl, err := h.firewallService.GetDomainList(c, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, DomainList{
		ID:          dl.ID,
		Name:        dl.Name,
		DomainCount: dl.DomainCount,
	})
}

type ListDomainListDomainsResponse struct {
	Domains []string `json:"domains"`
}

func (h *handler) ListDomainListDomains(c *gin.Context) {
	id, err := getIDParam(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	domains, err := h.firewallService.GetDomainListDomains(c, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ListDomainListDomainsResponse{
		Domains: domains,
	})
}

type ListDomainListsResponse struct {
	DomainLists []DomainList `json:"domainLists"`
}

func (h *handler) ListDomainLists(c *gin.Context) {
	lists, err := h.firewallService.GetDomainLists(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	domainLists := make([]DomainList, 0, len(lists))
	for _, list := range lists {
		domainLists = append(domainLists, DomainList{
			ID:          list.ID,
			Name:        list.Name,
			DomainCount: list.DomainCount,
		})
	}

	c.JSON(http.StatusOK, ListDomainListsResponse{
		DomainLists: domainLists,
	})
}

type CreateFirewallRuleRequest struct {
	Name              string             `json:"name"                        binding:"required"`
	DomainListID      uuid.UUID          `json:"domainListId"                binding:"required"`
	Action            string             `json:"action"                      binding:"required"`
	BlockResponseType *string            `json:"blockResponseType,omitempty"`
	BlockResponse     *ResourceRecordSet `json:"blockResponse,omitempty"`
	Priority          uint               `json:"priority"                    binding:"required"`
}

func (h *handler) CreateFirewallRule(c *gin.Context) {
	var req CreateFirewallRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err)
		return
	}

	rule, err := h.firewallService.CreateFirewallRule(c, &model.FirewallRule{
		Name:              req.Name,
		DomainListID:      req.DomainListID,
		Action:            getFirewallRuleAction(req.Action),
		BlockResponseType: getFirewallRuleBlockResponseType(req.BlockResponseType),
		BlockResponse:     convertAPIResourceRecordSetToModel(req.BlockResponse),
		Priority:          req.Priority,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertModelFirewallRuleToAPI(rule))
}

func (h *handler) DeleteFirewallRule(c *gin.Context) {
	id, err := getIDParam(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	err = h.firewallService.DeleteFirewallRule(c, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *handler) GetFirewallRule(c *gin.Context) {
	id, err := getIDParam(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	rule, err := h.firewallService.GetFirewallRule(c, id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertModelFirewallRuleToAPI(rule))
}

type ListFirewallRulesResponse struct {
	Rules []FirewallRule `json:"rules"`
}

func (h *handler) ListFirewallRules(c *gin.Context) {
	rules, err := h.firewallService.GetFirewallRules(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	apiRules := make([]FirewallRule, 0, len(rules))
	for _, rule := range rules {
		apiRules = append(apiRules, *convertModelFirewallRuleToAPI(&rule))
	}

	c.JSON(http.StatusOK, ListFirewallRulesResponse{
		Rules: apiRules,
	})
}

type UpdateFirewallRuleRequest struct {
	Name              string             `json:"name"                        binding:"required"`
	Action            string             `json:"action"                      binding:"required"`
	DomainListID      uuid.UUID          `json:"domainListId"                binding:"required"`
	BlockResponseType *string            `json:"blockResponseType,omitempty"`
	BlockResponse     *ResourceRecordSet `json:"blockResponse,omitempty"`
	Priority          uint               `json:"priority"                    binding:"required"`
}

func (h *handler) UpdateFirewallRule(c *gin.Context) {
	id, err := getIDParam(c)
	if err != nil {
		h.handleError(c, err)
		return
	}

	var req UpdateFirewallRuleRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err)
		return
	}

	rule, err := h.firewallService.UpdateFirewallRule(c, &model.FirewallRule{
		ID:                id,
		Name:              req.Name,
		DomainListID:      req.DomainListID,
		Action:            getFirewallRuleAction(req.Action),
		BlockResponseType: getFirewallRuleBlockResponseType(req.BlockResponseType),
		BlockResponse:     convertAPIResourceRecordSetToModel(req.BlockResponse),
		Priority:          req.Priority,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, convertModelFirewallRuleToAPI(rule))
}
