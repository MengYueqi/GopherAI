package rag

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type OllamaConfig struct {
	ModelName string
	BaseURL   string
}

type RAGDatabase interface {
	AddOneData(ctx context.Context, content string) error                // 向数据库添加一条消息
	GetEmbedding(ctx context.Context, content string) ([]float32, error) // 获取内容的向量表示
}

type RedisRAG struct {
	RedisClient  *redis.Client
	OllamaConfig OllamaConfig
}

type OllamaResponseBody struct {
	Embedding []float32 `json:"embedding"`
}

type OllamaRequestBody struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

func InitRedisRAG(redisConfig RedisConfig, ollamaConfig OllamaConfig) *RedisRAG {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})

	return &RedisRAG{
		RedisClient:  rdb,
		OllamaConfig: ollamaConfig,
	}
}

func (redisRag *RedisRAG) AddOneData(ctx context.Context, content string) error {
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is empty")
	}
	if redisRag.RedisClient == nil {
		return fmt.Errorf("redis client is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取内容的向量表示
	embedding, err := redisRag.GetEmbedding(ctx, content)
	if err != nil {
		return err
	}
	// 将 float32 切片转换为字节切片
	vectorBytes := float32SliceToBytes(embedding)
	key := fmt.Sprintf("rag:data:%d", time.Now().UnixNano())

	if err := redisRag.RedisClient.HSet(ctx, key, map[string]interface{}{
		"content":   content,
		"embedding": vectorBytes,
	}).Err(); err != nil {
		return fmt.Errorf("store rag data failed: %w", err)
	}

	return nil
}

func float32SliceToBytes(values []float32) []byte {
	buf := make([]byte, 4*len(values))
	for i, v := range values {
		binary.LittleEndian.PutUint32(buf[i*4:(i+1)*4], math.Float32bits(v))
	}
	return buf
}
func (redisRag *RedisRAG) GetEmbedding(ctx context.Context, content string) ([]float32, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content is empty")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// 获取 ollama 配置
	baseURL := strings.TrimRight(redisRag.OllamaConfig.BaseURL, "/")
	if baseURL == "" {
		return nil, fmt.Errorf("ollama baseURL is empty")
	}
	modelName := strings.TrimSpace(redisRag.OllamaConfig.ModelName)
	if modelName == "" {
		modelName = "nomic-embed-text"
	}

	// 构造请求体
	reqBody := OllamaRequestBody{
		Model:  modelName,
		Prompt: content,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request failed: %w", err)
	}

	// 发送请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/embeddings", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create embedding request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 超时设置为30秒
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("embedding request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// 解析响应体
	var respBody OllamaResponseBody

	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("decode embedding response failed: %w", err)
	}
	if len(respBody.Embedding) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}

	log.Println("Origin Word: ", content, " Received embedding:", respBody.Embedding)

	return respBody.Embedding, nil
}
