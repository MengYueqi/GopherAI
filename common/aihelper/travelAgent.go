package aihelper

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// 总体路线构建 Agent
func (o *OpenAIModel) NewOverallRoutePlannerAgent(ctx context.Context, tools []tool.BaseTool) (adk.Agent, error) {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "OverallRoutePlanner",
		Description: "构建总体旅行路线与行程节奏",
		Instruction: "你是总体路线规划专家。基于用户需求给出完整的路线框架、城市/区域顺序、交通方式选择和节奏建议，确保可执行且逻辑清晰。",
		Model:       o.llm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
		MaxIterations: 50,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// 机票推荐及价格评估 Agent
func (o *OpenAIModel) NewFlightAdvisorAgent(ctx context.Context, tools []tool.BaseTool) (adk.Agent, error) {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "FlightAdvisor",
		Description: "推荐机票选择并评估价格区间",
		Instruction: "你是机票推荐与价格评估助手。根据行程时间与出发/到达城市，给出航班选择建议、价格区间判断、购票时机与省钱策略，避免虚构具体航班号。",
		Model:       o.llm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
		MaxIterations: 50,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// 重要景点介绍生成 Agent
func (o *OpenAIModel) NewAttractionHighlightsAgent(ctx context.Context, tools []tool.BaseTool) (adk.Agent, error) {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "DailyItineraryBuilder",
		Description: "生成重要景点介绍",
		Instruction: "你是重要景点介绍助手。根据目的地与用户偏好，精选关键景点在特定的部分进行介绍。介绍要尽量详细，包含历史背景、文化意义和独特体验，帮助用户了解景点亮点。",
		Model:       o.llm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
		MaxIterations: 50,
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

// // 旅游攻略推荐 SequentialAgent
// func (o *OpenAIModel) NewTravelGuideRecommendationAgent(ctx context.Context, tools []tool.BaseTool) (adk.Agent, error) {
// 	// TODO: 增加携程旅行等获取票价的 tools
// 	// TODO: 增加每个 agent 不同的 tools 配置
// 	analyzer, err := o.NewOverallRoutePlannerAgent(ctx, tools)
// 	if err != nil {
// 		return nil, err
// 	}
// 	summarizer, err := o.NewFlightAdvisorAgent(ctx, tools)
// 	if err != nil {
// 		return nil, err
// 	}
// 	generator, err := o.NewAttractionHighlightsAgent(ctx, tools)
// 	if err != nil {
// 		return nil, err
// 	}

// 	sequentialAgent, err := adk.NewSequentialAgent(ctx, &adk.SequentialAgentConfig{
// 		Name:        "TravelGuideRecommendationPipeline",
// 		Description: "你需要合理规划多个 agent 的执行顺序，以生成完整的旅游攻略推荐。攻略包括路线规划、机票建议和景点介绍等方面内容。",
// 		SubAgents:   []adk.Agent{analyzer, summarizer, generator},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return sequentialAgent, nil
// }

// func main() {
// 	ctx := context.Background()

// 	o, err := NewOpenAIModel(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	sequentialAgent, err := o.NewTravelGuideRecommendationAgent(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// 创建 Runner
// 	runner := adk.NewRunner(ctx, adk.RunnerConfig{
// 		Agent: sequentialAgent,
// 	})

// 	// 执行旅游攻略推荐流程
// 	input := "我计划从上海出发去日本东京和京都，6天，预算1.5万人民币，希望节奏适中，想体验当地美食和文化。"

// 	fmt.Println("开始执行文档处理流水线...")
// 	iter := runner.Query(ctx, input)

// 	stepCount := 1
// 	for {
// 		event, ok := iter.Next()
// 		if !ok {
// 			break
// 		}

// 		if event.Err != nil {
// 			log.Fatal(event.Err)
// 		}

// 		if event.Output != nil && event.Output.MessageOutput != nil {
// 			fmt.Printf("\n=== 步骤 %d: %s ===\n", stepCount, event.AgentName)
// 			fmt.Printf("%s\n", event.Output.MessageOutput.Message.Content)
// 			stepCount++
// 		}
// 	}

// 	fmt.Println("\n文档处理流水线执行完成！")
// }
