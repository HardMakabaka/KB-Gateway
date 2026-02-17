package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTP   HTTPConfig
	Qdrant QdrantConfig
	Embed  EmbedConfig
	Chunk  ChunkConfig
	Limits LimitsConfig
}

type HTTPConfig struct {
	Addr string `envconfig:"HTTP_ADDR" default:":8080"`
}

type QdrantConfig struct {
	URL        string        `envconfig:"QDRANT_URL" default:"http://localhost:6333"`
	Collection string        `envconfig:"QDRANT_COLLECTION" default:"kb_chunks"`
	Timeout    time.Duration `envconfig:"QDRANT_TIMEOUT" default:"10s"`
}

type EmbedConfig struct {
	Provider        string `envconfig:"EMBED_PROVIDER" default:"openai"`
	Model           string `envconfig:"EMBED_MODEL" default:"text-embedding-3-small"`
	APIKey          string `envconfig:"OPENAI_API_KEY" default:""`
	Dim             int    `envconfig:"EMBED_DIM" default:"1536"`
	OllamaURL       string `envconfig:"OLLAMA_URL" default:"http://localhost:11434"`
	OllamaModel     string `envconfig:"OLLAMA_MODEL" default:"bge-m3"`
	CompatibleURL   string `envconfig:"COMPATIBLE_URL" default:"http://localhost:8000/v1"`
	CompatibleModel string `envconfig:"COMPATIBLE_MODEL" default:"bge-m3"`
}

type ChunkConfig struct {
	MaxChars  int `envconfig:"CHUNK_MAX_CHARS" default:"1200"`
	Overlap   int `envconfig:"CHUNK_OVERLAP" default:"200"`
	MinChars  int `envconfig:"CHUNK_MIN_CHARS" default:"200"`
	HardLimit int `envconfig:"CHUNK_HARD_LIMIT" default:"200"`
}

type LimitsConfig struct {
	MaxContentBytes int `envconfig:"MAX_CONTENT_BYTES" default:"5242880"`
}

func Load() (Config, error) {
	var cfg Config
	if err := envconfig.Process("KBG", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
