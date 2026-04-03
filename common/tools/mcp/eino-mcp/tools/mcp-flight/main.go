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
	serpAPIKey string
)

var (
	serverName    = "github.com/songjiayang/current-time"
	serverVersion = "1.0.0"
)

func init() {
	serpAPIKey = os.Getenv("SERPAPI_API_KEY")
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

	// Add google flights tool if api key is provided
	googleFlightsTool := mcp.NewTool(
		"google_flights",
		mcp.WithDescription("Fetch flight info via SerpAPI Google Flights"),
		mcp.WithString(
			"departure_id",
			mcp.Required(),
			mcp.Description("departure airport IATA code"),
		),
		mcp.WithString(
			"arrival_id",
			mcp.Required(),
			mcp.Description("arrival airport IATA code"),
		),
		mcp.WithString(
			"outbound_date",
			mcp.Required(),
			mcp.Description("outbound date in YYYY-MM-DD, Please after the date of today: "+time.Now().Format("2006-01-02")),
		),
		mcp.WithString(
			"currency",
			mcp.Description("currency code, e.g. CNY"),
		),
		mcp.WithString(
			"type",
			mcp.Description("only support 2 for one way"),
		),
	)
	if serpAPIKey == "" {
		log.Println("WARNING: SERPAPI_API_KEY is not set, google_flights tool will not be available")
	}
	s.AddTool(googleFlightsTool, googleFlightsHandler)

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

func googleFlightsHandler(
	ctx context.Context,
	request mcp.CallToolRequest,
) (*mcp.CallToolResult, error) {
	if serpAPIKey == "" {
		return nil, errors.New("SERPAPI_API_KEY is not set")
	}

	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("arguments must be an object")
	}
	departureID, ok := args["departure_id"].(string)
	if !ok || departureID == "" {
		return nil, errors.New("departure_id must be a non-empty string")
	}
	arrivalID, ok := args["arrival_id"].(string)
	if !ok || arrivalID == "" {
		return nil, errors.New("arrival_id must be a non-empty string")
	}
	outboundDate, ok := args["outbound_date"].(string)
	if !ok || outboundDate == "" {
		return nil, errors.New("outbound_date must be a non-empty string")
	}

	currency, _ := args["currency"].(string)
	flightType, _ := args["type"].(string)
	if flightType == "" {
		flightType = "2"
	}

	log.Printf(
		"google_flights params: departure_id=%s arrival_id=%s outbound_date=%s currency=%s type=%s",
		departureID,
		arrivalID,
		outboundDate,
		currency,
		flightType,
	)

	result, err := googleFlights(ctx, departureID, arrivalID, outboundDate, currency, flightType)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

func googleFlights(
	ctx context.Context,
	departureID string,
	arrivalID string,
	outboundDate string,
	currency string,
	flightType string,
) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://serpapi.com/search",
		nil,
	)

	if err != nil {
		return "", err
	}

	q := req.URL.Query()
	q.Set("engine", "google_flights")
	q.Set("departure_id", departureID)
	q.Set("arrival_id", arrivalID)
	q.Set("outbound_date", outboundDate)
	if currency != "" {
		q.Set("currency", currency)
	}
	if flightType != "" {
		q.Set("type", flightType)
	}
	q.Set("api_key", serpAPIKey)
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

	var data struct {
		BestFlights []struct {
			Flights []struct {
				DepartureAirport struct {
					Name string `json:"name"`
					ID   string `json:"id"`
					Time string `json:"time"`
				} `json:"departure_airport"`
				ArrivalAirport struct {
					Name string `json:"name"`
					ID   string `json:"id"`
					Time string `json:"time"`
				} `json:"arrival_airport"`
				Duration     int      `json:"duration"`
				Airline      string   `json:"airline"`
				FlightNumber string   `json:"flight_number"`
				Extensions   []string `json:"extensions"`
			} `json:"flights"`
			Layovers []struct {
				Duration int    `json:"duration"`
				Name     string `json:"name"`
				ID       string `json:"id"`
			} `json:"layovers"`
			TotalDuration int      `json:"total_duration"`
			Price         int      `json:"price"`
			Type          string   `json:"type"`
			Extensions    []string `json:"extensions"`
		} `json:"best_flights"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.BestFlights) == 0 {
		return "no flight results", nil
	}

	var sb strings.Builder
	maxResults := 2
	if len(data.BestFlights) < maxResults {
		maxResults = len(data.BestFlights)
	}
	for i := 0; i < maxResults; i++ {
		item := data.BestFlights[i]
		sb.WriteString(fmt.Sprintf("%d. %s | Price: %d | Total: %d min\n", i+1, item.Type, item.Price, item.TotalDuration))
		for _, flight := range item.Flights {
			sb.WriteString(fmt.Sprintf(
				"  - %s %s: %s (%s) -> %s (%s) | %d min\n",
				flight.Airline,
				flight.FlightNumber,
				flight.DepartureAirport.Time,
				flight.DepartureAirport.ID,
				flight.ArrivalAirport.Time,
				flight.ArrivalAirport.ID,
				flight.Duration,
			))
		}
		if len(item.Layovers) > 0 {
			sb.WriteString("  - Layovers: ")
			for j, layover := range item.Layovers {
				if j > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString(fmt.Sprintf("%s (%s) %d min", layover.Name, layover.ID, layover.Duration))
			}
			sb.WriteString("\n")
		}
		if len(item.Extensions) > 0 {
			sb.WriteString("  - Notes: ")
			sb.WriteString(strings.Join(item.Extensions, "; "))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}
