//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HardMakabaka/KB-Gateway/internal/qdrant"
	"github.com/google/uuid"
)

func main() {
	cli := qdrant.New("http://localhost:6333", 10*time.Second)
	vec := make([]float32, 384)
	p := qdrant.Point{ID: uuid.NewString(), Vector: vec, Payload: map[string]any{"text": "hi", "deleted": false}}
	if err := cli.Upsert(context.Background(), "kb_chunks", []qdrant.Point{p}); err != nil {
		panic(err)
	}
	b, _ := json.Marshal(p)
	fmt.Println("ok upsert one, point payload bytes", len(b))
}
