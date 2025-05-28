package responsepolicy

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/davidseybold/beacondns/internal/dnsstore"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

type EventProcessor struct {
	repository repository.TransactorRegistry
	logger     *slog.Logger
	store      dnsstore.ResponsePolicyWriter
}

type EventProcessorDeps struct {
	Repository repository.TransactorRegistry
	Logger     *slog.Logger
	DNSStore   dnsstore.ResponsePolicyWriter
}

func (d *EventProcessorDeps) Validate() error {
	if d.Repository == nil {
		return fmt.Errorf("repository is required")
	}

	if d.Logger == nil {
		return fmt.Errorf("logger is required")
	}

	if d.DNSStore == nil {
		return fmt.Errorf("dns store is required")
	}

	return nil
}

func NewEventProcessor(deps *EventProcessorDeps) (*EventProcessor, error) {
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return &EventProcessor{
		repository: deps.Repository,
		logger:     deps.Logger,
		store:      deps.DNSStore,
	}, nil
}

func (p *EventProcessor) Events() []string {
	return []string{
		EventTypeCreateResponsePolicy,
		EventTypeUpdateResponsePolicy,
		EventTypeDeleteResponsePolicy,
		EventTypeCreateResponsePolicyRule,
		EventTypeUpdateResponsePolicyRule,
		EventTypeDeleteResponsePolicyRule,
		EventTypeEnableResponsePolicy,
		EventTypeDisableResponsePolicy,
	}
}

func (p *EventProcessor) ProcessEvent(ctx context.Context, event *model.Event) error {
	switch event.Type {
	case EventTypeCreateResponsePolicy:
		return p.processCreateResponsePolicyEvent(ctx, event)
	case EventTypeUpdateResponsePolicy:
		return p.processUpdateResponsePolicyEvent(ctx, event)
	case EventTypeDeleteResponsePolicy:
		return p.processDeleteResponsePolicyEvent(ctx, event)
	case EventTypeCreateResponsePolicyRule:
		return p.processCreateResponsePolicyRuleEvent(ctx, event)
	case EventTypeUpdateResponsePolicyRule:
		return p.processUpdateResponsePolicyRuleEvent(ctx, event)
	case EventTypeDeleteResponsePolicyRule:
		return p.processDeleteResponsePolicyRuleEvent(ctx, event)
	case EventTypeEnableResponsePolicy:
		return p.processEnableResponsePolicyEvent(ctx, event)
	case EventTypeDisableResponsePolicy:
		return p.processDisableResponsePolicyEvent(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (p *EventProcessor) processCreateResponsePolicyEvent(_ context.Context, _ *model.Event) error {
	p.logger.Debug("nothing to do")
	return nil
}

func (p *EventProcessor) processUpdateResponsePolicyEvent(_ context.Context, _ *model.Event) error {
	p.logger.Debug("nothing to do")
	return nil
}

func (p *EventProcessor) processDeleteResponsePolicyEvent(ctx context.Context, event *model.Event) error {
	var deleteEvent DeleteResponsePolicyEvent
	if err := json.Unmarshal(event.Payload, &deleteEvent); err != nil {
		return err
	}

	return p.store.DeleteResponsePolicyRulesForPolicy(ctx, deleteEvent.ResponsePolicy.ID)
}

func (p *EventProcessor) processCreateResponsePolicyRuleEvent(ctx context.Context, event *model.Event) error {
	var createEvent CreateResponsePolicyRuleEvent
	if err := json.Unmarshal(event.Payload, &createEvent); err != nil {
		return err
	}

	rule := createEvent.ResponsePolicyRule

	storeResponsePolicyRule := dnsstore.ResponsePolicyRule{
		ResponsePolicyRule: *rule,
		Priority:           createEvent.ResponsePolicy.Priority,
		Meta: dnsstore.ResponsePolicyRuleMeta{
			PolicyID: createEvent.ResponsePolicy.ID,
		},
	}

	return p.store.PutResponsePolicyRule(ctx, &storeResponsePolicyRule)
}

func (p *EventProcessor) processUpdateResponsePolicyRuleEvent(ctx context.Context, event *model.Event) error {
	var updateEvent UpdateResponsePolicyRuleEvent
	if err := json.Unmarshal(event.Payload, &updateEvent); err != nil {
		return err
	}

	rule := updateEvent.ResponsePolicyRule
	storeResponsePolicyRule := dnsstore.ResponsePolicyRule{
		ResponsePolicyRule: *rule,
		Priority:           updateEvent.ResponsePolicy.Priority,
		Meta: dnsstore.ResponsePolicyRuleMeta{
			PolicyID: updateEvent.ResponsePolicy.ID,
		},
	}

	return p.store.PutResponsePolicyRule(ctx, &storeResponsePolicyRule)
}

func (p *EventProcessor) processDeleteResponsePolicyRuleEvent(ctx context.Context, event *model.Event) error {
	var deleteEvent DeleteResponsePolicyRuleEvent
	if err := json.Unmarshal(event.Payload, &deleteEvent); err != nil {
		return err
	}

	storeResponsePolicyRule := dnsstore.ResponsePolicyRule{
		ResponsePolicyRule: *deleteEvent.ResponsePolicyRule,
		Priority:           deleteEvent.ResponsePolicy.Priority,
		Meta: dnsstore.ResponsePolicyRuleMeta{
			PolicyID: deleteEvent.ResponsePolicy.ID,
		},
	}
	return p.store.DeleteResponsePolicyRule(ctx, &storeResponsePolicyRule)
}

func (p *EventProcessor) processEnableResponsePolicyEvent(ctx context.Context, event *model.Event) error {
	var enableEvent EnableResponsePolicyEvent
	if err := json.Unmarshal(event.Payload, &enableEvent); err != nil {
		return err
	}

	if len(enableEvent.ResponsePolicyRuleIDs) == 0 {
		return nil
	}

	dbPolicyRules, err := p.repository.GetResponsePolicyRepository().GetResponsePolicies(ctx, enableEvent.ResponsePolicy.ID, enableEvent.ResponsePolicyRuleIDs)
	if err != nil {
		return err
	}

	rules := make([]dnsstore.ResponsePolicyRule, len(dbPolicyRules))
	for i, rule := range dbPolicyRules {
		rules[i] = dnsstore.ResponsePolicyRule{
			ResponsePolicyRule: rule,
			Priority:           enableEvent.ResponsePolicy.Priority,
			Meta: dnsstore.ResponsePolicyRuleMeta{
				PolicyID: enableEvent.ResponsePolicy.ID,
			},
		}
	}

	return p.store.PutResponsePolicyRules(ctx, rules)
}

func (p *EventProcessor) processDisableResponsePolicyEvent(ctx context.Context, event *model.Event) error {
	var disableEvent DisableResponsePolicyEvent
	if err := json.Unmarshal(event.Payload, &disableEvent); err != nil {
		return err
	}

	if len(disableEvent.ResponsePolicyRuleIDs) == 0 {
		return nil
	}

	return p.store.DeleteResponsePolicyRulesForPolicy(ctx, disableEvent.ResponsePolicy.ID)
}
