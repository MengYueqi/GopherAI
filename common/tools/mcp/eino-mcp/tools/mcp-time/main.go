package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	transport    string
	serverlisten string
)

var (
	googleAPIKey       string
	googleSearchEngine string
)

var (
	serverName    = "github.com/songjiayang/current-time"
	serverVersion = "1.0.0"
)

func init() {
	googleAPIKey = os.Getenv("GOOGLE_API_KEY")
	googleSearchEngine = os.Getenv("GOOGLE_SEARCH_ENGINE_ID")
	flag.StringVar(&transport, "transport", "sse", "The transport to use, should be \"stdio\" or \"sse\"")
	flag.StringVar(&serverlisten, "server_listen", "localhost:8080", "The sse server listen address")
	flag.Parse()
}

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		serverName,
		serverVersion,
	)
	// Add tool
	tool := mcp.NewTool("current time",
		mcp.WithDescription("Get current time with timezone, Asia/Shanghai is default"),
		mcp.WithString("timezone",
			mcp.Required(),
			mcp.Description("current time timezone"),
		),
	)
	// Add tool handler
	s.AddTool(tool, currentTimeHandler)

	// Add google search tool if api key and search engine id are provided
	googleTool := mcp.NewTool(
		"google_search",
		mcp.WithDescription("Search information via Google Custom Search API"),
		mcp.WithString(
			"query",
			mcp.Required(),
			mcp.Description("search query"),
		),
	)
	if googleAPIKey == "" || googleSearchEngine == "" {
		log.Println("WARNING: GOOGLE_API_KEY or GOOGLE_SEARCH_ENGINE_ID is not set, google_search tool will not be available")
	}
	s.AddTool(googleTool, googleSearchHandler)

	// Only check for "sse" since stdio is the default
	if transport == "sse" {
		serverUrl := "http://" + serverlisten
		sseServer := server.NewSSEServer(s, server.WithBaseURL(serverUrl))
		log.Printf("SSE server listening on %s", serverlisten)
		if err := sseServer.Start(serverlisten); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(s); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

func currentTimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("arguments must be an object")
	}
	timezone, ok := args["timezone"].(string)
	if !ok {
		return nil, errors.New("timezone must be a string")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("parse timezone with error: %v", err)
	}
	return mcp.NewToolResultText(fmt.Sprintf(`current time is %s`, time.Now().In(loc))), nil
}

func googleSearchHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {

	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("arguments must be an object")
	}
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, errors.New("query must be a non-empty string")
	}

	result, err := googleSearch(ctx, query)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

func googleSearch(ctx context.Context, query string) (string, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://www.googleapis.com/customsearch/v1",
		nil,
	)

	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("key", googleAPIKey)
	q.Set("cx", googleSearchEngine)
	q.Set("q", query)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("google api error: %s", body)
	}

	// ===== parse response =====
	var data struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.Items) == 0 {
		return "no search results", nil
	}

	var sb strings.Builder
	for i, item := range data.Items {
		sb.WriteString(fmt.Sprintf(
			"%d. %s\n%s\n%s\n\n",
			i+1,
			item.Title,
			item.Snippet,
			item.Link,
		))
	}

	return sb.String(), nil
}
