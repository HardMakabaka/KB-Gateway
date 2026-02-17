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

// OpenAIEmbedder calls the official OpenAI /v1/embeddings endpoint.
type OpenAIEmbedder struct {
	apiKey string
	model  string
	dim    int
	client *http.Client
}

func NewOpenAI(apiKey, model string, dim int) *OpenAIEmbedder {
	return &OpenAIEmbedder{
		apiKey: apiKey,
		model:  model,
		dim:    dim,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *OpenAIEmbedder) Dim() int { return e.dim }

type openAIEmbedReq struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type openAIEmbedResp struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	body, err := json.Marshal(openAIEmbedReq{Input: inputs, Model: e.model})
	if err != nil {
		return nil, fmt.Errorf("openai embed: marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("openai embed: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.apiKey)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embed: request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai embed: read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai embed: status %d: %s", resp.StatusCode, string(respBody))
	}

	var result openAIEmbedResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("openai embed: unmarshal: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("openai embed: api error: %s", result.Error.Message)
	}
	if len(result.Data) != len(inputs) {
		return nil, fmt.Errorf("openai embed: expected %d embeddings, got %d", len(inputs), len(result.Data))
	}

	out := make([][]float32, len(result.Data))
	for i, d := range result.Data {
		out[i] = d.Embedding
	}
	return out, nil
}
