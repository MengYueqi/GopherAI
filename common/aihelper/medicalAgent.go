package aihelper

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const medicalRedFlagPromptPath = "common/tools/prompt/medical_red_flag_system.txt"

func loadMedicalRedFlagPrompt() (string, error) {
	data, err := os.ReadFile(medicalRedFlagPromptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", medicalRedFlagPromptPath, err)
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", fmt.Errorf("prompt file %s is empty", medicalRedFlagPromptPath)
	}
	return content, nil
}

func (o *OpenAIModel) MedicalAgentResp(ctx context.Context, description string) (*schema.Message, error) {
	// 构建医疗建议的响应逻辑
	g := compose.NewGraph[map[string]any, *schema.Message]()
	systemPrompt, err := loadMedicalRedFlagPrompt()
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	red_flag_pt := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.UserMessage("症状描述：{symptom}"),
	)

	_ = g.AddChatTemplateNode("nodeOfRedFlagPrompt", red_flag_pt)
	_ = g.AddChatModelNode("distinctRedFlag", o.llm, compose.WithNodeName("ChatModel"))
	// 对症状进行红旗分类
	_ = g.AddEdge(compose.START, "nodeOfRedFlagPrompt")
	_ = g.AddEdge("nodeOfRedFlagPrompt", "distinctRedFlag")
	_ = g.AddEdge("distinctRedFlag", compose.END)

	r, err := g.Compile(ctx)
	if err != nil {
		log.Printf("ERROR in compile the graph: %v \n", err)
		// panic(err)
	}

	in := map[string]any{
		"symptom": description,
	}
	ret, err := r.Invoke(ctx, in)
	if err != nil {
		log.Printf("ERROR in invoke the graph: %v \n", err)
		return nil, err
	}
	fmt.Println("invoke result: ", ret)
	return ret, nil
}
