package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFormatResultsEmpty(t *testing.T) {
	result := formatResults(nil)
	if result != "No results found" {
		t.Errorf("Expected 'No results found', got: %s", result)
	}

	result = formatResults([]map[string]interface{}{})
	if result != "No results found" {
		t.Errorf("Expected 'No results found', got: %s", result)
	}
}

func TestFormatResultsSingle(t *testing.T) {
	results := []map[string]interface{}{
		{"name": "test", "value": 42},
	}
	result := formatResults(results)
	if !contains(result, "name: test") {
		t.Errorf("Expected 'name: test' in result, got: %s", result)
	}
	if !contains(result, "value: 42") {
		t.Errorf("Expected 'value: 42' in result, got: %s", result)
	}
}

func TestFormatResultsMultiple(t *testing.T) {
	results := []map[string]interface{}{
		{"name": "first", "count": 1},
		{"name": "second", "count": 2},
	}
	result := formatResults(results)
	if !contains(result, "---") {
		t.Errorf("Expected separator between results, got: %s", result)
	}
	if !contains(result, "first") || !contains(result, "second") {
		t.Errorf("Expected both items in result, got: %s", result)
	}
}

func TestFormatSingleResultEmpty(t *testing.T) {
	result := formatSingleResult(nil)
	if result != "" {
		t.Errorf("Expected empty string for nil, got: %s", result)
	}
	result = formatSingleResult(map[string]interface{}{})
	if result != "" {
		t.Errorf("Expected empty string for empty map, got: %s", result)
	}
}

func TestFormatSingleResultWithData(t *testing.T) {
	data := map[string]interface{}{
		"guid": "abc123",
		"name": "My Dashboard",
	}
	result := formatSingleResult(data)
	if !contains(result, "guid: abc123") {
		t.Errorf("Expected 'guid: abc123' in result, got: %s", result)
	}
	if !contains(result, "name: My Dashboard") {
		t.Errorf("Expected 'name: My Dashboard' in result, got: %s", result)
	}
}

func TestNerdGraphQueryNoData(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer mockNR.Close()

	client := NewClientWithEndpoint("test-key", mockNR.URL)
	_, err := client.nerdGraphQuery(context.Background(), "{ actor { account { id } } }")
	if err == nil {
		t.Fatal("Expected error for empty response")
	}
	if err.Error() != "no data in response" {
		t.Errorf("Expected 'no data in response', got: %v", err)
	}
}

func TestNerdGraphQueryNoActor(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{},
		})
	}))
	defer mockNR.Close()

	client := NewClientWithEndpoint("test-key", mockNR.URL)
	_, err := client.nerdGraphQuery(context.Background(), "{ actor { account { id } } }")
	if err == nil {
		t.Fatal("Expected error for missing actor")
	}
	if err.Error() != "no actor in response" {
		t.Errorf("Expected 'no actor in response', got: %v", err)
	}
}

func TestNerdGraphQueryNoAccount(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{},
			},
		})
	}))
	defer mockNR.Close()

	client := NewClientWithEndpoint("test-key", mockNR.URL)
	_, err := client.nerdGraphQuery(context.Background(), "{ actor { account { id } } }")
	if err == nil {
		t.Fatal("Expected error for missing account")
	}
	if err.Error() != "no account in response" {
		t.Errorf("Expected 'no account in response', got: %v", err)
	}
}

func TestNerdGraphQuerySuccess(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"id": "12345",
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	client := NewClientWithEndpoint("test-key", mockNR.URL)
	acct, err := client.nerdGraphQuery(context.Background(), "{ actor { account { id } } }")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if acct["id"] != "12345" {
		t.Errorf("Expected account id '12345', got: %v", acct["id"])
	}
}

func TestGetAccountIDCached(t *testing.T) {
	client := NewClient("test-key")
	client.accountID = "99999"
	id, err := client.GetAccountID(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if id != "99999" {
		t.Errorf("Expected cached 99999, got: %s", id)
	}
}

func TestGetAccountIDNoAccounts(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"accounts": []interface{}{},
				},
			},
		})
	}))
	defer mockNR.Close()

	client := NewClientWithEndpoint("test-key", mockNR.URL)
	_, err := client.GetAccountID(context.Background())
	if err == nil {
		t.Fatal("Expected error for no accounts")
	}
	if !contains(err.Error(), "no accounts found") {
		t.Errorf("Expected 'no accounts found', got: %v", err)
	}
}

func TestGetOrDetectAccountIDProvided(t *testing.T) {
	client := NewClient("test-key")
	id, err := client.getOrDetectAccountID(context.Background(), "provided-id")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if id != "provided-id" {
		t.Errorf("Expected 'provided-id', got: %s", id)
	}
}

func TestExecuteNRQLErrorPaths(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]interface{}
		errMsg   string
	}{
		{
			name:     "no data",
			response: map[string]interface{}{},
			errMsg:   "no data in response",
		},
		{
			name: "no actor",
			response: map[string]interface{}{
				"data": map[string]interface{}{},
			},
			errMsg: "no actor in response",
		},
		{
			name: "no account",
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"actor": map[string]interface{}{},
				},
			},
			errMsg: "no account in response",
		},
		{
			name: "no nrql result",
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"actor": map[string]interface{}{
						"account": map[string]interface{}{},
					},
				},
			},
			errMsg: "no nrql result in response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer mockNR.Close()

			client := NewClientWithEndpoint("test-key", mockNR.URL)
			_, err := client.executeNRQL(context.Background(), "12345", "SELECT count(*) FROM Transaction")
			if err == nil {
				t.Fatal("Expected error")
			}
		})
	}
}

func TestExecuteNRQLNRQLError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"error": "syntax error at line 1",
						},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	client := NewClientWithEndpoint("test-key", mockNR.URL)
	_, err := client.executeNRQL(context.Background(), "12345", "INVALID QUERY")
	if err == nil {
		t.Fatal("Expected NRQL error")
	}
}

func TestGetAlertViolationsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{
								{
									"violationId":    "v123",
									"policyName":     "Test Policy",
									"conditionName":  "High CPU",
									"priority":       "CRITICAL",
									"openedAt":       "2024-01-01T00:00:00Z",
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
	result, err := server.ExecuteTool(ctx, "get_alert_violations", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("Get alert violations failed: %v", err)
	}
	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}
	if !contains(result.RawText, "v123") {
		t.Errorf("Expected result to contain 'v123', got: %s", result.RawText)
	}
	if !contains(result.RawText, "CRITICAL") {
		t.Errorf("Expected result to contain 'CRITICAL', got: %s", result.RawText)
	}
}

func TestGetAlertViolationsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{},
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
	result, err := server.ExecuteTool(ctx, "get_alert_violations", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("Get alert violations failed: %v", err)
	}
	if !contains(result.RawText, "No alert violations") {
		t.Errorf("Expected 'No alert violations', got: %s", result.RawText)
	}
}

func TestTailLogsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{
								{
									"timestamp": float64(1234567890),
									"message":   "User logged in",
									"level":     "INFO",
									"service":   "auth",
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
	result, err := server.ExecuteTool(ctx, "tail_logs", map[string]interface{}{
		"account_id": "12345",
		"query":      "level:INFO",
		"limit":      float64(10),
	})
	if err != nil {
		t.Fatalf("Tail logs failed: %v", err)
	}
	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}
	if !contains(result.RawText, "User logged in") {
		t.Errorf("Expected 'User logged in' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "auth") {
		t.Errorf("Expected 'auth' in result, got: %s", result.RawText)
	}
}

func TestTailLogsToolEmptyQuery(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{
								{"message": "test"},
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
	result, err := server.ExecuteTool(ctx, "tail_logs", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("Tail logs failed: %v", err)
	}
	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}
}

func TestGetInfrastructureMetricsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{
								{
									"hostname":      "web-01",
									"cpuPercent":    45.2,
									"memoryPercent": 62.1,
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
	result, err := server.ExecuteTool(ctx, "get_infrastructure_metrics", map[string]interface{}{
		"account_id": "12345",
		"hostname":   "web-01",
		"metric_type": "cpu",
		"duration":   "30 minutes",
	})
	if err != nil {
		t.Fatalf("Get infrastructure metrics failed: %v", err)
	}
	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}
	if !contains(result.RawText, "web-01") {
		t.Errorf("Expected 'web-01' in result, got: %s", result.RawText)
	}
}

func TestGetInfrastructureMetricsToolNoArgs(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{
								{"hostname": "server-1", "cpuPercent": 10.0},
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
	result, err := server.ExecuteTool(ctx, "get_infrastructure_metrics", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("Get infra metrics failed: %v", err)
	}
	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}
}

func TestGetInfrastructureMetricsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{},
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
	result, err := server.ExecuteTool(ctx, "get_infrastructure_metrics", map[string]interface{}{
		"account_id": "12345",
		"hostname":   "nonexistent",
	})
	if err != nil {
		t.Fatalf("Get infra metrics failed: %v", err)
	}
	if !contains(result.RawText, "No infrastructure metrics") {
		t.Errorf("Expected 'No infrastructure metrics', got: %s", result.RawText)
	}
}

func TestListDashboardsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []map[string]interface{}{
								{
									"guid": "guid-123",
									"name": "Production Overview",
								},
								{
									"guid": "guid-456",
									"name": "API Monitoring",
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
	result, err := server.ExecuteTool(ctx, "list_dashboards", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("List dashboards failed: %v", err)
	}
	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}
	if !contains(result.RawText, "Production Overview") {
		t.Errorf("Expected 'Production Overview' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "API Monitoring") {
		t.Errorf("Expected 'API Monitoring' in result, got: %s", result.RawText)
	}
}

func TestListDashboardsToolNoResults(t *testing.T) {
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
	result, err := server.ExecuteTool(ctx, "list_dashboards", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("List dashboards failed: %v", err)
	}
	if !contains(result.RawText, "No dashboards") {
		t.Errorf("Expected 'No dashboards', got: %s", result.RawText)
	}
}

func TestGetDashboardDataTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": map[string]interface{}{
						"guid":        "guid-123",
						"name":        "Production Overview",
						"description": "Main production dashboard",
						"pages": []map[string]interface{}{
							{
								"widgets": []map[string]interface{}{
									{"id": "w1", "title": "CPU Usage"},
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
	result, err := server.ExecuteTool(ctx, "get_dashboard_data", map[string]interface{}{
		"account_id":     "12345",
		"dashboard_guid": "guid-123",
	})
	if err != nil {
		t.Fatalf("Get dashboard data failed: %v", err)
	}
	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}
	if !contains(result.RawText, "Production Overview") {
		t.Errorf("Expected 'Production Overview' in result, got: %s", result.RawText)
	}
}

func TestGetDashboardDataToolNotFound(t *testing.T) {
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
	result, err := server.ExecuteTool(ctx, "get_dashboard_data", map[string]interface{}{
		"account_id":     "12345",
		"dashboard_guid": "nonexistent-guid",
	})
	if err != nil {
		t.Fatalf("Get dashboard data failed: %v", err)
	}
	if !contains(result.RawText, "not found") {
		t.Errorf("Expected 'not found', got: %s", result.RawText)
	}
}

func TestGetApplicationMetricsEscapesAppName(t *testing.T) {
	var capturedQuery string
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		capturedQuery = string(buf[:n])
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"results": []map[string]interface{}{
								{"throughput": 100, "errorRate": 0.5},
							},
						},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()

	_, err := server.ExecuteTool(ctx, "get_application_metrics", map[string]interface{}{
		"account_id": "12345",
		"app_name":   "AppWith'Quote",
	})
	if err != nil {
		t.Fatalf("Get app metrics failed: %v", err)
	}

	if !strings.Contains(capturedQuery, "AppWith''Quote") {
		t.Errorf("Expected escaped app name (''), got query: %s", capturedQuery)
	}
	if strings.Contains(capturedQuery, "AppWith'Quote") && !strings.Contains(capturedQuery, "AppWith''Quote") {
		t.Errorf("App name with single quote not escaped, got query: %s", capturedQuery)
	}
}

func TestFormatValueComplexTypes(t *testing.T) {
	nested := map[string]interface{}{
		"nested_key": "value",
	}
	result := formatValue(nested)
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Errorf("Expected JSON for map, got: %s", result)
	}
	if decoded["nested_key"] != "value" {
		t.Errorf("Expected 'value', got: %v", decoded["nested_key"])
	}
	if contains(result, "map[") || contains(result, "%!(EXTRA") {
		t.Errorf("Result should not contain Go formatting artifacts, got: %s", result)
	}

	arr := []interface{}{"a", "b", "c"}
	result = formatValue(arr)
	var decodedArr []interface{}
	if err := json.Unmarshal([]byte(result), &decodedArr); err != nil {
		t.Errorf("Expected JSON array, got: %s", result)
	}
}

func TestExecuteNRQLQueryError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mockNR.Close()

	client := NewClientWithEndpoint("test-key", mockNR.URL)
	_, err := client.executeNRQL(context.Background(), "12345", "SELECT * FROM Transaction")
	if err == nil {
		t.Fatal("Expected error from HTTP 500")
	}
}

func TestExecuteNRQLNRQLErrorResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"nrql": map[string]interface{}{
							"error": "Query timed out",
							"results": []map[string]interface{}{},
						},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	client := NewClientWithEndpoint("test-key", mockNR.URL)
	_, err := client.executeNRQL(context.Background(), "12345", "SELECT * FROM Transaction")
	if err == nil {
		t.Fatal("Expected NRQL error")
	}
	if !contains(err.Error(), "NRQL error") {
		t.Errorf("Expected 'NRQL error' in message, got: %v", err)
	}
}

func TestNewServerWithEndpointWriteEnabled(t *testing.T) {
	server := NewServerWithEndpoint("test-key", "http://localhost:9999", true)
	if server == nil {
		t.Fatal("Expected server to be created")
	}
	if !server.IsWriteEnabled() {
		t.Error("Expected writeEnabled to be true")
	}
}

func TestListDashboardsToolMissingAccountID(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"accounts": []map[string]interface{}{
						{"id": float64(12345), "name": "Test Account"},
					},
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []map[string]interface{}{
								{"guid": "g1", "name": "Auto Dashboard"},
							},
						},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "list_dashboards", map[string]interface{}{})
	if err != nil {
		t.Fatalf("List dashboards with auto-detect failed: %v", err)
	}
	if result.RawText == "" {
		t.Error("Expected non-empty result")
	}
	if !contains(result.RawText, "Auto Dashboard") {
		t.Errorf("Expected 'Auto Dashboard' in result, got: %s", result.RawText)
	}
}
