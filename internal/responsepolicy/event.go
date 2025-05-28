package responsepolicy

import (
	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/model"
)

const (
	EventTypeCreateResponsePolicy     = "CREATE_RESPONSE_POLICY"
	EventTypeUpdateResponsePolicy     = "UPDATE_RESPONSE_POLICY"
	EventTypeDeleteResponsePolicy     = "DELETE_RESPONSE_POLICY"
	EventTypeEnableResponsePolicy     = "ENABLE_RESPONSE_POLICY"
	EventTypeDisableResponsePolicy    = "DISABLE_RESPONSE_POLICY"
	EventTypeCreateResponsePolicyRule = "CREATE_RESPONSE_POLICY_RULE"
	EventTypeUpdateResponsePolicyRule = "UPDATE_RESPONSE_POLICY_RULE"
	EventTypeDeleteResponsePolicyRule = "DELETE_RESPONSE_POLICY_RULE"
)

type CreateResponsePolicyEvent struct {
	ResponsePolicy *model.ResponsePolicy `json:"responsePolicy"`
}

func NewCreateResponsePolicyEvent(responsePolicy *model.ResponsePolicy) *model.Event {
	return model.NewEvent(EventTypeCreateResponsePolicy, &CreateResponsePolicyEvent{
		ResponsePolicy: responsePolicy,
	})
}

type UpdateResponsePolicyEvent struct {
	ResponsePolicy *model.ResponsePolicy `json:"responsePolicy"`
}

func NewUpdateResponsePolicyEvent(responsePolicy *model.ResponsePolicy) *model.Event {
	return model.NewEvent(EventTypeUpdateResponsePolicy, &UpdateResponsePolicyEvent{
		ResponsePolicy: responsePolicy,
	})
}

type DeleteResponsePolicyEvent struct {
	ResponsePolicy *model.ResponsePolicy `json:"responsePolicy"`
}

func NewDeleteResponsePolicyEvent(responsePolicy *model.ResponsePolicy) *model.Event {
	return model.NewEvent(EventTypeDeleteResponsePolicy, &DeleteResponsePolicyEvent{
		ResponsePolicy: responsePolicy,
	})
}

type EnableResponsePolicyEvent struct {
	ResponsePolicy        *model.ResponsePolicy `json:"responsePolicy"`
	ResponsePolicyRuleIDs []uuid.UUID           `json:"responsePolicyRuleIds"`
}

func NewEnableResponsePolicyEvent(responsePolicy *model.ResponsePolicy, responsePolicyRuleIDs []uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeEnableResponsePolicy, &EnableResponsePolicyEvent{
		ResponsePolicy:        responsePolicy,
		ResponsePolicyRuleIDs: responsePolicyRuleIDs,
	})
}

type DisableResponsePolicyEvent struct {
	ResponsePolicy        *model.ResponsePolicy `json:"responsePolicy"`
	ResponsePolicyRuleIDs []uuid.UUID           `json:"responsePolicyRuleIds"`
}

func NewDisableResponsePolicyEvent(responsePolicy *model.ResponsePolicy, responsePolicyRuleIDs []uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeDisableResponsePolicy, &DisableResponsePolicyEvent{
		ResponsePolicy:        responsePolicy,
		ResponsePolicyRuleIDs: responsePolicyRuleIDs,
	})
}

type CreateResponsePolicyRuleEvent struct {
	ResponsePolicy     *model.ResponsePolicy     `json:"responsePolicy"`
	ResponsePolicyRule *model.ResponsePolicyRule `json:"responsePolicyRule"`
}

func NewCreateResponsePolicyRuleEvent(
	responsePolicy *model.ResponsePolicy,
	responsePolicyRule *model.ResponsePolicyRule,
) *model.Event {
	return model.NewEvent(EventTypeCreateResponsePolicyRule, &CreateResponsePolicyRuleEvent{
		ResponsePolicy:     responsePolicy,
		ResponsePolicyRule: responsePolicyRule,
	})
}

type DeleteResponsePolicyRuleEvent struct {
	ResponsePolicy     *model.ResponsePolicy     `json:"responsePolicy"`
	ResponsePolicyRule *model.ResponsePolicyRule `json:"responsePolicyRule"`
}

func NewDeleteResponsePolicyRuleEvent(
	responsePolicy *model.ResponsePolicy,
	responsePolicyRule *model.ResponsePolicyRule,
) *model.Event {
	return model.NewEvent(EventTypeDeleteResponsePolicyRule, &DeleteResponsePolicyRuleEvent{
		ResponsePolicy:     responsePolicy,
		ResponsePolicyRule: responsePolicyRule,
	})
}

type UpdateResponsePolicyRuleEvent struct {
	ResponsePolicy     *model.ResponsePolicy     `json:"responsePolicy"`
	ResponsePolicyRule *model.ResponsePolicyRule `json:"responsePolicyRule"`
}

func NewUpdateResponsePolicyRuleEvent(
	responsePolicy *model.ResponsePolicy,
	responsePolicyRule *model.ResponsePolicyRule,
) *model.Event {
	return model.NewEvent(EventTypeUpdateResponsePolicyRule, &UpdateResponsePolicyRuleEvent{
		ResponsePolicy:     responsePolicy,
		ResponsePolicyRule: responsePolicyRule,
	})
}
