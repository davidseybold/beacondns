package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
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

func (c *Client) CreateZone(ctx context.Context, name string) (*Zone, error) {
	req := createZoneRequest{Name: name}
	var resp Zone
	if err := c.postRequest(ctx, "/v1/zones", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	var resp listZonesResponse
	if err := c.getRequest(ctx, "/v1/zones", &resp); err != nil {
		return nil, err
	}
	return resp.Zones, nil
}

func (c *Client) GetZone(ctx context.Context, name string) (*Zone, error) {
	var resp Zone
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/zones/%s", name), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteZone(ctx context.Context, name string) error {
	return c.deleteRequest(ctx, fmt.Sprintf("/v1/zones/%s", name))
}

func (c *Client) ListResourceRecordSets(ctx context.Context, zoneName string) ([]ResourceRecordSet, error) {
	var resp listResourceRecordSetsResponse
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/zones/%s/rrsets", zoneName), &resp); err != nil {
		return nil, err
	}
	return resp.ResourceRecordSets, nil
}

func (c *Client) UpsertResourceRecordSet(
	ctx context.Context,
	zoneName string,
	rrSet ResourceRecordSet,
) (*ResourceRecordSet, error) {
	req := upsertResourceRecordSetRequest{ResourceRecordSet: rrSet}
	var resp ResourceRecordSet
	if err := c.postRequest(ctx, fmt.Sprintf("/v1/zones/%s/rrsets", zoneName), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetResourceRecordSet(
	ctx context.Context,
	zoneName string,
	name string,
	rrType string,
) (*ResourceRecordSet, error) {
	var resp ResourceRecordSet
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/zones/%s/rrsets/%s/%s", zoneName, name, rrType), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) DeleteResourceRecordSet(ctx context.Context, zoneID string, name string, rrType string) error {
	return c.deleteRequest(ctx, fmt.Sprintf("/v1/zones/%s/rrsets/%s/%s", zoneID, name, rrType))
}

func (c *Client) CreateFirewallRule(ctx context.Context, req CreateFirewallRuleRequest) (*FirewallRule, error) {
	var resp FirewallRule
	if err := c.postRequest(ctx, "/v1/firewall/rules", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetFirewallRules(ctx context.Context) ([]FirewallRule, error) {
	var resp getFirewallRulesResponse
	if err := c.getRequest(ctx, "/v1/firewall/rules", &resp); err != nil {
		return nil, err
	}
	return resp.Rules, nil
}

func (c *Client) DeleteFirewallRule(ctx context.Context, id uuid.UUID) error {
	return c.deleteRequest(ctx, fmt.Sprintf("/v1/firewall/rules/%s", id))
}

func (c *Client) GetFirewallRule(ctx context.Context, id uuid.UUID) (*FirewallRule, error) {
	var resp FirewallRule
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/firewall/rules/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateFirewallRule(
	ctx context.Context,
	id uuid.UUID,
	req UpdateFirewallRuleRequest,
) (*FirewallRule, error) {
	var resp FirewallRule
	if err := c.postRequest(ctx, fmt.Sprintf("/v1/firewall/rules/%s", id), req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CreateDomainList(ctx context.Context, req CreateDomainListRequest) (*DomainList, error) {
	var resp DomainList
	if err := c.postRequest(ctx, "/v1/firewall/domain-lists", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetDomainLists(ctx context.Context) ([]DomainList, error) {
	var resp listDomainListsResponse
	if err := c.getRequest(ctx, "/v1/firewall/domain-lists", &resp); err != nil {
		return nil, err
	}
	return resp.DomainLists, nil
}

func (c *Client) GetDomainList(ctx context.Context, id uuid.UUID) (*DomainList, error) {
	var resp DomainList
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/firewall/domain-lists/%s", id), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetDomainListDomains(ctx context.Context, id uuid.UUID) ([]string, error) {
	var resp getDomainListDomainsResponse
	if err := c.getRequest(ctx, fmt.Sprintf("/v1/firewall/domain-lists/%s/domains", id), &resp); err != nil {
		return nil, err
	}
	return resp.Domains, nil
}

func (c *Client) AddDomainsToDomainList(ctx context.Context, id uuid.UUID, domains []string) error {
	return c.postRequest(
		ctx,
		fmt.Sprintf("/v1/firewall/domain-lists/%s/domains", id),
		addDomainsToDomainListRequest{Domains: domains},
		nil,
	)
}

func (c *Client) RemoveDomainsFromDomainList(ctx context.Context, id uuid.UUID, domains []string) error {
	return c.postRequest(
		ctx,
		fmt.Sprintf("/v1/firewall/domain-lists/%s/domains", id),
		removeDomainsFromDomainListRequest{Domains: domains},
		nil,
	)
}

func (c *Client) DeleteDomainList(ctx context.Context, id uuid.UUID) error {
	return c.deleteRequest(ctx, fmt.Sprintf("/v1/firewall/domain-lists/%s", id))
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

func (c *Client) deleteRequest(ctx context.Context, path string) error {
	return c.doRequest(ctx, "DELETE", path, nil, nil)
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
		return c.handleError(resp)
	}

	if result != nil {
		if decodeErr := json.NewDecoder(resp.Body).Decode(result); decodeErr != nil {
			return fmt.Errorf("failed to decode response: %w", decodeErr)
		}
	}

	return nil
}

func (c *Client) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResponse errorResponse
	if err := json.Unmarshal(body, &errResponse); err != nil {
		return fmt.Errorf("failed to unmarshal error response: %w", err)
	}
	return parseError(errResponse)
}
