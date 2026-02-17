package embed

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/HardMakabaka/KB-Gateway/internal/config"
)

type Embedder interface {
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
	Dim() int
}

const fakeDim = 384

func New(cfg config.EmbedConfig) (Embedder, error) {
	switch cfg.Provider {
	case "openai":
		if cfg.APIKey == "" {
			log.Printf("warning: OPENAI_API_KEY not set; using fake embedder")
			return NewFake(fakeDim), nil
		}
		return NewOpenAI(cfg.APIKey, cfg.Model, cfg.Dim), nil
	case "ollama":
		return NewOllama(cfg.OllamaURL, cfg.OllamaModel, cfg.Dim), nil
	case "openai-compatible":
		return NewCompatible(cfg.CompatibleURL, cfg.CompatibleModel, cfg.Dim), nil
	default:
		return nil, fmt.Errorf("unknown embed provider: %q (supported: openai, ollama, openai-compatible)", cfg.Provider)
	}
}

// FakeEmbedder is a deterministic embedder for local dev/tests when no API key is configured.
// It should not be used in production.
type FakeEmbedder struct{ dim int }

func NewFake(dim int) *FakeEmbedder { return &FakeEmbedder{dim: dim} }

func (f *FakeEmbedder) Dim() int { return f.dim }

func (f *FakeEmbedder) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	_ = ctx
	out := make([][]float32, 0, len(inputs))
	for _, in := range inputs {
		h := sha256.Sum256([]byte(in))
		vec := make([]float32, f.dim)
		for i := 0; i < f.dim; i++ {
			b := binary.LittleEndian.Uint16(h[(i*2)%len(h):])
			vec[i] = float32(int(b)%2000-1000) / 1000.0
		}
		out = append(out, vec)
	}
	return out, nil
}
