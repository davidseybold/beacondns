package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Beacon DNS API client
type Client struct {
	host       string
	httpClient *http.Client
}

// Request types
type CreateZoneRequest struct {
	Name string `json:"name"`
}

type ChangeResourceRecordSetsRequest struct {
	Changes []Change `json:"changes"`
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

// doRequest performs an HTTP request and handles common response processing
func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.host+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// CreateZone creates a new DNS zone
func (c *Client) CreateZone(ctx context.Context, name string) (*CreateZoneResponse, error) {
	req := CreateZoneRequest{Name: name}
	var resp CreateZoneResponse
	if err := c.doRequest(ctx, "POST", "/v1/zones", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListZones lists all DNS zones
func (c *Client) ListZones(ctx context.Context) (*ListZonesResponse, error) {
	var resp ListZonesResponse
	if err := c.doRequest(ctx, "GET", "/v1/zones", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetZone gets information about a specific zone
func (c *Client) GetZone(ctx context.Context, id string) (*Zone, error) {
	var resp Zone
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/zones/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteZone deletes a DNS zone
func (c *Client) DeleteZone(ctx context.Context, id string) (*ChangeInfo, error) {
	var resp ChangeInfo
	if err := c.doRequest(ctx, "DELETE", fmt.Sprintf("/v1/zones/%s", id), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListResourceRecordSets lists all resource record sets in a zone
func (c *Client) ListResourceRecordSets(ctx context.Context, zoneID string) (*ListResourceRecordSetsResponse, error) {
	var resp ListResourceRecordSetsResponse
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/zones/%s/rrsets", zoneID), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ChangeResourceRecordSets creates, updates, or deletes resource record sets
func (c *Client) ChangeResourceRecordSets(ctx context.Context, zoneID string, changes []Change) (*ChangeInfo, error) {
	req := ChangeResourceRecordSetsRequest{Changes: changes}
	var resp ChangeInfo
	if err := c.doRequest(ctx, "POST", fmt.Sprintf("/v1/zones/%s/rrsets", zoneID), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetChange gets information about a change
func (c *Client) GetChange(ctx context.Context, changeID string) (*ChangeInfo, error) {
	var resp ChangeInfo
	if err := c.doRequest(ctx, "GET", fmt.Sprintf("/v1/changes/%s", changeID), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
