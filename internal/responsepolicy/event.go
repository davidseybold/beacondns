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
	ResponsePolicyID      uuid.UUID   `json:"responsePolicyId"`
	ResponsePolicyRuleIDs []uuid.UUID `json:"responsePolicyRuleIds"`
}

func NewEnableResponsePolicyEvent(responsePolicyID uuid.UUID, responsePolicyRuleIDs []uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeEnableResponsePolicy, &EnableResponsePolicyEvent{
		ResponsePolicyID:      responsePolicyID,
		ResponsePolicyRuleIDs: responsePolicyRuleIDs,
	})
}

type DisableResponsePolicyEvent struct {
	ResponsePolicyID      uuid.UUID   `json:"responsePolicyId"`
	ResponsePolicyRuleIDs []uuid.UUID `json:"responsePolicyRuleIds"`
}

func NewDisableResponsePolicyEvent(responsePolicyID uuid.UUID, responsePolicyRuleIDs []uuid.UUID) *model.Event {
	return model.NewEvent(EventTypeDisableResponsePolicy, &DisableResponsePolicyEvent{
		ResponsePolicyID:      responsePolicyID,
		ResponsePolicyRuleIDs: responsePolicyRuleIDs,
	})
}

type CreateResponsePolicyRuleEvent struct {
	ResponsePolicyID   uuid.UUID                 `json:"responsePolicyId"`
	ResponsePolicyRule *model.ResponsePolicyRule `json:"responsePolicyRule"`
}

func NewCreateResponsePolicyRuleEvent(
	responsePolicyID uuid.UUID,
	responsePolicyRule *model.ResponsePolicyRule,
) *model.Event {
	return model.NewEvent(EventTypeCreateResponsePolicyRule, &CreateResponsePolicyRuleEvent{
		ResponsePolicyID:   responsePolicyID,
		ResponsePolicyRule: responsePolicyRule,
	})
}

type DeleteResponsePolicyRuleEvent struct {
	ResponsePolicyID   uuid.UUID                 `json:"responsePolicyId"`
	ResponsePolicyRule *model.ResponsePolicyRule `json:"responsePolicyRule"`
}

func NewDeleteResponsePolicyRuleEvent(
	responsePolicyID uuid.UUID,
	responsePolicyRule *model.ResponsePolicyRule,
) *model.Event {
	return model.NewEvent(EventTypeDeleteResponsePolicyRule, &DeleteResponsePolicyRuleEvent{
		ResponsePolicyID:   responsePolicyID,
		ResponsePolicyRule: responsePolicyRule,
	})
}

type UpdateResponsePolicyRuleEvent struct {
	ResponsePolicyID   uuid.UUID                 `json:"responsePolicyId"`
	ResponsePolicyRule *model.ResponsePolicyRule `json:"responsePolicyRule"`
}

func NewUpdateResponsePolicyRuleEvent(
	responsePolicyID uuid.UUID,
	responsePolicyRule *model.ResponsePolicyRule,
) *model.Event {
	return model.NewEvent(EventTypeUpdateResponsePolicyRule, &UpdateResponsePolicyRuleEvent{
		ResponsePolicyID:   responsePolicyID,
		ResponsePolicyRule: responsePolicyRule,
	})
}
