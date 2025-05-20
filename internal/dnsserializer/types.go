package dnsserializer

import (
	"github.com/miekg/dns"
	"github.com/vmihailenco/msgpack/v5"
)

type RRSet struct {
	RRs []dns.RR `msg:"rrs"`
}

func (s *RRSet) EncodeMsgpack(enc *msgpack.Encoder) error {
	data := make([][]byte, len(s.RRs))
	for i, rr := range s.RRs {
		buf := make([]byte, 4096)
		n, err := dns.PackRR(rr, buf, 0, nil, false)
		if err != nil {
			return err
		}
		data[i] = buf[:n]
	}
	return enc.Encode(data)
}

func (s *RRSet) DecodeMsgpack(dec *msgpack.Decoder) error {
	data := make([][]byte, 0)
	err := dec.Decode(&data)
	if err != nil {
		return err
	}

	s.RRs = make([]dns.RR, len(data))
	for i, d := range data {
		rr, _, err := dns.UnpackRR(d, 0)
		if err != nil {
			return err
		}
		s.RRs[i] = rr
	}
	return nil
}

func MarshalRRSet(rrset *RRSet) ([]byte, error) {
	return msgpack.Marshal(rrset)
}

func UnmarshalRRSet(data []byte) (*RRSet, error) {
	rrset := &RRSet{}
	err := msgpack.Unmarshal(data, rrset)
	if err != nil {
		return nil, err
	}
	return rrset, nil
}
