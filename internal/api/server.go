package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/HardMakabaka/KB-Gateway/internal/chunk"
	"github.com/HardMakabaka/KB-Gateway/internal/config"
	"github.com/HardMakabaka/KB-Gateway/internal/embed"
	"github.com/HardMakabaka/KB-Gateway/internal/qdrant"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	cfg      config.Config
	qdrant   *qdrant.Client
	embedder embed.Embedder
	chunkCfg chunk.Config
	docLocks KeyedMutex
}

func NewServer(cfg config.Config) http.Handler {
	s := &Server{cfg: cfg}
	s.qdrant = qdrant.New(cfg.Qdrant.URL, cfg.Qdrant.Timeout)

	// v1: use fake embedder if no API key configured to keep local dev unblocked.
	if cfg.Embed.APIKey == "" {
		log.Printf("warning: OPENAI_API_KEY not set; using fake embedder")
		s.embedder = embed.NewFake(384)
	} else {
		// TODO: implement OpenAI embedder
		log.Printf("warning: OpenAI embedder not implemented yet; using fake embedder")
		s.embedder = embed.NewFake(384)
	}

	s.chunkCfg = chunk.Config{MaxChars: cfg.Chunk.MaxChars, Overlap: cfg.Chunk.Overlap, MinChars: cfg.Chunk.MinChars, HardLimit: cfg.Chunk.HardLimit}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	r.Route("/v1", func(r chi.Router) {
		r.Post("/docs/ingest", s.handleIngest)
		r.Post("/docs/activate", s.handleActivate)
		r.Post("/search", s.handleSearch)
	})

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.qdrant.EnsureCollection(ctx, cfg.Qdrant.Collection, s.embedder.Dim()); err != nil {
			log.Printf("ensure qdrant collection failed: %v", err)
		}
		// Ensure deleted=false is present for new docs; we rely on matchBool("deleted", false).
		// (If missing, qdrant match will not match; v1 requires deleted field to be always set.)
	}()

	return r
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
