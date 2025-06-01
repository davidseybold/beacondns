package beaconfirewall

import (
	"context"
	"errors"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/dnsstore"
	"github.com/davidseybold/beacondns/internal/model"
)

func (b *BeaconFirewall) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qName := state.QName()

	blog.Debugf("looking up rules for %s", qName)

	ruleIDs, ok := b.ruleLookup.FindMatchingRules(qName)
	if !ok {
		blog.Debugf("no rules found for %s", qName)
		return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
	}

	rule, err := b.getRuleToApply(ctx, ruleIDs)
	if err != nil {
		blog.Debugf("error getting rule to apply: %v", err)
		return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
	}

	switch rule.Action {
	case model.FirewallRuleActionAllow:
		return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
	case model.FirewallRuleActionAlert:
		// TODO: alert
		return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
	case model.FirewallRuleActionBlock:
		return b.blockResponse(state, w, rule)
	}

	return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
}

func (b *BeaconFirewall) blockResponse(
	state request.Request,
	w dns.ResponseWriter,
	rule *dnsstore.FirewallRule,
) (int, error) {
	msg := new(dns.Msg)
	msg.Authoritative, msg.RecursionAvailable, msg.Compress = true, false, true

	if rule.BlockResponseType == nil {
		return dns.RcodeServerFailure, errors.New("block response type is nil")
	}
	var rCode int
	switch *rule.BlockResponseType {
	case model.FirewallRuleBlockResponseTypeNXDOMAIN:
		blog.Debugf("blocking %s with NXDOMAIN", state.QName())
		rCode = dns.RcodeNameError
	case model.FirewallRuleBlockResponseTypeNODATA:
		blog.Debugf("blocking %s with NODATA", state.QName())
		rCode = dns.RcodeSuccess
	case model.FirewallRuleBlockResponseTypeOverride:
		blog.Debugf("blocking %s with OVERRIDE", state.QName())
		msg.Answer = rule.BlockResponse
		rCode = dns.RcodeSuccess
	default:
		blog.Errorf("unknown block response type: %s", *rule.BlockResponseType)
		return dns.RcodeServerFailure, errors.New("unknown block response type")
	}

	msg.Ns = []dns.RR{createSOARecord()}

	msg.SetRcode(state.Req, rCode)
	state.SizeAndDo(msg)
	msg = state.Scrub(msg)

	err := w.WriteMsg(msg)
	if err != nil {
		return dns.RcodeServerFailure, err
	}

	return rCode, nil
}

func createSOARecord() *dns.SOA {
	soa := new(dns.SOA)
	soa.Hdr = dns.RR_Header{
		Name:   "firewall.beacondns.org.",
		Rrtype: dns.TypeSOA,
		Class:  dns.ClassINET,
		Ttl:    3600,
	}

	soa.Mbox = "hostmaster.beacondns.org."
	soa.Ns = "ns1.beacondns.org."
	soa.Serial = 1
	soa.Refresh = 3600
	soa.Retry = 600
	soa.Expire = 86400

	return soa
}
