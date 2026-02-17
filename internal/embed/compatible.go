package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CompatibleEmbedder calls any OpenAI-compatible /v1/embeddings endpoint (vLLM, LocalAI, etc.).
type CompatibleEmbedder struct {
	baseURL string
	model   string
	dim     int
	client  *http.Client
}

func NewCompatible(baseURL, model string, dim int) *CompatibleEmbedder {
	return &CompatibleEmbedder{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		dim:     dim,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *CompatibleEmbedder) Dim() int { return e.dim }

func (e *CompatibleEmbedder) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	url := e.baseURL + "/embeddings"

	body, err := json.Marshal(openAIEmbedReq{Input: inputs, Model: e.model})
	if err != nil {
		return nil, fmt.Errorf("compatible embed: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("compatible embed: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("compatible embed: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("compatible embed: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("compatible embed: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result openAIEmbedResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("compatible embed: unmarshal: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("compatible embed: api error: %s", result.Error.Message)
	}
	if len(result.Data) != len(inputs) {
		return nil, fmt.Errorf("compatible embed: expected %d embeddings, got %d", len(inputs), len(result.Data))
	}

	out := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		out[i] = d.Embedding
	}
	return out, nil
}
