package recursive

import (
	"net"

	"github.com/miekg/dns"
)

const (
	rootTTL = 518400

	aRootNameserver = "a.root-servers.net."
	bRootNameserver = "b.root-servers.net."
	cRootNameserver = "c.root-servers.net."
	dRootNameserver = "d.root-servers.net."
	eRootNameserver = "e.root-servers.net."
	fRootNameserver = "f.root-servers.net."
	gRootNameserver = "g.root-servers.net."
	hRootNameserver = "h.root-servers.net."
	iRootNameserver = "i.root-servers.net."
	jRootNameserver = "j.root-servers.net."
	kRootNameserver = "k.root-servers.net."
	lRootNameserver = "l.root-servers.net."
	mRootNameserver = "m.root-servers.net."
)

var rootNameserverDomains = []string{
	aRootNameserver,
	bRootNameserver,
	cRootNameserver,
	dRootNameserver,
	eRootNameserver,
	fRootNameserver,
	gRootNameserver,
	hRootNameserver,
	iRootNameserver,
	jRootNameserver,
	kRootNameserver,
	lRootNameserver,
	mRootNameserver,
}

var ipv4RootHints = []dns.RR{
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   aRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("198.41.0.4"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   bRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("170.247.170.2"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   cRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("192.33.4.12"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   dRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("199.7.91.13"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   eRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("192.203.230.10"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   fRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("192.5.5.241"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   gRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    518400,
		},
		A: net.ParseIP("192.112.36.4"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   hRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("198.97.190.53"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   iRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("192.36.148.17"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   jRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("192.58.128.30"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   kRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("193.0.14.129"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   lRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("199.7.83.42"),
	},
	&dns.A{
		Hdr: dns.RR_Header{
			Name:   mRootNameserver,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		A: net.ParseIP("202.12.27.33"),
	},
}

var ipv6RootHints = []dns.RR{
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   aRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2001:503:ba3e::2:30"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   bRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2801:1b8:10::b"),
	},

	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   cRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2001:500:2::c"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   dRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    518400,
		},
		AAAA: net.ParseIP("2001:500:2d::d"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   eRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    518400,
		},
		AAAA: net.ParseIP("2001:500:a8::e"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   fRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    518400,
		},
		AAAA: net.ParseIP("2001:500:2f::f"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   gRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2001:500:12::d0d"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   hRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2001:500:1::53"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   iRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2001:7fe::53"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   jRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2001:503:c27::2:30"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   kRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2001:7fd::1"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   lRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    518400,
		},
		AAAA: net.ParseIP("2001:500:9f::42"),
	},
	&dns.AAAA{
		Hdr: dns.RR_Header{
			Name:   mRootNameserver,
			Rrtype: dns.TypeAAAA,
			Class:  dns.ClassINET,
			Ttl:    rootTTL,
		},
		AAAA: net.ParseIP("2001:dc3::35"),
	},
}

var rootHints = append(ipv4RootHints, ipv6RootHints...)
