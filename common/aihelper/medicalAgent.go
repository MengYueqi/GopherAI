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
const medicalNoRedFlagNoticePath = "common/tools/prompt/medical_no_red_flag_notice.txt"
const medicalSummaryPromptPath = "common/tools/prompt/medical_summary_prompt.txt"
const myBaseURL = "http://localhost:8081/sse"

type ModelJudgment struct {
	IsRedFlag   bool   `json:"red_flag"`
	Description string `json:"description"`
	Address     string `json:"address,omitempty"`
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
	// 构建旅游路径规划的响应逻辑
	g := compose.NewGraph[map[string]any, *schema.Message]()
	systemPrompt, err := loadPromptFile(medicalRedFlagPromptPath)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	red_flag_pt := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.UserMessage("旅行需求描述：{description}"),
	)
	summaryPrompt, err := loadPromptFile(medicalSummaryPromptPath)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	summary_pt := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(summaryPrompt),
		schema.UserMessage("规划内容：{content}"),
	)

	_ = g.AddChatTemplateNode("nodeOfPlanFeasibilityPrompt", red_flag_pt)
	_ = g.AddChatModelNode("distinctPlanFeasibility", o.llm, compose.WithNodeName("ChatModel"))
	_ = g.AddLambdaNode("parsePlanDecisionJSON", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (ModelJudgment, error) {
		// 解析模型输出，提取可行性判断
		outputContent := input.Content
		log.Printf("Model output for plan feasibility classification: %s\n", outputContent)
		// 根据输入的 JSON 字符串，提取 red_flag 字段
		var mj ModelJudgment
		err = json.Unmarshal([]byte(outputContent), &mj)
		log.Printf("Parsed ModelJudgment: %+v\n", mj)
		return mj, nil
	}))

	divide_plan_feasibility_condition := compose.NewGraphBranch(
		isRedFlag,
		map[string]bool{
			"plan_blocked_condition": true,
			"plan_allowed_condition": true,
		},
	)

	// 占位 plan_blocked_condition 和 plan_allowed_condition 节点
	_ = g.AddLambdaNode("plan_blocked_condition", compose.InvokableLambda(func(ctx context.Context, input ModelJudgment) (res []*schema.Message, err error) {
		log.Printf("Final plan feasibility result (blocked): %+v\n", input)
		noticeTemplate, err := loadPromptFile(medicalRedFlagNoticePath)
		if err != nil {
			return nil, err
		}
		content := strings.ReplaceAll(noticeTemplate, "{description}", input.Description)
		content = strings.ReplaceAll(content, "{address}", input.Address)
		return []*schema.Message{
			{
				Role:    schema.Assistant,
				Content: content,
			},
		}, nil
	}))

	// // 构建旅游路径规划的 agent
	cli, err := initMCPClient(ctx, myBaseURL)
	if err != nil {
		log.Printf("ERROR initializing MCP client: %v\n", err)
		return nil, err
	}
	tools, err := mcp.GetTools(ctx, &mcp.Config{Cli: cli})
	if err != nil {
		log.Printf("ERROR getting MCP tools: %v\n", err)
		return nil, err
	}
	red_flag_agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TravelFeasibilityAdvisor",
		Description: "旅游行程可行性评估与需求完善助手",
		Instruction: `你是旅游行程可行性评估助手。根据用户现有描述指出无法规划的原因，提出具体修改建议，并给出清晰的补充信息清单与追问问题。`,
		Model:       o.llm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
	})
	if err != nil {
		log.Printf("ERROR creating red flag agent: %v\n", err)
		return nil, err
	}
	// no_red_flag_agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
	// 	Name:        "TravelRoutePlanner",
	// 	Description: "旅游路径规划与行程设计助手",
	// 	Instruction: `你是旅游路径规划助手。根据用户需求给出清晰、可执行的行程路线规划，包含交通衔接、时间分配与预算重点，并提供备选方案。`,
	// 	Model:       o.llm,
	// 	ToolsConfig: adk.ToolsConfig{
	// 		ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
	// 	},
	// })

	// no_red_flag_agent, err := o.NewTravelGuideRecommendationAgent(ctx, tools)

	overallRoutePlanner, err := o.NewOverallRoutePlannerAgent(ctx, tools)
	if err != nil {
		log.Printf("ERROR creating overall route planner: %v\n", err)
		return nil, err
	}
	flightPlanner, err := o.NewFlightAdvisorAgent(ctx, tools)
	if err != nil {
		log.Printf("ERROR creating flight planner: %v\n", err)
		return nil, err
	}
	attractionPlanner, err := o.NewAttractionHighlightsAgent(ctx, tools)
	if err != nil {
		log.Printf("ERROR creating attraction planner: %v\n", err)
		return nil, err
	}

	_ = g.AddLambdaNode("plan_blocked_deal", compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (res *schema.Message, err error) {
		log.Printf("Handling plan blocked case with request: %+v\n", input)
		// 使用 agent 处理不可规划情况
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
			log.Printf("\nmessage:\n%+v\n======", msg)
		}
		return &schema.Message{Content: allmsg}, nil
	}))

	// _ = g.AddChatModelNode("red_flag_deal", o.llm, compose.WithNodeName("RedFlagChatModel"))

	_ = g.AddLambdaNode("plan_allowed_condition", compose.InvokableLambda(func(ctx context.Context, input ModelJudgment) (res []*schema.Message, err error) {
		log.Printf("Final plan feasibility result (allowed): %+v\n", input)
		noticeTemplate, err := loadPromptFile(medicalNoRedFlagNoticePath)
		if err != nil {
			return nil, err
		}
		content := strings.ReplaceAll(noticeTemplate, "{description}", input.Description)
		content = strings.ReplaceAll(content, "{address}", input.Address)
		return []*schema.Message{
			{
				Role:    schema.Assistant,
				Content: content,
			},
		}, nil
	}))

	// 绑定 overallRoutePlanner agent 节点
	_ = g.AddLambdaNode("overall_route_planner", compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (res *schema.Message, err error) {
		log.Printf("Handling overall route planning with request: %+v\n", input)
		runner := adk.NewRunner(ctx, adk.RunnerConfig{
			Agent: overallRoutePlanner,
		})
		iter := runner.Query(ctx, input[0].Content)
		var allmsg string
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				return nil, event.Err
			}
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				return nil, err
			}
			allmsg = msg.Content
		}
		return &schema.Message{Content: allmsg}, nil
	}))

	// 绑定 flightPlanner agent 节点
	_ = g.AddLambdaNode("flight_planner", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (res map[string]any, err error) {
		log.Printf("Handling flight planning with request: %+v\n", input)
		runner := adk.NewRunner(ctx, adk.RunnerConfig{
			Agent: flightPlanner,
		})
		iter := runner.Query(ctx, input.Content)
		var allmsg string
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				return nil, event.Err
			}
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				return nil, err
			}
			allmsg = msg.Content
		}
		return map[string]any{"航班规划": allmsg}, nil
	}))

	// 绑定 attractionPlanner agent 节点
	_ = g.AddLambdaNode("attraction_planner", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (res map[string]any, err error) {
		log.Printf("Handling attraction highlights with request: %+v\n", input)
		runner := adk.NewRunner(ctx, adk.RunnerConfig{
			Agent: attractionPlanner,
		})
		iter := runner.Query(ctx, input.Content)
		var allmsg string
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				return nil, event.Err
			}
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				return nil, err
			}
			allmsg = msg.Content
		}
		return map[string]any{"重点景点规划": allmsg}, nil
	}))
	_ = g.AddLambdaNode("summary_prompt_input", compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		// 整合各部分内容到一个字符串
		content := ""
		for k, v := range input {
			content += fmt.Sprintf("%s：%s\n", k, v)
		}
		log.Printf("Summary prompt input content: %s\n", content)
		return map[string]any{"content": content}, nil
	}))
	_ = g.AddLambdaNode("change_overall_output", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (map[string]any, error) {
		return map[string]any{"overall": input.Content}, nil
	}))
	_ = g.AddChatTemplateNode("summary_prompt", summary_pt)
	_ = g.AddChatModelNode("summary_model", o.llm, compose.WithNodeName("ChatModel"))

	// 对旅行需求进行可行性分类
	_ = g.AddEdge(compose.START, "nodeOfPlanFeasibilityPrompt")
	_ = g.AddEdge("nodeOfPlanFeasibilityPrompt", "distinctPlanFeasibility")
	_ = g.AddEdge("distinctPlanFeasibility", "parsePlanDecisionJSON")
	_ = g.AddBranch("parsePlanDecisionJSON", divide_plan_feasibility_condition)
	// 不可规划情况处理
	_ = g.AddEdge("plan_blocked_condition", "plan_blocked_deal")
	_ = g.AddEdge("plan_blocked_deal", compose.END)
	// 可规划情况处理
	_ = g.AddEdge("plan_allowed_condition", "overall_route_planner")
	_ = g.AddEdge("overall_route_planner", "flight_planner")
	_ = g.AddEdge("overall_route_planner", "attraction_planner")
	_ = g.AddEdge("overall_route_planner", "change_overall_output")
	_ = g.AddEdge("flight_planner", "summary_prompt_input")
	_ = g.AddEdge("attraction_planner", "summary_prompt_input")
	_ = g.AddEdge("change_overall_output", "summary_prompt_input")
	_ = g.AddEdge("summary_prompt_input", "summary_prompt")
	_ = g.AddEdge("summary_prompt", "summary_model")
	_ = g.AddEdge("summary_model", compose.END)

	r, err := g.Compile(ctx)
	if err != nil {
		log.Printf("ERROR in compile the graph: %v \n", err)
		// panic(err)
	}

	in := map[string]any{
		"description": description,
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
	log.Printf("plan feasibility (blocked) result: %v", result)
	if result {
		return "plan_blocked_condition", nil
	} else {
		return "plan_allowed_condition", nil
	}
}
