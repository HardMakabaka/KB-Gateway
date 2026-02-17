package embed

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaEmbedder struct {
	baseURL string
	model   string
	dim     int
	client  *http.Client
}

func NewOllama(baseURL, model string, dim int) *OllamaEmbedder {
	return &OllamaEmbedder{
		baseURL: baseURL,
		model:   model,
		dim:     dim,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *OllamaEmbedder) Dim() int { return e.dim }

type ollamaEmbedReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbedResp struct {
	Embedding []float32 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}

func (e *OllamaEmbedder) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	url := e.baseURL + "/api/embeddings"
	out := make([][]float32, 0, len(inputs))

	for _, input := range inputs {
		body, err := json.Marshal(ollamaEmbedReq{Model: e.model, Prompt: input})
		if err != nil {
			return nil, fmt.Errorf("ollama embed: marshal: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("ollama embed: new request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ollama embed: request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("ollama embed: read body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("ollama embed: status %d: %s", resp.StatusCode, string(respBody))
		}

		var result ollamaEmbedResp
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("ollama embed: unmarshal: %w", err)
		}
		if result.Error != "" {
			return nil, fmt.Errorf("ollama embed: api error: %s", result.Error)
		}

		out = append(out, result.Embedding)
	}
	return out, nil
}
