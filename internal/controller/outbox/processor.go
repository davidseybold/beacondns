package outbox

import (
	"context"
	"log"
	"time"

	"github.com/davidseybold/beacondns/internal/controller/usecase"
)

const (
	defaultBatchSize = 10
)

type Processor struct {
	ctx           context.Context
	outboxService usecase.OutboxService
	batchSize     int
}

func NewProcessor(ctx context.Context, s usecase.OutboxService, batchSize int) *Processor {
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}

	return &Processor{
		ctx:           ctx,
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
			_, err := p.outboxService.ProcessNextBatch(context.Background(), p.batchSize)
			if err != nil {
				log.Println("Error processing")
			}
		}
	}
}
