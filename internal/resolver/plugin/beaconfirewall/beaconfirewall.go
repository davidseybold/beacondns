package beaconfirewall

import (
	"context"
	"errors"
	"fmt"
	"sort"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/google/uuid"

	"github.com/coredns/coredns/plugin"

	"github.com/davidseybold/beacondns/internal/dnsstore"
)

const (
	pluginName = "beaconfirewall"
)

//nolint:gochecknoglobals // used for logging
var blog = clog.NewWithPlugin(pluginName)

type Config struct {
	EtcdEndpoints []string
}

type BeaconFirewall struct {
	Next plugin.Handler

	config Config

	store dnsstore.DNSStore
	close func() error

	ruleLookup *DNTrie[uuid.UUID]
}

var _ plugin.Handler = (*BeaconFirewall)(nil)

func (b *BeaconFirewall) Name() string { return pluginName }

func (b *BeaconFirewall) getRuleToApply(ctx context.Context, ruleIDs []uuid.UUID) (*dnsstore.FirewallRule, error) {
	rules, err := b.store.GetFirewallRules(ctx, ruleIDs)
	if err != nil {
		return nil, fmt.Errorf("error getting firewall rules: %w", err)
	}

	if len(rules) == 0 {
		return nil, errors.New("no rules")
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	return &rules[0], nil
}

func (b *BeaconFirewall) loadRules() error {
	rules, err := b.store.GetFirewallRuleMappings(context.Background())
	if err != nil {
		return fmt.Errorf("error getting firewall rule mappings: %w", err)
	}

	for _, rule := range rules {
		blog.Debugf("loading rule %s for domain %s", rule.RuleID, rule.Domain)
		b.ruleLookup.Insert(rule.Domain, rule.RuleID)
	}

	return nil
}

func (b *BeaconFirewall) listenForRuleChanges(ctx context.Context, ch <-chan dnsstore.FirewallRuleMappingEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ch:
			switch event.Type {
			case dnsstore.FirewallRuleMappingEventTypeCreate:
				b.ruleLookup.Insert(event.Mapping.Domain, event.Mapping.RuleID)
			case dnsstore.FirewallRuleMappingEventTypeDelete:
				b.ruleLookup.RemoveByValue(event.Mapping.RuleID)
			}
		}
	}
}
