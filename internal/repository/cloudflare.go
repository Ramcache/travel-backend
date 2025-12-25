package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CloudflareRepository struct {
	http     *http.Client
	baseURL  string
	apiToken string
}

func NewCloudflareRepository(apiToken string) *CloudflareRepository {
	return &CloudflareRepository{
		http: &http.Client{
			Timeout: 20 * time.Second,
		},
		baseURL:  "https://api.cloudflare.com/client/v4",
		apiToken: apiToken,
	}
}

// Cloudflare purge request поддерживает разные варианты.
// В реальности вам чаще всего нужно purge_everything или files.
// Остальные поля (tags/hosts/prefixes) оставлены “на вырост”.
type CloudflarePurgeCacheRequest struct {
	PurgeEverything bool     `json:"purge_everything,omitempty"`
	Files           []string `json:"files,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Hosts           []string `json:"hosts,omitempty"`
	Prefixes        []string `json:"prefixes,omitempty"`
}

type cloudflareMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cloudflareResponse[T any] struct {
	Success  bool                `json:"success"`
	Errors   []cloudflareMessage `json:"errors"`
	Messages []cloudflareMessage `json:"messages"`
	Result   T                   `json:"result"`
}

type CloudflarePurgeCacheResult struct {
	ID string `json:"id"`
}

func (r *CloudflareRepository) PurgeCache(ctx context.Context, zoneID string, req CloudflarePurgeCacheRequest) (CloudflarePurgeCacheResult, error) {
	var zero CloudflarePurgeCacheResult

	if r.apiToken == "" {
		return zero, fmt.Errorf("cloudflare api token is empty")
	}
	if zoneID == "" {
		return zero, fmt.Errorf("cloudflare zoneID is empty")
	}
	if !req.PurgeEverything &&
		len(req.Files) == 0 &&
		len(req.Tags) == 0 &&
		len(req.Hosts) == 0 &&
		len(req.Prefixes) == 0 {
		return zero, fmt.Errorf("purge request is empty")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return zero, fmt.Errorf("marshal purge request: %w", err)
	}

	url := fmt.Sprintf("%s/zones/%s/purge_cache", r.baseURL, zoneID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return zero, fmt.Errorf("new request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+r.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := r.http.Do(httpReq)
	if err != nil {
		return zero, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, fmt.Errorf("cloudflare http %d: %s", resp.StatusCode, string(respBody))
	}

	var out cloudflareResponse[CloudflarePurgeCacheResult]
	if err := json.Unmarshal(respBody, &out); err != nil {
		return zero, fmt.Errorf("unmarshal response: %w; body=%s", err, string(respBody))
	}

	if !out.Success {
		return zero, fmt.Errorf("cloudflare success=false errors=%v", out.Errors)
	}

	return out.Result, nil
}
