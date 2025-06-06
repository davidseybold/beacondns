package dnsstore

import (
	"github.com/google/uuid"
	"github.com/miekg/dns"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/davidseybold/beacondns/internal/model"
)

const (
	rrBufSize = 4096
)

type rrSet struct {
	RRs []dns.RR `msg:"rrs"`
}

func (s *rrSet) EncodeMsgpack(enc *msgpack.Encoder) error {
	data := make([][]byte, len(s.RRs))
	for i, rr := range s.RRs {
		buf := make([]byte, rrBufSize)
		n, err := dns.PackRR(rr, buf, 0, nil, false)
		if err != nil {
			return err
		}
		data[i] = buf[:n]
	}
	return enc.Encode(data)
}

func (s *rrSet) DecodeMsgpack(dec *msgpack.Decoder) error {
	data := make([][]byte, 0)
	err := dec.Decode(&data)
	if err != nil {
		return err
	}

	s.RRs = make([]dns.RR, len(data))
	for i, d := range data {
		rr, _, unpackErr := dns.UnpackRR(d, 0)
		if unpackErr != nil {
			return unpackErr
		}
		s.RRs[i] = rr
	}
	return nil
}

func marshalRRSet(rrset *rrSet) ([]byte, error) {
	return msgpack.Marshal(rrset)
}

func unmarshalRRSet(data []byte) (*rrSet, error) {
	rrset := &rrSet{}
	err := msgpack.Unmarshal(data, rrset)
	if err != nil {
		return nil, err
	}
	return rrset, nil
}

type firewallRule struct {
	ID                uuid.UUID                            `msg:"id"`
	Action            model.FirewallRuleAction             `msg:"action"`
	BlockResponseType *model.FirewallRuleBlockResponseType `msg:"blockResponseType,omitempty"`
	BlockResponse     rrSet                                `msg:"blockResponse,omitempty"`
	Priority          uint                                 `msg:"priority"`
}

func marshalFirewallRule(rule *FirewallRule) ([]byte, error) {
	return msgpack.Marshal(rule)
}

func unmarshalFirewallRule(data []byte) (*FirewallRule, error) {
	rule := &firewallRule{}
	err := msgpack.Unmarshal(data, rule)
	if err != nil {
		return nil, err
	}
	return &FirewallRule{
		ID:                rule.ID,
		Action:            rule.Action,
		BlockResponseType: rule.BlockResponseType,
		BlockResponse:     rule.BlockResponse.RRs,
		Priority:          rule.Priority,
	}, nil
}
