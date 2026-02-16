//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]any         `json:"payload"`
}

func main() {
	vec := make([]float32, 384)
	p := Point{ID: uuid.NewString(), Vector: vec, Payload: map[string]any{"text": "hi", "deleted": false}}
	body := map[string]any{"points": []Point{p}}
	b, _ := json.Marshal(body)
	fmt.Println("bytes", len(b))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPut, "http://localhost:6333/collections/kb_chunks/points?wait=true", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	cli := &http.Client{Timeout: 10 * time.Second}
	resp, err := cli.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var out any
	_ = json.NewDecoder(resp.Body).Decode(&out)
	fmt.Println(resp.StatusCode, out)
}
