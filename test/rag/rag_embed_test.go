package rag_test

import (
	"context"
	"testing"

	"GopherAI/common/rag"
)

func TestGetEmbedding(t *testing.T) {
	baseURL := "http://localhost:11434"
	if baseURL == "" {
		t.Skip("ollamaConfig.baseURL not set in config/config.toml")
	}

	redisRag := rag.InitRedisRAG(rag.RedisConfig{}, rag.OllamaConfig{
		BaseURL:   baseURL,
		ModelName: "nomic-embed-text",
	})

	embedding, err := redisRag.GetEmbedding(context.Background(), "你好世界")
	if err != nil {
		t.Fatalf("GetEmbedding failed: %v", err)
	}
	if len(embedding) == 0 {
		t.Fatalf("GetEmbedding returned empty embedding")
	}
}
