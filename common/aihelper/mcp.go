package aihelper

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// 初始化 MCP 客户端
func initMCPClient(ctx context.Context, serverURL string) (*client.Client, error) {
	cli, err := client.NewSSEMCPClient(serverURL)
	if err != nil {
		return nil, err
	}
	err = cli.Start(ctx)
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "example-client",
		Version: "1.0.0",
	}
	_, err = cli.Initialize(ctx, initRequest)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
