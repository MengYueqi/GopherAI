package conversation_compression_test

import (
	"GopherAI/common/tools/conversation_compression"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func TestCompressMessagesIfNeeded(t *testing.T) {
	if os.Getenv("RUN_CONVERSATION_COMPRESSION_TEST") != "1" {
		t.Skip("manual integration test; set RUN_CONVERSATION_COMPRESSION_TEST=1 to enable")
	}

	ctx := context.Background()
	llm := newCompressionModelFromEnv(t, ctx)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个编程助手，需要根据上下文连续回答问题。"),
		schema.UserMessage("我在做一个 Go 项目，需要给聊天历史增加自动压缩能力，避免上下文越来越长。"),
		schema.AssistantMessage("明白，可以在达到阈值时压缩较早消息，并保留最近几轮对话不参与压缩。", nil),
		schema.UserMessage("这个功能要基于 eino 的 schema.Message 和 ToolCallingChatModel 来实现。"),
		schema.AssistantMessage("可以把较早消息整理成摘要，再以前置 assistant 消息的方式插回消息列表。", nil),
		schema.UserMessage("压缩摘要里要保留用户目标、约束条件、已有结论，不要保留寒暄。"),
		schema.AssistantMessage("可以通过单独的 system prompt 约束模型输出成一段摘要。", nil),
		schema.UserMessage("最近一轮先不要压缩，因为后续回答要直接参考。"),
		schema.AssistantMessage("可以通过 keepRecentRounds 控制保留的最近轮次。", nil),
	}

	compressed, err := conversation_compression.CompressMessagesIfNeeded(ctx, messages, 40, 1, llm)
	if err != nil {
		t.Fatalf("CompressMessagesIfNeeded returned error: %v", err)
	}

	t.Log("compressed messages:")
	for i, msg := range compressed {
		if msg == nil {
			t.Logf("[%d] <nil>", i)
			continue
		}
		t.Logf("[%d] role=%s content=%s", i, msg.Role, msg.Content)
	}

	if len(compressed) < 3 {
		t.Fatalf("expected compressed messages length >= 3, got %d", len(compressed))
	}

	if compressed[0].Role != schema.System {
		t.Fatalf("expected first message to remain system, got %s", compressed[0].Role)
	}

	if compressed[1].Role != schema.Assistant {
		t.Fatalf("expected summary message role assistant, got %s", compressed[1].Role)
	}

	if !strings.HasPrefix(compressed[1].Content, "历史对话摘要：") {
		t.Fatalf("expected summary prefix, got %q", compressed[1].Content)
	}

	if compressed[len(compressed)-2].Content != "最近一轮先不要压缩，因为后续回答要直接参考。" {
		t.Fatalf("expected latest user round to be retained, got %q", compressed[len(compressed)-2].Content)
	}

	if compressed[len(compressed)-1].Content != "可以通过 keepRecentRounds 控制保留的最近轮次。" {
		t.Fatalf("expected latest assistant reply to be retained, got %q", compressed[len(compressed)-1].Content)
	}
}

func newCompressionModelFromEnv(t *testing.T, ctx context.Context) model.ToolCallingChatModel {
	t.Helper()

	apiKey := os.Getenv("OPENAI_API_KEY")
	modelName := os.Getenv("OPENAI_MODEL_NAME")
	baseURL := os.Getenv("OPENAI_BASE_URL_ALIYUN")

	if apiKey == "" {
		t.Fatal("OPENAI_API_KEY is required")
	}
	if modelName == "" {
		t.Fatal("OPENAI_MODEL_NAME is required")
	}
	if baseURL == "" {
		t.Fatal("OPENAI_BASE_URL_ALIYUN is required")
	}

	llm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  apiKey,
		Model:   modelName,
		BaseURL: baseURL,
	})
	if err != nil {
		t.Fatalf("create openai chat model failed: %v", err)
	}

	return llm
}
