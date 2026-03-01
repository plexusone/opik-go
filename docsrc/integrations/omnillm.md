# omnillm Integration

Integrate with [omnillm](https://github.com/plexusone/omnillm), a unified LLM wrapper library that supports multiple providers (OpenAI, Anthropic, Bedrock, Gemini, Ollama, xAI).

```go
import (
    "github.com/plexusone/omnillm"
    opikomnillm "github.com/plexusone/opik-go/integrations/omnillm"
)
```

## Three Integration Options

Choose the approach that best fits your use case:

| Option | Use Case | Complexity |
|--------|----------|------------|
| [Manual Span Wrapping](#option-1-manual-span-wrapping) | Fine-grained control | Low |
| [Tracing Wrapper](#option-2-tracing-wrapper) | Automatic tracing | Low |
| [Evaluation Provider](#option-3-evaluation-provider) | LLM-as-judge | Low |

---

## Option 1: Manual Span Wrapping

Wrap individual LLM calls with spans for maximum control.

```go
import (
    opik "github.com/plexusone/opik-go"
    "github.com/plexusone/omnillm"
)

func callLLM(ctx context.Context, client *omnillm.ChatClient, req *omnillm.ChatCompletionRequest) (*omnillm.ChatCompletionResponse, error) {
    // Get current span/trace from context
    var span *opik.Span
    var err error

    if parentSpan := opik.SpanFromContext(ctx); parentSpan != nil {
        span, err = parentSpan.Span(ctx, "llm.chat",
            opik.WithSpanType(opik.SpanTypeLLM),
            opik.WithSpanModel(req.Model),
            opik.WithSpanInput(map[string]any{
                "messages": req.Messages,
                "model":    req.Model,
            }),
        )
    } else if trace := opik.TraceFromContext(ctx); trace != nil {
        span, err = trace.Span(ctx, "llm.chat",
            opik.WithSpanType(opik.SpanTypeLLM),
            opik.WithSpanModel(req.Model),
            opik.WithSpanInput(map[string]any{
                "messages": req.Messages,
                "model":    req.Model,
            }),
        )
    }

    // Make the call
    startTime := time.Now()
    resp, respErr := client.CreateChatCompletion(ctx, req)
    duration := time.Since(startTime)

    // End span with output
    if span != nil && err == nil {
        endOpts := []opik.SpanOption{}

        if resp != nil {
            endOpts = append(endOpts, opik.WithSpanOutput(map[string]any{
                "content": resp.Choices[0].Message.Content,
                "model":   resp.Model,
            }))
            endOpts = append(endOpts, opik.WithSpanMetadata(map[string]any{
                "duration_ms":       duration.Milliseconds(),
                "prompt_tokens":     resp.Usage.PromptTokens,
                "completion_tokens": resp.Usage.CompletionTokens,
                "total_tokens":      resp.Usage.TotalTokens,
            }))
        }

        if respErr != nil {
            endOpts = append(endOpts, opik.WithSpanMetadata(map[string]any{
                "error": respErr.Error(),
            }))
        }

        span.End(ctx, endOpts...)
    }

    return resp, respErr
}
```

### When to Use

- You need custom span names or metadata
- You want to trace only specific calls
- You're integrating into existing code gradually

---

## Option 2: Tracing Wrapper

Use the built-in `TracingClient` wrapper for automatic tracing of all calls.

```go
import (
    opik "github.com/plexusone/opik-go"
    opikomnillm "github.com/plexusone/opik-go/integrations/omnillm"
    "github.com/plexusone/omnillm"
)

func main() {
    // Create omnillm client
    client, _ := omnillm.NewClient(omnillm.ClientConfig{
        Provider: omnillm.ProviderNameOpenAI,
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    })

    // Create Opik client
    opikClient, _ := opik.NewClient()

    // Wrap with tracing
    tracingClient := opikomnillm.NewTracingClient(client, opikClient)

    // Start a trace
    ctx, trace, _ := opik.StartTrace(ctx, opikClient, "my-task")
    defer trace.End(ctx)

    // All calls are automatically traced!
    resp, _ := tracingClient.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
        Model: "gpt-4o",
        Messages: []omnillm.Message{
            {Role: omnillm.RoleUser, Content: "Hello!"},
        },
    })

    fmt.Println(resp.Choices[0].Message.Content)
}
```

### TracingClient Methods

| Method | Description |
|--------|-------------|
| `CreateChatCompletion` | Traced chat completion |
| `CreateChatCompletionStream` | Traced streaming completion |
| `CreateChatCompletionWithMemory` | Traced completion with memory |
| `Close` | Close underlying client |
| `Client` | Access underlying omnillm client |

### Streaming Support

```go
stream, _ := tracingClient.CreateChatCompletionStream(ctx, req)
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    fmt.Print(chunk.Choices[0].Delta.Content)
}
// Span automatically ended with accumulated content
```

### Memory Support

```go
// Traced conversation with memory
resp, _ := tracingClient.CreateChatCompletionWithMemory(ctx, "session-123", req)
// Span includes session_id in metadata
```

### When to Use

- You want automatic tracing for all LLM calls
- You're building a new application
- You want consistent span formatting

---

## Option 3: Evaluation Provider

Use omnillm as an LLM provider for evaluation judges.

```go
import (
    opikomnillm "github.com/plexusone/opik-go/integrations/omnillm"
    "github.com/plexusone/opik-go/evaluation/llm"
    "github.com/plexusone/omnillm"
)

func main() {
    // Create omnillm client with any provider
    client, _ := omnillm.NewClient(omnillm.ClientConfig{
        Provider: omnillm.ProviderNameAnthropic,
        APIKey:   os.Getenv("ANTHROPIC_API_KEY"),
    })

    // Create evaluation provider
    provider := opikomnillm.NewProvider(client,
        opikomnillm.WithModel("claude-sonnet-4-20250514"),
        opikomnillm.WithTemperature(0.0),
    )

    // Use with evaluation metrics
    relevance := llm.NewAnswerRelevance(provider)
    hallucination := llm.NewHallucination(provider)
    coherence := llm.NewCoherence(provider)

    // Create evaluation engine
    engine := evaluation.NewEngine([]evaluation.Metric{
        relevance,
        hallucination,
        coherence,
    })

    // Evaluate
    input := evaluation.NewMetricInput(question, answer).
        WithExpected(expectedAnswer).
        WithContext(documents)

    result := engine.EvaluateOne(ctx, input)
}
```

### Provider Options

| Option | Description |
|--------|-------------|
| `WithModel(model)` | Set model name |
| `WithTemperature(temp)` | Set temperature |
| `WithMaxTokens(max)` | Set max tokens |

### When to Use

- Running evaluation experiments
- LLM-as-judge workflows
- Comparing outputs across models

---

## Supported Providers

omnillm supports these providers, all work with the Opik integration:

| Provider | Config |
|----------|--------|
| OpenAI | `ProviderNameOpenAI` |
| Anthropic | `ProviderNameAnthropic` |
| AWS Bedrock | `ProviderNameBedrock` |
| Google Gemini | `ProviderNameGemini` |
| Ollama | `ProviderNameOllama` |
| xAI (Grok) | `ProviderNameXAI` |

## Complete Example: All Three Options

```go
package main

import (
    "context"
    "fmt"

    opik "github.com/plexusone/opik-go"
    "github.com/plexusone/opik-go/evaluation"
    "github.com/plexusone/opik-go/evaluation/llm"
    opikomnillm "github.com/plexusone/opik-go/integrations/omnillm"
    "github.com/plexusone/omnillm"
)

func main() {
    ctx := context.Background()

    // Create omnillm client
    client, _ := omnillm.NewClient(omnillm.ClientConfig{
        Provider: omnillm.ProviderNameOpenAI,
        APIKey:   os.Getenv("OPENAI_API_KEY"),
    })

    // Create Opik client
    opikClient, _ := opik.NewClient()

    // OPTION 2: Tracing wrapper for automatic tracing
    tracingClient := opikomnillm.NewTracingClient(client, opikClient)

    // OPTION 3: Evaluation provider for LLM judges
    evalProvider := opikomnillm.NewProvider(client,
        opikomnillm.WithModel("gpt-4o"),
    )

    // Start trace
    ctx, trace, _ := opik.StartTrace(ctx, opikClient, "demo")
    defer trace.End(ctx)

    // Generate response (automatically traced)
    resp, _ := tracingClient.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
        Model:    "gpt-4o",
        Messages: []omnillm.Message{{Role: omnillm.RoleUser, Content: "What is 2+2?"}},
    })

    answer := resp.Choices[0].Message.Content

    // Evaluate response
    metrics := []evaluation.Metric{
        llm.NewAnswerRelevance(evalProvider),
        llm.NewCoherence(evalProvider),
    }
    engine := evaluation.NewEngine(metrics)

    input := evaluation.NewMetricInput("What is 2+2?", answer).WithExpected("4")
    result := engine.EvaluateOne(ctx, input)

    fmt.Printf("Response: %s\n", answer)
    fmt.Printf("Average score: %.2f\n", result.AverageScore())
}
```
