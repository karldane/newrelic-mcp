package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListApplicationsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []map[string]interface{}{
								{
									"guid":     "guid-1",
									"name":     "MyApp",
									"language": "java",
								},
								{
									"guid":     "guid-2",
									"name":     "AnotherApp",
									"language": "ruby",
								},
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
	result, err := server.ExecuteTool(ctx, "list_applications", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("List applications failed: %v", err)
	}

	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}

	if !contains(result.RawText, "MyApp") {
		t.Errorf("Expected result to contain 'MyApp', got: %s", result.RawText)
	}

	if !contains(result.RawText, "AnotherApp") {
		t.Errorf("Expected result to contain 'AnotherApp', got: %s", result.RawText)
	}
}

func TestListApplicationsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []map[string]interface{}{},
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
	result, err := server.ExecuteTool(ctx, "list_applications", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("List applications failed: %v", err)
	}

	if !contains(result.RawText, "No applications") {
		t.Errorf("Expected result to indicate no applications, got: %s", result.RawText)
	}
}

func TestGetAlertConditionsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"alerts": map[string]interface{}{
							"alertConditionsSearch": map[string]interface{}{
								"alertConditions": []map[string]interface{}{
									{
										"id":      "cond1",
										"name":    "High Error Rate",
										"type":    "NRQL",
										"enabled": true,
									},
								},
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
	result, err := server.ExecuteTool(ctx, "get_alert_conditions", map[string]interface{}{
		"account_id": "12345",
		"policy_id":  "12345",
	})

	if err != nil {
		t.Fatalf("Get alert conditions failed: %v", err)
	}

	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}

	if !contains(result.RawText, "High Error Rate") {
		t.Errorf("Expected result to contain 'High Error Rate', got: %s", result.RawText)
	}
}

func TestGetAlertConditionsToolMissingPolicyID(t *testing.T) {
	server := NewServer("test-key")

	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_alert_conditions", map[string]interface{}{
		"account_id": "12345",
	})

	if err == nil {
		t.Fatal("Expected error for missing policy_id")
	}

	if err.Error() != "tool get_alert_conditions: missing required parameter: policy_id" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestQueryTracesTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{
								{
									"traceId":     "abc123",
									"duration":    500.5,
									"entity.name": "MyService",
									"error":       true,
								},
							},
							"metadata": map[string]interface{}{
								"facets": []string{},
								"timeWindow": map[string]interface{}{
									"begin": "2024-01-01T00:00:00Z",
									"end":   "2024-01-01T01:00:00Z",
								},
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
	result, err := server.ExecuteTool(ctx, "query_traces", map[string]interface{}{
		"account_id": "12345",
		"duration":   "1 hour",
	})

	if err != nil {
		t.Fatalf("Query traces failed: %v", err)
	}

	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}

	if !contains(result.RawText, "abc123") {
		t.Errorf("Expected result to contain trace ID 'abc123', got: %s", result.RawText)
	}
}

func TestGetApplicationMetricsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{
								{
									"throughput":   120.5,
									"errorRate":    2.5,
									"responseTime": 150.3,
									"apdex":        0.95,
								},
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
	result, err := server.ExecuteTool(ctx, "get_application_metrics", map[string]interface{}{
		"account_id": "12345",
		"app_name":   "MyApp",
		"metrics":    []string{"throughput", "error_rate"},
	})

	if err != nil {
		t.Fatalf("Get application metrics failed: %v", err)
	}

	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}

	if !contains(result.RawText, "throughput") {
		t.Errorf("Expected result to contain 'throughput', got: %s", result.RawText)
	}
}

func TestGetTransactionTracesTool(t *testing.T) {
	server := NewServer("test-key")

	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "get_transaction_traces", map[string]interface{}{
		"account_id": "12345",
		"app_name":   "MyApp",
		"limit":      5,
	})

	if err != nil {
		t.Fatalf("Get transaction traces failed: %v", err)
	}

	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}

	if !contains(result.RawText, "MyApp") {
		t.Errorf("Expected result to contain 'MyApp', got: %s", result.RawText)
	}
}

func TestGetTransactionTracesToolMissingAppName(t *testing.T) {
	server := NewServer("test-key")

	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_transaction_traces", map[string]interface{}{
		"account_id": "12345",
	})

	if err == nil {
		t.Fatal("Expected error for missing app_name")
	}

	if err.Error() != "tool get_transaction_traces: missing required parameter: app_name" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestGetTraceDetailsTool(t *testing.T) {
	server := NewServer("test-key")

	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "get_trace_details", map[string]interface{}{
		"account_id": "12345",
		"trace_id":   "abc123def456",
	})

	if err != nil {
		t.Fatalf("Get trace details failed: %v", err)
	}

	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}

	if !contains(result.RawText, "abc123def456") {
		t.Errorf("Expected result to contain trace ID, got: %s", result.RawText)
	}
}

func TestGetTraceDetailsToolMissingTraceID(t *testing.T) {
	server := NewServer("test-key")

	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_trace_details", map[string]interface{}{
		"account_id": "12345",
	})

	if err == nil {
		t.Fatal("Expected error for missing trace_id")
	}

	if err.Error() != "tool get_trace_details: missing required parameter: trace_id" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
