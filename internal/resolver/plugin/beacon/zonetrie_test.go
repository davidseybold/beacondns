package beacon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewZoneTrie(t *testing.T) {
	trie := NewZoneTrie()
	assert.NotNil(t, trie)
	assert.NotNil(t, trie.root)
	assert.Empty(t, trie.root.ZoneName)
	assert.NotNil(t, trie.root.Children)
}

func TestZoneTrie_AddZone(t *testing.T) {
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
			trie := NewZoneTrie()
			for _, zone := range tt.zones {
				trie.AddZone(zone)
			}

			for zone, shouldExist := range tt.expected {
				exists := trie.Exists(zone)
				assert.Equal(t, shouldExist, exists, "zone %s should exist", zone)
			}
		})
	}
}

func TestZoneTrie_RemoveZone(t *testing.T) {
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
			trie := NewZoneTrie()
			for _, zone := range tt.add {
				trie.AddZone(zone)
			}

			trie.RemoveZone(tt.remove)

			for zone, shouldExist := range tt.expected {
				exists := trie.Exists(zone)
				assert.Equal(t, shouldExist, exists, "zone %s should exist", zone)
			}
		})
	}
}

func TestZoneTrie_FindLongestMatch(t *testing.T) {
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
			trie := NewZoneTrie()
			for _, zone := range tt.zones {
				trie.AddZone(zone)
			}

			match := trie.FindLongestMatch(tt.query)
			assert.Equal(t, tt.expected, match)
		})
	}
}

func TestZoneTrie_EdgeCases(t *testing.T) {
	trie := NewZoneTrie()

	// Test with non-FQDN input
	trie.AddZone("example.com")
	assert.True(t, trie.Exists("example.com"))

	// Test with empty string
	trie.AddZone("")
	assert.True(t, trie.Exists(""))

	// Test with root domain
	trie.AddZone(".")
	assert.True(t, trie.Exists("."))

	// Test with very long domain
	longDomain := "a." + string(make([]byte, 100)) + ".example.com."
	trie.AddZone(longDomain)
	assert.True(t, trie.Exists(longDomain))
}
