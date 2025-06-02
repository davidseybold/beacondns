package firewall

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetchDomainListFromSourceURL(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		serverDelay    time.Duration
		wantDomains    []string
		wantErr        bool
	}{
		{
			name: "successful fetch with valid domains",
			serverResponse: `example.com
test.com
# This is a comment
domain.com

another.com`,
			serverStatus: http.StatusOK,
			serverDelay:  0,
			wantDomains:  []string{"example.com", "test.com", "domain.com", "another.com"},
			wantErr:      false,
		},
		{
			name:           "empty response",
			serverResponse: "",
			serverStatus:   http.StatusOK,
			serverDelay:    0,
			wantDomains:    []string{},
			wantErr:        false,
		},
		{
			name:           "server error",
			serverResponse: "error",
			serverStatus:   http.StatusInternalServerError,
			serverDelay:    0,
			wantDomains:    nil,
			wantErr:        true,
		},
		{
			name:           "timeout",
			serverResponse: "timeout",
			serverStatus:   http.StatusOK,
			serverDelay:    11 * time.Second,
			wantDomains:    nil,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				time.Sleep(tt.serverDelay)
				w.WriteHeader(tt.serverStatus)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Call the function
			got, err := fetchDomainListFromSourceURL(t.Context(), server.URL)

			// Check error
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			// Check domains
			assert.NoError(t, err)
			assert.Equal(t, tt.wantDomains, got)
		})
	}
}
