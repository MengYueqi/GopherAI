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
	"strconv"
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
	unsplashAccessKey string
)

var (
	serverName    = "github.com/songjiayang/mcp-photo"
	serverVersion = "1.0.0"
)

func init() {
	unsplashAccessKey = os.Getenv("UNSPLASH_ACCESS_KEY")
	flag.StringVar(&transport, "transport", "sse", "The transport to use, should be \"stdio\" or \"sse\"")
	flag.StringVar(&serverlisten, "server_listen", "localhost:8080", "The sse server listen address")
	flag.Parse()
}

func main() {
	s := server.NewMCPServer(serverName, serverVersion)

	tool := mcp.NewTool(
		"search_photos",
		mcp.WithDescription("Search photos via Unsplash"),
		mcp.WithString(
			"query",
			mcp.Required(),
			mcp.Description("search keywords"),
		),
		mcp.WithNumber(
			"page",
			mcp.Description("page number to retrieve, default is 1"),
		),
		mcp.WithNumber(
			"per_page",
			mcp.Description("number of items per page, default is 10, max is 30"),
		),
		mcp.WithString(
			"order_by",
			mcp.Description("sort order, valid values are latest and relevant"),
		),
		mcp.WithString(
			"collections",
			mcp.Description("comma-separated collection IDs"),
		),
		mcp.WithString(
			"content_filter",
			mcp.Description("content safety filter, valid values are low and high"),
		),
		mcp.WithString(
			"color",
			mcp.Description("color filter"),
		),
		mcp.WithString(
			"orientation",
			mcp.Description("orientation filter, valid values are landscape, portrait, squarish"),
		),
	)
	if unsplashAccessKey == "" {
		log.Println("WARNING: UNSPLASH_ACCESS_KEY is not set, search_photos tool will not be available")
	}
	s.AddTool(tool, searchPhotosHandler)

	if transport == "sse" {
		serverURL := "http://" + serverlisten
		sseServer := server.NewSSEServer(s, server.WithBaseURL(serverURL))
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

func searchPhotosHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	if unsplashAccessKey == "" {
		return nil, errors.New("UNSPLASH_ACCESS_KEY is not set")
	}

	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("arguments must be an object")
	}

	params, err := parseSearchPhotosArgs(args)
	if err != nil {
		return nil, err
	}

	log.Printf(
		"search_photos params: query=%q page=%d per_page=%d order_by=%s collections=%s content_filter=%s color=%s orientation=%s",
		params.Query,
		params.Page,
		params.PerPage,
		params.OrderBy,
		params.Collections,
		params.ContentFilter,
		params.Color,
		params.Orientation,
	)

	result, err := searchPhotos(ctx, params)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

type photoSearchParams struct {
	Query         string
	Page          int
	PerPage       int
	OrderBy       string
	Collections   string
	ContentFilter string
	Color         string
	Orientation   string
}

func parseSearchPhotosArgs(args map[string]any) (photoSearchParams, error) {
	query, ok := args["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return photoSearchParams{}, errors.New("query must be a non-empty string")
	}

	params := photoSearchParams{
		Query:         strings.TrimSpace(query),
		Page:          1,
		PerPage:       10,
		OrderBy:       "relevant",
		ContentFilter: "low",
	}

	if v, ok, err := getOptionalInt(args, "page"); err != nil {
		return photoSearchParams{}, err
	} else if ok {
		if v < 1 {
			return photoSearchParams{}, errors.New("page must be greater than 0")
		}
		params.Page = v
	}

	if v, ok, err := getOptionalInt(args, "per_page"); err != nil {
		return photoSearchParams{}, err
	} else if ok {
		if v < 1 || v > 30 {
			return photoSearchParams{}, errors.New("per_page must be between 1 and 30")
		}
		params.PerPage = v
	}

	if value, ok := args["order_by"]; ok {
		orderBy, ok := value.(string)
		if !ok {
			return photoSearchParams{}, errors.New("order_by must be a string")
		}
		orderBy = strings.TrimSpace(orderBy)
		if orderBy != "latest" && orderBy != "relevant" {
			return photoSearchParams{}, errors.New("order_by must be one of: latest, relevant")
		}
		params.OrderBy = orderBy
	}

	if value, ok := args["collections"]; ok {
		collections, ok := value.(string)
		if !ok {
			return photoSearchParams{}, errors.New("collections must be a string")
		}
		params.Collections = strings.TrimSpace(collections)
	}

	if value, ok := args["content_filter"]; ok {
		contentFilter, ok := value.(string)
		if !ok {
			return photoSearchParams{}, errors.New("content_filter must be a string")
		}
		contentFilter = strings.TrimSpace(contentFilter)
		if contentFilter != "low" && contentFilter != "high" {
			return photoSearchParams{}, errors.New("content_filter must be one of: low, high")
		}
		params.ContentFilter = contentFilter
	}

	if value, ok := args["color"]; ok {
		color, ok := value.(string)
		if !ok {
			return photoSearchParams{}, errors.New("color must be a string")
		}
		color = strings.TrimSpace(color)
		if color != "" && !isAllowedValue(color, map[string]struct{}{
			"black_and_white": {},
			"black":           {},
			"white":           {},
			"yellow":          {},
			"orange":          {},
			"red":             {},
			"purple":          {},
			"magenta":         {},
			"green":           {},
			"teal":            {},
			"blue":            {},
		}) {
			return photoSearchParams{}, errors.New("color has an invalid value")
		}
		params.Color = color
	}

	if value, ok := args["orientation"]; ok {
		orientation, ok := value.(string)
		if !ok {
			return photoSearchParams{}, errors.New("orientation must be a string")
		}
		orientation = strings.TrimSpace(orientation)
		if orientation != "" && !isAllowedValue(orientation, map[string]struct{}{
			"landscape": {},
			"portrait":  {},
			"squarish":  {},
		}) {
			return photoSearchParams{}, errors.New("orientation must be one of: landscape, portrait, squarish")
		}
		params.Orientation = orientation
	}

	return params, nil
}

func getOptionalInt(args map[string]any, key string) (int, bool, error) {
	value, ok := args[key]
	if !ok {
		return 0, false, nil
	}

	switch v := value.(type) {
	case float64:
		if v != float64(int(v)) {
			return 0, true, fmt.Errorf("%s must be an integer", key)
		}
		return int(v), true, nil
	case int:
		return v, true, nil
	case int32:
		return int(v), true, nil
	case int64:
		return int(v), true, nil
	case json.Number:
		n, err := strconv.Atoi(v.String())
		if err != nil {
			return 0, true, fmt.Errorf("%s must be an integer", key)
		}
		return n, true, nil
	default:
		return 0, true, fmt.Errorf("%s must be a number", key)
	}
}

func isAllowedValue(value string, allowed map[string]struct{}) bool {
	_, ok := allowed[value]
	return ok
}

func searchPhotos(ctx context.Context, params photoSearchParams) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.unsplash.com/search/photos",
		nil,
	)
	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("query", params.Query)
	q.Set("page", strconv.Itoa(params.Page))
	q.Set("per_page", strconv.Itoa(params.PerPage))
	q.Set("order_by", params.OrderBy)
	q.Set("content_filter", params.ContentFilter)
	if params.Collections != "" {
		q.Set("collections", params.Collections)
	}
	if params.Color != "" {
		q.Set("color", params.Color)
	}
	if params.Orientation != "" {
		q.Set("orientation", params.Orientation)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", "Client-ID "+unsplashAccessKey)
	req.Header.Set("Accept-Version", "v1")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unsplash api error: %s", strings.TrimSpace(string(body)))
	}

	var data struct {
		Total      int `json:"total"`
		TotalPages int `json:"total_pages"`
		Results    []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
			AltDesc     string `json:"alt_description"`
			URLs        struct {
				Regular string `json:"regular"`
			} `json:"urls"`
			User struct {
				Name     string `json:"name"`
				Username string `json:"username"`
			} `json:"user"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.Results) == 0 {
		return fmt.Sprintf("no photo results for query %q", params.Query), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"Found %d photos across %d pages for query %q.\n\n",
		data.Total,
		data.TotalPages,
		params.Query,
	))

	for i, item := range data.Results {
		description := strings.TrimSpace(item.Description)
		if description == "" {
			description = strings.TrimSpace(item.AltDesc)
		}
		if description == "" {
			description = "No description"
		}

		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, description))
		sb.WriteString(fmt.Sprintf("ID: %s\n", item.ID))
		sb.WriteString(fmt.Sprintf("Photographer: %s (@%s)\n", item.User.Name, item.User.Username))
		sb.WriteString(fmt.Sprintf("Regular: %s\n", item.URLs.Regular))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
