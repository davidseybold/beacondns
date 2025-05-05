package outbox

import (
	"context"
	"log"
	"time"

	"github.com/davidseybold/beacondns/internal/controller/repository"
	"github.com/davidseybold/beacondns/internal/controller/usecase"
	"github.com/davidseybold/beacondns/internal/libs/supervisor"
	"github.com/google/uuid"
)

type Processor struct {
	repoRegistry  repository.Registry
	outboxService usecase.OutboxService

	cancel    context.CancelFunc
	done      chan struct{}
	batchSize int
}

var _ supervisor.Process = (*Processor)(nil)

func NewProcessor(r repository.Registry, s usecase.OutboxService, batchSize int) *Processor {
	return &Processor{
		repoRegistry:  r,
		outboxService: s,
		batchSize:     batchSize,
	}
}

func (p *Processor) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.done = make(chan struct{})

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer func() {
			ticker.Stop()
			close(p.done)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.process()
			}
		}
	}()

	return nil
}

func (p *Processor) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	<-p.done
	return nil
}

func (p *Processor) process() {
	pendingMsgs, err := p.repoRegistry.GetOutboxRepository().GetPendingMessages(context.Background(), p.batchSize)
	if err != nil {
		log.Printf("Error fetching pending messages: %v", err)
		return
	}

	if len(pendingMsgs) == 0 {
		log.Println("No pending messages to process")
		return
	}

	msgsIDsToDelete := make([]uuid.UUID, 0, len(pendingMsgs))

	for _, msg := range pendingMsgs {
		if err := p.outboxService.SendOutboxMessage(context.Background(), msg); err != nil {
			log.Printf("Error sending message %s: %v", msg.ID, err)
			continue
		}
		log.Printf("Successfully sent message %s", msg.ID)
		msgsIDsToDelete = append(msgsIDsToDelete, msg.ID)
	}

	if err := p.repoRegistry.GetOutboxRepository().DeleteMessages(context.Background(), msgsIDsToDelete); err != nil {
		log.Printf("Error deleting sent messages: %v", err)
		return
	}
}
