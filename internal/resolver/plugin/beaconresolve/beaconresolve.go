package beaconresolve

import (
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/davidseybold/beacondns/internal/recursive"

	"github.com/coredns/coredns/plugin"
)

const (
	pluginName = "beaconresolve"
)

//nolint:gochecknoglobals // used for logging
var blog = clog.NewWithPlugin(pluginName)

type Config struct {
}

type BeaconResolve struct {
	Next plugin.Handler

	config Config

	resolver *recursive.Resolver
}

var _ plugin.Handler = (*BeaconResolve)(nil)

func (b *BeaconResolve) Name() string { return pluginName }
