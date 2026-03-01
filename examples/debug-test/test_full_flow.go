//go:build ignore

// This test creates a trace with span and ends both, showing all HTTP traffic
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	opik "github.com/plexusone/opik-go"
)

type debugTransport struct {
	transport http.RoundTripper
}

func (d *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(req, true) // true = include body
	fmt.Printf("\n=== REQUEST ===\n%s\n", string(dump))

	resp, err := d.transport.RoundTrip(req)
	if err != nil {
		fmt.Printf("=== ERROR: %v ===\n", err)
		return nil, err
	}

	fmt.Printf("=== RESPONSE Status: %d ===\n", resp.StatusCode)

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

	fmt.Printf("=== Full Flow Debug Test ===\n")
	fmt.Printf("Workspace: %s, Project: %s\n\n", workspace, project)

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
		fmt.Printf("ERROR creating client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Step 1: Create trace
	fmt.Println("\n>>> STEP 1: Creating trace...")
	ctx, trace, err := opik.StartTrace(ctx, client, "debug-test-trace",
		opik.WithTraceInput(map[string]string{"test": "input"}),
	)
	if err != nil {
		fmt.Printf("ERROR creating trace: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Trace created: %s\n", trace.ID())

	// Step 2: Create span
	fmt.Println("\n>>> STEP 2: Creating span...")
	ctx, span, err := opik.StartSpan(ctx, "debug-test-span",
		opik.WithSpanType("llm"),
		opik.WithSpanInput(map[string]string{"prompt": "test"}),
		opik.WithSpanModel("test-model"),
		opik.WithSpanProvider("test-provider"),
	)
	if err != nil {
		fmt.Printf("ERROR creating span: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Span created: %s\n", span.ID())

	// Small delay
	time.Sleep(100 * time.Millisecond)

	// Step 3: End span
	fmt.Println("\n>>> STEP 3: Ending span...")
	err = span.End(ctx, opik.WithSpanOutput(map[string]string{"result": "output"}))
	if err != nil {
		fmt.Printf("ERROR ending span: %v\n", err)
	} else {
		fmt.Println("Span ended successfully")
	}

	// Step 4: End trace
	fmt.Println("\n>>> STEP 4: Ending trace...")
	err = trace.End(ctx, opik.WithTraceOutput(map[string]string{"final": "output"}))
	if err != nil {
		fmt.Printf("ERROR ending trace: %v\n", err)
	} else {
		fmt.Println("Trace ended successfully")
	}

	fmt.Printf("\n=== COMPLETE ===\nTrace ID: %s\n", trace.ID())
}
