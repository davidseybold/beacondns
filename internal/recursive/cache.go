package recursive

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/miekg/dns"
)

type dnsCacheValue interface {
	isCacheValue()
}

type cacheEntry struct {
	RRs []dns.RR
}

func (c *cacheEntry) isCacheValue() {}

type negativeCacheEntry struct {
	Rcode int
	SOA   *dns.SOA
}

func (c *negativeCacheEntry) isCacheValue() {}

type dnsCache struct {
	cache     *ristretto.Cache[string, dnsCacheValue]
	rootHints map[string]*dns.RR
}

func newDNSCache() (*dnsCache, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, dnsCacheValue]{
		NumCounters: 1e7,     // 10M entries
		MaxCost:     1 << 30, // 1GB
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	hints := make(map[string]*dns.RR)
	for _, rr := range rootHints {
		hints[key(rr.Header().Name, rr.Header().Rrtype)] = &rr
	}

	return &dnsCache{cache: cache, rootHints: hints}, nil
}

func (c *dnsCache) Close() {
	c.cache.Close()
}

func key(name string, qtype uint16) string {
	return name + ":" + dns.Type(qtype).String()
}

func negativeKey(name string, qtype uint16) string {
	return "negative:" + name + ":" + dns.Type(qtype).String()
}

func (c *dnsCache) Get(name string, qtype uint16) ([]dns.RR, bool) {
	k := key(name, qtype)

	if rr, ok := c.rootHints[k]; ok {
		return []dns.RR{*rr}, true
	}

	val, ok := c.cache.Get(k)
	if !ok {
		return nil, false
	}

	entry, ok := val.(*cacheEntry)
	if !ok {
		return nil, false
	}

	ttl, ok := c.cache.GetTTL(k)
	if !ok {
		return nil, false
	}

	for _, rr := range entry.RRs {
		rr.Header().Ttl = uint32(ttl.Seconds())
	}

	return entry.RRs, true
}

func (c *dnsCache) Put(name string, qtype uint16, rrs []dns.RR) {
	if len(rrs) == 0 {
		return
	}

	k := key(name, qtype)
	minTTL := rrs[0].Header().Ttl
	for _, rr := range rrs[1:] {
		if rr.Header().Ttl < minTTL {
			minTTL = rr.Header().Ttl
		}
	}
	c.cache.SetWithTTL(k, &cacheEntry{RRs: rrs}, 1, time.Duration(minTTL)*time.Second)
}

func (c *dnsCache) PutRecords(rrs []dns.RR) {
	if len(rrs) == 0 {
		return
	}

	// Group records by name, type and class
	grouped := make(map[string][]dns.RR)
	for _, rr := range rrs {
		k := key(rr.Header().Name, rr.Header().Rrtype)
		grouped[k] = append(grouped[k], rr)
	}

	// Store each group with shortest TTL
	for k, records := range grouped {
		minTTL := records[0].Header().Ttl
		for _, rr := range records[1:] {
			if rr.Header().Ttl < minTTL {
				minTTL = rr.Header().Ttl
			}
		}

		// Skip caching records with TTL 0
		if minTTL == 0 {
			continue
		}

		c.cache.SetWithTTL(k, &cacheEntry{RRs: records}, 1, time.Duration(minTTL)*time.Second)
	}
}

func (c *dnsCache) PutNegative(name string, qtype uint16, rcode int, soa *dns.SOA) {
	if soa == nil {
		return
	}

	k := negativeKey(name, qtype)
	ttl := soa.Header().Ttl

	c.cache.SetWithTTL(k, &negativeCacheEntry{Rcode: rcode, SOA: soa}, 1, time.Duration(ttl)*time.Second)
}

func (c *dnsCache) GetNegative(name string, qtype uint16) (int, *dns.SOA, bool) {
	k := negativeKey(name, qtype)
	val, ok := c.cache.Get(k)
	if !ok {
		return -1, nil, false
	}

	entry, ok := val.(*negativeCacheEntry)
	if !ok {
		return -1, nil, false
	}

	ttl, ok := c.cache.GetTTL(k)
	if !ok {
		return -1, nil, false
	}

	entry.SOA.Header().Ttl = uint32(ttl.Seconds())

	return entry.Rcode, entry.SOA, true
}
