package beacon

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/logger"
	"github.com/davidseybold/beacondns/internal/messaging"
)

//nolint:gochecknoinits // used for plugin registration
func init() {
	log.Info("registering beacon plugin")
	plugin.Register("beacon", setup)
}

func setup(c *caddy.Controller) error {
	beacon, err := beaconParse(c)
	if err != nil {
		return plugin.Error("beacon", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		beacon.Next = next
		return beacon
	})

	c.OnStartup(beacon.OnStartup)
	c.OnFinalShutdown(beacon.OnFinalShutdown)

	return nil
}

func (b *Beacon) OnStartup() error {
	log.Info("starting beacon")
	b.logger.Info("starting beacon")

	store, err := kvstore.NewBadgerKVStore(b.dbPath)
	if err != nil {
		return fmt.Errorf("error creating Badger KV store: %w", err)
	}
	b.store = store

	publishConn, err := messaging.DialAMQP(b.rabbitmqConnString)
	if err != nil {
		return fmt.Errorf("error creating RabbitMQ connection: %w", err)
	}

	publisher, err := messaging.NewRabbitMQPublisher(publishConn, "beacon")
	if err != nil {
		return fmt.Errorf("error creating RabbitMQ publisher: %w", err)
	}

	consumeConn, err := messaging.DialAMQP(b.rabbitmqConnString)
	if err != nil {
		return fmt.Errorf("error creating RabbitMQ connection: %w", err)
	}

	consumer := messaging.NewRabbitMQConsumer(fmt.Sprintf("resolver.%s", b.hostName), consumeConn)

	err = messaging.SetupAMQPTopology(consumeConn, messaging.RabbitMQTopology{
		Exchange: messaging.RabbitMQExchange{
			Name: b.rabbitmqExchange,
			Kind: "topic",
		},
		Queues: []string{
			b.changeQueue,
		},
	})
	if err != nil {
		return fmt.Errorf("error setting up RabbitMQ topology: %w", err)
	}

	changeListener := NewChangeListener(ChangeListenerConfig{
		Consumer:     consumer,
		Publisher:    publisher,
		ChangeQueue:  b.changeQueue,
		OnZoneChange: b.handleZoneChange,
		Logger:       b.logger,
	})

	consumerCtx, consumerCancel := context.WithCancel(context.Background())

	go func() {
		runErr := changeListener.Run(consumerCtx)
		if runErr != nil && runErr != context.Canceled {
			log.Error("error consuming messages: %w", runErr)
		}
	}()

	b.close = func() error {
		consumerCancel()

		var errs []error
		if publishErr := publishConn.Close(); publishErr != nil {
			errs = append(errs, fmt.Errorf("publish connection: %w", publishErr))
		}

		if consumeErr := consumeConn.Close(); consumeErr != nil {
			errs = append(errs, fmt.Errorf("consume connection: %w", consumeErr))
		}

		if storeErr := b.store.Close(); storeErr != nil {
			errs = append(errs, fmt.Errorf("store: %w", storeErr))
		}

		if len(errs) > 0 {
			return fmt.Errorf("failed to close: %v", errs)
		}

		return nil
	}

	err = b.loadZones()
	if err != nil {
		return fmt.Errorf("error loading zones: %w", err)
	}

	return nil
}

func (b *Beacon) OnFinalShutdown() error {
	b.logger.Info("shutting down beacon")
	return b.close()
}

func (b *Beacon) loadZones() error {
	log.Info("loading zones")
	items, err := b.store.GetPrefix([]byte("/zones"))
	if err != nil {
		return fmt.Errorf("error getting zones: %w", err)
	}

	log.Info("loading zones", "count", len(items))

	for _, item := range items {
		log.Info("loading zone", "zone", string(item.Value))
		b.zoneTrie.AddZone(string(item.Value))
	}
	return nil
}

func beaconParse(c *caddy.Controller) (*Beacon, error) {
	if c.Next() {
		if c.NextBlock() {
			return parseAttributes(c)
		}
	}
	return &Beacon{}, nil
}

func parseAttributes(c *caddy.Controller) (*Beacon, error) {
	beacon := &Beacon{
		zoneTrie: NewZoneTrie(),
		logger:   logger.NewJSONLogger(slog.LevelInfo, os.Stdout),
	}

	for {
		switch c.Val() {
		case "hostname":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			beacon.hostName = c.Val()
		case "db_path":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			beacon.dbPath = c.Val()
		case "rabbitmq_conn_string":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			beacon.rabbitmqConnString = c.Val()
		case "rabbitmq_exchange":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			beacon.rabbitmqExchange = c.Val()
		case "change_queue":
			if !c.NextArg() {
				return nil, c.ArgErr()
			}
			beacon.changeQueue = c.Val()
		default:
			if c.Val() != "}" {
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
		if !c.Next() {
			break
		}
	}

	return beacon, nil
}
