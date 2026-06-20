package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListWorkloadsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"workload": map[string]interface{}{
							"workloads": []map[string]interface{}{
								{"guid": "wl-1", "name": "Production Stack"},
								{"guid": "wl-2", "name": "Staging Stack"},
							},
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
	result, err := server.ExecuteTool(ctx, "list_workloads", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("list_workloads failed: %v", err)
	}
	if !contains(result.RawText, "Production Stack") {
		t.Errorf("Expected 'Production Stack' in result, got: %s", result.RawText)
	}
}

func TestListWorkloadsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"workload": map[string]interface{}{
							"workloads": []interface{}{},
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
	result, err := server.ExecuteTool(ctx, "list_workloads", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_workloads should not error on empty: %v", err)
	}
	if !contains(result.RawText, "No workloads found") {
		t.Errorf("Expected 'No workloads found', got: %s", result.RawText)
	}
}

func TestGetWorkloadTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": map[string]interface{}{
						"guid": "wl-1",
						"name": "Production Stack",
						"workload": map[string]interface{}{
							"status": "HEALTHY",
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
	result, err := server.ExecuteTool(ctx, "get_workload", map[string]interface{}{
		"workload_guid": "wl-1",
	})

	if err != nil {
		t.Fatalf("get_workload failed: %v", err)
	}
	if !contains(result.RawText, "Production Stack") {
		t.Errorf("Expected 'Production Stack' in result, got: %s", result.RawText)
	}
}

func TestGetWorkloadToolNotFound(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "get_workload", map[string]interface{}{
		"workload_guid": "bad-guid",
	})
	if err != nil {
		t.Fatalf("get_workload should not error on not found: %v", err)
	}
	if !contains(result.RawText, "not found") {
		t.Errorf("Expected 'not found' in result, got: %s", result.RawText)
	}
}

func TestGetWorkloadToolMissingGUID(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_workload", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing workload_guid")
	}
}

func TestListWorkloadsEmptyAccountResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "list_workloads", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_workloads should not error on empty response: %v", err)
	}
	if !contains(result.RawText, "No workloads found") {
		t.Errorf("Expected 'No workloads found', got: %s", result.RawText)
	}
}
