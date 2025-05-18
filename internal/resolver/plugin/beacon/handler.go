package beacon

import (
	"context"
	"errors"
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"

	"github.com/davidseybold/beacondns/internal/model"
)

func (b *Beacon) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qname := state.Name()
	qtype := state.QType()

	zone := b.zoneTrie.FindLongestMatch(qname)

	if zone == "" {
		log.Info("no zone found, forwarding to next plugin")
		return plugin.NextOrFailure(b.Name(), b.Next, ctx, w, r)
	}

	rrset, ok := b.lookup(zone, qname, model.RRType(state.Type()))
	if !ok {
		return b.errorResponse(state, dns.RcodeServerFailure, errors.New("no rrset found"))
	}

	var answers []dns.RR
	var extras []dns.RR
	switch qtype {
	case dns.TypeA:
		answers, extras = b.A(rrset)
	case dns.TypeAAAA:
		answers, extras = b.AAAA(rrset)
	case dns.TypeCNAME:
		answers, extras = b.CNAME(rrset)
	case dns.TypeNS:
		answers, extras = b.NS(rrset)
	case dns.TypeSOA:
		answers, extras = b.SOA(rrset)
	default:
		return b.errorResponse(
			state,
			dns.RcodeNotImplemented,
			fmt.Errorf("unsupported query type: %s", state.Type()),
		)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	state.SizeAndDo(m)
	m = state.Scrub(m)
	_ = w.WriteMsg(m)

	return dns.RcodeSuccess, nil
}

func (b *Beacon) Name() string { return "beacon" }

func (b *Beacon) errorResponse(state request.Request, rcode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	state.SizeAndDo(m)
	_ = state.W.WriteMsg(m)
	return dns.RcodeSuccess, err
}
