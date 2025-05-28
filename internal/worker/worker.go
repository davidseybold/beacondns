package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/davidseybold/beacondns/internal/beaconerr"
	"github.com/davidseybold/beacondns/internal/log"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

var (
	errNoProcessorFound = errors.New("no processor found for event type")
)

type EventProcessor interface {
	ProcessEvent(ctx context.Context, event *model.Event) error
	Events() []string
}

type Worker struct {
	logger           *slog.Logger
	eventToProcessor map[string]EventProcessor
	registry         repository.TransactorRegistry
}

func New(registry repository.TransactorRegistry, l *slog.Logger, eventProcessors []EventProcessor) *Worker {
	if l == nil {
		l = log.NewDiscardLogger()
	}

	eventToProcessor := make(map[string]EventProcessor)
	for _, processor := range eventProcessors {
		for _, event := range processor.Events() {
			eventToProcessor[event] = processor
		}
	}

	return &Worker{
		logger:           l,
		eventToProcessor: eventToProcessor,
		registry:         registry,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := w.processEvents(ctx)
			if err != nil {
				w.logger.ErrorContext(ctx, "failed to process change", "error", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (w *Worker) processEvents(ctx context.Context) error {
	err := w.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) error {
		event, err := r.GetEventRepository().GetEventWithLock(ctx)
		if err != nil {
			if beaconerr.IsNoSuchError(err) {
				return nil
			}
			return err
		}

		if event == nil {
			return nil
		}

		processor, ok := w.eventToProcessor[event.Type]
		if !ok {
			return errNoProcessorFound
		}

		err = processor.ProcessEvent(ctx, event)
		if err != nil {
			return fmt.Errorf("error processing event: %w", err)
		}

		err = r.GetEventRepository().DeleteEvent(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("failed to delete event: %w", err)
		}

		w.logger.Info("processed event", "event", event.Type)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to process event: %w", err)
	}

	return nil
}
