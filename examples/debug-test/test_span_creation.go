//go:build ignore

// This file tests span creation directly to verify the SDK works.
// Run with: go run test_span_creation.go
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

	fmt.Printf("Testing span creation...\n")
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

	// Step 1: Create a trace using context helper
	fmt.Println("\n1. Creating trace via opik.StartTrace...")
	ctx, trace, err := opik.StartTrace(ctx, client, "test-span-creation-trace",
		opik.WithTraceInput(map[string]string{"test": "input"}),
	)
	if err != nil {
		fmt.Printf("ERROR creating trace: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   Trace created: %s\n", trace.ID())

	// Step 2: Verify trace is in context
	fmt.Println("\n2. Verifying trace in context...")
	traceFromCtx := opik.TraceFromContext(ctx)
	if traceFromCtx == nil {
		fmt.Println("   ERROR: TraceFromContext returned nil!")
		os.Exit(1)
	}
	fmt.Printf("   TraceFromContext returned: %s\n", traceFromCtx.ID())

	// Step 3: Create a span using context helper
	fmt.Println("\n3. Creating span via opik.StartSpan...")
	ctx, span, err := opik.StartSpan(ctx, "test-span",
		opik.WithSpanType("llm"),
		opik.WithSpanInput(map[string]string{"prompt": "test"}),
		opik.WithSpanModel("test-model"),
		opik.WithSpanProvider("test-provider"),
	)
	if err != nil {
		fmt.Printf("   ERROR creating span: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("   Span created: %s (traceID=%s)\n", span.ID(), span.TraceID())

	// Step 4: Verify span is in context
	fmt.Println("\n4. Verifying span in context...")
	spanFromCtx := opik.SpanFromContext(ctx)
	if spanFromCtx == nil {
		fmt.Println("   ERROR: SpanFromContext returned nil!")
		os.Exit(1)
	}
	fmt.Printf("   SpanFromContext returned: %s\n", spanFromCtx.ID())

	// Step 5: End the span
	fmt.Println("\n5. Ending span...")
	time.Sleep(100 * time.Millisecond) // Small delay to simulate work
	err = span.End(ctx, opik.WithSpanOutput(map[string]string{"result": "test output"}))
	if err != nil {
		fmt.Printf("   ERROR ending span: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("   Span ended successfully")

	// Step 6: End the trace
	fmt.Println("\n6. Ending trace...")
	err = trace.End(ctx, opik.WithTraceOutput(map[string]string{"final": "output"}))
	if err != nil {
		fmt.Printf("   ERROR ending trace: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("   Trace ended successfully")

	fmt.Println("\n=== SUCCESS ===")
	fmt.Println("Trace and span created and ended successfully.")
	fmt.Printf("Check Comet.com for trace ID: %s\n", trace.ID())
}
