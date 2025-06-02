package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/davidseybold/beacondns/internal/db/postgres"
	"github.com/davidseybold/beacondns/internal/model"
)

const (
	createFirewallRuleQuery = `
		INSERT INTO firewall_rules (name, domain_list_id, action, block_response_type, block_response, priority)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, domain_list_id, action, block_response_type, block_response, priority
	`

	updateFirewallRuleQuery = `
		UPDATE firewall_rules
		SET name = $2, domain_list_id = $3, action = $4, block_response_type = $5, block_response = $6, priority = $7
		WHERE id = $1
		RETURNING id, name, domain_list_id, action, block_response_type, block_response, priority
	`

	createDomainListQuery = `
		INSERT INTO domain_lists (name, is_managed, source_url)
		VALUES ($1, $2, $3)
		RETURNING id, name, is_managed, source_url, updated_at
	`

	createDomainListDomainsQuery = `
		INSERT INTO domain_list_domains (domain_list_id, domain)
		VALUES ($1, $2)
	`

	deleteDomainListDomainsQuery = `
		DELETE FROM domain_list_domains
		WHERE domain_list_id = $1
		AND domain = ANY($2)
	`

	deleteAllDomainListDomainsQuery = `
		DELETE FROM domain_list_domains
		WHERE domain_list_id = $1
	`

	deleteDomainListQuery = `
		DELETE FROM domain_lists
		WHERE id = $1
	`

	deleteFirewallRuleQuery = `
		DELETE FROM firewall_rules
		WHERE id = $1
	`

	getDomainListQuery = `
		SELECT id, name, is_managed, source_url, updated_at
		FROM domain_lists
		WHERE id = $1
	`

	getDomainListInfoQuery = `
		SELECT dl.id, dl.name, dl.is_managed, dl.source_url, dl.updated_at, (SELECT COUNT(*) FROM domain_list_domains dld WHERE dld.domain_list_id = dl.id) AS domain_count
		FROM domain_lists dl
		WHERE id = $1
	`

	getFirewallRuleQuery = `
		SELECT id, name, domain_list_id, action, block_response_type, block_response, priority
		FROM firewall_rules
		WHERE id = $1
	`

	listFirewallRulesQuery = `
		SELECT id, name, domain_list_id, action, block_response_type, block_response, priority
		FROM firewall_rules
		ORDER BY priority ASC
	`

	listDomainListsQuery = `
		SELECT dl.id, dl.name, dl.is_managed, dl.source_url, COUNT(dld.domain) AS domain_count
		FROM domain_lists dl
		INNER JOIN domain_list_domains dld ON dl.id = dld.domain_list_id
		GROUP BY dl.id, dl.name
		ORDER BY name ASC
	`

	getDomainListDomainsQuery = `
		SELECT domain	
		FROM domain_list_domains
		WHERE domain_list_id = $1
	`

	getDomainListLinkedRulesQuery = `
		SELECT fr.id
		FROM firewall_rules fr
		WHERE fr.domain_list_id = $1
	`

	getFirewallRulesByDomainListIDQuery = `
		SELECT id, name, domain_list_id, action, block_response_type, block_response, priority
		FROM firewall_rules
		WHERE domain_list_id = $1
	`
)

type FirewallRepository interface {
	CreateFirewallRule(ctx context.Context, rule *model.FirewallRule) (*model.FirewallRule, error)
	UpdateFirewallRule(ctx context.Context, rule *model.FirewallRule) (*model.FirewallRule, error)
	DeleteFirewallRule(ctx context.Context, id uuid.UUID) error
	GetFirewallRule(ctx context.Context, id uuid.UUID) (*model.FirewallRule, error)
	ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error)
	GetFirewallRulesByDomainListID(ctx context.Context, domainListID uuid.UUID) ([]model.FirewallRule, error)

	CreateDomainList(ctx context.Context, list *model.DomainList) (*model.DomainListInfo, error)
	DeleteDomainList(ctx context.Context, id uuid.UUID) error
	GetDomainList(ctx context.Context, id uuid.UUID) (*model.DomainList, error)
	GetDomainListInfo(ctx context.Context, id uuid.UUID) (*model.DomainListInfo, error)
	GetDomainListDomains(ctx context.Context, id uuid.UUID) ([]string, error)
	ListDomainLists(ctx context.Context) ([]model.DomainListInfo, error)
	OverwriteDomainListDomains(ctx context.Context, id uuid.UUID, domains []string) (*model.DomainListInfo, error)

	AddDomainsToDomainList(ctx context.Context, id uuid.UUID, domains []string) error
	RemoveDomainsFromDomainList(ctx context.Context, id uuid.UUID, domains []string) error
}

var _ FirewallRepository = (*PostgresFirewallRepository)(nil)

type PostgresFirewallRepository struct {
	db postgres.Queryer
}

func (p *PostgresFirewallRepository) CreateDomainList(
	ctx context.Context,
	list *model.DomainList,
) (*model.DomainListInfo, error) {
	row := p.db.QueryRow(ctx, createDomainListQuery, list.Name, list.IsManaged, list.SourceURL)

	var info model.DomainListInfo
	err := row.Scan(&info.ID, &info.Name, &info.IsManaged, &info.SourceURL, &info.LastUpdated)
	if err != nil {
		return nil, handleError(err, "failed to scan domain list: %w", err)
	}

	for _, domain := range list.Domains {
		_, err = p.db.Exec(ctx, createDomainListDomainsQuery, info.ID, domain)
		if err != nil {
			return nil, handleError(err, "failed to execute create domain list domains query: %w", err)
		}
	}

	info.DomainCount = len(list.Domains)

	return &info, nil
}

func (p *PostgresFirewallRepository) insertDomainListDomains(
	ctx context.Context,
	id uuid.UUID,
	domains []string,
) error {
	_, err := p.db.CopyFrom(
		ctx,
		pgx.Identifier{"domain_list_domains"},
		[]string{"domain_list_id", "domain"},
		pgx.CopyFromSlice(len(domains), func(i int) ([]any, error) {
			return []any{id, domains[i]}, nil
		}),
	)
	if err != nil {
		return handleError(err, "failed to execute insert domain list domains query: %w", err)
	}

	return nil
}

func (p *PostgresFirewallRepository) DeleteDomainList(ctx context.Context, id uuid.UUID) error {
	_, err := p.db.Exec(ctx, deleteDomainListQuery, id)
	if err != nil {
		return handleError(err, "failed to execute delete domain list query: %w", err)
	}

	return nil
}

func (p *PostgresFirewallRepository) GetDomainList(ctx context.Context, id uuid.UUID) (*model.DomainList, error) {
	row := p.db.QueryRow(ctx, getDomainListQuery, id)

	var list model.DomainList
	if err := row.Scan(&list.ID, &list.Name, &list.IsManaged, &list.SourceURL, &list.LastUpdated); err != nil {
		return nil, handleError(err, "failed to scan domain list: %w", err)
	}

	domains, err := p.GetDomainListDomains(ctx, list.ID)
	if err != nil {
		return nil, err
	}

	list.Domains = domains

	return &list, nil
}

func (p *PostgresFirewallRepository) GetDomainListInfo(
	ctx context.Context,
	id uuid.UUID,
) (*model.DomainListInfo, error) {
	row := p.db.QueryRow(ctx, getDomainListInfoQuery, id)

	var info model.DomainListInfo
	if err := row.Scan(&info.ID, &info.Name, &info.IsManaged, &info.SourceURL, &info.LastUpdated, &info.DomainCount); err != nil {
		return nil, handleError(err, "failed to scan domain list info: %w", err)
	}

	rows, err := p.db.Query(ctx, getDomainListLinkedRulesQuery, id)
	if err != nil {
		return nil, handleError(err, "failed to execute get domain list linked rules query: %w", err)
	}

	linkedRules := []uuid.UUID{}
	for rows.Next() {
		var ruleID uuid.UUID
		if err = rows.Scan(&ruleID); err != nil {
			return nil, handleError(err, "failed to scan linked rule: %w", err)
		}
		linkedRules = append(linkedRules, ruleID)
	}

	info.LinkedRules = linkedRules

	return &info, nil
}

func (p *PostgresFirewallRepository) ListDomainLists(ctx context.Context) ([]model.DomainListInfo, error) {
	rows, err := p.db.Query(ctx, listDomainListsQuery)
	if err != nil {
		return nil, handleError(err, "failed to execute list domain lists query: %w", err)
	}

	lists := []model.DomainListInfo{}
	for rows.Next() {
		var info model.DomainListInfo
		if err = rows.Scan(&info.ID, &info.Name, &info.IsManaged, &info.SourceURL, &info.DomainCount); err != nil {
			return nil, handleError(err, "failed to scan domain list: %w", err)
		}
		lists = append(lists, info)
	}

	return lists, nil
}

func (p *PostgresFirewallRepository) AddDomainsToDomainList(ctx context.Context, id uuid.UUID, domains []string) error {
	for _, domain := range domains {
		_, err := p.db.Exec(ctx, createDomainListDomainsQuery, id, domain)
		if err != nil {
			return handleError(err, "failed to execute create domain list domains query: %w", err)
		}
	}

	return nil
}

func (p *PostgresFirewallRepository) CreateFirewallRule(
	ctx context.Context,
	rule *model.FirewallRule,
) (*model.FirewallRule, error) {
	var blockResponse []byte
	if rule.BlockResponse != nil {
		var err error
		blockResponse, err = json.Marshal(rule.BlockResponse)
		if err != nil {
			return nil, handleError(err, "failed to marshal block response: %w", err)
		}
	}

	_, err := p.db.Exec(
		ctx,
		createFirewallRuleQuery,
		rule.Name,
		rule.DomainListID,
		rule.Action,
		rule.BlockResponseType,
		blockResponse,
		rule.Priority,
	)
	if err != nil {
		return nil, handleError(err, "failed to execute create firewall rule query: %w", err)
	}

	return rule, nil
}

func (p *PostgresFirewallRepository) DeleteFirewallRule(ctx context.Context, id uuid.UUID) error {
	_, err := p.db.Exec(ctx, deleteFirewallRuleQuery, id)
	return handleError(err, "failed to execute delete firewall rule query: %w", err)
}

func (p *PostgresFirewallRepository) GetFirewallRule(ctx context.Context, id uuid.UUID) (*model.FirewallRule, error) {
	row := p.db.QueryRow(ctx, getFirewallRuleQuery, id)

	var rule model.FirewallRule
	err := row.Scan(
		&rule.ID,
		&rule.Name,
		&rule.DomainListID,
		&rule.Action,
		&rule.BlockResponseType,
		&rule.BlockResponse,
		&rule.Priority,
	)
	if err != nil {
		return nil, handleError(err, "failed to scan firewall rule: %w", err)
	}

	return &rule, nil
}

func (p *PostgresFirewallRepository) GetFirewallRulesByDomainListID(
	ctx context.Context,
	domainListID uuid.UUID,
) ([]model.FirewallRule, error) {
	rows, err := p.db.Query(ctx, getFirewallRulesByDomainListIDQuery, domainListID)
	if err != nil {
		return nil, handleError(err, "failed to execute get firewall rules by domain list id query: %w", err)
	}

	rules := []model.FirewallRule{}
	for rows.Next() {
		var rule model.FirewallRule
		if err = rows.Scan(&rule.ID, &rule.Name, &rule.DomainListID, &rule.Action, &rule.BlockResponseType, &rule.BlockResponse, &rule.Priority); err != nil {
			return nil, handleError(err, "failed to scan firewall rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (p *PostgresFirewallRepository) ListFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	rows, err := p.db.Query(ctx, listFirewallRulesQuery)
	if err != nil {
		return nil, handleError(err, "failed to execute list firewall rules query: %w", err)
	}

	rules := []model.FirewallRule{}
	for rows.Next() {
		var rule model.FirewallRule
		if err = rows.Scan(&rule.ID, &rule.Name, &rule.DomainListID, &rule.Action, &rule.BlockResponseType, &rule.BlockResponse, &rule.Priority); err != nil {
			return nil, handleError(err, "failed to scan firewall rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

func (p *PostgresFirewallRepository) RemoveDomainsFromDomainList(
	ctx context.Context,
	id uuid.UUID,
	domains []string,
) error {
	_, err := p.db.Exec(ctx, deleteDomainListDomainsQuery, id, domains)
	if err != nil {
		return handleError(err, "failed to execute delete domain list domains query: %w", err)
	}

	return nil
}

func (p *PostgresFirewallRepository) UpdateFirewallRule(
	ctx context.Context,
	rule *model.FirewallRule,
) (*model.FirewallRule, error) {
	var blockResponse []byte
	if rule.BlockResponse != nil {
		var err error
		blockResponse, err = json.Marshal(rule.BlockResponse)
		if err != nil {
			return nil, handleError(err, "failed to marshal block response: %w", err)
		}
	}

	_, err := p.db.Exec(
		ctx,
		updateFirewallRuleQuery,
		rule.ID,
		rule.Name,
		rule.DomainListID,
		rule.Action,
		rule.BlockResponseType,
		blockResponse,
		rule.Priority,
	)
	if err != nil {
		return nil, handleError(err, "failed to execute update firewall rule query: %w", err)
	}

	return rule, nil
}

func (p *PostgresFirewallRepository) GetDomainListDomains(ctx context.Context, id uuid.UUID) ([]string, error) {
	rows, err := p.db.Query(ctx, getDomainListDomainsQuery, id)
	if err != nil {
		return nil, handleError(err, "failed to execute get domain list domains query: %w", err)
	}

	domains := []string{}
	for rows.Next() {
		var domain string
		if err = rows.Scan(&domain); err != nil {
			return nil, handleError(err, "failed to scan domain: %w", err)
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

func (p *PostgresFirewallRepository) OverwriteDomainListDomains(
	ctx context.Context,
	id uuid.UUID,
	domains []string,
) (*model.DomainListInfo, error) {
	_, err := p.db.Exec(ctx, deleteAllDomainListDomainsQuery, id)
	if err != nil {
		return nil, handleError(err, "failed to execute delete domain list domains query: %w", err)
	}

	err = p.insertDomainListDomains(ctx, id, domains)
	if err != nil {
		return nil, handleError(err, "failed to execute insert domain list domains query: %w", err)
	}

	info, err := p.GetDomainListInfo(ctx, id)
	if err != nil {
		return nil, handleError(err, "failed to get domain list info: %w", err)
	}

	return info, nil
}
