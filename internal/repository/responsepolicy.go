package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"github.com/davidseybold/beacondns/internal/db/postgres"
	"github.com/davidseybold/beacondns/internal/model"
)

const (
	insertResponsePolicy = `
		INSERT INTO response_policies (id, name, description, priority, enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, priority, enabled
	`
	insertResponsePolicyRule = `
		INSERT INTO response_policy_rules (id, response_policy_id, name, trigger_type, trigger_value, action_type, local_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, trigger_type, trigger_value, action_type, local_data
	`

	deleteResponsePolicy = `
		DELETE FROM response_policies WHERE id = $1
	`
	deleteResponsePolicyRule = `
		DELETE FROM response_policy_rules WHERE id = $1
	`

	getResponsePolicy = `
		SELECT id, name, description, priority, enabled
		FROM response_policies
		WHERE id = $1
	`
	getResponsePolicyRule = `
		SELECT id, name, trigger_type, trigger_value, action_type, local_data
		FROM response_policy_rules
		WHERE id = $1
	`
	listResponsePolicies = `
		SELECT id, name, description, priority, enabled
		FROM response_policies
	`
	listResponsePolicyRules = `
		SELECT id, name, trigger_type, trigger_value, action_type, local_data
		FROM response_policy_rules
		WHERE response_policy_id = $1
	`

	updateResponsePolicyRule = `
		UPDATE response_policy_rules
		SET name = $2, trigger_type = $3, trigger_value = $4, action_type = $5, local_data = $6
		WHERE id = $1
		RETURNING id, name, trigger_type, trigger_value, action_type, local_data
	`
)

type ResponsePolicyRepository interface {
	CreateResponsePolicy(ctx context.Context, policy model.ResponsePolicy) (model.ResponsePolicy, error)
	GetResponsePolicy(ctx context.Context, id uuid.UUID) (model.ResponsePolicy, error)
	ListResponsePolicies(ctx context.Context) ([]model.ResponsePolicy, error)
	DeleteResponsePolicy(ctx context.Context, id uuid.UUID) error
	CreateResponsePolicyRule(ctx context.Context, rule model.ResponsePolicyRule) (model.ResponsePolicyRule, error)
	GetResponsePolicyRule(ctx context.Context, id uuid.UUID) (model.ResponsePolicyRule, error)
	UpdateResponsePolicyRule(ctx context.Context, rule model.ResponsePolicyRule) (model.ResponsePolicyRule, error)
	DeleteResponsePolicyRule(ctx context.Context, id uuid.UUID) error
	ListResponsePolicyRules(ctx context.Context, policyID uuid.UUID) ([]model.ResponsePolicyRule, error)
}

var _ ResponsePolicyRepository = (*PostgresResponsePolicyRepository)(nil)

type PostgresResponsePolicyRepository struct {
	db postgres.Queryer
}

func (p *PostgresResponsePolicyRepository) CreateResponsePolicy(
	ctx context.Context,
	policy model.ResponsePolicy,
) (model.ResponsePolicy, error) {
	row := p.db.QueryRow(
		ctx,
		insertResponsePolicy,
		policy.ID,
		policy.Name,
		policy.Description,
		policy.Priority,
		policy.Enabled,
	)
	if err := row.Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Enabled); err != nil {
		return model.ResponsePolicy{}, err
	}

	return policy, nil
}

func (p *PostgresResponsePolicyRepository) CreateResponsePolicyRule(
	ctx context.Context,
	rule model.ResponsePolicyRule,
) (model.ResponsePolicyRule, error) {
	localDataBlob, err := json.Marshal(rule.LocalData)
	if err != nil {
		return model.ResponsePolicyRule{}, err
	}

	row := p.db.QueryRow(
		ctx,
		insertResponsePolicyRule,
		rule.ID,
		rule.Name,
		rule.TriggerType,
		rule.TriggerValue,
		rule.ActionType,
		localDataBlob,
	)
	if err = row.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.TriggerValue, &rule.ActionType, &rule.LocalData); err != nil {
		return model.ResponsePolicyRule{}, err
	}

	return rule, nil
}

func (p *PostgresResponsePolicyRepository) DeleteResponsePolicy(ctx context.Context, id uuid.UUID) error {
	_, err := p.db.Exec(ctx, deleteResponsePolicy, id)
	return err
}

func (p *PostgresResponsePolicyRepository) DeleteResponsePolicyRule(ctx context.Context, id uuid.UUID) error {
	_, err := p.db.Exec(ctx, deleteResponsePolicyRule, id)
	return err
}

func (p *PostgresResponsePolicyRepository) GetResponsePolicy(
	ctx context.Context,
	id uuid.UUID,
) (model.ResponsePolicy, error) {
	var policy model.ResponsePolicy
	row := p.db.QueryRow(ctx, getResponsePolicy, id)
	if err := row.Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Enabled); err != nil {
		return model.ResponsePolicy{}, err
	}

	return policy, nil
}

func (p *PostgresResponsePolicyRepository) GetResponsePolicyRule(
	ctx context.Context,
	id uuid.UUID,
) (model.ResponsePolicyRule, error) {
	var rule model.ResponsePolicyRule
	row := p.db.QueryRow(ctx, getResponsePolicyRule, id)
	if err := row.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.TriggerValue, &rule.ActionType, &rule.LocalData); err != nil {
		return model.ResponsePolicyRule{}, err
	}

	return rule, nil
}

func (p *PostgresResponsePolicyRepository) ListResponsePolicies(ctx context.Context) ([]model.ResponsePolicy, error) {
	rows, err := p.db.Query(ctx, listResponsePolicies)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []model.ResponsePolicy
	for rows.Next() {
		var policy model.ResponsePolicy
		if err = rows.Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Enabled); err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

func (p *PostgresResponsePolicyRepository) ListResponsePolicyRules(
	ctx context.Context,
	policyID uuid.UUID,
) ([]model.ResponsePolicyRule, error) {
	rows, err := p.db.Query(ctx, listResponsePolicyRules, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.ResponsePolicyRule
	for rows.Next() {
		var rule model.ResponsePolicyRule
		if err = rows.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.TriggerValue, &rule.ActionType, &rule.LocalData); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (p *PostgresResponsePolicyRepository) UpdateResponsePolicyRule(
	ctx context.Context,
	rule model.ResponsePolicyRule,
) (model.ResponsePolicyRule, error) {
	localDataBlob, err := json.Marshal(rule.LocalData)
	if err != nil {
		return model.ResponsePolicyRule{}, err
	}

	row := p.db.QueryRow(
		ctx,
		updateResponsePolicyRule,
		rule.ID,
		rule.Name,
		rule.TriggerType,
		rule.TriggerValue,
		rule.ActionType,
		localDataBlob,
	)
	if err = row.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.TriggerValue, &rule.ActionType, &rule.LocalData); err != nil {
		return model.ResponsePolicyRule{}, err
	}

	return rule, nil
}
