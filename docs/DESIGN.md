# KB Gateway - Design (v1)

## Overview
KB Gateway is a Go HTTP service that sits between callers (apps, scripts, future Dify, customer support) and the index store (Qdrant) plus external model APIs (embedding/LLM, optional in v1).

v1 focuses on:
- ingestion (normalized text pushed in)
- chunking
- embedding
- Qdrant write + active version switching
- secure search with ACL filtering

## Architecture
- HTTP API (JSON)
- Core services:
  - Chunker
  - Embedder (OpenAI text-embedding-3-small)
  - Qdrant client
  - Version manager (activate/rollback)
  - ACL filter builder
- Storage:
  - Qdrant (vectors + payload)
  - Optional: Postgres/SQLite for ingest jobs/audit (defer; start with file/stdout logs)

## Data Model

### Qdrant Collection: `kb_chunks`
Payload fields (subset):
- project_id (string)
- doc_id (string)
- doc_version (string)
- doc_version_ts (int)
- is_active (bool)
- chunk_id (int)
- source (string)
- title (string)
- path_or_url (string)
- content_hash (string)
- acl_public (bool)
- acl_allow ([string])
- acl_external_public (bool, default false)
- created_at (int)
- updated_at (int)
- deleted (bool, optional)

Vector:
- embedding (float[])

## ACL Semantics
- Internal principal can access a chunk if:
  - acl_public == true OR intersects(acl_allow, principal.groups)
- Customer principal can access a chunk if:
  - acl_external_public == true OR intersects(acl_allow, principal.groups)
  - (acl_public does NOT grant customer access)

## Versioning
- Ingest creates new doc_version V2 with is_active=false.
- After successful upsert of all chunks, activate V2:
  - set is_active=false for currently active version (doc_id filter)
  - set is_active=true for V2
- Rollback is activate(target_version).

Concurrency:
- Serialize operations per (project_id, doc_id). Approach: in-memory lock map (single instance). Future: distributed lock if multi-replica.

## API (Draft)

### POST /v1/docs/ingest
Input:
- project_id
- doc_id
- title
- source
- path_or_url
- content (plain text or markdown)
- acl_public
- acl_allow[]

Output:
- doc_version
- chunks_written

### POST /v1/docs/activate
Input:
- project_id
- doc_id
- doc_version

### POST /v1/search
Input:
- query
- project_scope[]
- principal {type,id,groups[]}
- top_k

Output:
- results[] with citations

### POST /v1/docs/delete
Input:
- project_id
- doc_id
- hard (bool)

Behavior:
- `hard=false`: soft delete (set `deleted=true`, `is_active=false` for all versions)
- `hard=true`: delete points by filter

### POST /v1/docs/rollback
Input:
- project_id
- doc_id
- target_doc_version

Behavior:
- Alias of activate semantics: deactivate current active, activate target version

## Chunking
- v1: recursive text splitting with overlap.
- markdown: header-aware splitting (best-effort) before recursive fallback.

## Future Enhancements
- Hybrid search (BM25) + reranker.
- Audit storage and analytics.
- Dify integration.
