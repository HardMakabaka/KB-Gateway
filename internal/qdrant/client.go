package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	if resp.StatusCode == 200 || resp.StatusCode == 409 {
		return nil
	}
	return fmt.Errorf("ensure collection: status %d", resp.StatusCode)
}

func (c *Client) Upsert(ctx context.Context, collection string, points []Point) error {
	// Qdrant expects PUT /points for upsert.
	body := map[string]any{"points": points}
	return c.put(ctx, fmt.Sprintf("/collections/%s/points?wait=true", collection), body, nil)
}

func (c *Client) SetPayload(ctx context.Context, collection string, payload map[string]any, filter Filter) error {
	body := map[string]any{
		"payload": payload,
		"filter":  filter,
	}
	return c.post(ctx, fmt.Sprintf("/collections/%s/points/payload?wait=true", collection), body, nil)
}

type SearchResult struct {
	ID      any            `json:"id"`
	Score   float64        `json:"score"`
	Payload map[string]any `json:"payload"`
}

type SearchResponse struct {
	Result []SearchResult `json:"result"`
}

func (c *Client) Search(ctx context.Context, collection string, vector []float32, filter Filter, limit int) ([]SearchResult, error) {
	body := map[string]any{
		"vector":       vector,
		"limit":        limit,
		"with_payload": true,
		"filter":       filter,
	}
	var out SearchResponse
	if err := c.post(ctx, fmt.Sprintf("/collections/%s/points/search", collection), body, &out); err != nil {
		return nil, err
	}
	return out.Result, nil
}

func (c *Client) post(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPost, path, body, out)
}

func (c *Client) put(ctx context.Context, path string, body any, out any) error {
	return c.do(ctx, http.MethodPut, path, body, out)
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		x, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qdrant %s %s status %d: %s", method, path, resp.StatusCode, string(x))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
