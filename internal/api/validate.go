package api

import (
	"github.com/go-playground/validator/v10"

	"github.com/davidseybold/beacondns/internal/model"
)

var validators = map[string]validator.Func{
	"responsePolicyRuleTriggerType": responsePolicyRuleTriggerTypeValidator,
	"responsePolicyRuleActionType":  responsePolicyRuleActionTypeValidator,
	"responsePolicyRuleLocalData":   responsePolicyRuleLocalDataValidator,
}

func responsePolicyRuleTriggerTypeValidator(fl validator.FieldLevel) bool {
	triggerType := fl.Field().String()
	return model.ResponsePolicyRuleTriggerType(triggerType).IsValid()
}

func responsePolicyRuleActionTypeValidator(fl validator.FieldLevel) bool {
	actionType := fl.Field().String()
	return model.ResponsePolicyRuleActionType(actionType).IsValid()
}

func responsePolicyRuleLocalDataValidator(fl validator.FieldLevel) bool {

	actionType := fl.Parent().FieldByName("ActionType").String()

	if actionType != model.ResponsePolicyRuleActionTypeLOCALDATA.String() {
		return true
	}

	localData, ok := fl.Field().Interface().([]model.ResourceRecordSet)
	if !ok {
		return false
	}

	if len(localData) == 0 {
		return false
	}

	return true
}
