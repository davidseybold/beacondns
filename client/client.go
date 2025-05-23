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

var (
	httpClientTimeout = 30 * time.Second
)

type Client struct {
	host       string
	httpClient *http.Client
}

func New(host string) *Client {
	return &Client{
		host: host,
		httpClient: &http.Client{
			Timeout: httpClientTimeout,
		},
	}
}

func (c *Client) CreateZone(ctx context.Context, name string) (*CreateZoneResponse, error) {
	req := CreateZoneRequest{Name: name}
	var resp CreateZoneResponse
	if err := c.postRequest(ctx, "/v1/zones", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListZones(ctx context.Context) (*ListZonesResponse, error) {
	var resp ListZonesResponse
	if err := c.getRequest(ctx, "/v1/zones", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetZone(ctx context.Context, id string) (*Zone, error) {
	var resp Zone
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/zones/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteZone(ctx context.Context, id string) (*ChangeInfo, error) {
	var resp ChangeInfo
	if err := c.deleteRequest(ctx, fmt.Sprintf("/v1/zones/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListResourceRecordSets(ctx context.Context, zoneID string) (*ListResourceRecordSetsResponse, error) {
	var resp ListResourceRecordSetsResponse
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/zones/%s/rrsets", zoneID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ChangeResourceRecordSets(ctx context.Context, zoneID string, changes []Change) (*ChangeInfo, error) {
	req := ChangeResourceRecordSetsRequest{Changes: changes}
	var resp ChangeInfo
	if err := c.postRequest(ctx, fmt.Sprintf("/v1/zones/%s/rrsets", zoneID), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetChange(ctx context.Context, changeID string) (*ChangeInfo, error) {
	var resp ChangeInfo
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/changes/%s", changeID), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) getRequest(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, "GET", path, nil, result)
}

func (c *Client) postRequest(ctx context.Context, path string, body any, result any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.doRequest(ctx, "POST", path, bytes.NewReader(jsonBody), result)
}

func (c *Client) deleteRequest(ctx context.Context, path string, result any) error {
	return c.doRequest(ctx, "DELETE", path, nil, result)
}

func (c *Client) doRequest(ctx context.Context, method, path string, bodyReader io.Reader, result any) error {
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

	if resp.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if decodeErr := json.NewDecoder(resp.Body).Decode(result); decodeErr != nil {
			return fmt.Errorf("failed to decode response: %w", decodeErr)
		}
	}

	return nil
}
