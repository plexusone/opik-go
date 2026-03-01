// Package omnillm provides Opik integration for the omnillm LLM wrapper library.
//
// This package provides two main features:
//
//  1. Automatic tracing: Wrap omnillm.ChatClient to automatically create spans
//     for all LLM calls with input/output/usage tracking.
//
//  2. Evaluation provider: Use omnillm as an LLM provider for evaluation metrics
//     that require LLM-as-judge capabilities.
//
// # Automatic Tracing
//
// Wrap your omnillm.ChatClient to automatically trace all LLM calls:
//
//	import (
//	    "github.com/plexusone/omnillm"
//	    opik "github.com/plexusone/opik-go"
//	    opikomnillm "github.com/plexusone/opik-go/integrations/omnillm"
//	)
//
//	// Create omnillm client
//	client, _ := omnillm.NewClient(omnillm.ClientConfig{
//	    Provider: omnillm.ProviderNameOpenAI,
//	    APIKey:   os.Getenv("OPENAI_API_KEY"),
//	})
//
//	// Create Opik client
//	opikClient, _ := opik.NewClient()
//
//	// Wrap for tracing
//	tracingClient := opikomnillm.NewTracingClient(client, opikClient)
//
//	// Use within a trace context
//	ctx, trace, _ := opik.StartTrace(ctx, opikClient, "my-task")
//	defer trace.End(ctx)
//
//	// All calls are automatically traced
//	resp, _ := tracingClient.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
//	    Model:    "gpt-4o",
//	    Messages: []omnillm.Message{{Role: omnillm.RoleUser, Content: "Hello!"}},
//	})
//
// # Evaluation Provider
//
// Use omnillm as a provider for LLM-based evaluation metrics:
//
//	import (
//	    "github.com/plexusone/omnillm"
//	    opikomnillm "github.com/plexusone/opik-go/integrations/omnillm"
//	    "github.com/plexusone/opik-go/evaluation/llm"
//	)
//
//	// Create omnillm client
//	client, _ := omnillm.NewClient(omnillm.ClientConfig{
//	    Provider: omnillm.ProviderNameAnthropic,
//	    APIKey:   os.Getenv("ANTHROPIC_API_KEY"),
//	})
//
//	// Create evaluation provider
//	provider := opikomnillm.NewProvider(client,
//	    opikomnillm.WithModel("claude-3-opus-20240229"),
//	)
//
//	// Use with evaluation metrics
//	relevance := llm.NewAnswerRelevance(provider)
//	hallucination := llm.NewHallucination(provider)
//
// # Streaming Support
//
// The tracing client also supports streaming with automatic span capture:
//
//	stream, _ := tracingClient.CreateChatCompletionStream(ctx, req)
//	defer stream.Close()
//
//	for {
//	    chunk, err := stream.Recv()
//	    if err == io.EOF {
//	        break
//	    }
//	    // Process chunk...
//	}
//	// Span is automatically ended with complete response when stream closes
package omnillm
