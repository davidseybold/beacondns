package responsepolicy

import (
	"context"

	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

type Service interface {
	CreateResponsePolicy(ctx context.Context, policy *model.ResponsePolicy) (*model.ResponsePolicy, error)
	GetResponsePolicy(ctx context.Context, id uuid.UUID) (*model.ResponsePolicy, error)
	ListResponsePolicies(ctx context.Context) ([]model.ResponsePolicy, error)
	UpdateResponsePolicy(ctx context.Context, mod *model.ResponsePolicy) (*model.ResponsePolicy, error)
	DeleteResponsePolicy(ctx context.Context, id uuid.UUID) error

	CreateResponsePolicyRule(
		ctx context.Context,
		policyID uuid.UUID,
		rule *model.ResponsePolicyRule,
	) (*model.ResponsePolicyRule, error)
	GetResponsePolicyRule(ctx context.Context, id uuid.UUID) (*model.ResponsePolicyRule, error)
	ListResponsePolicyRules(ctx context.Context, policyID uuid.UUID) ([]model.ResponsePolicyRule, error)
	UpdateResponsePolicyRule(ctx context.Context, rule *model.ResponsePolicyRule) (*model.ResponsePolicyRule, error)
	DeleteResponsePolicyRule(ctx context.Context, id uuid.UUID) error
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

	newPolicy, err := d.registry.GetResponsePolicyRepository().CreateResponsePolicy(ctx, policy)
	if err != nil {
		return nil, err
	}

	return newPolicy, nil
}

func (d *DefaultService) CreateResponsePolicyRule(
	ctx context.Context,
	policyID uuid.UUID,
	rule *model.ResponsePolicyRule,
) (*model.ResponsePolicyRule, error) {
	if rule.ID == uuid.Nil {
		rule.ID = uuid.New()
	}

	newRule, err := d.registry.GetResponsePolicyRepository().CreateResponsePolicyRule(ctx, policyID, rule)
	if err != nil {
		return nil, err
	}

	return newRule, nil
}

func (d *DefaultService) DeleteResponsePolicy(ctx context.Context, id uuid.UUID) error {
	return d.registry.GetResponsePolicyRepository().DeleteResponsePolicy(ctx, id)
}

func (d *DefaultService) DeleteResponsePolicyRule(ctx context.Context, id uuid.UUID) error {
	return d.registry.GetResponsePolicyRepository().DeleteResponsePolicyRule(ctx, id)
}

func (d *DefaultService) GetResponsePolicy(ctx context.Context, id uuid.UUID) (*model.ResponsePolicy, error) {
	return d.registry.GetResponsePolicyRepository().GetResponsePolicy(ctx, id)
}

func (d *DefaultService) GetResponsePolicyRule(ctx context.Context, id uuid.UUID) (*model.ResponsePolicyRule, error) {
	return d.registry.GetResponsePolicyRepository().GetResponsePolicyRule(ctx, id)
}

func (d *DefaultService) ListResponsePolicies(ctx context.Context) ([]model.ResponsePolicy, error) {
	return d.registry.GetResponsePolicyRepository().ListResponsePolicies(ctx)
}

func (d *DefaultService) ListResponsePolicyRules(
	ctx context.Context,
	policyID uuid.UUID,
) ([]model.ResponsePolicyRule, error) {
	return d.registry.GetResponsePolicyRepository().ListResponsePolicyRules(ctx, policyID)
}

func (d *DefaultService) UpdateResponsePolicy(
	ctx context.Context,
	mod *model.ResponsePolicy,
) (*model.ResponsePolicy, error) {
	return d.registry.GetResponsePolicyRepository().UpdateResponsePolicy(ctx, mod)
}

func (d *DefaultService) UpdateResponsePolicyRule(
	ctx context.Context,
	rule *model.ResponsePolicyRule,
) (*model.ResponsePolicyRule, error) {
	return d.registry.GetResponsePolicyRepository().UpdateResponsePolicyRule(ctx, rule)
}
