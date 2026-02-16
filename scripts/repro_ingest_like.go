//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HardMakabaka/KB-Gateway/internal/chunk"
	"github.com/HardMakabaka/KB-Gateway/internal/embed"
	"github.com/HardMakabaka/KB-Gateway/internal/qdrant"
	"github.com/google/uuid"
)

type ChunkPayload struct {
	ProjectID         string   `json:"project_id"`
	DocID             string   `json:"doc_id"`
	DocVersion        string   `json:"doc_version"`
	DocVersionTS      int64    `json:"doc_version_ts"`
	IsActive          bool     `json:"is_active"`
	ChunkID           int      `json:"chunk_id"`
	Text              string   `json:"text"`
	Deleted           bool     `json:"deleted"`
	ACLPublic         bool     `json:"acl_public"`
	ACLExternalPublic bool     `json:"acl_external_public"`
	ACLAllow          []string `json:"acl_allow"`
	CreatedAt         int64    `json:"created_at"`
	UpdatedAt         int64    `json:"updated_at"`
}

func main() {
	chunks := chunk.Split(chunk.Config{MaxChars: 1200, Overlap: 200, MinChars: 1, HardLimit: 200}, "Hello world.\n\nThis is internal public.")
	em := embed.NewFake(384)
	vecs, _ := em.Embed(context.Background(), chunks)
	cli := qdrant.New("http://localhost:6333", 10*time.Second)

	now := time.Now().UTC()
	ver := now.Format("20060102T150405Z")
	ts := now.Unix()

	var points []qdrant.Point
	for i, c := range chunks {
		pl := ChunkPayload{ProjectID: "proj1", DocID: "docA", DocVersion: ver, DocVersionTS: ts, IsActive: false, ChunkID: i, Text: c, Deleted: false, ACLPublic: true, ACLExternalPublic: false, ACLAllow: nil, CreatedAt: ts, UpdatedAt: ts}
		m := map[string]any{}
		b, _ := json.Marshal(pl)
		_ = json.Unmarshal(b, &m)
		points = append(points, qdrant.Point{ID: uuid.NewString(), Vector: vecs[i], Payload: m})
	}

	fmt.Println("points", len(points), "vecdim", len(points[0].Vector))
	if err := cli.Upsert(context.Background(), "kb_chunks", points); err != nil {
		panic(err)
	}
	fmt.Println("upsert ok")
}
