package qdrant

type Point struct {
	ID      any            `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

type Filter map[string]any
