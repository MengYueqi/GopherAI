package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
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
	serverName    = "github.com/songjiayang/mcp-flight-mock"
	serverVersion = "1.0.0"
)

func init() {
	flag.StringVar(&transport, "transport", "sse", "The transport to use, should be \"stdio\" or \"sse\"")
	flag.StringVar(&serverlisten, "server_listen", "localhost:8082", "The sse server listen address")
	flag.Parse()
}

func main() {
	s := server.NewMCPServer(serverName, serverVersion)

	timeTool := mcp.NewTool("current time",
		mcp.WithDescription("Get current time with timezone, Asia/Shanghai is default"),
		mcp.WithString("timezone",
			mcp.Required(),
			mcp.Description("current time timezone"),
		),
	)
	s.AddTool(timeTool, currentTimeHandler)

	flightTool := mcp.NewTool(
		"google_flights",
		mcp.WithDescription("Mock flight info for local development without calling SerpAPI"),
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
			mcp.Description("outbound date in YYYY-MM-DD"),
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
	s.AddTool(flightTool, googleFlightsHandler)

	if transport == "sse" {
		serverURL := "http://" + serverlisten
		sseServer := server.NewSSEServer(s, server.WithBaseURL(serverURL))
		log.Printf("mock mcp-flight SSE server listening on %s", serverlisten)
		if err := sseServer.Start(serverlisten); err != nil {
			log.Fatalf("Server error: %v", err)
		}
		return
	}

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func currentTimeHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("arguments must be an object")
	}
	timezone, ok := args["timezone"].(string)
	if !ok || timezone == "" {
		return nil, errors.New("timezone must be a non-empty string")
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("parse timezone with error: %v", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("current time is %s", time.Now().In(loc).Format(time.RFC3339))), nil
}

func googleFlightsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return nil, errors.New("arguments must be an object")
	}

	departureID, ok := args["departure_id"].(string)
	if !ok || strings.TrimSpace(departureID) == "" {
		return nil, errors.New("departure_id must be a non-empty string")
	}
	arrivalID, ok := args["arrival_id"].(string)
	if !ok || strings.TrimSpace(arrivalID) == "" {
		return nil, errors.New("arrival_id must be a non-empty string")
	}
	outboundDate, ok := args["outbound_date"].(string)
	if !ok || strings.TrimSpace(outboundDate) == "" {
		return nil, errors.New("outbound_date must be a non-empty string")
	}

	currency, _ := args["currency"].(string)
	if strings.TrimSpace(currency) == "" {
		currency = "CNY"
	}
	flightType, _ := args["type"].(string)
	if strings.TrimSpace(flightType) == "" {
		flightType = "2"
	}

	log.Printf(
		"mock google_flights params: departure_id=%s arrival_id=%s outbound_date=%s currency=%s type=%s",
		departureID, arrivalID, outboundDate, currency, flightType,
	)

	result := buildMockFlightResult(strings.ToUpper(departureID), strings.ToUpper(arrivalID), outboundDate, strings.ToUpper(currency))
	return mcp.NewToolResultText(result), nil
}

func buildMockFlightResult(departureID, arrivalID, outboundDate, currency string) string {
	switch {
	case departureID == "SHA" && arrivalID == "NRT":
		return fmt.Sprintf(
			"1. One way | Price: %s 2350 | Total: 205 min\n  - China Eastern MU 523: %s 09:20 (%s) -> %s 13:45 (%s) | 205 min\n  - Notes: Mock fare; Carry-on included; Best for nonstop arrival\n\n2. One way | Price: %s 1980 | Total: 410 min\n  - Spring Airlines 9C 8517: %s 07:10 (%s) -> %s 10:55 (PVG) | 105 min\n  - Japan Airlines JL 874: %s 13:20 (PVG) -> %s 16:00 (%s) | 160 min\n  - Layovers: Shanghai Pudong (PVG) 145 min\n  - Notes: Mock fare; Cheapest option; self-transfer risk\n",
			currency, outboundDate, departureID, outboundDate, arrivalID,
			currency, outboundDate, departureID, outboundDate, outboundDate, outboundDate, arrivalID,
		)
	case departureID == "SHA" && arrivalID == "KIX":
		return fmt.Sprintf(
			"1. One way | Price: %s 1890 | Total: 175 min\n  - China Eastern MU 747: %s 08:30 (%s) -> %s 12:25 (%s) | 175 min\n  - Notes: Mock fare; Nonstop; Good for Kansai itinerary\n\n2. One way | Price: %s 1620 | Total: 370 min\n  - Juneyao HO 1335: %s 06:50 (%s) -> %s 09:25 (PVG) | 95 min\n  - Peach MM 080: %s 11:40 (PVG) -> %s 14:00 (%s) | 140 min\n  - Layovers: Shanghai Pudong (PVG) 135 min\n  - Notes: Mock fare; Budget option\n",
			currency, outboundDate, departureID, outboundDate, arrivalID,
			currency, outboundDate, departureID, outboundDate, outboundDate, outboundDate, arrivalID,
		)
	default:
		return fmt.Sprintf(
			"1. One way | Price: %s 2680 | Total: 420 min\n  - MockAir MK 101: %s 09:00 (%s) -> %s 12:10 (HKG) | 190 min\n  - MockAir MK 202: %s 14:20 (HKG) -> %s 16:00 (%s) | 100 min\n  - Layovers: Hong Kong (HKG) 130 min\n  - Notes: Mock fare; Generated for local development only\n\n2. One way | Price: %s 3120 | Total: 255 min\n  - MockAir MK 888: %s 13:40 (%s) -> %s 17:55 (%s) | 255 min\n  - Notes: Mock fare; Fastest option; Generated for local development only\n",
			currency, outboundDate, departureID, outboundDate, outboundDate, outboundDate, arrivalID,
			currency, outboundDate, departureID, outboundDate, arrivalID,
		)
	}
}
