package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HardMakabaka/KB-Gateway/internal/chunk"
	"github.com/HardMakabaka/KB-Gateway/internal/qdrant"
	"github.com/HardMakabaka/KB-Gateway/pkg/types"
	"github.com/google/uuid"
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

	unlock := s.docLocks.Lock(req.ProjectID + ":" + req.DocID)
	defer unlock()

	now := time.Now().UTC()
	docVersion := now.Format("20060102T150405Z")
	docVersionTS := now.Unix()

	chunks := chunk.Split(s.chunkCfg, req.Content)
	// Safety fallback: for very short docs, the chunker may return 0 chunks due to MinChars.
	// We still want to index the content rather than sending an empty upsert to Qdrant.
	if len(chunks) == 0 {
		chunks = []string{strings.TrimSpace(req.Content)}
	}
	vecs, err := s.embedder.Embed(r.Context(), chunks)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "embed_failed"})
		return
	}

	points := make([]qdrant.Point, 0, len(chunks))
	for i, c := range chunks {
		payload := ChunkPayload{
			ProjectID:         req.ProjectID,
			DocID:             req.DocID,
			DocVersion:        docVersion,
			DocVersionTS:      docVersionTS,
			IsActive:          false,
			ChunkID:           i,
			Source:            req.Source,
			Title:             req.Title,
			PathOrURL:         req.PathOrURL,
			Text:              c,
			ContentHash:       hashString(c),
			ACLPublic:         req.ACLPublic,
			ACLExternalPublic: false,
			ACLAllow:          req.ACLAllow,
			CreatedAt:         docVersionTS,
			UpdatedAt:         docVersionTS,
			Deleted:           false,
		}
		p := map[string]any{}
		b, _ := json.Marshal(payload)
		_ = json.Unmarshal(b, &p)

		points = append(points, qdrant.Point{
			ID:      uuid.NewString(),
			Vector:  vecs[i],
			Payload: p,
		})
	}

	if err := s.qdrant.Upsert(r.Context(), s.cfg.Qdrant.Collection, points); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "qdrant_upsert_failed", "detail": err.Error()})
		return
	}

	if err := s.activateLocked(r.Context(), req.ProjectID, req.DocID, docVersion); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "activate_failed", "detail": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, ingestResponse{DocVersion: docVersion, ChunksWritten: len(chunks)})
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
	if req.ProjectID == "" || req.DocID == "" || req.DocVersion == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_fields"})
		return
	}

	unlock := s.docLocks.Lock(req.ProjectID + ":" + req.DocID)
	defer unlock()

	if err := s.activateLocked(r.Context(), req.ProjectID, req.DocID, req.DocVersion); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "activate_failed", "detail": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) activateLocked(ctx context.Context, projectID, docID, docVersion string) error {
	fDeactivate := qdrant.Filter{"must": []any{
		map[string]any{"key": "project_id", "match": map[string]any{"value": projectID}},
		map[string]any{"key": "doc_id", "match": map[string]any{"value": docID}},
		map[string]any{"key": "is_active", "match": map[string]any{"value": true}},
	}}
	if err := s.qdrant.SetPayload(ctx, s.cfg.Qdrant.Collection, map[string]any{"is_active": false}, fDeactivate); err != nil {
		return err
	}

	fActivate := qdrant.Filter{"must": []any{
		map[string]any{"key": "project_id", "match": map[string]any{"value": projectID}},
		map[string]any{"key": "doc_id", "match": map[string]any{"value": docID}},
		map[string]any{"key": "doc_version", "match": map[string]any{"value": docVersion}},
	}}
	return s.qdrant.SetPayload(ctx, s.cfg.Qdrant.Collection, map[string]any{"is_active": true, "updated_at": time.Now().UTC().Unix()}, fActivate)
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
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid_json"})
		return
	}
	if strings.TrimSpace(req.Query) == "" || len(req.ProjectScope) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "missing_fields"})
		return
	}
	limit := req.TopK
	if limit <= 0 {
		limit = 10
	}

	vecs, err := s.embedder.Embed(r.Context(), []string{req.Query})
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "embed_failed"})
		return
	}

	f := andFilters(buildBaseFilter(req.ProjectScope), buildACLFilter(req.Principal))
	res, err := s.qdrant.Search(r.Context(), s.cfg.Qdrant.Collection, vecs[0], f, limit)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "qdrant_search_failed", "detail": err.Error()})
		return
	}

	out := make([]searchResult, 0, len(res))
	for _, it := range res {
		p := it.Payload
		out = append(out, searchResult{
			Text:       toString(p["text"]),
			Score:      it.Score,
			ProjectID:  toString(p["project_id"]),
			DocID:      toString(p["doc_id"]),
			DocVersion: toString(p["doc_version"]),
			ChunkID:    toInt(p["chunk_id"]),
			Title:      toString(p["title"]),
			PathOrURL:  toString(p["path_or_url"]),
		})
	}
	writeJSON(w, http.StatusOK, searchResponse{Results: out})
}

func toString(v any) string {
	s, _ := v.(string)
	return s
}

func toInt(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case string:
		i, _ := strconv.Atoi(x)
		return i
	default:
		return 0
	}
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
