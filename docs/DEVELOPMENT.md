# KB Gateway - Development Guide

## Prereqs
- Go 1.22+
- Docker (for Qdrant)

## Local Dev

### 1) Start Qdrant
```bash
docker compose up -d
```

### 2) Run KB Gateway
```bash
make run
# or
KBG_HTTP_ADDR=:8080 \
KBG_QDRANT_URL=http://localhost:6333 \
KBG_QDRANT_COLLECTION=kb_chunks \
make run
```

> If `OPENAI_API_KEY` is not set, the service uses a deterministic **FakeEmbedder** (dev-only) so you can still run end-to-end.

### 3) Smoke test
Ingest a short internal-public doc (short docs are accepted; chunker fallback ensures we never send an empty upsert).

```bash
curl -sS -X POST http://localhost:8080/v1/docs/ingest \
  -H 'Content-Type: application/json' \
  -d '{"project_id":"proj1","doc_id":"docA","title":"Doc A","source":"markdown","path_or_url":"README.md","content":"Short doc.","acl_public":true,"acl_allow":[]}' | jq .

# internal can see
curl -sS -X POST http://localhost:8080/v1/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"Short","project_scope":["proj1"],"principal":{"type":"internal_user","id":"u1","groups":[]},"top_k":5}' | jq '.results|length'

# customer cannot see internal-public
curl -sS -X POST http://localhost:8080/v1/search \
  -H 'Content-Type: application/json' \
  -d '{"query":"Short","project_scope":["proj1"],"principal":{"type":"customer_user","id":"c1","groups":["customer:acme"]},"top_k":5}' | jq '.results|length'
```

## Testing
```bash
make test
```

Includes unit tests for:
- ACL semantics (customer does NOT inherit `acl_public`)
- Qdrant filter shape

## CI (planned)
- `go test ./...`
- build

