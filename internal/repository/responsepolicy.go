package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"

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
		DELETE FROM response_policy_rules WHERE id = $1 AND response_policy_id = $2
	`

	getResponsePolicy = `
		SELECT id, name, description, priority, enabled
		FROM response_policies
		WHERE id = $1
	`
	getResponsePolicyRule = `
		SELECT id, name, trigger_type, trigger_value, action_type, local_data
		FROM response_policy_rules
		WHERE id = $1 AND response_policy_id = $2
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
		SET name = $3, trigger_type = $4, trigger_value = $5, action_type = $6, local_data = $7
		WHERE id = $2 AND response_policy_id = $1
		RETURNING id, name, trigger_type, trigger_value, action_type, local_data
	`
	updateResponsePolicy = `
		UPDATE response_policies
		SET name = $2, description = $3, priority = $4, enabled = $5
		WHERE id = $1
		RETURNING id, name, description, priority, enabled
	`

	toggleResponsePolicy = `
		UPDATE response_policies
		SET enabled = $2
		WHERE id = $1
		RETURNING id, name, description, priority, enabled
	`

	getResponsePolicies = `
		SELECT id, name, trigger_type, trigger_value, action_type, local_data
		FROM response_policy_rules
		WHERE response_policy_id = $1 AND id = ANY($2)
	`
)

type ResponsePolicyRepository interface {
	CreateResponsePolicy(ctx context.Context, policy *model.ResponsePolicy) (*model.ResponsePolicy, error)
	UpdateResponsePolicy(ctx context.Context, policy *model.ResponsePolicy) (*model.ResponsePolicy, error)
	GetResponsePolicy(ctx context.Context, id uuid.UUID) (*model.ResponsePolicy, error)
	ListResponsePolicies(ctx context.Context) ([]model.ResponsePolicy, error)
	DeleteResponsePolicy(ctx context.Context, id uuid.UUID) error
	ToggleResponsePolicy(ctx context.Context, id uuid.UUID, enabled bool) error

	CreateResponsePolicyRule(
		ctx context.Context,
		policyID uuid.UUID,
		rule *model.ResponsePolicyRule,
	) (*model.ResponsePolicyRule, error)
	GetResponsePolicyRule(ctx context.Context, policyID uuid.UUID, id uuid.UUID) (*model.ResponsePolicyRule, error)
	UpdateResponsePolicyRule(
		ctx context.Context,
		policyID uuid.UUID,
		rule *model.ResponsePolicyRule,
	) (*model.ResponsePolicyRule, error)
	DeleteResponsePolicyRule(ctx context.Context, policyID uuid.UUID, id uuid.UUID) error
	ListResponsePolicyRules(ctx context.Context, policyID uuid.UUID) ([]model.ResponsePolicyRule, error)
	GetResponsePolicies(ctx context.Context, policyID uuid.UUID, ruleIDs []uuid.UUID) ([]model.ResponsePolicyRule, error)
}

var _ ResponsePolicyRepository = (*PostgresResponsePolicyRepository)(nil)

type PostgresResponsePolicyRepository struct {
	db postgres.Queryer
}

func (p *PostgresResponsePolicyRepository) CreateResponsePolicy(
	ctx context.Context,
	policy *model.ResponsePolicy,
) (*model.ResponsePolicy, error) {
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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, ErrEntityAlreadyExists
			}
		}

		return nil, fmt.Errorf("failed to scan response policy row: %w", err)
	}

	return policy, nil
}

func (p *PostgresResponsePolicyRepository) UpdateResponsePolicy(
	ctx context.Context,
	policy *model.ResponsePolicy,
) (*model.ResponsePolicy, error) {
	row := p.db.QueryRow(
		ctx,
		updateResponsePolicy,
		policy.ID,
		policy.Name,
		policy.Description,
		policy.Priority,
		policy.Enabled,
	)
	if err := row.Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Enabled); err != nil {
		return nil, fmt.Errorf("failed to scan response policy row: %w", err)
	}

	return policy, nil
}

func (p *PostgresResponsePolicyRepository) ToggleResponsePolicy(ctx context.Context, id uuid.UUID, enabled bool) error {
	_, err := p.db.Exec(ctx, toggleResponsePolicy, id, enabled)
	if err != nil {
		return fmt.Errorf("failed to toggle response policy: %w", err)
	}

	return nil
}

func (p *PostgresResponsePolicyRepository) CreateResponsePolicyRule(
	ctx context.Context,
	policyID uuid.UUID,
	rule *model.ResponsePolicyRule,
) (*model.ResponsePolicyRule, error) {
	localDataBlob, err := json.Marshal(rule.LocalData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response policy rule local data: %w", err)
	}

	row := p.db.QueryRow(
		ctx,
		insertResponsePolicyRule,
		rule.ID,
		policyID,
		rule.Name,
		rule.TriggerType,
		rule.TriggerValue,
		rule.ActionType,
		localDataBlob,
	)
	if err = row.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.TriggerValue, &rule.ActionType, &rule.LocalData); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, ErrEntityAlreadyExists
			}
		}

		return nil, fmt.Errorf("failed to scan response policy rule row: %w", err)
	}

	return rule, nil
}

func (p *PostgresResponsePolicyRepository) DeleteResponsePolicy(ctx context.Context, id uuid.UUID) error {
	ct, err := p.db.Exec(ctx, deleteResponsePolicy, id)
	if err != nil {
		return fmt.Errorf("failed to delete response policy: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return ErrEntityNotFound
	}

	return nil
}

func (p *PostgresResponsePolicyRepository) DeleteResponsePolicyRule(
	ctx context.Context,
	policyID uuid.UUID,
	id uuid.UUID,
) error {
	ct, err := p.db.Exec(ctx, deleteResponsePolicyRule, id, policyID)
	if err != nil {
		return fmt.Errorf("failed to delete response policy rule: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return ErrEntityNotFound
	}

	return nil
}

func (p *PostgresResponsePolicyRepository) GetResponsePolicy(
	ctx context.Context,
	id uuid.UUID,
) (*model.ResponsePolicy, error) {
	var policy model.ResponsePolicy
	row := p.db.QueryRow(ctx, getResponsePolicy, id)
	if err := row.Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEntityNotFound
		}

		return nil, fmt.Errorf("failed to scan response policy row: %w", err)
	}

	return &policy, nil
}

func (p *PostgresResponsePolicyRepository) GetResponsePolicyRule(
	ctx context.Context,
	policyID uuid.UUID,
	id uuid.UUID,
) (*model.ResponsePolicyRule, error) {
	var rule model.ResponsePolicyRule
	row := p.db.QueryRow(ctx, getResponsePolicyRule, id, policyID)
	if err := row.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.TriggerValue, &rule.ActionType, &rule.LocalData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEntityNotFound
		}

		return nil, fmt.Errorf("failed to scan response policy rule row: %w", err)
	}

	return &rule, nil
}

func (p *PostgresResponsePolicyRepository) ListResponsePolicies(ctx context.Context) ([]model.ResponsePolicy, error) {
	rows, err := p.db.Query(ctx, listResponsePolicies)
	if err != nil {
		return nil, fmt.Errorf("failed to query response policies: %w", err)
	}
	defer rows.Close()

	var policies []model.ResponsePolicy
	for rows.Next() {
		var policy model.ResponsePolicy
		if err = rows.Scan(&policy.ID, &policy.Name, &policy.Description, &policy.Priority, &policy.Enabled); err != nil {
			return nil, fmt.Errorf("failed to scan response policy row: %w", err)
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
		return nil, fmt.Errorf("failed to query response policy rules: %w", err)
	}
	defer rows.Close()

	var rules []model.ResponsePolicyRule
	for rows.Next() {
		var rule model.ResponsePolicyRule
		if err = rows.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.TriggerValue, &rule.ActionType, &rule.LocalData); err != nil {
			return nil, fmt.Errorf("failed to scan response policy rule row: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (p *PostgresResponsePolicyRepository) UpdateResponsePolicyRule(
	ctx context.Context,
	policyID uuid.UUID,
	rule *model.ResponsePolicyRule,
) (*model.ResponsePolicyRule, error) {
	localDataBlob, err := json.Marshal(rule.LocalData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response policy rule local data: %w", err)
	}

	row := p.db.QueryRow(
		ctx,
		updateResponsePolicyRule,
		policyID,
		rule.ID,
		rule.Name,
		rule.TriggerType,
		rule.TriggerValue,
		rule.ActionType,
		localDataBlob,
	)
	if err = row.Scan(&rule.ID, &rule.Name, &rule.TriggerType, &rule.TriggerValue, &rule.ActionType, &rule.LocalData); err != nil {
		return nil, fmt.Errorf("failed to scan response policy rule row: %w", err)
	}

	return rule, nil
}

func (p *PostgresResponsePolicyRepository) GetResponsePolicies(
	ctx context.Context,
	policyID uuid.UUID,
	ruleIDs []uuid.UUID,
) ([]model.ResponsePolicyRule, error) {
	rows, err := p.db.Query(ctx, getResponsePolicies, policyID, ruleIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query response policies: %w", err)
	}
	defer rows.Close()

	var policies []model.ResponsePolicyRule
	for rows.Next() {
		var policy model.ResponsePolicyRule
		if err = rows.Scan(&policy.ID, &policy.Name, &policy.TriggerType, &policy.TriggerValue, &policy.ActionType, &policy.LocalData); err != nil {
			return nil, fmt.Errorf("failed to scan response policy rule row: %w", err)
		}
		policies = append(policies, policy)
	}

	return policies, nil
}
