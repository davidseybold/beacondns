package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/recursive"
)

func main() {
	resolver, err := recursive.NewResolver(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err != nil {
		slog.Error("failed to create resolver", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if len(os.Args) < 3 {
		fmt.Println("Usage: dug <domain> <type>")
		return
	}

	domain := os.Args[1]
	domain = dns.Fqdn(domain)

	qType, ok := dns.StringToType[strings.ToUpper(os.Args[2])]
	if !ok {
		fmt.Println("Invalid type")
		return
	}

	result, err := resolver.Resolve(ctx, domain, qType)
	if err != nil {
		slog.Error("failed to resolve", "error", err)
		return
	}

	fmt.Printf("%+v\n", result.AnswerPacket)
	fmt.Printf(";; Query time: %v\n", result.Rtt)
}
