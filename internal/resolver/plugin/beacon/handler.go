package beacon

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

func (b *Beacon) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qname := state.Name()
	qtype := state.QType()

	zone := b.zoneTrie.FindLongestMatch(qname)

	if zone == "" {
		blog.Debug("no zone found, forwarding to next plugin")
		return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
	}

	answers, ok := b.lookup(zone, qname, dns.Type(qtype))
	if !ok {
		blog.Debug("no answers found, returning not found response")
		return b.notFoundResponse(zone, state), nil
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	m.Answer = append(m.Answer, answers...)

	state.SizeAndDo(m)
	m = state.Scrub(m)
	_ = w.WriteMsg(m)

	return dns.RcodeSuccess, nil
}

func (b *Beacon) notFoundResponse(zone string, state request.Request) int {
	m := new(dns.Msg)
	m.SetRcode(state.Req, dns.RcodeNameError)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	soa, ok := b.lookup(zone, zone, dns.Type(dns.TypeSOA))
	if ok {
		m.Ns = append(m.Ns, soa...)
	}

	state.SizeAndDo(m)
	_ = state.W.WriteMsg(m)
	return dns.RcodeSuccess
}
