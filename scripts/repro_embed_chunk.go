//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"

	"github.com/HardMakabaka/KB-Gateway/internal/chunk"
	"github.com/HardMakabaka/KB-Gateway/internal/embed"
)

func main() {
	chunks := chunk.Split(chunk.Config{MaxChars: 1200, Overlap: 200, MinChars: 1, HardLimit: 200}, "Hello world.\n\nThis is internal public.")
	fmt.Println("chunks", len(chunks))
	em := embed.NewFake(384)
	vecs, err := em.Embed(context.Background(), chunks)
	if err != nil {
		panic(err)
	}
	fmt.Println("vecs", len(vecs), "dim", len(vecs[0]))
}
