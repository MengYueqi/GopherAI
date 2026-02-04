package rag_test

import (
	"context"
	"log"
	"os"
	"testing"

	"GopherAI/common/rag"
)

func countRAGKeys(ctx context.Context, redisRag *rag.RedisRAG) (int, error) {
	var cursor uint64
	total := 0
	for {
		keys, next, err := redisRag.RedisClient.Scan(ctx, cursor, "rag:data:*", 100).Result()
		if err != nil {
			return 0, err
		}
		total += len(keys)
		if next == 0 {
			return total, nil
		}
		cursor = next
	}
}

func TestAddOneData(t *testing.T) {
	ctx := context.Background()
	redisAddr := os.Getenv("REDIS_RAG_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6381"
	}
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	redisRag := rag.InitRedisRAG(rag.RedisConfig{Addr: redisAddr}, rag.OllamaConfig{
		BaseURL:   baseURL,
		ModelName: "nomic-embed-text",
	})
	if err := redisRag.RedisClient.Ping(ctx).Err(); err != nil {
		t.Skipf("redis not available at %s: %v", redisAddr, err)
	}

	before, err := countRAGKeys(ctx, redisRag)
	if err != nil {
		t.Fatalf("count rag keys before failed: %v", err)
	}

	if err := redisRag.AddOneData(ctx, "你好，世界"); err != nil {
		t.Fatalf("AddOneData failed: %v", err)
	}
	// 输出嵌入之前 key 的数量
	log.Printf("RAG keys before: %d", before)

	after, err := countRAGKeys(ctx, redisRag)
	if err != nil {
		t.Fatalf("count rag keys after failed: %v", err)
	}
	if after <= before {
		t.Fatalf("expected rag key count to increase, before=%d after=%d", before, after)
	}
	log.Printf("RAG keys after: %d", after)
}
