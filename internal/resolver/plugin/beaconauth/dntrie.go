package beaconauth

import (
	"github.com/miekg/dns"
)

type DNTrie struct {
	root      *DNTrieNode
	zoneCount int
}

type DNTrieNode struct {
	DomainName *string
	Children   map[string]*DNTrieNode
}

func NewDNTrie() *DNTrie {
	return &DNTrie{
		root: &DNTrieNode{
			DomainName: nil,
			Children:   make(map[string]*DNTrieNode),
		},
		zoneCount: 0,
	}
}

func (t *DNTrie) Insert(domainName string) {
	domainName = dns.Fqdn(domainName)
	t.insert(domainName)
}

func (t *DNTrie) Remove(domainName string) {
	domainName = dns.Fqdn(domainName)
	t.remove(domainName)
}

func (t *DNTrie) Contains(domainName string) bool {
	domainName = dns.Fqdn(domainName)
	return t.contains(domainName)
}

func (t *DNTrie) FindLongestMatch(domainName string) string {
	domainName = dns.Fqdn(domainName)
	return t.findLongestMatch(domainName)
}

func (t *DNTrie) Count() int {
	return t.zoneCount
}

func (t *DNTrie) insert(domainName string) {
	parts := dns.SplitDomainName(domainName)
	current := t.root

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			current.Children[part] = &DNTrieNode{
				DomainName: nil,
				Children:   make(map[string]*DNTrieNode),
			}
		}
		current = current.Children[part]
	}

	if current.DomainName == nil {
		t.zoneCount++
	}
	current.DomainName = &domainName
}

func (t *DNTrie) remove(domainName string) {
	parts := dns.SplitDomainName(domainName)

	path := make([]*DNTrieNode, 0, len(parts)+1)
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

	if current.DomainName != nil {
		t.zoneCount--
	}

	current.DomainName = nil

	for i := len(path) - 1; i > 0; i-- {
		node := path[i]
		parent := path[i-1]
		if len(node.Children) == 0 && node.DomainName == nil {
			delete(parent.Children, parts[len(parts)-i])
		} else {
			break
		}
	}
}

func (t *DNTrie) contains(domainName string) bool {
	parts := dns.SplitDomainName(domainName)
	current := t.root

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			return false
		}
		current = current.Children[part]
	}

	return current.DomainName != nil
}

func (t *DNTrie) findLongestMatch(domainName string) string {
	parts := dns.SplitDomainName(domainName)
	current := t.root
	var match string

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			break
		}
		current = current.Children[part]
		if current.DomainName != nil {
			match = *current.DomainName
		}
	}

	return match
}
