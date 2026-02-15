package embed

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
)

type Embedder interface {
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
	Dim() int
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
