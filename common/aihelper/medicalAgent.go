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
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const medicalRedFlagPromptPath = "common/tools/prompt/medical_red_flag_system.txt"
const medicalRedFlagNoticePath = "common/tools/prompt/medical_red_flag_notice.txt"
const medicalNoRedFlagNoticePath = "common/tools/prompt/medical_no_red_flag_notice.txt"
const medicalSummaryPromptPath = "common/tools/prompt/medical_summary_prompt.txt"
const myBaseURL = "http://localhost:8081/sse"
const flightBaseURL = "http://localhost:8082/sse"

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

func (o *OpenAIModel) MedicalAgentResp(ctx context.Context, description string, progressCb TravelPlanningProgressCallback) (*schema.Message, error) {
	g := compose.NewGraph[map[string]any, *schema.Message]()
	systemPrompt, err := loadPromptFile(medicalRedFlagPromptPath)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	summaryPrompt, err := loadPromptFile(medicalSummaryPromptPath)
	if err != nil {
		log.Printf("ERROR: %v\n", err)
		return nil, err
	}
	redFlagPrompt := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(systemPrompt),
		schema.UserMessage("旅行需求描述：{description}"),
	)
	summaryPromptTemplate := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(summaryPrompt),
		schema.UserMessage("规划内容：{content}"),
	)

	_ = g.AddChatTemplateNode("nodeOfPlanFeasibilityPrompt", redFlagPrompt, compose.WithNodeName("plan_feasibility_prompt"))
	_ = g.AddChatModelNode("distinctPlanFeasibility", o.llm, compose.WithNodeName("feasibility_check"))
	_ = g.AddLambdaNode("parsePlanDecisionJSON", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (ModelJudgment, error) {
		outputContent := input.Content
		log.Printf("Model output for plan feasibility classification: %s\n", outputContent)
		var mj ModelJudgment
		err = json.Unmarshal([]byte(outputContent), &mj)
		log.Printf("Parsed ModelJudgment: %+v\n", mj)
		return mj, nil
	}), compose.WithNodeName("parse_plan_decision"))

	dividePlanFeasibilityCondition := compose.NewGraphBranch(
		isRedFlag,
		map[string]bool{
			"plan_blocked_condition": true,
			"plan_allowed_condition": true,
		},
	)

	_ = g.AddLambdaNode("plan_blocked_condition", compose.InvokableLambda(func(ctx context.Context, input ModelJudgment) (res []*schema.Message, err error) {
		log.Printf("Final plan feasibility result (blocked): %+v\n", input)
		noticeTemplate, err := loadPromptFile(medicalRedFlagNoticePath)
		if err != nil {
			return nil, err
		}
		content := strings.ReplaceAll(noticeTemplate, "{description}", input.Description)
		content = strings.ReplaceAll(content, "{address}", input.Address)
		return []*schema.Message{{
			Role:    schema.Assistant,
			Content: content,
		}}, nil
	}), compose.WithNodeName("prepare_blocked_notice"))

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
	flightCli, err := initMCPClient(ctx, flightBaseURL)
	if err != nil {
		log.Printf("ERROR initializing flight MCP client: %v\n", err)
		return nil, err
	}
	flightTools, err := mcp.GetTools(ctx, &mcp.Config{Cli: flightCli})
	if err != nil {
		log.Printf("ERROR getting flight MCP tools: %v\n", err)
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
	flightPlanner, err := o.NewFlightAdvisorAgent(ctx, flightTools)
	if err != nil {
		log.Printf("ERROR creating flight planner: %v\n", err)
		return nil, err
	}
	attractionPlanner, err := o.NewAttractionHighlightsAgent(ctx, tools)
	if err != nil {
		log.Printf("ERROR creating attraction planner: %v\n", err)
		return nil, err
	}
	travelJSONFormatter, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TravelPlanJSONFormatter",
		Description: "将旅行规划摘要转换为结构化 JSON",
		Instruction: `你是旅行规划 JSON 转换助手。你会收到一份已经整理好的旅行方案摘要，请将其严格转换为一个合法 JSON 对象，不要输出 Markdown，不要输出解释，不要输出代码块。

必须输出以下结构：
{
  "mode": "plan",
  "overall_summary": "整体路线、节奏、适合人群与核心建议的简洁概括",
  "flight_price": {
    "summary": "机票价格与选择建议的总结",
    "currency": "币种，如 CNY",
    "price_range": "价格区间，如 1800-2600",
    "booking_tips": ["购票建议1", "购票建议2"],
    "raw_text": "机票规划原始摘要，尽量保留原信息"
  },
  "daily_plans": [
    {
      "day": 1,
      "title": "当天标题",
      "route": "当天路线顺序",
      "transport": "主要交通方式与衔接",
      "summary": "当天安排概述",
      "attractions": [
        {
          "name": "景点名称",
          "description": "景点介绍",
          "highlights": ["亮点1", "亮点2"],
          "images": [
            {
              "title": "图片标题或主题",
              "url": "图片直链",
              "source": "来源名称",
              "source_url": "来源页面链接"
            }
          ]
        }
      ],
      "tips": ["当天提示1", "当天提示2"]
    }
  ],
  "sources": ["来源1", "来源2"],
  "notice": "",
  "raw_text": ""
}

约束：
1. 所有字段都必须输出；没有信息时返回空字符串、空数组或合理默认值。
2. daily_plans 必须是数组。
3. 如果摘要中包含真实图片链接、来源链接或引用链接，必须原样保留，不要改写 URL。
4. 如果某个景点没有图片，则 images 返回空数组。
5. 如果行程按天组织，则每天都要保留当天景点及其图片信息。
6. 只输出 JSON 对象本身。`,
		Model: o.llm,
	})
	if err != nil {
		log.Printf("ERROR creating travel json formatter: %v\n", err)
		return nil, err
	}
	_ = g.AddLambdaNode("plan_blocked_deal", compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (res *schema.Message, err error) {
		log.Printf("Handling plan blocked case with request: %+v\n", input)
		return runAgentQuery(ctx, red_flag_agent, input[0].Content)
	}), compose.WithNodeName("requirements_feedback"))

	_ = g.AddLambdaNode("plan_allowed_condition", compose.InvokableLambda(func(ctx context.Context, input ModelJudgment) (res []*schema.Message, err error) {
		log.Printf("Final plan feasibility result (allowed): %+v\n", input)
		noticeTemplate, err := loadPromptFile(medicalNoRedFlagNoticePath)
		if err != nil {
			return nil, err
		}
		content := strings.ReplaceAll(noticeTemplate, "{description}", input.Description)
		content = strings.ReplaceAll(content, "{address}", input.Address)
		return []*schema.Message{{
			Role:    schema.Assistant,
			Content: content,
		}}, nil
	}), compose.WithNodeName("prepare_allowed_notice"))

	_ = g.AddLambdaNode("overall_route_planner", compose.InvokableLambda(func(ctx context.Context, input []*schema.Message) (res *schema.Message, err error) {
		log.Printf("Handling overall route planning with request: %+v\n", input)
		return runAgentQuery(ctx, overallRoutePlanner, input[0].Content)
	}), compose.WithNodeName("overall_route"))

	_ = g.AddLambdaNode("flight_planner", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (res map[string]any, err error) {
		log.Printf("Handling flight planning with request: %+v\n", input)
		msg, err := runAgentQuery(ctx, flightPlanner, input.Content)
		if err != nil {
			return nil, err
		}
		return map[string]any{"航班规划": msg.Content}, nil
	}), compose.WithNodeName("flight_planning"))

	_ = g.AddLambdaNode("attraction_planner", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (res map[string]any, err error) {
		log.Printf("Handling attraction highlights with request: %+v\n", input)
		msg, err := runAgentQuery(ctx, attractionPlanner, input.Content)
		if err != nil {
			return nil, err
		}
		return map[string]any{"重点景点规划": msg.Content}, nil
	}), compose.WithNodeName("attraction_planning"))

	_ = g.AddLambdaNode("summary_prompt_input", compose.InvokableLambda(func(ctx context.Context, input map[string]any) (map[string]any, error) {
		content := ""
		for k, v := range input {
			content += fmt.Sprintf("%s：%s\n", k, v)
		}
		log.Printf("Summary prompt input content: %s\n", content)
		return map[string]any{"content": content}, nil
	}), compose.WithNodeName("prepare_summary_prompt"))
	_ = g.AddLambdaNode("change_overall_output", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (map[string]any, error) {
		return map[string]any{"overall": input.Content}, nil
	}), compose.WithNodeName("capture_overall_output"))
	_ = g.AddChatTemplateNode("summary_prompt", summaryPromptTemplate, compose.WithNodeName("summary_prompt"))
	_ = g.AddChatModelNode("summary_model", o.llm, compose.WithNodeName("plan_summary"))
	_ = g.AddLambdaNode("travel_json_formatter", compose.InvokableLambda(func(ctx context.Context, input *schema.Message) (*schema.Message, error) {
		return runAgentQuery(ctx, travelJSONFormatter, input.Content)
	}), compose.WithNodeName("json_structuring"))

	_ = g.AddEdge(compose.START, "nodeOfPlanFeasibilityPrompt")
	_ = g.AddEdge("nodeOfPlanFeasibilityPrompt", "distinctPlanFeasibility")
	_ = g.AddEdge("distinctPlanFeasibility", "parsePlanDecisionJSON")
	_ = g.AddBranch("parsePlanDecisionJSON", dividePlanFeasibilityCondition)
	_ = g.AddEdge("plan_blocked_condition", "plan_blocked_deal")
	_ = g.AddEdge("plan_blocked_deal", compose.END)
	_ = g.AddEdge("plan_allowed_condition", "overall_route_planner")
	_ = g.AddEdge("overall_route_planner", "flight_planner")
	_ = g.AddEdge("overall_route_planner", "attraction_planner")
	_ = g.AddEdge("overall_route_planner", "change_overall_output")
	_ = g.AddEdge("flight_planner", "summary_prompt_input")
	_ = g.AddEdge("attraction_planner", "summary_prompt_input")
	_ = g.AddEdge("change_overall_output", "summary_prompt_input")
	_ = g.AddEdge("summary_prompt_input", "summary_prompt")
	_ = g.AddEdge("summary_prompt", "summary_model")
	_ = g.AddEdge("summary_model", "travel_json_formatter")
	_ = g.AddEdge("travel_json_formatter", compose.END)

	r, err := g.Compile(ctx)
	if err != nil {
		log.Printf("ERROR in compile the graph: %v \n", err)
		return nil, err
	}

	in := map[string]any{
		"description": description,
	}
	ret, err := r.Invoke(ctx, in, compose.WithCallbacks(buildTravelPlanningCallback(progressCb)))
	if err != nil {
		log.Printf("ERROR in invoke the graph: %v \n", err)
		return nil, err
	}
	return ret, nil
}

func buildTravelPlanningCallback(progressCb TravelPlanningProgressCallback) callbacks.Handler {
	stageMeta := map[string]TravelPlanningProgress{
		"feasibility_check":     {Stage: "feasibility_check", Label: "可行性评估", Detail: "正在判断需求是否足以开始规划。", Percent: 20},
		"requirements_feedback": {Stage: "requirements_feedback", Label: "需求补充建议", Detail: "正在生成补充信息与修改建议。", Percent: 100},
		"overall_route":         {Stage: "overall_route", Label: "总体路线设计", Detail: "正在规划整体路线与行程节奏。", Percent: 45},
		"flight_planning":       {Stage: "flight_planning", Label: "机票信息分析", Detail: "正在评估航班价格与购票建议。", Percent: 68},
		"attraction_planning":   {Stage: "attraction_planning", Label: "景点亮点规划", Detail: "正在补充每日景点和亮点内容。", Percent: 82},
		"plan_summary":          {Stage: "plan_summary", Label: "行程汇总成文", Detail: "正在汇总多阶段结果。", Percent: 94},
		"json_structuring":      {Stage: "json_structuring", Label: "结构化整理", Detail: "正在整理为前端可直接渲染的结构化结果。", Percent: 100},
	}

	send := func(status string, info *callbacks.RunInfo, detail string) context.Context {
		if progressCb == nil || info == nil {
			return context.Background()
		}
		meta, ok := stageMeta[info.Name]
		if !ok {
			return context.Background()
		}
		progressCb(TravelPlanningProgress{
			Stage:   meta.Stage,
			Label:   meta.Label,
			Status:  status,
			Detail:  detail,
			Percent: meta.Percent,
		})
		return context.Background()
	}

	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			meta, ok := stageMeta[info.Name]
			if !ok {
				return ctx
			}
			progressCb(TravelPlanningProgress{
				Stage:   meta.Stage,
				Label:   meta.Label,
				Status:  "running",
				Detail:  meta.Detail,
				Percent: max(1, meta.Percent-10),
			})
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if _, ok := stageMeta[info.Name]; !ok {
				return ctx
			}
			return send("completed", info, stageMeta[info.Name].Detail)
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			if _, ok := stageMeta[info.Name]; !ok {
				return ctx
			}
			if progressCb != nil {
				meta := stageMeta[info.Name]
				progressCb(TravelPlanningProgress{
					Stage:   meta.Stage,
					Label:   meta.Label,
					Status:  "failed",
					Detail:  err.Error(),
					Percent: meta.Percent,
				})
			}
			return ctx
		}).
		Build()
}

func isRedFlag(ctx context.Context, prevJ ModelJudgment) (string, error) {
	result := prevJ.IsRedFlag
	log.Printf("plan feasibility (blocked) result: %v", result)
	if result {
		return "plan_blocked_condition", nil
	}
	return "plan_allowed_condition", nil
}

func runAgentQuery(ctx context.Context, agent adk.Agent, query string) (*schema.Message, error) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: agent,
	})
	iter := runner.Query(ctx, query)
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
}
