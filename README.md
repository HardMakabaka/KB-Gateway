# KB Gateway

KB Gateway is a performance-first Go HTTP service that provides a unified ingestion/search/versioning API for RAG knowledge bases.

## Embedding Providers

KB Gateway supports multiple embedding backends via `KBG_EMBED_PROVIDER`:

| Provider | Env Var | Default |
|---|---|---|
| `openai` (default) | `KBG_OPENAI_API_KEY`, `KBG_EMBED_MODEL` | `text-embedding-3-small` |
| `ollama` | `KBG_OLLAMA_URL`, `KBG_OLLAMA_MODEL` | `http://localhost:11434`, `bge-m3` |
| `openai-compatible` | `KBG_COMPATIBLE_URL`, `KBG_COMPATIBLE_MODEL` | `http://localhost:8000/v1`, `bge-m3` |

All providers share `KBG_EMBED_DIM` (default: `1536`) to set the vector dimension.

If `KBG_EMBED_PROVIDER=openai` and no API key is set, a deterministic FakeEmbedder is used for local dev.

## Docs
- docs/REQUIREMENTS.md
- docs/DESIGN.md
- docs/DEVELOPMENT.md
