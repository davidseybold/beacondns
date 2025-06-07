//nolint:forbidigo
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/recursive"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	resolver, err := recursive.NewResolver(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err != nil {
		return fmt.Errorf("failed to create resolver: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if len(os.Args) < 3 {
		return errors.New("usage: dug <domain> <type>")
	}

	domain := os.Args[1]
	domain = dns.Fqdn(domain)

	qType, ok := dns.StringToType[strings.ToUpper(os.Args[2])]
	if !ok {
		return errors.New("invalid type")
	}

	result, err := resolver.Resolve(ctx, domain, qType)
	if err != nil {
		return fmt.Errorf("failed to resolve: %w", err)
	}

	fmt.Println(result.AnswerPacket)
	fmt.Println(";; Query time:", result.Rtt)

	return nil
}
