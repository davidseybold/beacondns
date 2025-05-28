package responsepolicy

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

type Service interface {
	CreateResponsePolicy(ctx context.Context, policy *model.ResponsePolicy) (*model.ResponsePolicy, error)
	GetResponsePolicy(ctx context.Context, id uuid.UUID) (*model.ResponsePolicy, error)
	ListResponsePolicies(ctx context.Context) ([]model.ResponsePolicy, error)
	UpdateResponsePolicy(ctx context.Context, mod *model.ResponsePolicy) (*model.ResponsePolicy, error)
	DeleteResponsePolicy(ctx context.Context, id uuid.UUID) error
	ToggleResponsePolicy(ctx context.Context, id uuid.UUID, enabled bool) error

	CreateResponsePolicyRule(
		ctx context.Context,
		policyID uuid.UUID,
		rule *model.ResponsePolicyRule,
	) (*model.ResponsePolicyRule, error)
	GetResponsePolicyRule(ctx context.Context, policyID uuid.UUID, id uuid.UUID) (*model.ResponsePolicyRule, error)
	ListResponsePolicyRules(ctx context.Context, policyID uuid.UUID) ([]model.ResponsePolicyRule, error)
	UpdateResponsePolicyRule(
		ctx context.Context,
		policyID uuid.UUID,
		rule *model.ResponsePolicyRule,
	) (*model.ResponsePolicyRule, error)
	DeleteResponsePolicyRule(ctx context.Context, policyID uuid.UUID, id uuid.UUID) error
}

var _ Service = (*DefaultService)(nil)

type DefaultService struct {
	registry repository.TransactorRegistry
}

func NewService(registry repository.TransactorRegistry) *DefaultService {
	return &DefaultService{
		registry: registry,
	}
}

func (d *DefaultService) CreateResponsePolicy(
	ctx context.Context,
	policy *model.ResponsePolicy,
) (*model.ResponsePolicy, error) {
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}

	var newPolicy *model.ResponsePolicy
	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var err error
		newPolicy, err = d.registry.GetResponsePolicyRepository().CreateResponsePolicy(ctx, policy)
		if err != nil {
			return err
		}

		event := NewCreateResponsePolicyEvent(newPolicy)

		err = r.GetEventRepository().CreateEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil && errors.Is(err, repository.ErrEntityAlreadyExists) {
		return nil, beaconerr.ErrResponsePolicyAlreadyExists("response policy already exists")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to create response policy", err)
	}

	return newPolicy, nil
}

func (d *DefaultService) DeleteResponsePolicy(ctx context.Context, id uuid.UUID) error {
	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		policy, err := r.GetResponsePolicyRepository().GetResponsePolicy(ctx, id)
		if err != nil {
			return err
		}

		err = r.GetResponsePolicyRepository().DeleteResponsePolicy(ctx, id)
		if err != nil {
			return err
		}

		event := NewDeleteResponsePolicyEvent(policy)

		err = r.GetEventRepository().CreateEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchResponsePolicy("response policy not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to delete response policy", err)
	}

	return nil
}

func (d *DefaultService) GetResponsePolicy(ctx context.Context, id uuid.UUID) (*model.ResponsePolicy, error) {
	policy, err := d.registry.GetResponsePolicyRepository().GetResponsePolicy(ctx, id)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return nil, beaconerr.ErrNoSuchResponsePolicy("response policy not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get response policy", err)
	}

	return policy, nil
}

func (d *DefaultService) ListResponsePolicies(ctx context.Context) ([]model.ResponsePolicy, error) {
	policies, err := d.registry.GetResponsePolicyRepository().ListResponsePolicies(ctx)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to list response policies", err)
	}

	return policies, nil
}

func (d *DefaultService) UpdateResponsePolicy(
	ctx context.Context,
	mod *model.ResponsePolicy,
) (*model.ResponsePolicy, error) {
	var updatedPolicy *model.ResponsePolicy
	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var err error
		updatedPolicy, err = r.GetResponsePolicyRepository().UpdateResponsePolicy(ctx, mod)
		if err != nil {
			return err
		}

		event := NewUpdateResponsePolicyEvent(updatedPolicy)

		err = r.GetEventRepository().CreateEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to update response policy", err)
	}

	return updatedPolicy, nil
}

func (d *DefaultService) ToggleResponsePolicy(ctx context.Context, id uuid.UUID, enabled bool) error {
	responsePolicy, err := d.registry.GetResponsePolicyRepository().GetResponsePolicy(ctx, id)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchResponsePolicy("response policy not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to get response policy", err)
	}

	// If the response policy is already in the desired state, do nothing.
	if responsePolicy.Enabled == enabled {
		return nil
	}

	responsePolicyRules, err := d.registry.GetResponsePolicyRepository().ListResponsePolicyRules(ctx, id)
	if err != nil {
		return beaconerr.ErrInternalError("failed to list response policy rules", err)
	}

	responsePolicyRuleIDs := make([]uuid.UUID, len(responsePolicyRules))
	for i, rule := range responsePolicyRules {
		responsePolicyRuleIDs[i] = rule.ID
	}

	return d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		err = r.GetResponsePolicyRepository().ToggleResponsePolicy(ctx, id, enabled)
		if err != nil {
			return beaconerr.ErrInternalError("failed to toggle response policy", err)
		}

		var event *model.Event
		if enabled {
			event = NewEnableResponsePolicyEvent(id, responsePolicyRuleIDs)
		} else {
			event = NewDisableResponsePolicyEvent(id, responsePolicyRuleIDs)
		}

		err = r.GetEventRepository().CreateEvent(ctx, event)
		if err != nil {
			return beaconerr.ErrInternalError("failed to toggle response policy", err)
		}

		return nil
	})
}

func (d *DefaultService) CreateResponsePolicyRule(
	ctx context.Context,
	policyID uuid.UUID,
	rule *model.ResponsePolicyRule,
) (*model.ResponsePolicyRule, error) {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	var newRule *model.ResponsePolicyRule
	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var err error
		newRule, err = d.registry.GetResponsePolicyRepository().CreateResponsePolicyRule(ctx, policyID, rule)
		if err != nil {
			return err
		}

		event := NewCreateResponsePolicyRuleEvent(policyID, newRule)

		err = r.GetEventRepository().CreateEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil && errors.Is(err, repository.ErrEntityAlreadyExists) {
		return nil, beaconerr.ErrResponsePolicyRuleAlreadyExists("response policy rule already exists")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to create response policy rule", err)
	}

	return newRule, nil
}

func (d *DefaultService) DeleteResponsePolicyRule(ctx context.Context,
	policyID uuid.UUID,
	id uuid.UUID,
) error {
	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		rule, err := r.GetResponsePolicyRepository().GetResponsePolicyRule(ctx, policyID, id)
		if err != nil {
			return err
		}

		err = r.GetResponsePolicyRepository().DeleteResponsePolicyRule(ctx, policyID, id)
		if err != nil {
			return err
		}

		event := NewDeleteResponsePolicyRuleEvent(policyID, rule)

		err = r.GetEventRepository().CreateEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchResponsePolicyRule("response policy rule not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to delete response policy rule", err)
	}

	return nil
}

func (d *DefaultService) GetResponsePolicyRule(
	ctx context.Context,
	policyID uuid.UUID,
	id uuid.UUID,
) (*model.ResponsePolicyRule, error) {
	rule, err := d.registry.GetResponsePolicyRepository().GetResponsePolicyRule(ctx, policyID, id)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return nil, beaconerr.ErrNoSuchResponsePolicyRule("response policy rule not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get response policy rule", err)
	}
	return rule, nil
}

func (d *DefaultService) ListResponsePolicyRules(
	ctx context.Context,
	policyID uuid.UUID,
) ([]model.ResponsePolicyRule, error) {
	rules, err := d.registry.GetResponsePolicyRepository().ListResponsePolicyRules(ctx, policyID)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to list response policy rules", err)
	}

	return rules, nil
}

func (d *DefaultService) UpdateResponsePolicyRule(
	ctx context.Context,
	policyID uuid.UUID,
	rule *model.ResponsePolicyRule,
) (*model.ResponsePolicyRule, error) {
	var updatedRule *model.ResponsePolicyRule
	err := d.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var err error
		updatedRule, err = r.GetResponsePolicyRepository().UpdateResponsePolicyRule(ctx, policyID, rule)
		if err != nil {
			return err
		}

		event := NewUpdateResponsePolicyRuleEvent(policyID, updatedRule)

		err = r.GetEventRepository().CreateEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to update response policy rule", err)
	}

	return updatedRule, nil
}
