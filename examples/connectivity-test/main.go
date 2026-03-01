// Package main provides a connectivity test with HTTP debug output.
//
// This is useful for debugging API communication issues by showing
// the raw HTTP requests being sent to the Opik API.
//
// Usage:
//
//	export OPIK_API_KEY="your-api-key"
//	export OPIK_WORKSPACE="your-workspace"
//	export OPIK_PROJECT="your-project"  # optional
//	go run .
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	opik "github.com/plexusone/opik-go"
)

type debugTransport struct {
	transport http.RoundTripper
}

func (d *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(req, false)
	fmt.Printf("=== REQUEST ===\n%s\n", string(dump))

	resp, err := d.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

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

	httpClient := &http.Client{
		Transport: &debugTransport{transport: http.DefaultTransport},
	}

	client, err := opik.NewClient(
		opik.WithAPIKey(apiKey),
		opik.WithWorkspace(workspace),
		opik.WithProjectName(project),
		opik.WithHTTPClient(httpClient),
	)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	traces, err := client.ListTraces(ctx, 1, 2)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n=== TRACES ===\n")
	for _, t := range traces {
		data, _ := json.MarshalIndent(t, "", "  ")
		fmt.Printf("%s\n", data)
	}
}
