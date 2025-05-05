package controller

import (
	"context"
	"fmt"
	"net/http"

	"github.com/davidseybold/beacondns/internal/controller/api"
	"github.com/davidseybold/beacondns/internal/controller/outbox"
	"github.com/davidseybold/beacondns/internal/controller/repository"
	"github.com/davidseybold/beacondns/internal/controller/usecase"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
	"github.com/davidseybold/beacondns/internal/libs/server"
	"github.com/davidseybold/beacondns/internal/libs/supervisor"
)

type ControllerSettings struct {
	Port                   int
	DBHost                 string
	DBUser                 string
	DBPassword             string
	DBPort                 int
	DBName                 string
	OutboxBatchSize        int
	OutboxProcessorEnabled bool
}

func NewServer(ctx context.Context, s ControllerSettings) (*supervisor.Supervisor, error) {

	db, err := postgres.NewConnectionPool(context.Background(), postgres.Config{
		Host:     s.DBHost,
		DBName:   s.DBName,
		User:     s.DBUser,
		Password: s.DBPassword,
		Port:     s.DBPort,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating connection pool: %w", err)
	}

	repoRegistry := repository.NewPostgresRepositoryRegistry(db)

	controllerService := usecase.NewControllerService(repoRegistry)
	outboxService := usecase.NewOutboxService()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: api.NewHTTPHandler(controllerService),
	}

	supervisorOpts := []supervisor.SupervisorOption{
		supervisor.WithProcess(server.NewHTTPServer(httpServer)),
	}

	if s.OutboxProcessorEnabled {
		if s.OutboxBatchSize <= 0 {
			return nil, fmt.Errorf("outbox batch size must be greater than 0")
		}
		outboxProcess := supervisor.WithProcess(outbox.NewProcessor(repoRegistry, outboxService, s.OutboxBatchSize))
		supervisorOpts = append(supervisorOpts, outboxProcess)
	}

	return supervisor.New(supervisorOpts...), nil
}
