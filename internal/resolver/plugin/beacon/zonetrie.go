package beacon

import (
	"github.com/miekg/dns"
)

type ZoneTrie struct {
	root *ZoneNode
}

type ZoneNode struct {
	ZoneName string
	Children map[string]*ZoneNode
}

func NewZoneTrie() *ZoneTrie {
	return &ZoneTrie{
		root: &ZoneNode{
			ZoneName: "",
			Children: make(map[string]*ZoneNode),
		},
	}
}

func (t *ZoneTrie) AddZone(zoneName string) {
	zoneName = dns.Fqdn(zoneName)
	parts := dns.SplitDomainName(zoneName)
	current := t.root

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			current.Children[part] = &ZoneNode{
				ZoneName: "",
				Children: make(map[string]*ZoneNode),
			}
		}
		current = current.Children[part]
	}
	current.ZoneName = zoneName
}

func (t *ZoneTrie) RemoveZone(zoneName string) {
	zoneName = dns.Fqdn(zoneName)
	parts := dns.SplitDomainName(zoneName)

	path := make([]*ZoneNode, 0, len(parts)+1)
	current := t.root
	path = append(path, current)

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			return
		}
		current = current.Children[part]
		path = append(path, current)
	}

	current.ZoneName = ""

	for i := len(path) - 1; i > 0; i-- {
		node := path[i]
		parent := path[i-1]
		if len(node.Children) == 0 && node.ZoneName == "" {
			delete(parent.Children, parts[len(parts)-i])
		} else {
			break
		}
	}
}

func (t *ZoneTrie) Exists(domainName string) bool {
	domainName = dns.Fqdn(domainName)
	parts := dns.SplitDomainName(domainName)
	current := t.root

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			return false
		}
		current = current.Children[part]
	}

	return current.ZoneName != ""
}

func (t *ZoneTrie) FindLongestMatch(domainName string) string {
	domainName = dns.Fqdn(domainName)
	parts := dns.SplitDomainName(domainName)
	current := t.root
	var match string

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			break
		}
		current = current.Children[part]
		if current.ZoneName != "" {
			match = current.ZoneName
		}
	}

	return match
}
