package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/davidseybold/beacondns/internal/dnsstore"
	"github.com/davidseybold/beacondns/internal/logger"
	"github.com/davidseybold/beacondns/internal/model"
	"github.com/davidseybold/beacondns/internal/repository"
)

type Worker struct {
	logger   *slog.Logger
	registry repository.TransactorRegistry
	store    dnsstore.DNSStore
}

func New(registry repository.TransactorRegistry, store dnsstore.DNSStore, l *slog.Logger) *Worker {
	if l == nil {
		l = logger.NewDiscardLogger()
	}

	return &Worker{
		registry: registry,
		store:    store,
		logger:   l,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := w.processChange(ctx)
			if err != nil {
				w.logger.Error("failed to process change", "error", err)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (w *Worker) processChange(ctx context.Context) error {
	_, err := w.registry.InTx(ctx, func(ctx context.Context, r repository.Registry) (any, error) {
		change, err := r.GetChangeRepository().GetChangeToProcess(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get change to process: %w", err)
		}

		if change == nil {
			return nil, nil
		}

		switch change.Type {
		case model.ChangeTypeZone:
			err = w.processZoneChange(ctx, change)
			if err != nil {
				return nil, fmt.Errorf("failed to process zone change: %w", err)
			}
		}

		w.logger.Info("processed change", "change", change.ZoneChange.ZoneName)

		err = r.GetChangeRepository().UpdateChangeStatus(ctx, change.ID, model.ChangeStatusInSync)
		if err != nil {
			return nil, fmt.Errorf("failed to update change status: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("failed to process change: %w", err)
	}

	return nil
}
