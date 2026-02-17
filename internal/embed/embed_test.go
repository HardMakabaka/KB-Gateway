package embed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HardMakabaka/KB-Gateway/internal/config"
)

func TestNewFactory_OpenAI_NoKey_FallsFakeEmbedder(t *testing.T) {
	e, err := New(config.EmbedConfig{Provider: "openai", APIKey: ""})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := e.(*FakeEmbedder); !ok {
		t.Fatalf("expected FakeEmbedder, got %T", e)
	}
}

func TestNewFactory_UnknownProvider(t *testing.T) {
	_, err := New(config.EmbedConfig{Provider: "nope"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestOpenAIEmbedder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var req openAIEmbedReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := openAIEmbedResp{}
		for range req.Input {
			resp.Data = append(resp.Data, struct {
				Embedding []float32 `json:"embedding"`
			}{Embedding: []float32{0.1, 0.2, 0.3}})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e := NewOpenAI("test-key", "test-model", 3)
	if e.Dim() != 3 {
		t.Fatalf("expected dim 3, got %d", e.Dim())
	}
}

func TestOllamaEmbedder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			http.Error(w, "wrong path: "+r.URL.Path, http.StatusNotFound)
			return
		}
		var req ollamaEmbedReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := ollamaEmbedResp{Embedding: []float32{0.1, 0.2, 0.3}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e := NewOllama(srv.URL, "test-model", 3)
	if e.Dim() != 3 {
		t.Fatalf("expected dim 3, got %d", e.Dim())
	}

	vecs, err := e.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vecs))
	}
	if len(vecs[0]) != 3 {
		t.Fatalf("expected dim 3, got %d", len(vecs[0]))
	}
}

func TestOllamaEmbedder_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ollamaEmbedResp{Error: "model not found"})
	}))
	defer srv.Close()

	e := NewOllama(srv.URL, "bad-model", 3)
	_, err := e.Embed(context.Background(), []string{"hello"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCompatibleEmbedder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			http.Error(w, "wrong path: "+r.URL.Path, http.StatusNotFound)
			return
		}
		var req openAIEmbedReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := openAIEmbedResp{}
		for range req.Input {
			resp.Data = append(resp.Data, struct {
				Embedding []float32 `json:"embedding"`
			}{Embedding: []float32{0.4, 0.5, 0.6}})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	e := NewCompatible(srv.URL+"/v1", "test-model", 3)
	if e.Dim() != 3 {
		t.Fatalf("expected dim 3, got %d", e.Dim())
	}

	vecs, err := e.Embed(context.Background(), []string{"hello", "world"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vecs))
	}
	if vecs[0][0] != 0.4 {
		t.Fatalf("expected 0.4, got %f", vecs[0][0])
	}
}

func TestCompatibleEmbedder_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	e := NewCompatible(srv.URL+"/v1", "test-model", 3)
	_, err := e.Embed(context.Background(), []string{"hello"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFakeEmbedder_Deterministic(t *testing.T) {
	e := NewFake(256)
	v1, _ := e.Embed(context.Background(), []string{"hello"})
	v2, _ := e.Embed(context.Background(), []string{"hello"})
	for i := range v1[0] {
		if v1[0][i] != v2[0][i] {
			t.Fatalf("FakeEmbedder not deterministic at index %d", i)
		}
	}
}
