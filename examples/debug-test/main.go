// Package main provides diagnostic tools for testing and debugging Opik traces and spans.
//
// Usage:
//
//	export OPIK_API_KEY="your-api-key"
//	export OPIK_WORKSPACE="your-workspace"
//	export OPIK_PROJECT="your-project"  # optional, defaults to "stats-agent-team"
//	go run .
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	opik "github.com/plexusone/opik-go"
	"github.com/plexusone/opik-go/internal/api"
)

func main() {
	apiKey := os.Getenv("OPIK_API_KEY")
	workspace := os.Getenv("OPIK_WORKSPACE")
	project := os.Getenv("OPIK_PROJECT")
	if project == "" {
		project = "stats-agent-team"
	}

	if apiKey == "" {
		fmt.Println("ERROR: OPIK_API_KEY not set")
		os.Exit(1)
	}

	fmt.Printf("Connecting to Comet Opik...\n")
	fmt.Printf("  Workspace: %s\n", workspace)
	fmt.Printf("  Project: %s\n", project)

	client, err := opik.NewClient(
		opik.WithAPIKey(apiKey),
		opik.WithWorkspace(workspace),
		opik.WithProjectName(project),
	)
	if err != nil {
		fmt.Printf("ERROR creating client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	apiClient := client.API()

	// List traces
	fmt.Println("\n--- Recent Traces ---")
	traces, err := client.ListTraces(ctx, 1, 10)
	if err != nil {
		fmt.Printf("ERROR listing traces: %v\n", err)
		os.Exit(1)
	}

	if len(traces) == 0 {
		fmt.Println("No traces found")
		os.Exit(0)
	}

	for i, t := range traces {
		fmt.Printf("\n%d. Trace: %s\n", i+1, t.Name)
		fmt.Printf("   ID: %s\n", t.ID)
		fmt.Printf("   Start: %s\n", t.StartTime.Format(time.RFC3339))
		if !t.EndTime.IsZero() {
			fmt.Printf("   End: %s\n", t.EndTime.Format(time.RFC3339))
			fmt.Printf("   Duration: %s\n", t.EndTime.Sub(t.StartTime))
		} else {
			fmt.Printf("   End: NOT SET (trace still open)\n")
		}

		// Query spans for this trace using the API directly
		traceUUID, err := uuid.Parse(t.ID)
		if err != nil {
			fmt.Printf("   ERROR parsing trace UUID: %v\n", err)
			continue
		}

		spans, err := apiClient.GetSpansByProject(ctx, api.GetSpansByProjectParams{
			ProjectName: api.NewOptString(project),
			TraceID:     api.NewOptUUID(traceUUID),
			Size:        api.NewOptInt32(20),
		})
		if err != nil {
			fmt.Printf("   ERROR listing spans: %v\n", err)
			continue
		}

		if len(spans.Content) == 0 {
			fmt.Printf("   Spans: NONE\n")
		} else {
			fmt.Printf("   Spans: %d found\n", len(spans.Content))
			for j, s := range spans.Content {
				name := "<unnamed>"
				if s.Name.Set {
					name = s.Name.Value
				}
				spanType := "<unknown>"
				if s.Type.Set {
					spanType = string(s.Type.Value)
				}
				fmt.Printf("     %d. %s (type=%s)\n", j+1, name, spanType)
				if s.Model.Set && s.Model.Value != "" {
					fmt.Printf("        Model: %s\n", s.Model.Value)
				}
				if s.Provider.Set && s.Provider.Value != "" {
					fmt.Printf("        Provider: %s\n", s.Provider.Value)
				}
			}
		}
	}
}
