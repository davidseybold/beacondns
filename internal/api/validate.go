package api

import (
	"github.com/go-playground/validator/v10"

	"github.com/davidseybold/beacondns/internal/model"
)

var validators = map[string]validator.Func{
	"firewallRuleAction":            validateFirewallRuleAction,
	"firewallRuleBlockResponseType": validateFirewallRuleBlockResponseType,
}

func validateFirewallRuleAction(fl validator.FieldLevel) bool {
	action := fl.Field().String()
	_, ok := model.ValidFirewallRuleActions[model.FirewallRuleAction(action)]
	return ok
}

func validateFirewallRuleBlockResponseType(fl validator.FieldLevel) bool {
	responseType := fl.Field().String()
	_, ok := model.ValidFirewallRuleBlockResponseTypes[model.FirewallRuleBlockResponseType(responseType)]
	return ok
}
