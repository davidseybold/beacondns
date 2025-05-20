package resolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
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
	Type          Type
	Forwarder     *string
	EtcdEndpoints []string
}

func (c *Config) Validate() error {
	if c.Type == TypeForwarder && c.Forwarder == nil {
		return errors.New("forwarder is required when resolver type is forwarder")
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
		etcd_endpoints {{ join .EtcdEndpoints " " }}
	}
}`

//nolint:gochecknoglobals // used for template parsing
var coreTemplate = template.Must(template.New("corefile").Funcs(template.FuncMap{
	"join": strings.Join,
}).Parse(corefile))

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
