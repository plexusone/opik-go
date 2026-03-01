//go:build ignore

// Get detailed info about a specific trace
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

	traceIDStr := "019b840a-ac41-7d07-8fea-cc6a612cc80b" // The test trace
	if len(os.Args) > 1 {
		traceIDStr = os.Args[1]
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

	// Get trace by ID
	traceUUID, err := uuid.Parse(traceIDStr)
	if err != nil {
		fmt.Printf("ERROR parsing trace ID: %v\n", err)
		os.Exit(1)
	}

	trace, err := apiClient.GetTraceById(ctx, api.GetTraceByIdParams{
		ID: traceUUID,
	})
	if err != nil {
		fmt.Printf("ERROR getting trace: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Trace Details ===")
	data, _ := json.MarshalIndent(trace, "", "  ")
	fmt.Println(string(data))

	// Get ALL spans for project (no trace filter)
	fmt.Println("\n=== All Spans in Project ===")
	spans, err := apiClient.GetSpansByProject(ctx, api.GetSpansByProjectParams{
		ProjectName: api.NewOptString("stats-agent-team"),
		Size:        api.NewOptInt32(50),
	})
	if err != nil {
		fmt.Printf("ERROR getting spans: %v\n", err)
		os.Exit(1)
	}

	if len(spans.Content) == 0 {
		fmt.Println("No spans found")
	} else {
		for i, s := range spans.Content {
			fmt.Printf("\n--- Span %d ---\n", i+1)
			data, _ := json.MarshalIndent(s, "", "  ")
			fmt.Println(string(data))
		}
	}
}
