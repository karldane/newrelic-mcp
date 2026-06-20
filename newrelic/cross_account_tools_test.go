package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCrossAccountNRQLTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"nrql": map[string]interface{}{
						"results": []map[string]interface{}{
							{"appName": "my-app", "count": float64(42)},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "cross_account_nrql", map[string]interface{}{
		"query":       "SELECT count(*) FROM Transaction",
		"account_ids": "12345,67890",
	})

	if err != nil {
		t.Fatalf("cross_account_nrql failed: %v", err)
	}
	if !contains(result.RawText, "my-app") {
		t.Errorf("Expected 'my-app' in result, got: %s", result.RawText)
	}
}

func TestCrossAccountNRQLToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"nrql": map[string]interface{}{
						"results": []interface{}{},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "cross_account_nrql", map[string]interface{}{
		"query":       "SELECT count(*) FROM Transaction",
		"account_ids": "12345",
	})
	if err != nil {
		t.Fatalf("cross_account_nrql should not error on empty: %v", err)
	}
	if !contains(result.RawText, "No results") {
		t.Errorf("Expected 'No results' in result, got: %s", result.RawText)
	}
}

func TestCrossAccountNRQLToolMissingParams(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "cross_account_nrql", map[string]interface{}{
		"query": "SELECT * FROM Transaction",
	})
	if err == nil {
		t.Fatal("Expected error for missing account_ids")
	}
}

func TestCrossAccountNRQLToolMissingQuery(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "cross_account_nrql", map[string]interface{}{
		"account_ids": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing query")
	}
}
