package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/HardMakabaka/KB-Gateway/internal/chunk"
	"github.com/HardMakabaka/KB-Gateway/pkg/types"
)

type ingestRequest struct {
	ProjectID string   `json:"project_id"`
	DocID     string   `json:"doc_id"`
	Title     string   `json:"title"`
	Source    string   `json:"source"`
	PathOrURL string   `json:"path_or_url"`
	Content   string   `json:"content"`
	ACLPublic bool     `json:"acl_public"`
	ACLAllow  []string `json:"acl_allow"`
}

type ingestResponse struct {
	DocVersion    string `json:"doc_version"`
	ChunksWritten int    `json:"chunks_written"`
	Note          string `json:"note"`
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	var req ingestRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, int64(s.cfg.Limits.MaxContentBytes))).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	if req.ProjectID == "" || req.DocID == "" || strings.TrimSpace(req.Content) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_fields"})
		return
	}

	chunks := chunk.Split(s.chunkCfg, req.Content)
	// Embedding + Qdrant upsert is TODO in v1 skeleton.
	_ = chunks

	docVersion := time.Now().UTC().Format("20060102T150405Z")
	writeJSON(w, http.StatusOK, ingestResponse{DocVersion: docVersion, ChunksWritten: len(chunks), Note: "qdrant upsert not implemented yet"})
}

type activateRequest struct {
	ProjectID  string `json:"project_id"`
	DocID      string `json:"doc_id"`
	DocVersion string `json:"doc_version"`
}

func (s *Server) handleActivate(w http.ResponseWriter, r *http.Request) {
	var req activateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "note": "activate not implemented yet"})
}

type searchRequest struct {
	Query        string          `json:"query"`
	ProjectScope []string        `json:"project_scope"`
	Principal    types.Principal `json:"principal"`
	TopK         int             `json:"top_k"`
}

type searchResult struct {
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
	ProjectID  string  `json:"project_id"`
	DocID      string  `json:"doc_id"`
	DocVersion string  `json:"doc_version"`
	ChunkID    int     `json:"chunk_id"`
	Title      string  `json:"title"`
	PathOrURL  string  `json:"path_or_url"`
}

type searchResponse struct {
	Results []searchResult `json:"results"`
	Note    string         `json:"note"`
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	writeJSON(w, http.StatusOK, searchResponse{Results: nil, Note: "qdrant search not implemented yet"})
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
