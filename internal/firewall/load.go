package firewall

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func fetchDomainListFromSourceURL(ctx context.Context, sourceURL string) ([]string, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch domain list from source URL: %s: status %s", sourceURL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	domains := []string{}
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		} else if strings.HasPrefix(line, "#") {
			continue
		}

		domains = append(domains, line)
	}

	return domains, nil
}
