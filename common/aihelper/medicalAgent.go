package aihelper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const medicalRedFlagPromptPath = "common/tools/prompt/medical_red_flag_system.txt"

type ModelJudgment struct {
	IsRedFlag bool   `json:"red_flag"`
	Symptoms  string `json:"symptoms"`
}

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
	_ = g.AddLambdaNode("getJSON", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (ModelJudgment, error) {
		// 解析模型输出，提取红旗信息
		outputContent := input.Content
		log.Printf("Model output for red flag classification: %s\n", outputContent)
		// 根据输入的 JSON 字符串，提取 is_red_flag 字段
		var mj ModelJudgment
		err = json.Unmarshal([]byte(outputContent), &mj)
		log.Printf("Parsed ModelJudgment: %+v\n", mj)
		return mj, nil
	}))

	divide_red_flag_condition := compose.NewGraphBranch(
		isRedFlag,
		map[string]bool{
			"red_flag_condition":    true,
			"no_red_flag_condition": true,
		},
	)

	// 占位 red_flag_condition 和 no_red_flag_condition 节点
	_ = g.AddLambdaNode("red_flag_condition", compose.InvokableLambda(func(ctx context.Context, input ModelJudgment) (res *schema.Message, err error) {
		log.Printf("Final red flag classification result: %+v\n", input)
		return &schema.Message{Content: "紧急情况，请立即就医"}, nil
	}))
	_ = g.AddLambdaNode("no_red_flag_condition", compose.InvokableLambda(func(ctx context.Context, input ModelJudgment) (res *schema.Message, err error) {
		log.Printf("Final no red flag classification result: %+v\n", input)
		return &schema.Message{Content: "非紧急情况，请预约医生进行进一步检查"}, nil
	}))

	// 对症状进行红旗分类
	_ = g.AddEdge(compose.START, "nodeOfRedFlagPrompt")
	_ = g.AddEdge("nodeOfRedFlagPrompt", "distinctRedFlag")
	_ = g.AddEdge("distinctRedFlag", "getJSON")
	_ = g.AddBranch("getJSON", divide_red_flag_condition)
	_ = g.AddEdge("red_flag_condition", compose.END)
	_ = g.AddEdge("no_red_flag_condition", compose.END)

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

func isRedFlag(ctx context.Context, prevJ ModelJudgment) (string, error) {
	result := prevJ.IsRedFlag
	log.Printf("isRedFlag result: %v", result)
	if result {
		return "red_flag_condition", nil
	} else {
		return "no_red_flag_condition", nil
	}
}
