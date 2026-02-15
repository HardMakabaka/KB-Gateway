package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, httpClient: &http.Client{Timeout: timeout}}
}

func (c *Client) EnsureCollection(ctx context.Context, name string, vectorDim int) error {
	// Best-effort: try create; if already exists, ignore.
	body := map[string]any{
		"vectors": map[string]any{
			"size":     vectorDim,
			"distance": "Cosine",
		},
	}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/collections/%s", c.baseURL, name), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	if resp.StatusCode == 409 {
		return nil
	}
	return fmt.Errorf("ensure collection: status %d", resp.StatusCode)
}
