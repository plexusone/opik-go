//go:build ignore

// This test lists traces and their spans using the SDK's public methods.
// Useful for verifying that traces and spans are properly linked.
//
// Run with: go run check_spans.go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	opik "github.com/plexusone/opik-go"
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

		// Query spans for this trace using the new ListSpans method
		spans, err := client.ListSpans(ctx, t.ID, 1, 20)
		if err != nil {
			fmt.Printf("   ERROR listing spans: %v\n", err)
			continue
		}

		if len(spans) == 0 {
			fmt.Printf("   Spans: NONE\n")
		} else {
			fmt.Printf("   Spans: %d found\n", len(spans))
			for j, s := range spans {
				fmt.Printf("     %d. %s (type=%s, id=%s)\n", j+1, s.Name, s.Type, s.ID)
				if s.Model != "" {
					fmt.Printf("        Model: %s\n", s.Model)
				}
				if s.Provider != "" {
					fmt.Printf("        Provider: %s\n", s.Provider)
				}
			}
		}
	}
}
