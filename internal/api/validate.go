package api

import (
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/davidseybold/beacondns/internal/model"
)

var validators = map[string]validator.Func{
	"firewallRuleAction":            validateFirewallRuleAction,
	"firewallRuleBlockResponseType": validateFirewallRuleBlockResponseType,
}

func validateFirewallRuleAction(fl validator.FieldLevel) bool {
	action := strings.ToUpper(fl.Field().String())
	_, ok := model.ValidFirewallRuleActions[model.FirewallRuleAction(action)]

	return ok
}

func validateFirewallRuleBlockResponseType(fl validator.FieldLevel) bool {
	responseType := strings.ToUpper(fl.Field().String())
	_, ok := model.ValidFirewallRuleBlockResponseTypes[model.FirewallRuleBlockResponseType(responseType)]

	return ok
}

func registerStructValidators(v *validator.Validate) {
	v.RegisterStructValidation(firewallRuleRequestStructValidation, FirewallRuleRequest{})
}

func firewallRuleRequestStructValidation(sl validator.StructLevel) {
	req, ok := sl.Current().Interface().(FirewallRuleRequest)
	if !ok {
		return
	}

	if strings.ToUpper(req.Action) == string(model.FirewallRuleActionBlock) && req.BlockResponseType == nil {
		sl.ReportError(req.BlockResponseType, "blockResponseType", "BlockResponseType", "required_if_action_block", "")
	} else if strings.ToUpper(req.Action) == string(model.FirewallRuleActionBlock) && strings.ToUpper(*req.BlockResponseType) == string(model.FirewallRuleBlockResponseTypeOverride) && req.BlockResponse == nil {
		sl.ReportError(req.BlockResponse, "blockResponse", "BlockResponse", "required_if_action_block_and_block_response_type_override", "")
	}
}
