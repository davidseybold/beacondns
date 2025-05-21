package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents a Beacon DNS API client
type Client struct {
	host       string
	httpClient *http.Client
}

// New creates a new Beacon DNS API client
func New(host string) *Client {
	return &Client{
		host: host,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Zone represents a DNS zone
type Zone struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	ResourceRecordSetCount int    `json:"resourceRecordSetCount"`
}

// ResourceRecordSet represents a DNS record set
type ResourceRecordSet struct {
	Name            string           `json:"name"`
	Type            string           `json:"type"`
	TTL             uint32           `json:"ttl"`
	ResourceRecords []ResourceRecord `json:"resourceRecords"`
}

// ResourceRecord represents a single DNS record
type ResourceRecord struct {
	Value string `json:"value"`
}

// Change represents a change to a resource record set
type Change struct {
	Action            string            `json:"action"`
	ResourceRecordSet ResourceRecordSet `json:"resourceRecordSet"`
}

// ChangeInfo represents information about a change
type ChangeInfo struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	SubmittedAt string `json:"submittedAt"`
}

// CreateZoneResponse represents the response from creating a zone
type CreateZoneResponse struct {
	ChangeInfo ChangeInfo `json:"changeInfo"`
	Zone       Zone       `json:"zone"`
}

// ListZonesResponse represents the response from listing zones
type ListZonesResponse struct {
	Zones []Zone `json:"zones"`
}

// ListResourceRecordSetsResponse represents the response from listing resource record sets
type ListResourceRecordSetsResponse struct {
	ResourceRecordSets []ResourceRecordSet `json:"resourceRecordSets"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CreateZone creates a new DNS zone
func (c *Client) CreateZone(ctx context.Context, name string) (*CreateZoneResponse, error) {
	reqBody := map[string]string{"name": name}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.host+"/v1/zones", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s - %s", errResp.Code, errResp.Message)
	}

	var response CreateZoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ListZones lists all DNS zones
func (c *Client) ListZones(ctx context.Context) (*ListZonesResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.host+"/v1/zones", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s - %s", errResp.Code, errResp.Message)
	}

	var response ListZonesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// GetZone gets information about a specific zone
func (c *Client) GetZone(ctx context.Context, zoneID string) (*Zone, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/zones/%s", c.host, zoneID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s - %s", errResp.Code, errResp.Message)
	}

	var zone Zone
	if err := json.NewDecoder(resp.Body).Decode(&zone); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &zone, nil
}

// ChangeResourceRecordSets changes resource record sets in a zone
func (c *Client) ChangeResourceRecordSets(ctx context.Context, zoneID string, changes []Change) (*ChangeInfo, error) {
	reqBody := map[string][]Change{"changes": changes}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/v1/zones/%s/rrsets", c.host, zoneID), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s - %s", errResp.Code, errResp.Message)
	}

	var changeInfo ChangeInfo
	if err := json.NewDecoder(resp.Body).Decode(&changeInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &changeInfo, nil
}

// ListResourceRecordSets lists all resource record sets in a zone
func (c *Client) ListResourceRecordSets(ctx context.Context, zoneID string) (*ListResourceRecordSetsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/v1/zones/%s/rrsets", c.host, zoneID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s - %s", errResp.Code, errResp.Message)
	}

	var response ListResourceRecordSetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// DeleteZone deletes a DNS zone
func (c *Client) DeleteZone(ctx context.Context, zoneID string) (*ChangeInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/v1/zones/%s", c.host, zoneID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s - %s", errResp.Code, errResp.Message)
	}

	var changeInfo ChangeInfo
	if err := json.NewDecoder(resp.Body).Decode(&changeInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &changeInfo, nil
}
