//go:build ignore

// This file tests JSON encoding of API types.
// Useful for debugging JSON serialization issues.
//
// Run with: go run test_json_encode.go
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/plexusone/opik-go/internal/api"
)

func main() {
	traceUUID, _ := uuid.NewV7()

	// Test trace write encoding
	fmt.Println("=== TraceBatchWrite ===")
	traceReq := api.TraceBatchWrite{
		Traces: []api.TraceWrite{{
			ID:          api.NewOptUUID(traceUUID),
			ProjectName: api.NewOptString("test-project"),
			Name:        api.NewOptString("test-trace"),
			StartTime:   time.Now(),
			Tags:        []string{"test"},
		}},
	}

	data, err := json.MarshalIndent(traceReq, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	// Test trace update encoding (the problematic case)
	fmt.Println("\n=== TraceBatchUpdate (with null JSON fields) ===")
	nullJSON := api.JsonListString([]byte("null"))
	outputJSON := api.JsonListString([]byte(`{"result":"test"}`))

	updateReq := api.TraceBatchUpdate{
		Ids: []uuid.UUID{traceUUID},
		Update: api.TraceUpdate{
			EndTime:  api.NewOptDateTime(time.Now()),
			Input:    nullJSON,
			Output:   outputJSON,
			Metadata: nullJSON,
		},
	}

	data, err = json.MarshalIndent(updateReq, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	// Test what happens with empty JsonListString (the bug case)
	fmt.Println("\n=== TraceBatchUpdate (with EMPTY JSON fields - BUG) ===")
	bugReq := api.TraceBatchUpdate{
		Ids: []uuid.UUID{traceUUID},
		Update: api.TraceUpdate{
			EndTime:  api.NewOptDateTime(time.Now()),
			Input:    api.JsonListString{}, // Empty - causes malformed JSON!
			Output:   outputJSON,
			Metadata: api.JsonListString{}, // Empty - causes malformed JSON!
		},
	}

	data, err = json.MarshalIndent(bugReq, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	fmt.Println("\nNote: The above may produce malformed JSON if JsonListString is empty!")
}
