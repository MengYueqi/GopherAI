package rag_test

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"GopherAI/common/rag"
)

func TestGetKNN(t *testing.T) {
	ctx := context.Background()
	redisAddr := os.Getenv("REDIS_RAG_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6381"
	}
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if baseURL == "" {
		t.Skip("ollama baseURL is empty")
	}

	// Initialize RAG client with Redis and Ollama settings.
	redisRag := rag.InitRedisRAG(rag.RedisConfig{Addr: redisAddr}, rag.OllamaConfig{
		BaseURL:   baseURL,
		ModelName: "nomic-embed-text",
	})
	if err := redisRag.RedisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available at %s: %v", redisAddr, err)
	}
	// Ensure the RediSearch index exists before running KNN search.
	if err := redisRag.RedisClient.Do(ctx, "FT.INFO", "idx:rag_data").Err(); err != nil {
		t.Skipf("rag index not available (run common/rag/redis_init.sh): %v", err)
	}
	// Query nearest neighbors for the target text.
	results, err := redisRag.GetKNN(ctx, "小宝宝有点发烧，什么情况", 5)
	if err != nil {
		t.Fatalf("GetKNN failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("GetKNN returned no results")
	}
	if strings.TrimSpace(results[0].Content) == "" {
		t.Fatalf("GetKNN returned empty content")
	}
	for _, res := range results {
		log.Printf("GetKNN result content: %s, score: %s", res.Content, strconv.FormatFloat(res.Score, 'f', 6, 64))
	}
}
