//go:build ignore

// Get span by ID directly and also try querying by trace ID
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"

	opik "github.com/plexusone/opik-go"
	"github.com/plexusone/opik-go/internal/api"
)

func main() {
	apiKey := os.Getenv("OPIK_API_KEY")
	workspace := os.Getenv("OPIK_WORKSPACE")

	// From the test run
	spanIDStr := "019b8424-f897-71a1-9210-dcaa279ac65a"
	traceIDStr := "019b8424-f6e5-7a44-bad7-a4f3d9705d7f"

	if len(os.Args) > 1 {
		spanIDStr = os.Args[1]
	}
	if len(os.Args) > 2 {
		traceIDStr = os.Args[2]
	}

	client, err := opik.NewClient(
		opik.WithAPIKey(apiKey),
		opik.WithWorkspace(workspace),
	)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	apiClient := client.API()

	// Try to get span by ID directly
	fmt.Println("=== Getting Span by ID ===")
	spanUUID, err := uuid.Parse(spanIDStr)
	if err != nil {
		fmt.Printf("ERROR parsing span ID: %v\n", err)
		os.Exit(1)
	}

	spanRes, err := apiClient.GetSpanById(ctx, api.GetSpanByIdParams{
		ID: spanUUID,
	})
	if err != nil {
		fmt.Printf("ERROR getting span: %v\n", err)
	} else {
		switch v := spanRes.(type) {
		case *api.GetSpanByIdOK:
			data, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(data))
		case *api.GetSpanByIdNotFound:
			fmt.Println("Span not found!")
		default:
			fmt.Printf("Unexpected response type: %T\n", v)
		}
	}

	// Try querying spans by trace ID
	fmt.Println("\n=== Querying Spans by Trace ID ===")
	traceUUID, err := uuid.Parse(traceIDStr)
	if err != nil {
		fmt.Printf("ERROR parsing trace ID: %v\n", err)
		os.Exit(1)
	}

	spans, err := apiClient.GetSpansByProject(ctx, api.GetSpansByProjectParams{
		ProjectName: api.NewOptString("stats-agent-team"),
		TraceID:     api.NewOptUUID(traceUUID),
		Size:        api.NewOptInt32(50),
	})
	if err != nil {
		fmt.Printf("ERROR getting spans: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Spans found: %d\n", len(spans.Content))
	for i, s := range spans.Content {
		fmt.Printf("\n--- Span %d ---\n", i+1)
		data, _ := json.MarshalIndent(s, "", "  ")
		fmt.Println(string(data))
	}
}
