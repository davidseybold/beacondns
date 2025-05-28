package responsepolicy

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

type EventProcessor struct {
	repository repository.TransactorRegistry
	logger     *slog.Logger
}

type EventProcessorDeps struct {
	Repository repository.TransactorRegistry
	Logger     *slog.Logger
}

func NewEventProcessor(deps *EventProcessorDeps) *EventProcessor {
	return &EventProcessor{
		repository: deps.Repository,
		logger:     deps.Logger,
	}
}

func (p *EventProcessor) Events() []string {
	return []string{
		EventTypeCreateResponsePolicy,
		EventTypeUpdateResponsePolicy,
		EventTypeDeleteResponsePolicy,
		EventTypeCreateResponsePolicyRule,
		EventTypeUpdateResponsePolicyRule,
		EventTypeDeleteResponsePolicyRule,
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
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (p *EventProcessor) processCreateResponsePolicyEvent(_ context.Context, _ *model.Event) error {
	return nil
}

func (p *EventProcessor) processUpdateResponsePolicyEvent(_ context.Context, _ *model.Event) error {
	return nil
}

func (p *EventProcessor) processDeleteResponsePolicyEvent(_ context.Context, _ *model.Event) error {
	return nil
}

func (p *EventProcessor) processCreateResponsePolicyRuleEvent(_ context.Context, _ *model.Event) error {
	return nil
}

func (p *EventProcessor) processUpdateResponsePolicyRuleEvent(_ context.Context, _ *model.Event) error {
	return nil
}

func (p *EventProcessor) processDeleteResponsePolicyRuleEvent(_ context.Context, _ *model.Event) error {
	return nil
}
