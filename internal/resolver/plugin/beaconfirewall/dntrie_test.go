package beaconfirewall

// import (
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/stretchr/testify/assert"
// )

// func TestDNTrie_InsertAndFindMatchingRule(t *testing.T) {
// 	id1 := uuid.New()
// 	id2 := uuid.New()
// 	id3 := uuid.New()

// 	tests := []struct {
// 		name      string
// 		domain    string
// 		wantID    uuid.UUID
// 		wantFound bool
// 		populate  func(t *DNTrie[uuid.UUID])
// 	}{
// 		{
// 			name:      "empty trie lookup",
// 			domain:    "example.com",
// 			wantID:    uuid.UUID{},
// 			wantFound: false,
// 			populate:  func(trie *DNTrie[uuid.UUID]) {},
// 		},
// 		{
// 			name:      "exact match root domain",
// 			domain:    "example.com",
// 			wantID:    id1,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 			},
// 		},
// 		{
// 			name:      "exact match subdomain",
// 			domain:    "sub.example.com",
// 			wantID:    id2,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("sub.example.com", 2, id2)
// 			},
// 		},
// 		{
// 			name:      "exact match nested subdomain",
// 			domain:    "sub.sub.example.com",
// 			wantID:    id3,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("sub.sub.example.com", 3, id3)
// 			},
// 		},
// 		{
// 			name:      "non-matching domain",
// 			domain:    "other.com",
// 			wantID:    uuid.UUID{},
// 			wantFound: false,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("sub.example.com", 2, id2)
// 			},
// 		},
// 		{
// 			name:      "wildcard match",
// 			domain:    "www.example.com",
// 			wantID:    id1,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("*.example.com", 1, id1)
// 			},
// 		},
// 		{
// 			name:      "wildcard match with subdomain",
// 			domain:    "www.sub.example.com",
// 			wantID:    id1,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("*.example.com", 1, id1)
// 				trie.Insert("sub.example.com", 2, id2)
// 			},
// 		},
// 		{
// 			name:      "wildcard with specific subdomain",
// 			domain:    "www.sub.example.com",
// 			wantID:    id2,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("*.example.com", 1, id1)
// 				trie.Insert("www.sub.example.com", 2, id2)
// 			},
// 		},
// 		{
// 			name:      "more specific wildcard",
// 			domain:    "www.sub.example.com",
// 			wantID:    id3,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("*.example.com", 1, id1)
// 				trie.Insert("*.sub.example.com", 3, id3)
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			trie := NewDNTrie[uuid.UUID]()
// 			tt.populate(trie)
// 			gotID, gotFound := trie.FindMatchingRule(tt.domain)
// 			assert.Equal(t, tt.wantFound, gotFound)
// 			if tt.wantFound {
// 				assert.Equal(t, tt.wantID, gotID)
// 			}
// 		})
// 	}
// }

// func TestDNTrie_Priority(t *testing.T) {
// 	id1 := uuid.New()
// 	id2 := uuid.New()
// 	id3 := uuid.New()

// 	tests := []struct {
// 		name      string
// 		domain    string
// 		wantID    uuid.UUID
// 		wantFound bool
// 		populate  func(t *DNTrie[uuid.UUID])
// 	}{
// 		{
// 			name:      "highest priority rule",
// 			domain:    "example.com",
// 			wantID:    id2, // priority 3
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("example.com", 3, id2)
// 				trie.Insert("example.com", 2, id3)
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			trie := NewDNTrie[uuid.UUID]()
// 			tt.populate(trie)
// 			gotID, gotFound := trie.FindMatchingRule(tt.domain)
// 			assert.Equal(t, tt.wantFound, gotFound)
// 			if tt.wantFound {
// 				assert.Equal(t, tt.wantID, gotID)
// 			}
// 		})
// 	}
// }

// func TestDNTrie_Contains(t *testing.T) {
// 	id1 := uuid.New()

// 	tests := []struct {
// 		name     string
// 		domain   string
// 		want     bool
// 		populate func(t *DNTrie[uuid.UUID])
// 	}{
// 		{
// 			name:   "empty trie check",
// 			domain: "example.com",
// 			want:   false,
// 			populate: func(trie *DNTrie[uuid.UUID]) {

// 			},
// 		},
// 		{
// 			name:   "exact match root domain",
// 			domain: "example.com",
// 			want:   true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 			},
// 		},
// 		{
// 			name:   "exact match subdomain",
// 			domain: "sub.example.com",
// 			want:   true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("sub.example.com", 1, id1)
// 			},
// 		},
// 		{
// 			name:   "non-matching domain",
// 			domain: "other.com",
// 			want:   false,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 			},
// 		},
// 		{
// 			name:   "non-matching subdomain",
// 			domain: "sub.other.com",
// 			want:   false,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			trie := NewDNTrie[uuid.UUID]()
// 			tt.populate(trie)
// 			got := trie.Contains(tt.domain)
// 			assert.Equal(t, tt.want, got)
// 		})
// 	}
// }

// func TestDNTrie_WildcardMatching(t *testing.T) {
// 	id1 := uuid.New()
// 	id2 := uuid.New()
// 	id3 := uuid.New()
// 	id4 := uuid.New()

// 	tests := []struct {
// 		name      string
// 		domain    string
// 		wantID    uuid.UUID
// 		wantFound bool
// 		populate  func(t *DNTrie[uuid.UUID])
// 	}{
// 		{
// 			name:      "exact match root domain",
// 			domain:    "example.com",
// 			wantID:    id1,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("sub.example.com", 3, id3)
// 				trie.Insert("*.sub.example.com", 4, id4)
// 			},
// 		},
// 		{
// 			name:      "wildcard match for subdomain",
// 			domain:    "www.example.com",
// 			wantID:    id2,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("sub.example.com", 3, id3)
// 				trie.Insert("*.sub.example.com", 4, id4)
// 			},
// 		},
// 		{
// 			name:      "exact match subdomain",
// 			domain:    "sub.example.com",
// 			wantID:    id3,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("sub.example.com", 3, id3)
// 				trie.Insert("*.sub.example.com", 4, id4)
// 			},
// 		},
// 		{
// 			name:      "wildcard match for nested subdomain",
// 			domain:    "www.sub.example.com",
// 			wantID:    id4,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("sub.example.com", 3, id3)
// 				trie.Insert("*.sub.example.com", 4, id4)
// 			},
// 		},
// 		{
// 			name:      "no match for non-matching domain",
// 			domain:    "other.com",
// 			wantID:    uuid.UUID{},
// 			wantFound: false,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("sub.example.com", 3, id3)
// 				trie.Insert("*.sub.example.com", 4, id4)
// 			},
// 		},
// 		{
// 			name:      "no match for non-matching subdomain",
// 			domain:    "sub.other.com",
// 			wantID:    uuid.UUID{},
// 			wantFound: false,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("sub.example.com", 3, id3)
// 				trie.Insert("*.sub.example.com", 4, id4)
// 			},
// 		},
// 		{
// 			name:      "no match for partial wildcard",
// 			domain:    "test.example.com",
// 			wantID:    id2,
// 			wantFound: true,
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("sub.example.com", 3, id3)
// 				trie.Insert("*.sub.example.com", 4, id4)
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			trie := NewDNTrie[uuid.UUID]()
// 			tt.populate(trie)
// 			gotID, gotFound := trie.FindMatchingRule(tt.domain)
// 			assert.Equal(t, tt.wantFound, gotFound)
// 			if tt.wantFound {
// 				assert.Equal(t, tt.wantID, gotID)
// 			}
// 		})
// 	}
// }

// func TestDNTrie_RemoveByID(t *testing.T) {
// 	id1 := uuid.New()
// 	id2 := uuid.New()
// 	id3 := uuid.New()
// 	id4 := uuid.New()
// 	id5 := uuid.New()
// 	id6 := uuid.New()

// 	tests := []struct {
// 		name        string
// 		populate    func(t *DNTrie[uuid.UUID])
// 		removeID    uuid.UUID
// 		checkDomain string
// 		wantID      uuid.UUID
// 		wantFound   bool
// 		wantRemoved bool
// 	}{
// 		{
// 			name: "remove non-existent ID",
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 			},
// 			removeID:    id2,
// 			checkDomain: "example.com",
// 			wantID:      id1,
// 			wantFound:   true,
// 			wantRemoved: false,
// 		},
// 		{
// 			name: "remove single value from node",
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 			},
// 			removeID:    id1,
// 			checkDomain: "example.com",
// 			wantID:      uuid.UUID{},
// 			wantFound:   false,
// 			wantRemoved: true,
// 		},
// 		{
// 			name: "remove one of multiple values from node",
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("example.com", 2, id2)
// 				trie.Insert("example.com", 3, id3)
// 			},
// 			removeID:    id2,
// 			checkDomain: "example.com",
// 			wantID:      id3, // highest priority remaining
// 			wantFound:   true,
// 			wantRemoved: true,
// 		},
// 		{
// 			name: "remove middle node with children",
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("sub.example.com", 2, id2)
// 				trie.Insert("sub.sub.example.com", 3, id3)
// 			},
// 			removeID:    id2,
// 			checkDomain: "sub.example.com",
// 			wantID:      uuid.UUID{},
// 			wantFound:   false,
// 			wantRemoved: true,
// 		},
// 		{
// 			name: "remove leaf node",
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("sub.example.com", 2, id2)
// 				trie.Insert("sub.sub.example.com", 3, id3)
// 			},
// 			removeID:    id3,
// 			checkDomain: "sub.sub.example.com",
// 			wantID:      uuid.UUID{},
// 			wantFound:   false,
// 			wantRemoved: true,
// 		},
// 		{
// 			name: "remove wildcard node",
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("sub.example.com", 3, id3)
// 			},
// 			removeID:    id2,
// 			checkDomain: "www.example.com",
// 			wantID:      id1,
// 			wantFound:   false,
// 			wantRemoved: true,
// 		},
// 		{
// 			name: "remove node with multiple wildcards",
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("*.sub.example.com", 3, id3)
// 				trie.Insert("sub.example.com", 4, id4)
// 			},
// 			removeID:    id3,
// 			checkDomain: "www.sub.example.com",
// 			wantID:      id2,
// 			wantFound:   true,
// 			wantRemoved: true,
// 		},
// 		{
// 			name: "remove node with multiple values and wildcards",
// 			populate: func(trie *DNTrie[uuid.UUID]) {
// 				trie.Insert("example.com", 1, id1)
// 				trie.Insert("*.example.com", 2, id2)
// 				trie.Insert("*.example.com", 3, id3)
// 				trie.Insert("sub.example.com", 4, id4)
// 				trie.Insert("sub.example.com", 5, id5)
// 				trie.Insert("sub.example.com", 6, id6)
// 			},
// 			removeID:    id5,
// 			checkDomain: "sub.example.com",
// 			wantID:      id6, // highest priority remaining
// 			wantFound:   true,
// 			wantRemoved: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			trie := NewDNTrie[uuid.UUID]()
// 			tt.populate(trie)

// 			// Verify initial state
// 			_, initialFound := trie.FindMatchingRule(tt.checkDomain)
// 			assert.True(t, initialFound, "Initial domain should exist")

// 			// Remove the ID
// 			removed := trie.RemoveByValue(tt.removeID)
// 			assert.Equal(t, tt.wantRemoved, removed, "RemoveByID return value mismatch")

// 			// Verify final state
// 			gotID, gotFound := trie.FindMatchingRule(tt.checkDomain)
// 			assert.Equal(t, tt.wantFound, gotFound, "FindMatchingRule found mismatch")
// 			if tt.wantFound {
// 				assert.Equal(t, tt.wantID, gotID, "FindMatchingRule ID mismatch")
// 			}
// 		})
// 	}
// }
