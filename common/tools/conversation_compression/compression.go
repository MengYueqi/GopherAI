package conversation_compression

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const compressionPrompt = `你是一个对话压缩助手。请将提供的历史对话压缩成一段简洁、连贯的中文摘要。

要求：
1. 只保留对后续对话有帮助的信息，例如用户目标、背景事实、约束条件、偏好、已经完成的结论、待解决问题。
2. 删除寒暄、重复表达、无关细节。
3. 不要使用列表，不要分段，输出一段话即可。
4. 使用第三人称或客观陈述方式总结，不要写“以下是总结”之类的引导语。
5. 如果历史对话中存在明确结论、代码决策、接口约束、业务规则，务必保留。
6. 不要虚构对话中没有出现的信息。`

// CompressMessagesIfNeeded 会在消息总 token 估算值达到阈值时压缩较早的对话，
// 并将压缩结果作为一条 assistant 消息插入到消息前部。
//
// 说明：
// 1. token 统计为近似估算，不是模型侧精确 tokenizer。
// 2. keepRecentRounds 表示保留最近多少轮用户发言及其后续回复不参与压缩。
// 3. 前置 system 消息不会被压缩，压缩摘要会插入在所有前置 system 消息之后。
func CompressMessagesIfNeeded(
	ctx context.Context,
	messages []*schema.Message,
	triggerTokens int,
	keepRecentRounds int,
	compressModel model.ToolCallingChatModel,
) ([]*schema.Message, error) {
	if len(messages) == 0 || compressModel == nil || triggerTokens <= 0 {
		return cloneMessages(messages), nil
	}

	if estimateMessagesTokens(messages) < triggerTokens {
		return cloneMessages(messages), nil
	}

	systemPrefixEnd := countLeadingSystemMessages(messages)
	compressEnd := findCompressionEnd(messages, keepRecentRounds)
	if compressEnd <= systemPrefixEnd {
		return cloneMessages(messages), nil
	}

	toCompress := cloneMessages(messages[systemPrefixEnd:compressEnd])
	if len(toCompress) == 0 {
		return cloneMessages(messages), nil
	}

	summary, err := summarizeMessages(ctx, compressModel, toCompress)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(summary) == "" {
		return cloneMessages(messages), nil
	}

	result := make([]*schema.Message, 0, len(messages)-(compressEnd-systemPrefixEnd)+1)
	result = append(result, cloneMessages(messages[:systemPrefixEnd])...)
	result = append(result, &schema.Message{
		Role:    schema.Assistant,
		Content: "历史对话摘要：" + strings.TrimSpace(summary),
	})
	result = append(result, cloneMessages(messages[compressEnd:])...)

	return result, nil
}

func summarizeMessages(
	ctx context.Context,
	compressModel model.ToolCallingChatModel,
	messages []*schema.Message,
) (string, error) {
	var conversation strings.Builder
	for _, msg := range messages {
		if msg == nil {
			continue
		}

		content := extractMessageText(msg)
		if strings.TrimSpace(content) == "" {
			continue
		}

		conversation.WriteString(roleLabel(msg.Role))
		conversation.WriteString("：")
		conversation.WriteString(content)
		conversation.WriteString("\n")
	}

	if conversation.Len() == 0 {
		return "", nil
	}

	resp, err := compressModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(compressionPrompt),
		schema.UserMessage("请压缩下面的历史对话：\n" + conversation.String()),
	}, model.WithMaxTokens(300))
	if err != nil {
		return "", fmt.Errorf("compress conversation failed: %w", err)
	}

	return strings.TrimSpace(resp.Content), nil
}

func findCompressionEnd(messages []*schema.Message, keepRecentRounds int) int {
	if keepRecentRounds <= 0 {
		return len(messages)
	}

	userRoundsSeen := 0
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg == nil {
			continue
		}
		if msg.Role == schema.User {
			userRoundsSeen++
			if userRoundsSeen >= keepRecentRounds {
				return i
			}
		}
	}

	return 0
}

func countLeadingSystemMessages(messages []*schema.Message) int {
	for i, msg := range messages {
		if msg == nil || msg.Role != schema.System {
			return i
		}
	}
	return len(messages)
}

func estimateMessagesTokens(messages []*schema.Message) int {
	total := 0
	for _, msg := range messages {
		total += estimateMessageTokens(msg)
	}
	return total
}

func estimateMessageTokens(msg *schema.Message) int {
	if msg == nil {
		return 0
	}

	text := extractMessageText(msg)
	if text == "" {
		return 8
	}

	return estimateTextTokens(text) + 8
}

func extractMessageText(msg *schema.Message) string {
	if msg == nil {
		return ""
	}

	var parts []string
	if strings.TrimSpace(msg.Content) != "" {
		parts = append(parts, msg.Content)
	}
	if strings.TrimSpace(msg.ReasoningContent) != "" {
		parts = append(parts, msg.ReasoningContent)
	}
	for _, part := range msg.UserInputMultiContent {
		if part.Type == schema.ChatMessagePartTypeText && strings.TrimSpace(part.Text) != "" {
			parts = append(parts, part.Text)
		}
	}
	for _, part := range msg.AssistantGenMultiContent {
		if part.Type == schema.ChatMessagePartTypeText && strings.TrimSpace(part.Text) != "" {
			parts = append(parts, part.Text)
		}
	}
	for _, tc := range msg.ToolCalls {
		if strings.TrimSpace(tc.Function.Name) != "" {
			parts = append(parts, tc.Function.Name)
		}
		if strings.TrimSpace(tc.Function.Arguments) != "" {
			parts = append(parts, tc.Function.Arguments)
		}
	}
	if strings.TrimSpace(msg.ToolName) != "" {
		parts = append(parts, msg.ToolName)
	}
	if strings.TrimSpace(msg.ToolCallID) != "" {
		parts = append(parts, msg.ToolCallID)
	}

	return strings.Join(parts, "\n")
}

func estimateTextTokens(text string) int {
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}

	cjkCount := 0
	wordCount := 0
	inWord := false
	punctCount := 0

	for _, r := range text {
		switch {
		case isCJK(r):
			cjkCount++
			inWord = false
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if !inWord {
				wordCount++
				inWord = true
			}
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			punctCount++
			inWord = false
		default:
			inWord = false
		}
	}

	runeCount := utf8.RuneCountInString(text)
	estimated := cjkCount + wordCount + punctCount/4
	if estimated <= 0 {
		estimated = runeCount/2 + 1
	}
	if estimated < runeCount/6 {
		estimated = runeCount / 6
	}
	if estimated == 0 {
		estimated = 1
	}
	return estimated
}

func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r)
}

func roleLabel(role schema.RoleType) string {
	switch role {
	case schema.System:
		return "system"
	case schema.User:
		return "user"
	case schema.Assistant:
		return "assistant"
	case schema.Tool:
		return "tool"
	default:
		return string(role)
	}
}

func cloneMessages(messages []*schema.Message) []*schema.Message {
	if len(messages) == 0 {
		return nil
	}

	cloned := make([]*schema.Message, 0, len(messages))
	for _, msg := range messages {
		if msg == nil {
			cloned = append(cloned, nil)
			continue
		}
		msgCopy := *msg
		cloned = append(cloned, &msgCopy)
	}
	return cloned
}
