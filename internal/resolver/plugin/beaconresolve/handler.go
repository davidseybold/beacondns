package beaconresolve

import (
	"context"

	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

func (b *BeaconResolve) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qName := state.QName()
	qType := state.QType()

	result, err := b.resolver.Resolve(ctx, qName, qType)
	if err != nil {
		return dns.RcodeServerFailure, err
	}

	result.AnswerPacket.Id = r.Id

	err = w.WriteMsg(result.AnswerPacket)
	if err != nil {
		return dns.RcodeServerFailure, err
	}

	return result.AnswerPacket.Rcode, nil
}
