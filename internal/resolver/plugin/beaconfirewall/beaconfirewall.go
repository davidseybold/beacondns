package beaconfirewall

import (
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/google/uuid"

	"github.com/coredns/coredns/plugin"

	"github.com/davidseybold/beacondns/internal/dnsstore"
)

const (
	pluginName = "beaconfirewall"
)

//nolint:gochecknoglobals // used for logging
var blog = clog.NewWithPlugin(pluginName)

type Config struct {
	EtcdEndpoints []string
}

type ruleMeta struct {
	PolicyID uuid.UUID
	RuleID   uuid.UUID
}

type BeaconFirewall struct {
	Next plugin.Handler

	config Config

	store dnsstore.DNSStore
	close func() error

	ruleLookup *DNTrie[ruleMeta]
}

var _ plugin.Handler = (*BeaconFirewall)(nil)

func (b *BeaconFirewall) Name() string { return pluginName }
