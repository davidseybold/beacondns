package outbox

import (
	"context"
	"log"
	"time"

	"github.com/davidseybold/beacondns/internal/controller/repository"
	"github.com/davidseybold/beacondns/internal/controller/usecase"
	"github.com/google/uuid"
)

type Processor struct {
	ctx           context.Context
	repoRegistry  repository.Registry
	outboxService usecase.OutboxService
	batchSize     int
}

func NewProcessor(ctx context.Context, r repository.Registry, s usecase.OutboxService, batchSize int) *Processor {
	return &Processor{
		ctx:           ctx,
		repoRegistry:  r,
		outboxService: s,
		batchSize:     batchSize,
	}
}

func (p *Processor) Run() error {
	go p.process()
	<-p.ctx.Done()

	return p.ctx.Err()
}

func (p *Processor) process() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.processOutbox()
		}
	}
}

func (p *Processor) processOutbox() {
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
