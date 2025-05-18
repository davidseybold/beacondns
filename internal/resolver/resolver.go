package resolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"text/template"

	"github.com/coredns/caddy"

	// Used for plugin registration.

	"github.com/coredns/coredns/core/dnsserver"
	_ "github.com/coredns/coredns/plugin/any"
	_ "github.com/coredns/coredns/plugin/errors"
	_ "github.com/coredns/coredns/plugin/forward"
	_ "github.com/coredns/coredns/plugin/log"

	// Used for plugin registration.
	_ "github.com/davidseybold/beacondns/internal/resolver/plugin/beacon"
)

func init() {
	dnsserver.Directives = []string{
		"root",
		"metadata",
		"geoip",
		"cancel",
		"tls",
		"timeouts",
		"multisocket",
		"reload",
		"nsid",
		"bufsize",
		"bind",
		"debug",
		"trace",
		"ready",
		"health",
		"pprof",
		"prometheus",
		"errors",
		"log",
		"dnstap",
		"local",
		"dns64",
		"acl",
		"any",
		"chaos",
		"loadbalance",
		"tsig",
		"cache",
		"rewrite",
		"header",
		"dnssec",
		"autopath",
		"minimal",
		"template",
		"transfer",
		"hosts",
		"file",
		"auto",
		"secondary",
		"etcd",
		"loop",
		"beacon", // custom plugin
		"forward",
		"grpc",
		"erratic",
		"whoami",
		"on",
		"sign",
		"view",
	}
}

type Type string

const (
	TypeForwarder Type = "forwarder"
	TypeRecursive Type = "recursive"
)

type Config struct {
	Type               Type
	Forwarder          *string
	HostName           string
	DBPath             string
	RabbitMQConnString string
	RabbitExchange     string
	ChangeQueue        string
}

func (c *Config) Validate() error {
	if c.Type == TypeForwarder && c.Forwarder == nil {
		return errors.New("forwarder is required when resolver type is forwarder")
	}

	if c.RabbitExchange == "" {
		return errors.New("rabbitmq_exchange is required")
	}

	if c.ChangeQueue == "" {
		return errors.New("change_queue is required")
	}

	if c.RabbitMQConnString == "" {
		return errors.New("rabbitmq_conn_string is required")
	}

	if c.HostName == "" {
		return errors.New("hostname is required")
	}

	if c.DBPath == "" {
		return errors.New("db_path is required")
	}

	return nil
}

type Resolver struct {
	caddyInput *caddy.CaddyfileInput
}

func New(config *Config) (*Resolver, error) {
	caddyInput, err := loadCaddyInput(*config)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		caddyInput: caddyInput,
	}, nil
}

func (r *Resolver) Run(ctx context.Context) error {
	instance, err := caddy.Start(r.caddyInput)
	if err != nil {
		return err
	}

	<-ctx.Done()
	err = instance.Stop()
	if err != nil {
		return err
	}

	errs := instance.ShutdownCallbacks()

	// Wait for the instance to stop
	instance.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("error shutting down instance: %v", errs)
	}

	return ctx.Err()
}

const corefile = `
. {
    any
    {{ if eq .Type "forwarder" }}
    forward . {{ .Forwarder }}
    {{ end }}
    errors
    log
	debug
    beacon {
        hostname {{ .HostName }}
        db_path {{ .DBPath }}
        rabbitmq_conn_string {{ .RabbitMQConnString }}
        rabbitmq_exchange {{ .RabbitExchange }}
        change_queue {{ .ChangeQueue }}
}`

//nolint:gochecknoglobals // used for template parsing
var coreTemplate = template.Must(template.New("corefile").Parse(corefile))

func loadCaddyInput(config Config) (*caddy.CaddyfileInput, error) {
	var corefile bytes.Buffer
	if err := coreTemplate.Execute(&corefile, config); err != nil {
		return nil, err
	}

	return &caddy.CaddyfileInput{
		Contents:       corefile.Bytes(),
		Filepath:       "memory",
		ServerTypeName: "dns",
	}, nil
}
