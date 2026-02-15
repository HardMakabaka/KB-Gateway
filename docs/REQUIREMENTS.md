# KB Gateway - Requirements

## Background
We need a small-to-medium scale, multi-tech-stack RAG knowledge base capability.

Core approach:
- Qdrant as the long-term vector/index base.
- Single Qdrant collection with `project_id` filter for cross-project search.
- Multi-version per document with `is_active` switch for fast rollback.
- Dify may be added later as an optional UI/orchestration layer; it must not become the source of truth.

## Goals
- Provide a unified HTTP API for ingesting documents, searching chunks, and managing versions.
- Enforce access control at query-time to prevent data leaks across projects/customers.
- Support daily updates and rollback.
- Support sources: Git repositories / Markdown / Web pages / PDF (initially via external feeders pushing normalized text).

## Non-goals (v1)
- Full IAM/SSO integration.
- A full UI for knowledge base management.
- Advanced evaluation platform (only minimal eval harness is required).

## Key Decisions
- Language: Go (performance-first, simple deployment).
- Versioning: keep multiple versions, switch active version atomically.
- Access control semantics:
  - `acl_public=true` means **internal-public only**.
  - Customer visibility requires explicit allow via `acl_allow`.
  - Future: `acl_external_public=true` can be introduced for "public to all customers" with publishing workflow.

## Users / Personas
- Internal users (developers/ops) using KB for project work.
- Customer users viewing permitted project docs.
- Future: intelligent customer support (must be safe-by-default).

## Functional Requirements

### Document Ingestion
- Upsert a document into a project.
- Parse/normalize content into chunks.
- Create a new `doc_version` for each ingest.
- Write chunks to Qdrant with `is_active=false` then atomically activate the new version.
- Support delete and rollback.

### Search
- Vector search with filters:
  - `project_id` in scope
  - `is_active=true`
  - ACL enforcement
- Return results with full citation metadata.

### Access Control
- Request includes a `principal` (type/id/groups).
- ACL fields:
  - `acl_public` (internal-public)
  - `acl_allow` (groups)
  - optional `acl_external_public` (future)

### Versioning
- Support `activate(doc_id, doc_version)`.
- Ensure only one active version per `doc_id`.
- Concurrency: operations for the same `doc_id` must be serialized (implementation detail).

## Non-functional Requirements
- Performance: p95 search latency < 500ms at small/medium scale (excluding LLM calls).
- Safety: no cross-tenant leakage; audit logs (v1 minimal).
- Operability: dockerized Qdrant; app runnable via docker compose or binary.

## Acceptance Criteria (v1)
- Search results differ appropriately for internal vs customer principals.
- Customer principals cannot see internal-public docs.
- Customer A cannot see Customer B docs.
- Ingesting a new version switches citations to new `doc_version`.
- Rollback switches citations back to target `doc_version`.

