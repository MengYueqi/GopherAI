package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"GopherAI/common/rag"
	"GopherAI/config"
)

type SampleItem struct {
	ID      json.Number `json:"id"`
	Content string      `json:"content"`
}

// 读取 sample JSON（数组）并反序列化为结构体切片
func loadSample(path string) ([]SampleItem, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.UseNumber()
	var items []SampleItem
	if err := decoder.Decode(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func main() {
	// 读取配置文件作为默认值
	cfg := config.GetConfig()
	defaultRedisAddr := ""
	if cfg.RedisHost != "" && cfg.RedisPort != 0 {
		defaultRedisAddr = fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort)
	}

	//  example command:
	// 	go run data/script/addDataToRedis.go \
	//   --input data/all_medical_sample.json \
	//   --redis-addr 127.0.0.1:6379 \
	//   --ollama-base-url http://localhost:11434 \
	//   --ollama-model nomic-embed-text

	// 允许通过命令行覆盖配置
	inputPath := flag.String("input", "data/all_medical_sample.json", "Path to sample JSON file.")
	redisAddr := flag.String("redis-addr", defaultRedisAddr, "Redis addr, e.g. 127.0.0.1:6379.")
	redisPassword := flag.String("redis-password", cfg.RedisPassword, "Redis password.")
	redisDB := flag.Int("redis-db", cfg.RedisDb, "Redis DB.")
	ollamaBaseURL := flag.String("ollama-base-url", cfg.OllamaConfig.BaseURL, "Ollama base URL.")
	ollamaModel := flag.String("ollama-model", cfg.OllamaConfig.ModelName, "Ollama model name.")
	flag.Parse()

	if *redisAddr == "" {
		log.Fatal("redis addr is empty")
	}
	if *ollamaBaseURL == "" {
		log.Fatal("ollama base URL is empty")
	}

	items, err := loadSample(*inputPath)
	if err != nil {
		log.Fatalf("load sample failed: %v", err)
	}
	if len(items) == 0 {
		log.Fatal("sample file is empty")
	}

	// 初始化 RAG 组件
	redisRag := rag.InitRedisRAG(rag.RedisConfig{
		Addr:     *redisAddr,
		Password: *redisPassword,
		DB:       *redisDB,
	}, rag.OllamaConfig{
		BaseURL:   *ollamaBaseURL,
		ModelName: *ollamaModel,
	})

	ctx := context.Background()
	// 提前检查 Redis 可用性
	if err := redisRag.RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	inserted := 0
	skipped := 0
	for i, item := range items {
		// 跳过空内容
		if strings.TrimSpace(item.Content) == "" {
			skipped++
			continue
		}
		// 调用 rag 接口生成向量并写入 Redis
		if err := redisRag.AddOneData(ctx, item.Content); err != nil {
			log.Printf("add data failed at index %d (id=%s): %v", i, item.ID.String(), err)
			continue
		}
		inserted++
		if inserted%50 == 0 {
			log.Printf("embedded %d/%d", inserted, len(items))
		}
	}

	log.Printf("done. inserted=%d skipped=%d total=%d", inserted, skipped, len(items))
}
