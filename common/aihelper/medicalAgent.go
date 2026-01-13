package aihelper

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func (o *OpenAIModel) MedicalAgentResp(ctx context.Context, description string) (*schema.Message, error) {
	// 构建医疗建议的响应逻辑
	g := compose.NewGraph[map[string]any, *schema.Message]()
	pt := prompt.FromMessages(
		schema.FString,
		schema.UserMessage("请解释 {symptom} 是什么原因?"),
	)

	// 添加症状描述到提示模板，构建推理逻辑
	_ = g.AddChatTemplateNode("nodeOfPrompt", pt)
	_ = g.AddChatModelNode("nodeOfModel", o.llm, compose.WithNodeName("ChatModel"))
	_ = g.AddEdge(compose.START, "nodeOfPrompt")
	_ = g.AddEdge("nodeOfPrompt", "nodeOfModel")
	_ = g.AddEdge("nodeOfModel", compose.END)

	r, err := g.Compile(ctx)
	if err != nil {
		panic(err)
	}

	in := map[string]any{
		"symptom": description,
	}
	ret, err := r.Invoke(ctx, in)
	fmt.Println("invoke result: ", ret)
	return ret, nil
}
