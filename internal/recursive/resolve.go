package recursive

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"

	blog "github.com/davidseybold/beacondns/internal/log"
)

const (
	maxCnameDepth = 10
)

var (
	ErrCnameDepthExceeded    = errors.New("cname depth exceeded")
	ErrMaxIterationsExceeded = errors.New("max iterations exceeded")
)

type Result struct {
	Qname        string        // Text string, original question
	Qtype        uint16        // Type code asked for
	AnswerPacket *dns.Msg      // Full answer packet
	Rtt          time.Duration // Time the query took
}

type Resolver struct {
	udpClient dnsClient
	tcpClient dnsClient
	cache     *dnsCache
	logger    *slog.Logger
}

type dnsClient interface {
	ExchangeContext(ctx context.Context, msg *dns.Msg, address string) (*dns.Msg, time.Duration, error)
}

func NewResolver(logger *slog.Logger) (*Resolver, error) {
	udpClient := &dns.Client{}
	tcpClient := &dns.Client{
		Net: "tcp",
	}
	return newResolver(udpClient, tcpClient, logger)
}

func newResolver(udpClient, tcpClient dnsClient, logger *slog.Logger) (*Resolver, error) {
	if logger == nil {
		logger = blog.NewDiscardLogger()
	}
	cache, err := newDNSCache()
	if err != nil {
		return nil, err
	}
	return &Resolver{
		cache:     cache,
		udpClient: udpClient,
		tcpClient: tcpClient,
		logger:    logger,
	}, nil
}

func (r *Resolver) Close() {
	r.cache.Close()
}

func (r *Resolver) Resolve(ctx context.Context, qname string, qtype uint16) (*Result, error) {
	return r.resolve(ctx, qname, qtype) // TODO: Add qState
}

type qState struct {
	QName           string
	QType           uint16
	Nameservers     []string
	NameserverIndex int
}

func (r *Resolver) resolve(ctx context.Context, qName string, qType uint16) (*Result, error) {
	result := &Result{
		Qname: qName,
		Qtype: qType,
	}

	start := time.Now()
	cnameDepth := 0

	var resp *dns.Msg
	var err error

	currentQName := qName

	answers := make([]dns.RR, 0, maxCnameDepth)

	followCnames := qType != dns.TypeCNAME

	iteration := 0
	for iteration < maxCnameDepth {
		resp, err = r.iterativeResolve(ctx, currentQName, qType)
		if err != nil {
			return nil, err
		}

		if !followCnames {
			break
		}

		cnameFound := false
		for _, rr := range resp.Answer {
			if rr.Header().Rrtype == dns.TypeCNAME {
				cname, _ := rr.(*dns.CNAME)

				currentQName = cname.Target
				answers = append(answers, rr)
				cnameDepth++
				cnameFound = true
				break
			}
		}

		if cnameFound {
			iteration++
		} else {
			break
		}
	}

	answerPacket := new(dns.Msg)
	answerPacket.MsgHdr = dns.MsgHdr{
		Id:                 dns.Id(),
		Response:           true,
		Opcode:             dns.OpcodeQuery,
		Authoritative:      false,
		Truncated:          resp.Truncated,
		RecursionAvailable: true,
		Rcode:              resp.Rcode,
	}
	answerPacket.SetQuestion(qName, qType)
	answers = append(answers, resp.Answer...)
	answerPacket.Answer = answers

	answerPacket.Ns = resp.Ns

	result.Rtt = time.Since(start)
	result.AnswerPacket = answerPacket

	return result, nil
}

var stackPool = sync.Pool{
	New: func() any {
		return newQStateStack()
	},
}

func (r *Resolver) iterativeResolve(ctx context.Context, qName string, qType uint16) (*dns.Msg, error) {
	qStateStack, _ := stackPool.Get().(*qStateStack)
	defer stackPool.Put(qStateStack)
	qStateStack.Reset()

	qStateStack.Push(&qState{QName: qName, QType: qType, Nameservers: rootNameserverDomains})

	resp := new(dns.Msg)
	var err error

	for qStateStack.Peek() != nil {
		q := qStateStack.Peek()

		if q.NameserverIndex >= len(q.Nameservers) {
			return nil, ErrMaxIterationsExceeded
		}

		rrs, ok := r.cache.Get(q.QName, q.QType)
		if ok {
			resp = &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Opcode:   dns.OpcodeQuery,
					Rcode:    dns.RcodeSuccess,
				},
				Answer: rrs,
			}
			qStateStack.Pop()
			continue
		}

		rcode, soa, ok := r.cache.GetNegative(q.QName, q.QType)
		if ok {
			resp = &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Response: true,
					Opcode:   dns.OpcodeQuery,
					Rcode:    rcode,
				},

				Ns: []dns.RR{soa},
			}
			qStateStack.Pop()
			continue
		}

		nsDomain := q.Nameservers[q.NameserverIndex]
		nsIPs, ok := r.getNSIPsFromCache(nsDomain)
		if !ok {
			qStateStack.Push(&qState{
				QName:           nsDomain,
				QType:           dns.TypeA,
				Nameservers:     rootNameserverDomains,
				NameserverIndex: 0,
			})

			qStateStack.Push(&qState{
				QName:           nsDomain,
				QType:           dns.TypeAAAA,
				Nameservers:     rootNameserverDomains,
				NameserverIndex: 0,
			})
			continue
		}

		resp, err = r.queryNameservers(ctx, q.QName, q.QType, nsIPs)
		if err != nil {
			return nil, err
		}

		r.cacheDNSResponse(q.QName, q.QType, resp)

		if isReferral(resp) {
			referrals := getNameServerReferrals(resp)
			q.Nameservers = referrals
			q.NameserverIndex = 0
			continue
		}

		qStateStack.Pop()
	}

	return resp, err
}

func (r *Resolver) getNSIPsFromCache(nsName string) ([]string, bool) {
	aRecs, okA := r.cache.Get(nsName, dns.TypeA)
	aaaaRecs, okAAAA := r.cache.Get(nsName, dns.TypeAAAA)

	if !okA && !okAAAA {
		return nil, false
	}

	ips := extractIPs(aRecs, aaaaRecs)
	if len(ips) == 0 {
		return nil, false
	}

	rand.Shuffle(len(ips), func(i, j int) {
		ips[i], ips[j] = ips[j], ips[i]
	})

	return ips, true
}

func extractIPs(aRecs, aaaaRecs []dns.RR) []string {
	var ips []string
	for _, rr := range aRecs {
		if a, ok := rr.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}
	for _, rr := range aaaaRecs {
		if aaaa, ok := rr.(*dns.AAAA); ok {
			ips = append(ips, aaaa.AAAA.String())
		}
	}
	return ips
}

func (r *Resolver) queryNameservers(ctx context.Context, qName string, qType uint16, nsIPs []string) (*dns.Msg, error) {
	msg := new(dns.Msg)
	msg.SetQuestion(qName, qType)
	msg.RecursionDesired = false

	for _, nsIP := range nsIPs {
		resp, err := r.queryNameserver(ctx, nsIP, msg)
		if err != nil {
			continue
		}

		if isMalformedOrTruncated(resp) {
			continue
		}

		if resp.Rcode == dns.RcodeSuccess || resp.Rcode == dns.RcodeNameError {
			return resp, nil
		}
	}

	return nil, fmt.Errorf("failed to lookup %s: all nameservers failed", qName)
}

func (r *Resolver) queryNameserver(ctx context.Context, nsIP string, msg *dns.Msg) (*dns.Msg, error) {
	var resp *dns.Msg
	resp, err := r.requestUDP(ctx, nsIP, msg)
	if err == nil && !resp.Truncated {
		return resp, nil
	} else if err != nil {
		return nil, err
	}

	resp, err = r.requestTCP(ctx, nsIP, msg)
	if err == nil {
		return resp, nil
	}

	return nil, fmt.Errorf("failed to query nameserver %s", nsIP)
}

func (r *Resolver) requestUDP(ctx context.Context, ip string, msg *dns.Msg) (*dns.Msg, error) {
	msg.SetEdns0(1232, false)
	var resp *dns.Msg
	var err error
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	resp, _, err = r.udpClient.ExchangeContext(timeoutCtx, msg, net.JoinHostPort(ip, "53"))
	cancel()

	if err == nil {
		return resp, nil
	}
	return nil, err
}

func (r *Resolver) requestTCP(ctx context.Context, ip string, msg *dns.Msg) (*dns.Msg, error) {
	resp, _, err := r.tcpClient.ExchangeContext(ctx, msg, net.JoinHostPort(ip, "53"))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *Resolver) cacheDNSResponse(qName string, qType uint16, resp *dns.Msg) {
	if isMalformedOrTruncated(resp) {
		return
	} else if resp.Rcode != dns.RcodeSuccess && resp.Rcode != dns.RcodeNameError {
		return
	}

	if isNegative(resp) {
		r.cache.PutNegative(qName, qType, resp.Rcode, getSOA(resp))
		return
	}

	// Always cache answer records
	r.cache.Put(qName, qType, resp.Answer)

	glueTargets := map[string]struct{}{}

	recordsToCache := make([]dns.RR, 0, len(resp.Answer)+len(resp.Ns)+len(resp.Extra))
	// Include Authority records but only if they are delegations
	for _, rr := range resp.Ns {
		if rr.Header().Rrtype == dns.TypeNS && dns.IsSubDomain(rr.Header().Name, qName) {
			ns, _ := rr.(*dns.NS)
			recordsToCache = append(recordsToCache, ns)
			glueTargets[ns.Ns] = struct{}{}
		}
	}

	for _, rr := range resp.Extra {
		if rr.Header().Rrtype != dns.TypeA && rr.Header().Rrtype != dns.TypeAAAA {
			continue
		}

		if _, ok := glueTargets[rr.Header().Name]; ok {
			// Include glue records
			recordsToCache = append(recordsToCache, rr)
		}
	}

	r.cache.PutRecords(recordsToCache)
}

func getNameServerReferrals(resp *dns.Msg) []string {
	referrals := make([]string, 0, len(resp.Ns))
	for _, rr := range resp.Ns {
		if rr.Header().Rrtype == dns.TypeNS {
			ns, _ := rr.(*dns.NS)
			referrals = append(referrals, ns.Ns)
		}
	}
	return referrals
}

func isReferral(resp *dns.Msg) bool {
	return len(resp.Ns) > 0 && len(resp.Answer) == 0 && resp.Rcode == dns.RcodeSuccess && contains(resp.Ns, dns.TypeNS)
}

func isNegative(resp *dns.Msg) bool {
	return (resp.Rcode == dns.RcodeNameError || (len(resp.Answer) == 0 && resp.Rcode == dns.RcodeSuccess)) &&
		contains(resp.Ns, dns.TypeSOA)
}

func contains(rrs []dns.RR, rrtype uint16) bool {
	for _, rr := range rrs {
		if rr.Header().Rrtype == rrtype {
			return true
		}
	}
	return false
}

func getSOA(resp *dns.Msg) *dns.SOA {
	for _, rr := range resp.Ns {
		if rr.Header().Rrtype == dns.TypeSOA {
			soa, _ := rr.(*dns.SOA)
			return soa
		}
	}
	return nil
}

func isMalformedOrTruncated(msg *dns.Msg) bool {
	// 1. Null or incomplete message
	if msg == nil {
		return true
	}

	// 2. Truncated bit set — indicates incomplete UDP response
	if msg.Truncated {
		return true
	}

	// 3. Header opcode must be a query (typically)
	if msg.Opcode != dns.OpcodeQuery {
		return true
	}

	// 4. Weird or invalid RCODEs (optional stricter check)
	if msg.Rcode > dns.RcodeRefused {
		return true
	}

	// 5. Unexpected structure (e.g., question mismatch)
	if len(msg.Question) == 0 {
		return true
	}

	return false
}

type qStateStack struct {
	states []*qState
}

func newQStateStack() *qStateStack {
	return &qStateStack{
		states: make([]*qState, 0, 10),
	}
}

func (s *qStateStack) Push(state *qState) {
	s.states = append(s.states, state)
}

func (s *qStateStack) Pop() *qState {
	if len(s.states) == 0 {
		return nil
	}

	state := s.states[len(s.states)-1]
	s.states = s.states[:len(s.states)-1]
	return state
}

func (s *qStateStack) Peek() *qState {
	if len(s.states) == 0 {
		return nil
	}

	return s.states[len(s.states)-1]
}

func (s *qStateStack) Reset() {
	s.states = s.states[:0]
}
