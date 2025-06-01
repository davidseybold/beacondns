package beaconfirewall

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

func (b *BeaconFirewall) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	_ = request.Request{W: w, Req: r}

	rec := &responseRecorder{
		ResponseWriter: w,
	}

	rc, err := plugin.NextOrFailure(b.Name(), b.Next, ctx, rec, r)
	if err != nil || !plugin.ClientWrite(rc) {
		return rc, err
	}

	resMsg := rec.Msg()

	err = w.WriteMsg(resMsg)
	if err != nil {
		return dns.RcodeServerFailure, err
	}

	return rc, nil
}

type responseRecorder struct {
	dns.ResponseWriter
	msg *dns.Msg
}

func (r *responseRecorder) Msg() *dns.Msg {
	return r.msg
}

func (r *responseRecorder) WriteMsg(res *dns.Msg) error {
	r.msg = res
	return nil
}
