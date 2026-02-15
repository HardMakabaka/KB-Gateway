# KB Gateway - Development Guide

## Local Dev
- Run Qdrant via docker compose.
- Run kb-gateway Go service locally.

## Testing
- Unit tests for:
  - ACL filter logic
  - version activation rules
  - chunking determinism
- Integration tests against Qdrant (optional in CI, run in docker).

## CI
- Lint + go test
- Basic build

