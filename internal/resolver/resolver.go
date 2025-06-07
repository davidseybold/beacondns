package resolver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/coredns/caddy"

	"github.com/coredns/coredns/core/dnsserver"
	_ "github.com/coredns/coredns/plugin/any"     // used for plugin registration.
	_ "github.com/coredns/coredns/plugin/cache"   // used for plugin registration.
	_ "github.com/coredns/coredns/plugin/errors"  // used for plugin registration.
	_ "github.com/coredns/coredns/plugin/forward" // used for plugin registration.
	_ "github.com/coredns/coredns/plugin/log"     // used for plugin registration.

	_ "github.com/davidseybold/beacondns/internal/resolver/plugin/beaconauth"     // used for plugin registration.
	_ "github.com/davidseybold/beacondns/internal/resolver/plugin/beaconfirewall" // used for plugin registration.
	_ "github.com/davidseybold/beacondns/internal/resolver/plugin/beaconresolve"  // used for plugin registration.
)

//nolint:gochecknoinits // used for plugin registration.
func init() {
	//nolint:reassign // used to register custom plugin in addition to the default ones.
	dnsserver.Directives = []string{
		"debug",
		"errors",
		"log",
		"any",
		"cache",
		"beaconfirewall", // custom plugin
		"beaconauth",     // custom plugin
		"beaconresolve",  // custom plugin
		"forward",
	}
}

type Type string

const (
	TypeForwarder Type = "forwarder"
	TypeRecursive Type = "recursive"
)

type Config struct {
	Type          Type
	Forwarders    []string
	EtcdEndpoints []string
	DebugMode     bool
}

func (c *Config) Validate() error {
	if c.Type == TypeForwarder && len(c.Forwarders) == 0 {
		return errors.New("at least one forwarder is required when resolver type is forwarder")
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
    forward . {{ join .Forwarders " " }}
    {{ end }}
    {{ if eq .Type "recursive" }}
    beaconresolve
    {{ end }}
    errors
    log
	{{ if .DebugMode }}
	debug
	{{ end }}
    beaconauth {
		etcd_endpoints {{ join .EtcdEndpoints " " }}
	}
	beaconfirewall {
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
