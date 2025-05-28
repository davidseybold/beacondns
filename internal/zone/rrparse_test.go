package zone

import (
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"

	"github.com/davidseybold/beacondns/internal/model"
)

func TestA(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid A record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "A",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "192.168.1.1"},
				},
			},
			want: []dns.RR{
				&dns.A{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					A: net.ParseIP("192.168.1.1"),
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid IPv4 address",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "A",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "256.256.256.256"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidIPv4Address,
		},
		{
			name: "ipv6 address",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "A",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "2001:db8::1"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidIPv4Address,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "A",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := A(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAAAA(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid AAAA record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "AAAA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "2001:db8::1"},
				},
			},
			want: []dns.RR{
				&dns.AAAA{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeAAAA,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					AAAA: net.ParseIP("2001:db8::1"),
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid IPv6 address",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "AAAA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid-ipv6"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidIPv6Address,
		},
		{
			name: "ipv4 address",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "AAAA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "192.168.1.1"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidIPv6Address,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "AAAA",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AAAA(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCNAME(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid CNAME record",
			rrset: &model.ResourceRecordSet{
				Name: "www.example.com.",
				Type: "CNAME",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "example.com"},
				},
			},
			want: []dns.RR{
				&dns.CNAME{
					Hdr: dns.RR_Header{
						Name:   "www.example.com.",
						Rrtype: dns.TypeCNAME,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Target: "example.com.",
				},
			},
			wantErr: nil,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "www.example.com.",
				Type:            "CNAME",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CNAME(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMX(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid MX record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "MX",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "10 mail.example.com"},
				},
			},
			want: []dns.RR{
				&dns.MX{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeMX,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Preference: 10,
					Mx:         "mail.example.com.",
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid MX priority",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "MX",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid mail.example.com"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "invalid MX format",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "MX",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "10"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidMXFieldCount,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "MX",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MX(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNS(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid NS record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "NS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "ns1.example.com"},
				},
			},
			want: []dns.RR{
				&dns.NS{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeNS,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Ns: "ns1.example.com.",
				},
			},
			wantErr: nil,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "NS",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NS(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPTR(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid PTR record",
			rrset: &model.ResourceRecordSet{
				Name: "1.1.168.192.in-addr.arpa.",
				Type: "PTR",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "example.com"},
				},
			},
			want: []dns.RR{
				&dns.PTR{
					Hdr: dns.RR_Header{
						Name:   "1.1.168.192.in-addr.arpa.",
						Rrtype: dns.TypePTR,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Ptr: "example.com.",
				},
			},
			wantErr: nil,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "1.1.168.192.in-addr.arpa.",
				Type:            "PTR",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PTR(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSOA(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid SOA record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SOA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "ns1.example.com. admin.example.com. 2024031501 7200 3600 1209600 86400"},
				},
			},
			want: []dns.RR{
				&dns.SOA{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeSOA,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Ns:      "ns1.example.com.",
					Mbox:    "admin.example.com.",
					Serial:  2024031501,
					Refresh: 7200,
					Retry:   3600,
					Expire:  1209600,
					Minttl:  86400,
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid SOA field count",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SOA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "ns1.example.com. admin.example.com. 2024031501"},
				},
			},
			want:    nil,
			wantErr: ErrSOARecordFieldCount,
		},
		{
			name: "invalid SOA serial",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SOA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "ns1.example.com. admin.example.com. invalid 7200 3600 1209600 86400"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "SOA",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SOA(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSRV(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid SRV record",
			rrset: &model.ResourceRecordSet{
				Name: "_sip._tcp.example.com.",
				Type: "SRV",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "10 20 5060 sip.example.com"},
				},
			},
			want: []dns.RR{
				&dns.SRV{
					Hdr: dns.RR_Header{
						Name:   "_sip._tcp.example.com.",
						Rrtype: dns.TypeSRV,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Priority: 10,
					Weight:   20,
					Port:     5060,
					Target:   "sip.example.com",
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid SRV field count",
			rrset: &model.ResourceRecordSet{
				Name: "_sip._tcp.example.com.",
				Type: "SRV",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "10 20"},
				},
			},
			want:    nil,
			wantErr: ErrSRVRecordFieldCount,
		},
		{
			name: "invalid SRV priority",
			rrset: &model.ResourceRecordSet{
				Name: "_sip._tcp.example.com.",
				Type: "SRV",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid 20 5060 sip.example.com"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "invalid SRV weight",
			rrset: &model.ResourceRecordSet{
				Name: "_sip._tcp.example.com.",
				Type: "SRV",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "10 invalid 5060 sip.example.com"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "invalid SRV port",
			rrset: &model.ResourceRecordSet{
				Name: "_sip._tcp.example.com.",
				Type: "SRV",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "10 20 invalid sip.example.com"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "_sip._tcp.example.com.",
				Type:            "SRV",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SRV(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTXT(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "quoted TXT record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "TXT",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "\"v=spf1 ip4:192.168.1.1 -all\""},
				},
			},
			want: []dns.RR{
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Txt: []string{"v=spf1 ip4:192.168.1.1 -all"},
				},
			},
			wantErr: nil,
		},
		{
			name: "non-quoted TXT value",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "TXT",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "v=spf1 ip4:192.168.1.1 -all"},
				},
			},
			want: []dns.RR{
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Txt: []string{"v=spf1 ip4:192.168.1.1 -all"},
				},
			},
			wantErr: nil,
		},
		{
			name: "mixed quoted and non-quoted TXT value",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "TXT",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "\"v=spf1\" ip4:192.168.1.1 \"-all\""},
				},
			},
			want:    nil,
			wantErr: ErrInvalidCharacterString,
		},
		{
			name: "multiple TXT strings",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "TXT",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "\"v=spf1\" \"ip4:192.168.1.1\" \"-all\""},
				},
			},
			want: []dns.RR{
				&dns.TXT{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeTXT,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Txt: []string{"v=spf1", "ip4:192.168.1.1", "-all"},
				},
			},
			wantErr: nil,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "TXT",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TXT(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCAA(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid CAA record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "CAA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "0 issue \"ca.example.com\""},
				},
			},
			want: []dns.RR{
				&dns.CAA{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeCAA,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Flag:  0,
					Tag:   "issue",
					Value: "ca.example.com",
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid CAA field count",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "CAA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "0 issue"},
				},
			},
			want:    nil,
			wantErr: ErrCAARecordFieldCount,
		},
		{
			name: "invalid CAA flag",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "CAA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid issue \"ca.example.com\""},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "CAA",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CAA(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDS(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid DS record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "DS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "60485 8 2 E3D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291"},
				},
			},
			want: []dns.RR{
				&dns.DS{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeDS,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					KeyTag:     60485,
					Algorithm:  8,
					DigestType: 2,
					Digest:     "E3D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291",
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid DS field count",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "DS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "60485 8"},
				},
			},
			want:    nil,
			wantErr: ErrDSRecordFieldCount,
		},
		{
			name: "invalid DS key tag",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "DS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid 8 2 E3D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291F7D3C291"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "DS",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DS(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHTTPS(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid HTTPS record with all key values",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "HTTPS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{
						Value: `1 . alpn="h2,h3" port=443 ipv4hint="192.168.1.1,192.168.1.2" ipv6hint="2001:db8::1,2001:db8::2" mandatory="alpn,port,ipv4hint,ipv6hint" no-default-alpn`,
					},
				},
			},
			want: []dns.RR{
				&dns.HTTPS{
					SVCB: dns.SVCB{
						Hdr: dns.RR_Header{
							Name:   "example.com.",
							Rrtype: dns.TypeHTTPS,
							Class:  dns.ClassINET,
							Ttl:    300,
						},
						Priority: 1,
						Target:   ".",
						Value: []dns.SVCBKeyValue{
							&dns.SVCBAlpn{
								Alpn: []string{"h2", "h3"},
							},
							&dns.SVCBPort{
								Port: 443,
							},
							&dns.SVCBIPv4Hint{
								Hint: []net.IP{
									net.ParseIP("192.168.1.1"),
									net.ParseIP("192.168.1.2"),
								},
							},
							&dns.SVCBIPv6Hint{
								Hint: []net.IP{
									net.ParseIP("2001:db8::1"),
									net.ParseIP("2001:db8::2"),
								},
							},
							&dns.SVCBMandatory{
								Code: []dns.SVCBKey{
									dns.SVCB_ALPN,
									dns.SVCB_PORT,
									dns.SVCB_IPV4HINT,
									dns.SVCB_IPV6HINT,
								},
							},
							&dns.SVCBNoDefaultAlpn{},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid HTTPS record with minimal key values",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "HTTPS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . alpn=h2"},
				},
			},
			want: []dns.RR{
				&dns.HTTPS{
					SVCB: dns.SVCB{
						Hdr: dns.RR_Header{
							Name:   "example.com.",
							Rrtype: dns.TypeHTTPS,
							Class:  dns.ClassINET,
							Ttl:    300,
						},
						Priority: 1,
						Target:   ".",
						Value: []dns.SVCBKeyValue{
							&dns.SVCBAlpn{
								Alpn: []string{"h2"},
							},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid HTTPS field count",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "HTTPS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1"},
				},
			},
			want:    nil,
			wantErr: ErrSVCBRecordRequiredFieldCount,
		},
		{
			name: "invalid HTTPS priority",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "HTTPS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid . alpn=h2"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "invalid HTTPS port",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "HTTPS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . port=invalid"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "invalid HTTPS IPv4 hint",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "HTTPS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . ipv4hint=\"invalid\""},
				},
			},
			want:    nil,
			wantErr: ErrInvalidIPv4Address,
		},
		{
			name: "invalid HTTPS IPv6 hint",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "HTTPS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . ipv6hint=\"invalid\""},
				},
			},
			want:    nil,
			wantErr: ErrInvalidIPv6Address,
		},
		{
			name: "invalid HTTPS mandatory key",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "HTTPS",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . mandatory=invalid"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidMandatoryKey,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "HTTPS",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HTTPS(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNAPTR(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid NAPTR record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "NAPTR",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "100 50 \"s\" \"http+I2L+I2C+I2R\" \"\" ."},
				},
			},
			want: []dns.RR{
				&dns.NAPTR{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeNAPTR,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Order:       100,
					Preference:  50,
					Flags:       "s",
					Service:     "http+I2L+I2C+I2R",
					Regexp:      "",
					Replacement: ".",
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid NAPTR field count",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "NAPTR",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "100 50 \"s\""},
				},
			},
			want:    nil,
			wantErr: ErrNAPTRRecordFieldCount,
		},
		{
			name: "invalid NAPTR order",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "NAPTR",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid 50 \"s\" \"http+I2L+I2C+I2R\" \"\" _http._tcp.example.com."},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "NAPTR",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NAPTR(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSSHFP(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid SSHFP record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SSHFP",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 1 123456789abcdef67890123456789abcdef67890"},
				},
			},
			want: []dns.RR{
				&dns.SSHFP{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeSSHFP,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Algorithm:   1,
					Type:        1,
					FingerPrint: "123456789abcdef67890123456789abcdef67890",
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid SSHFP field count",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SSHFP",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 1"},
				},
			},
			want:    nil,
			wantErr: ErrSSHFPRecordFieldCount,
		},
		{
			name: "invalid SSHFP algorithm",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SSHFP",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid 1 123456789abcdef67890123456789abcdef67890"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "SSHFP",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SSHFP(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSVCB(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid SVCB record with all key values",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SVCB",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{
						Value: "1 . alpn=h2,h3 port=443 ipv4hint=\"192.168.1.1,192.168.1.2\" ipv6hint=\"2001:db8::1,2001:db8::2\" mandatory=alpn,port,ipv4hint,ipv6hint no-default-alpn",
					},
				},
			},
			want: []dns.RR{
				&dns.SVCB{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeSVCB,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Priority: 1,
					Target:   ".",
					Value: []dns.SVCBKeyValue{
						&dns.SVCBAlpn{
							Alpn: []string{"h2", "h3"},
						},
						&dns.SVCBPort{
							Port: 443,
						},
						&dns.SVCBIPv4Hint{
							Hint: []net.IP{
								net.ParseIP("192.168.1.1"),
								net.ParseIP("192.168.1.2"),
							},
						},
						&dns.SVCBIPv6Hint{
							Hint: []net.IP{
								net.ParseIP("2001:db8::1"),
								net.ParseIP("2001:db8::2"),
							},
						},
						&dns.SVCBMandatory{
							Code: []dns.SVCBKey{
								dns.SVCB_ALPN,
								dns.SVCB_PORT,
								dns.SVCB_IPV4HINT,
								dns.SVCB_IPV6HINT,
							},
						},
						&dns.SVCBNoDefaultAlpn{},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "valid SVCB record with minimal key values",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SVCB",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . alpn=h2"},
				},
			},
			want: []dns.RR{
				&dns.SVCB{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeSVCB,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Priority: 1,
					Target:   ".",
					Value: []dns.SVCBKeyValue{
						&dns.SVCBAlpn{
							Alpn: []string{"h2"},
						},
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid SVCB field count",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SVCB",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1"},
				},
			},
			want:    nil,
			wantErr: ErrSVCBRecordRequiredFieldCount,
		},
		{
			name: "invalid SVCB priority",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SVCB",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid . alpn=h2"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "invalid SVCB port",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SVCB",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . port=invalid"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "invalid SVCB IPv4 hint",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SVCB",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . ipv4hint=\"invalid\""},
				},
			},
			want:    nil,
			wantErr: ErrInvalidIPv4Address,
		},
		{
			name: "invalid SVCB IPv6 hint",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SVCB",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . ipv6hint=\"invalid\""},
				},
			},
			want:    nil,
			wantErr: ErrInvalidIPv6Address,
		},
		{
			name: "invalid SVCB mandatory key",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "SVCB",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "1 . mandatory=invalid"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidMandatoryKey,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "SVCB",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SVCB(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTLSA(t *testing.T) {
	tests := []struct {
		name    string
		rrset   *model.ResourceRecordSet
		want    []dns.RR
		wantErr error
	}{
		{
			name: "valid TLSA record",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "TLSA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "3 1 1 123456789abcdef67890123456789abcdef67890"},
				},
			},
			want: []dns.RR{
				&dns.TLSA{
					Hdr: dns.RR_Header{
						Name:   "example.com.",
						Rrtype: dns.TypeTLSA,
						Class:  dns.ClassINET,
						Ttl:    300,
					},
					Usage:        3,
					Selector:     1,
					MatchingType: 1,
					Certificate:  "123456789abcdef67890123456789abcdef67890",
				},
			},
			wantErr: nil,
		},
		{
			name: "invalid TLSA field count",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "TLSA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "3 1"},
				},
			},
			want:    nil,
			wantErr: ErrTLSARecordFieldCount,
		},
		{
			name: "invalid TLSA usage",
			rrset: &model.ResourceRecordSet{
				Name: "example.com.",
				Type: "TLSA",
				TTL:  300,
				ResourceRecords: []model.ResourceRecord{
					{Value: "invalid 1 1 123456789abcdef67890123456789abcdef67890"},
				},
			},
			want:    nil,
			wantErr: ErrInvalidInteger,
		},
		{
			name: "empty resource records",
			rrset: &model.ResourceRecordSet{
				Name:            "example.com.",
				Type:            "TLSA",
				TTL:             300,
				ResourceRecords: []model.ResourceRecord{},
			},
			want:    nil,
			wantErr: ErrNoResourceRecords,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TLSA(tt.rrset)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
