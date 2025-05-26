package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_CreateZone(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse *Zone
		serverStatus   int
		serverError    *errorResponse
		wantErr        bool
	}{
		{
			name: "success",
			serverResponse: &Zone{
				ID:   "test-zone",
				Name: "example.com",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "zone already exists",
			serverError: &errorResponse{
				Code:    "ZoneAlreadyExists",
				Message: "zone already exists",
			},
			serverStatus: http.StatusConflict,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/zones", r.URL.Path)

				if tt.serverError != nil {
					w.WriteHeader(tt.serverStatus)
					json.NewEncoder(w).Encode(tt.serverError)
					return
				}

				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := New(server.URL)
			zone, err := client.CreateZone(t.Context(), "example.com")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.serverError != nil {
					assert.IsType(t, &ZoneAlreadyExistsError{}, err)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse, zone)
		})
	}
}

func TestClient_GetZone(t *testing.T) {
	tests := []struct {
		name           string
		zoneID         string
		serverResponse *Zone
		serverStatus   int
		serverError    *errorResponse
		wantErr        bool
	}{
		{
			name:   "success",
			zoneID: "test-zone",
			serverResponse: &Zone{
				ID:   "test-zone",
				Name: "example.com",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:   "zone not found",
			zoneID: "non-existent",
			serverError: &errorResponse{
				Code:    "NoSuchZone",
				Message: "zone not found",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/zones/"+tt.zoneID, r.URL.Path)

				if tt.serverError != nil {
					w.WriteHeader(tt.serverStatus)
					json.NewEncoder(w).Encode(tt.serverError)
					return
				}

				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := New(server.URL)
			zone, err := client.GetZone(t.Context(), tt.zoneID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.serverError != nil {
					assert.IsType(t, &NoSuchZoneError{}, err)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse, zone)
		})
	}
}

func TestClient_ListZones(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse []Zone
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "success",
			serverResponse: []Zone{
				{ID: "zone1", Name: "example1.com"},
				{ID: "zone2", Name: "example2.com"},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:           "empty list",
			serverResponse: []Zone{},
			serverStatus:   http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/zones", r.URL.Path)

				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(listZonesResponse{Zones: tt.serverResponse})
			}))
			defer server.Close()

			client := New(server.URL)
			zones, err := client.ListZones(t.Context())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse, zones)
		})
	}
}

func TestClient_DeleteZone(t *testing.T) {
	tests := []struct {
		name         string
		zoneID       string
		serverStatus int
		serverError  *errorResponse
		wantErr      bool
	}{
		{
			name:         "success",
			zoneID:       "test-zone",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:   "zone not found",
			zoneID: "non-existent",
			serverError: &errorResponse{
				Code:    "NoSuchZone",
				Message: "zone not found",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Equal(t, "/v1/zones/"+tt.zoneID, r.URL.Path)

				if tt.serverError != nil {
					w.WriteHeader(tt.serverStatus)
					json.NewEncoder(w).Encode(tt.serverError)
					return
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			client := New(server.URL)
			err := client.DeleteZone(t.Context(), tt.zoneID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.serverError != nil {
					assert.IsType(t, &NoSuchZoneError{}, err)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestClient_ListResourceRecordSets(t *testing.T) {
	tests := []struct {
		name           string
		zoneID         string
		serverResponse []ResourceRecordSet
		serverStatus   int
		serverError    *errorResponse
		wantErr        bool
	}{
		{
			name:   "success",
			zoneID: "test-zone",
			serverResponse: []ResourceRecordSet{
				{
					Name:            "example.com",
					Type:            "A",
					TTL:             300,
					ResourceRecords: []ResourceRecord{{Value: "1.2.3.4"}},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:   "zone not found",
			zoneID: "non-existent",
			serverError: &errorResponse{
				Code:    "NoSuchZone",
				Message: "zone not found",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/zones/"+tt.zoneID+"/rrsets", r.URL.Path)

				if tt.serverError != nil {
					w.WriteHeader(tt.serverStatus)
					json.NewEncoder(w).Encode(tt.serverError)
					return
				}

				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(listResourceRecordSetsResponse{
					ResourceRecordSets: tt.serverResponse,
				})
			}))
			defer server.Close()

			client := New(server.URL)
			rrsets, err := client.ListResourceRecordSets(t.Context(), tt.zoneID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.serverError != nil {
					assert.IsType(t, &NoSuchZoneError{}, err)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse, rrsets)
		})
	}
}

func TestClient_UpsertResourceRecordSet(t *testing.T) {
	tests := []struct {
		name           string
		zoneID         string
		rrSet          ResourceRecordSet
		serverResponse *ResourceRecordSet
		serverStatus   int
		serverError    *errorResponse
		wantErr        bool
	}{
		{
			name:   "success",
			zoneID: "test-zone",
			rrSet: ResourceRecordSet{
				Name:            "example.com",
				Type:            "A",
				TTL:             300,
				ResourceRecords: []ResourceRecord{{Value: "1.2.3.4"}},
			},
			serverResponse: &ResourceRecordSet{
				Name:            "example.com",
				Type:            "A",
				TTL:             300,
				ResourceRecords: []ResourceRecord{{Value: "1.2.3.4"}},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:   "zone not found",
			zoneID: "non-existent",
			rrSet: ResourceRecordSet{
				Name:            "example.com",
				Type:            "A",
				TTL:             300,
				ResourceRecords: []ResourceRecord{{Value: "1.2.3.4"}},
			},
			serverError: &errorResponse{
				Code:    "NoSuchZone",
				Message: "zone not found",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v1/zones/"+tt.zoneID+"/rrsets", r.URL.Path)

				if tt.serverError != nil {
					w.WriteHeader(tt.serverStatus)
					json.NewEncoder(w).Encode(tt.serverError)
					return
				}

				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := New(server.URL)
			rrset, err := client.UpsertResourceRecordSet(t.Context(), tt.zoneID, tt.rrSet)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.serverError != nil {
					assert.IsType(t, &NoSuchZoneError{}, err)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse, rrset)
		})
	}
}

func TestClient_GetResourceRecordSet(t *testing.T) {
	tests := []struct {
		name           string
		zoneID         string
		recordName     string
		recordType     string
		serverResponse *ResourceRecordSet
		serverStatus   int
		serverError    *errorResponse
		wantErr        bool
	}{
		{
			name:       "success",
			zoneID:     "test-zone",
			recordName: "example.com",
			recordType: "A",
			serverResponse: &ResourceRecordSet{
				Name:            "example.com",
				Type:            "A",
				TTL:             300,
				ResourceRecords: []ResourceRecord{{Value: "1.2.3.4"}},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "record not found",
			zoneID:     "test-zone",
			recordName: "example.com",
			recordType: "A",
			serverError: &errorResponse{
				Code:    "NoSuchResourceRecordSet",
				Message: "record not found",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v1/zones/"+tt.zoneID+"/rrsets/"+tt.recordName+"/"+tt.recordType, r.URL.Path)

				if tt.serverError != nil {
					w.WriteHeader(tt.serverStatus)
					json.NewEncoder(w).Encode(tt.serverError)
					return
				}

				w.WriteHeader(tt.serverStatus)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := New(server.URL)
			rrset, err := client.GetResourceRecordSet(t.Context(), tt.zoneID, tt.recordName, tt.recordType)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.serverError != nil {
					assert.IsType(t, &NoSuchResourceRecordSetError{}, err)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.serverResponse, rrset)
		})
	}
}

func TestClient_DeleteResourceRecordSet(t *testing.T) {
	tests := []struct {
		name         string
		zoneID       string
		recordName   string
		recordType   string
		serverStatus int
		serverError  *errorResponse
		wantErr      bool
	}{
		{
			name:         "success",
			zoneID:       "test-zone",
			recordName:   "example.com",
			recordType:   "A",
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:       "record not found",
			zoneID:     "test-zone",
			recordName: "example.com",
			recordType: "A",
			serverError: &errorResponse{
				Code:    "NoSuchResourceRecordSet",
				Message: "record not found",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Equal(t, "/v1/zones/"+tt.zoneID+"/rrsets/"+tt.recordName+"/"+tt.recordType, r.URL.Path)

				if tt.serverError != nil {
					w.WriteHeader(tt.serverStatus)
					json.NewEncoder(w).Encode(tt.serverError)
					return
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			client := New(server.URL)
			err := client.DeleteResourceRecordSet(t.Context(), tt.zoneID, tt.recordName, tt.recordType)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.serverError != nil {
					assert.IsType(t, &NoSuchResourceRecordSetError{}, err)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}
