package beaconauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDNTrie(t *testing.T) {
	trie := NewDNTrie()
	assert.NotNil(t, trie)
	assert.NotNil(t, trie.root)
	assert.Empty(t, trie.root.DomainName)
	assert.NotNil(t, trie.root.Children)
}

func TestDNTrie_AddZone(t *testing.T) {
	tests := []struct {
		name     string
		zones    []string
		expected map[string]bool
	}{
		{
			name: "single zone",
			zones: []string{
				"example.com.",
			},
			expected: map[string]bool{
				"example.com.": true,
			},
		},
		{
			name: "multiple zones",
			zones: []string{
				"example.com.",
				"sub.example.com.",
			},
			expected: map[string]bool{
				"example.com.":     true,
				"sub.example.com.": true,
			},
		},
		{
			name: "duplicate zones",
			zones: []string{
				"example.com.",
				"example.com.",
			},
			expected: map[string]bool{
				"example.com.": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := NewDNTrie()
			for _, zone := range tt.zones {
				trie.Insert(zone)
			}

			for zone, shouldExist := range tt.expected {
				exists := trie.Contains(zone)
				assert.Equal(t, shouldExist, exists, "zone %s should exist", zone)
			}
		})
	}
}

func TestDNTrie_RemoveZone(t *testing.T) {
	tests := []struct {
		name     string
		add      []string
		remove   string
		expected map[string]bool
	}{
		{
			name: "remove existing zone",
			add: []string{
				"example.com.",
				"sub.example.com.",
			},
			remove: "sub.example.com.",
			expected: map[string]bool{
				"example.com.":     true,
				"sub.example.com.": false,
			},
		},
		{
			name: "remove non-existent zone",
			add: []string{
				"example.com.",
			},
			remove: "sub.example.com.",
			expected: map[string]bool{
				"example.com.":     true,
				"sub.example.com.": false,
			},
		},
		{
			name: "remove parent zone",
			add: []string{
				"example.com.",
				"sub.example.com.",
			},
			remove: "example.com.",
			expected: map[string]bool{
				"example.com.":     false,
				"sub.example.com.": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := NewDNTrie()
			for _, zone := range tt.add {
				trie.Insert(zone)
			}

			trie.Remove(tt.remove)

			for zone, shouldExist := range tt.expected {
				exists := trie.Contains(zone)
				assert.Equal(t, shouldExist, exists, "zone %s should exist", zone)
			}
		})
	}
}

func TestDNTrie_FindLongestMatch(t *testing.T) {
	tests := []struct {
		name     string
		zones    []string
		query    string
		expected string
	}{
		{
			name: "exact match",
			zones: []string{
				"example.com.",
			},
			query:    "example.com.",
			expected: "example.com.",
		},
		{
			name: "subdomain match",
			zones: []string{
				"example.com.",
				"sub.example.com.",
			},
			query:    "foo.sub.example.com.",
			expected: "sub.example.com.",
		},
		{
			name: "parent domain match",
			zones: []string{
				"example.com.",
				"sub.example.com.",
			},
			query:    "other.example.com.",
			expected: "example.com.",
		},
		{
			name: "no match",
			zones: []string{
				"example.com.",
			},
			query:    "other.com.",
			expected: "",
		},
		{
			name: "empty query",
			zones: []string{
				"example.com.",
			},
			query:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trie := NewDNTrie()
			for _, zone := range tt.zones {
				trie.Insert(zone)
			}

			match := trie.FindLongestMatch(tt.query)
			assert.Equal(t, tt.expected, match)
		})
	}
}

func TestDNTrie_EdgeCases(t *testing.T) {
	trie := NewDNTrie()

	// Test with non-FQDN input
	trie.Insert("example.com")
	assert.True(t, trie.Contains("example.com"))

	// Test with empty string
	trie.Insert("")
	assert.True(t, trie.Contains(""))

	// Test with root domain
	trie.Insert(".")
	assert.True(t, trie.Contains("."))

	// Test with very long domain
	longDomain := "a." + string(make([]byte, 100)) + ".example.com."
	trie.Insert(longDomain)
	assert.True(t, trie.Contains(longDomain))
}
