package aihelper

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const medicalRedFlagPromptPath = "common/tools/prompt/medical_red_flag_system.txt"
const medicalRedFlagNoticePath = "common/tools/prompt/medical_red_flag_notice.txt"
const myBaseURL = "http://localhost:8081/sse"

type ModelJudgment struct {
	IsRedFlag bool   `json:"red_flag"`
	Symptoms  string `json:"symptoms"`
	Address   string `json:"address,omitempty"`
}

func loadPromptFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", path, err)
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", fmt.Errorf("prompt file %s is empty", path)
	}
	return content, nil
}

func (o *OpenAIModel) MedicalAgentResp(ctx context.Context, description string) (*schema.Message, error) {
	// 构建医疗建议的响应逻辑
	g := compose.NewGraph[map[string]any, *schema.Message]()
	systemPrompt, err := loadPromptFile(medicalRedFlagPromptPath)
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
	_ = g.AddLambdaNode("red_flag_condition", compose.InvokableLambda(func(ctx context.Context, input ModelJudgment) (res []*schema.Message, err error) {
		log.Printf("Final red flag classification result: %+v\n", input)
		// 构造 prompt，使用模型生成紧急医疗建议
		pt := fmt.Sprintf("目前初步判断为危险事件，请根据初步的判断和症状给出建议：%+v", input)
		return []*schema.Message{
			{
				Role:    schema.Assistant,
				Content: pt,
			},
		}, nil
	}))

	// // 构建处理红色情况提示
	cli, err := initMCPClient(ctx, myBaseURL)
	if err != nil {
		log.Printf("ERROR initializing MCP client: %v\n", err)
		return nil, err
	}
	tools, err := mcp.GetTools(ctx, &mcp.Config{Cli: cli})
	red_flag_agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "EmergencyMedicalExpert",
		Description: "紧急医疗事件处理专家",
		Instruction: `你是紧急医疗事件处理专家。根据用户描述，快速判断风险并给出紧急处置建议；必要时提示立即联系急救或就医。`,
		Model:       o.llm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
	})

	_ = g.AddLambdaNode("red_flag_deal", compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (res *schema.Message, err error) {
		log.Printf("Handling red flag case with symptoms: %+v\n", input)
		// 使用 agent 处理红旗情况
		runner := adk.NewRunner(ctx, adk.RunnerConfig{
			Agent: red_flag_agent,
		})
		iter := runner.Query(ctx, input[0].Content)
		var allmsg string
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				log.Fatal(event.Err)
			}
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				log.Fatal(err)
			}
			allmsg = msg.Content
			fmt.Printf("\nmessage:\n%v\n======", msg.Content)
		}
		return &schema.Message{Content: allmsg}, nil
	}))

	// _ = g.AddChatModelNode("red_flag_deal", o.llm, compose.WithNodeName("RedFlagChatModel"))

	_ = g.AddLambdaNode("no_red_flag_condition", compose.InvokableLambda(func(ctx context.Context, input ModelJudgment) (res *schema.Message, err error) {
		log.Printf("Final no red flag classification result: %+v\n", input)
		noticeTemplate, err := loadPromptFile(medicalRedFlagNoticePath)
		if err != nil {
			return nil, err
		}
		content := strings.ReplaceAll(noticeTemplate, "{symptoms}", input.Symptoms)
		return &schema.Message{Content: content}, nil
	}))

	// 对症状进行红旗分类
	_ = g.AddEdge(compose.START, "nodeOfRedFlagPrompt")
	_ = g.AddEdge("nodeOfRedFlagPrompt", "distinctRedFlag")
	_ = g.AddEdge("distinctRedFlag", "getJSON")
	_ = g.AddBranch("getJSON", divide_red_flag_condition)
	// 红旗事件处理
	_ = g.AddEdge("red_flag_condition", "red_flag_deal")
	_ = g.AddEdge("red_flag_deal", compose.END)
	// 非红旗事件处理
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
