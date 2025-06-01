package beaconfirewall

import (
	"github.com/miekg/dns"
)

const (
	wildcardLabel = "*"
)

type valueList[T comparable] []T

func (v *valueList[T]) Vals() []T {
	return *v
}

func (v *valueList[T]) Push(val T) {
	*v = append(*v, val)
}

func (v *valueList[T]) Len() int {
	return len(*v)
}

func (v *valueList[T]) RemoveByValue(val T) valueList[T] {
	newVals := make([]T, 0, len(*v)-1)
	for _, v := range *v {
		if v != val {
			newVals = append(newVals, v)
		}
	}
	return newVals
}

type DNTrie[T comparable] struct {
	root        *DNTrieNode[T]
	valueToNode map[T]*DNTrieNode[T]
}

type DNTrieNode[T comparable] struct {
	DomainName *string
	Children   map[string]*DNTrieNode[T]
	values     valueList[T]
	isWildcard bool
	isTerminal bool
}

func NewDNTrie[T comparable]() *DNTrie[T] {
	return &DNTrie[T]{
		root: &DNTrieNode[T]{
			DomainName: nil,
			Children:   make(map[string]*DNTrieNode[T]),
			values:     make(valueList[T], 0),
			isWildcard: false,
			isTerminal: false,
		},
		valueToNode: make(map[T]*DNTrieNode[T]),
	}
}

func (t *DNTrie[T]) Insert(domainName string, value T) {
	domainName = dns.Fqdn(domainName)
	node := t.insert(domainName, value)
	t.valueToNode[value] = node
}

func (t *DNTrie[T]) RemoveByValue(val T) bool {
	node := t.valueToNode[val]
	if node == nil {
		return false
	}

	node.values = node.values.RemoveByValue(val)

	if node.values.Len() == 0 {
		node.isTerminal = false
	}

	if node.values.Len() == 0 && len(node.Children) == 0 {
		t.removeNode(node)
	}

	delete(t.valueToNode, val)

	return true
}

func (t *DNTrie[T]) removeNode(n *DNTrieNode[T]) {
	if n.DomainName == nil {
		return
	}

	domainName := *n.DomainName
	parts := dns.SplitDomainName(domainName)

	path := make([]*DNTrieNode[T], 0, len(parts)+1)
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

	current.DomainName = nil
	current.isWildcard = false

	for i := len(path) - 1; i > 0; i-- {
		node := path[i]
		parent := path[i-1]
		if len(node.Children) == 0 && node.DomainName == nil && node.values.Len() == 0 {
			delete(parent.Children, parts[len(parts)-i])
		} else {
			break
		}
	}
}

func (t *DNTrie[T]) Contains(domainName string) bool {
	domainName = dns.Fqdn(domainName)
	return t.contains(domainName)
}

func (t *DNTrie[T]) FindMatchingRules(domainName string) ([]T, bool) {
	domainName = dns.Fqdn(domainName)
	return t.findMatchingRules(domainName)
}

func (t *DNTrie[T]) insert(domainName string, value T) *DNTrieNode[T] {
	parts := dns.SplitDomainName(domainName)
	current := t.root

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			current.Children[part] = &DNTrieNode[T]{
				DomainName: nil,
				Children:   make(map[string]*DNTrieNode[T]),
				values:     make(valueList[T], 0),
				isWildcard: part == wildcardLabel,
				isTerminal: false,
			}
		}
		current = current.Children[part]
	}

	current.DomainName = &domainName
	current.isTerminal = true
	current.values.Push(value)

	return current
}

func (t *DNTrie[T]) contains(domainName string) bool {
	parts := dns.SplitDomainName(domainName)
	current := t.root

	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if current.Children[part] == nil {
			return false
		}
		current = current.Children[part]
	}

	return current.isTerminal
}

func (t *DNTrie[T]) findMatchingRules(domainName string) ([]T, bool) {
	parts := dns.SplitDomainName(domainName)

	var bestMatch *DNTrieNode[T]

	current := t.root
	i := len(parts)
	for i > 0 {
		part := parts[i-1]
		if current.Children[wildcardLabel] != nil {
			bestMatch = current.Children[wildcardLabel]
		}
		if current.Children[part] == nil {
			break
		}
		current = current.Children[part]
		i--
	}

	if current.isTerminal && i == 0 {
		return current.values.Vals(), true
	}

	if bestMatch != nil {
		return bestMatch.values.Vals(), true
	}

	return nil, false
}
