package firewall

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

type Service interface {
	CreateFirewallRule(ctx context.Context, rule *model.FirewallRule) (*model.FirewallRule, error)
	UpdateFirewallRule(ctx context.Context, rule *model.FirewallRule) (*model.FirewallRule, error)
	DeleteFirewallRule(ctx context.Context, id uuid.UUID) error
	GetFirewallRule(ctx context.Context, id uuid.UUID) (*model.FirewallRule, error)
	GetFirewallRules(ctx context.Context) ([]model.FirewallRule, error)

	CreateDomainList(ctx context.Context, name string, domains []string) (*model.DomainListInfo, error)
	DeleteDomainList(ctx context.Context, id uuid.UUID) error
	AddDomainsToDomainList(ctx context.Context, id uuid.UUID, domains []string) error
	RemoveDomainsFromDomainList(ctx context.Context, id uuid.UUID, domains []string) error
	GetDomainList(ctx context.Context, id uuid.UUID) (*model.DomainListInfo, error)
	GetDomainLists(ctx context.Context) ([]model.DomainListInfo, error)
	GetDomainListDomains(ctx context.Context, id uuid.UUID) ([]string, error)
}

type DefaultService struct {
	repReg repository.TransactorRegistry
}

var _ Service = (*DefaultService)(nil)

func NewService(repReg repository.TransactorRegistry) *DefaultService {
	return &DefaultService{
		repReg: repReg,
	}
}

func (d *DefaultService) AddDomainsToDomainList(ctx context.Context, id uuid.UUID, domains []string) error {
	info, err := d.repReg.GetFirewallRepository().GetDomainListInfo(ctx, id)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchDomainList("Domain list not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to add domains to domain list", err)
	}

	fqdnDomains := make([]string, 0, len(domains))
	for _, domain := range domains {
		fqdnDomains = append(fqdnDomains, dns.Fqdn(domain))
	}

	err = d.repReg.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		txErr := r.GetFirewallRepository().AddDomainsToDomainList(ctx, id, fqdnDomains)
		if txErr != nil {
			return txErr
		}

		event := NewDomainListDomainsAddedEvent(id, fqdnDomains, info.LinkedRules)
		if txErr = d.repReg.GetEventRepository().CreateEvent(ctx, event); txErr != nil {
			return fmt.Errorf("failed to save event: %w", txErr)
		}

		return nil
	})
	if err != nil && errors.Is(err, repository.ErrEntityAlreadyExists) {
		return beaconerr.ErrDomainExistsInDomainList("domain already exists in domain list")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to add domains to domain list", err)
	}

	return nil
}

func (d *DefaultService) CreateDomainList(
	ctx context.Context,
	name string,
	domains []string,
) (*model.DomainListInfo, error) {
	fqdnDomains := make([]string, 0, len(domains))
	for _, domain := range domains {
		fqdnDomains = append(fqdnDomains, dns.Fqdn(domain))
	}

	dl := &model.DomainList{
		ID:      uuid.New(),
		Name:    name,
		Domains: fqdnDomains,
	}

	var info *model.DomainListInfo
	err := d.repReg.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var txErr error
		info, txErr = r.GetFirewallRepository().CreateDomainList(ctx, dl)
		if txErr != nil {
			return fmt.Errorf("failed to create domain list: %w", txErr)
		}

		event := NewDomainListCreatedEvent(info.ID, fqdnDomains)
		if txErr = d.repReg.GetEventRepository().CreateEvent(ctx, event); txErr != nil {
			return fmt.Errorf("failed to save event: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to create domain list", err)
	}

	return info, nil
}

func (d *DefaultService) CreateFirewallRule(
	ctx context.Context,
	rule *model.FirewallRule,
) (*model.FirewallRule, error) {
	domains, err := d.repReg.GetFirewallRepository().GetDomainListDomains(ctx, rule.DomainListID)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get domain list domains", err)
	}

	rule.ID = uuid.New()

	var createdRule *model.FirewallRule
	err = d.repReg.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var txErr error
		createdRule, txErr = r.GetFirewallRepository().CreateFirewallRule(ctx, rule)
		if txErr != nil {
			return fmt.Errorf("failed to create firewall rule: %w", txErr)
		}

		event := NewRuleCreatedEvent(createdRule, domains)
		if txErr = d.repReg.GetEventRepository().CreateEvent(ctx, event); txErr != nil {
			return fmt.Errorf("failed to save event: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to create firewall rule", err)
	}

	return createdRule, nil
}

func (d *DefaultService) DeleteDomainList(ctx context.Context, id uuid.UUID) error {
	err := d.repReg.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		txErr := r.GetFirewallRepository().DeleteDomainList(ctx, id)
		if txErr != nil {
			return fmt.Errorf("failed to delete domain list: %w", txErr)
		}

		event := NewDomainListDeletedEvent(id)
		if txErr = d.repReg.GetEventRepository().CreateEvent(ctx, event); txErr != nil {
			return fmt.Errorf("failed to save event: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return beaconerr.ErrInternalError("failed to delete domain list", err)
	}

	return nil
}

func (d *DefaultService) DeleteFirewallRule(ctx context.Context, id uuid.UUID) error {
	rule, err := d.repReg.GetFirewallRepository().GetFirewallRule(ctx, id)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchFirewallRule("firewall rule not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to delete firewall rule", err)
	}

	err = d.repReg.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		txErr := r.GetFirewallRepository().DeleteFirewallRule(ctx, id)
		if txErr != nil {
			return fmt.Errorf("failed to delete firewall rule: %w", txErr)
		}

		event := NewRuleDeletedEvent(rule)
		if txErr = d.repReg.GetEventRepository().CreateEvent(ctx, event); txErr != nil {
			return fmt.Errorf("failed to save event: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return beaconerr.ErrInternalError("failed to delete firewall rule", err)
	}

	return nil
}

func (d *DefaultService) GetDomainList(ctx context.Context, id uuid.UUID) (*model.DomainListInfo, error) {
	info, err := d.repReg.GetFirewallRepository().GetDomainListInfo(ctx, id)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return nil, beaconerr.ErrNoSuchDomainList("domain list not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get domain list", err)
	}

	return info, nil
}

func (d *DefaultService) GetDomainListDomains(ctx context.Context, id uuid.UUID) ([]string, error) {
	domains, err := d.repReg.GetFirewallRepository().GetDomainListDomains(ctx, id)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get domain list domains", err)
	}

	return domains, nil
}

func (d *DefaultService) GetFirewallRule(ctx context.Context, id uuid.UUID) (*model.FirewallRule, error) {
	rule, err := d.repReg.GetFirewallRepository().GetFirewallRule(ctx, id)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return nil, beaconerr.ErrNoSuchFirewallRule("firewall rule not found")
	} else if err != nil {
		return nil, beaconerr.ErrInternalError("failed to get firewall rule", err)
	}

	return rule, nil
}

func (d *DefaultService) GetDomainLists(ctx context.Context) ([]model.DomainListInfo, error) {
	lists, err := d.repReg.GetFirewallRepository().ListDomainLists(ctx)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to list domain lists", err)
	}

	return lists, nil
}

func (d *DefaultService) GetFirewallRules(ctx context.Context) ([]model.FirewallRule, error) {
	rules, err := d.repReg.GetFirewallRepository().ListFirewallRules(ctx)
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to list firewall rules", err)
	}

	return rules, nil
}

func (d *DefaultService) RemoveDomainsFromDomainList(ctx context.Context, id uuid.UUID, domains []string) error {
	dl, err := d.repReg.GetFirewallRepository().GetDomainListInfo(ctx, id)
	if err != nil && errors.Is(err, repository.ErrEntityNotFound) {
		return beaconerr.ErrNoSuchDomainList("domain list not found")
	} else if err != nil {
		return beaconerr.ErrInternalError("failed to remove domains from domain list", err)
	}

	fqdnDomains := make([]string, 0, len(domains))
	for _, domain := range domains {
		fqdnDomains = append(fqdnDomains, dns.Fqdn(domain))
	}

	err = d.repReg.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		txErr := r.GetFirewallRepository().RemoveDomainsFromDomainList(ctx, id, fqdnDomains)
		if txErr != nil {
			return fmt.Errorf("failed to remove domains from domain list: %w", txErr)
		}

		event := NewDomainListDomainsRemovedEvent(id, fqdnDomains, dl.LinkedRules)
		if txErr = d.repReg.GetEventRepository().CreateEvent(ctx, event); txErr != nil {
			return fmt.Errorf("failed to save event: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return beaconerr.ErrInternalError("failed to remove domains from domain list", err)
	}

	return nil
}

func (d *DefaultService) UpdateFirewallRule(
	ctx context.Context,
	rule *model.FirewallRule,
) (*model.FirewallRule, error) {
	var updatedRule *model.FirewallRule
	err := d.repReg.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		var txErr error
		updatedRule, txErr = r.GetFirewallRepository().UpdateFirewallRule(ctx, rule)
		if txErr != nil {
			return fmt.Errorf("failed to update firewall rule: %w", txErr)
		}

		event := NewRuleUpdatedEvent(updatedRule)
		if txErr = d.repReg.GetEventRepository().CreateEvent(ctx, event); txErr != nil {
			return fmt.Errorf("failed to save event: %w", txErr)
		}

		return nil
	})
	if err != nil {
		return nil, beaconerr.ErrInternalError("failed to update firewall rule", err)
	}

	return updatedRule, nil
}
