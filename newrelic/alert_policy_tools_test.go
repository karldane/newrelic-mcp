package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAlertPolicyTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"alerts": map[string]interface{}{
							"policy": map[string]interface{}{
								"id":                 "123",
								"name":               "My Policy",
								"incidentPreference": "PER_CONDITION",
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
	result, err := server.ExecuteTool(ctx, "get_alert_policy", map[string]interface{}{
		"policy_id":  "123",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("get_alert_policy failed: %v", err)
	}
	if !contains(result.RawText, "My Policy") {
		t.Errorf("Expected 'My Policy' in result, got: %s", result.RawText)
	}
}

func TestGetAlertPolicyToolMissingPolicyID(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_alert_policy", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing policy_id")
	}
}

func TestGetAlertPolicyToolNotFound(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"alerts": map[string]interface{}{},
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
	result, err := server.ExecuteTool(ctx, "get_alert_policy", map[string]interface{}{
		"policy_id":  "999",
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("get_alert_policy should not error on not found: %v", err)
	}
	if !contains(result.RawText, "No alert policy found") {
		t.Errorf("Expected 'No alert policy found', got: %s", result.RawText)
	}
}

func TestCreateAlertPolicyWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsPolicyCreate": map[string]interface{}{
					"id":                 "456",
					"name":               "New Policy",
					"incidentPreference": "PER_CONDITION",
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_alert_policy", map[string]interface{}{
		"name":       "New Policy",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("create_alert_policy failed: %v", err)
	}
	if !contains(result.RawText, "Created alert policy") {
		t.Errorf("Expected 'Created alert policy' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "456") {
		t.Errorf("Expected ID '456' in result, got: %s", result.RawText)
	}
}

func TestCreateAlertPolicyWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_alert_policy", map[string]interface{}{
		"name":       "New Policy",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestCreateAlertPolicyMissingName(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_alert_policy", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing name")
	}
}

func TestCreateAlertPolicyUnexpectedResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsPolicyCreate": nil,
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_alert_policy", map[string]interface{}{
		"name":       "New Policy",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for nil response")
	}
}

func TestUpdateAlertPolicyWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsPolicyUpdate": map[string]interface{}{
					"id":                 "123",
					"name":               "Updated Policy",
					"incidentPreference": "PER_POLICY",
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "update_alert_policy", map[string]interface{}{
		"policy_id":  "123",
		"name":       "Updated Policy",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("update_alert_policy failed: %v", err)
	}
	if !contains(result.RawText, "Updated alert policy") {
		t.Errorf("Expected 'Updated alert policy' in result, got: %s", result.RawText)
	}
}

func TestUpdateAlertPolicyWithPreference(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsPolicyUpdate": map[string]interface{}{
					"id":                 "123",
					"name":               "Updated Policy",
					"incidentPreference": "PER_POLICY",
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "update_alert_policy", map[string]interface{}{
		"policy_id":           "123",
		"name":                "Updated Policy",
		"incident_preference": "PER_POLICY",
		"account_id":          "12345",
	})
	if err != nil {
		t.Fatalf("update_alert_policy failed: %v", err)
	}
	if !contains(result.RawText, "Updated alert policy") {
		t.Errorf("Expected 'Updated alert policy' in result, got: %s", result.RawText)
	}
}

func TestUpdateAlertPolicyWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_alert_policy", map[string]interface{}{
		"policy_id":  "123",
		"name":       "Updated Policy",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestUpdateAlertPolicyMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_alert_policy", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing params")
	}
}

func TestDeleteAlertPolicyWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsPolicyDelete": map[string]interface{}{
					"id": "123",
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "delete_alert_policy", map[string]interface{}{
		"policy_id":  "123",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("delete_alert_policy failed: %v", err)
	}
	if !contains(result.RawText, "Deleted alert policy") {
		t.Errorf("Expected 'Deleted alert policy' in result, got: %s", result.RawText)
	}
}

func TestDeleteAlertPolicyWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_alert_policy", map[string]interface{}{
		"policy_id":  "123",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestDeleteAlertPolicyMissingPolicyID(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_alert_policy", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing policy_id")
	}
}

func TestDeleteAlertPolicyUnexpectedResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsPolicyDelete": nil,
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_alert_policy", map[string]interface{}{
		"policy_id":  "999",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for nil response")
	}
}

func TestListNRQLAlertConditionsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"alerts": map[string]interface{}{
							"nrqlConditionsSearch": map[string]interface{}{
								"nrqlConditions": []map[string]interface{}{
									{
										"id":      "c1",
										"name":    "High Error Rate",
										"enabled": true,
										"nrql":    map[string]interface{}{"query": "SELECT count(*) FROM Transaction WHERE error IS TRUE"},
										"critical": map[string]interface{}{
											"thresholdDuration": 300,
											"duration":          float64(5),
										},
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
	result, err := server.ExecuteTool(ctx, "list_nrql_alert_conditions", map[string]interface{}{
		"policy_id":  "123",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("list_nrql_alert_conditions failed: %v", err)
	}
	if !contains(result.RawText, "High Error Rate") {
		t.Errorf("Expected 'High Error Rate' in result, got: %s", result.RawText)
	}
}

func TestListNRQLAlertConditionsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"alerts": map[string]interface{}{
							"nrqlConditionsSearch": map[string]interface{}{
								"nrqlConditions": []interface{}{},
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
	result, err := server.ExecuteTool(ctx, "list_nrql_alert_conditions", map[string]interface{}{
		"policy_id":  "999",
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_nrql_alert_conditions failed: %v", err)
	}
	if !contains(result.RawText, "No alert conditions found") {
		t.Errorf("Expected 'No alert conditions found', got: %s", result.RawText)
	}
}

func TestListNRQLAlertConditionsToolMissingPolicyID(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "list_nrql_alert_conditions", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing policy_id")
	}
}

func TestRealCreateAlertConditionWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsNrqlConditionCreate": map[string]interface{}{
					"id":      "cond-1",
					"name":    "Test Condition",
					"enabled": true,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_alert_condition", map[string]interface{}{
		"policy_id":          "123",
		"name":               "Test Condition",
		"nrql_query":         "SELECT count(*) FROM Transaction",
		"critical_threshold": float64(10),
		"account_id":         "12345",
	})

	if err != nil {
		t.Fatalf("create_alert_condition failed: %v", err)
	}
	if !contains(result.RawText, "Created alert condition") {
		t.Errorf("Expected 'Created alert condition' in result, got: %s", result.RawText)
	}
}

func TestRealCreateAlertConditionWithWarning(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsNrqlConditionCreate": map[string]interface{}{
					"id":      "cond-2",
					"name":    "Test With Warning",
					"enabled": true,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_alert_condition", map[string]interface{}{
		"policy_id":          "123",
		"name":               "Test With Warning",
		"nrql_query":         "SELECT count(*) FROM Transaction",
		"critical_threshold": float64(10),
		"warning_threshold":  float64(5),
		"duration_minutes":   float64(10),
		"account_id":         "12345",
	})
	if err != nil {
		t.Fatalf("create_alert_condition failed: %v", err)
	}
	if !contains(result.RawText, "Created alert condition") {
		t.Errorf("Expected 'Created alert condition' in result, got: %s", result.RawText)
	}
}

func TestRealCreateAlertConditionWriteDisabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_alert_condition", map[string]interface{}{
		"policy_id":          "123",
		"name":               "Test Condition",
		"nrql_query":         "SELECT count(*) FROM Transaction",
		"critical_threshold": float64(10),
		"account_id":         "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestRealCreateAlertConditionMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_alert_condition", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing params")
	}
}

func TestRealCreateAlertConditionUnexpectedResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"alertsNrqlConditionCreate": nil,
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_alert_condition", map[string]interface{}{
		"policy_id":          "123",
		"name":               "Test",
		"nrql_query":         "SELECT count(*) FROM Transaction",
		"critical_threshold": float64(10),
		"account_id":         "12345",
	})
	if err == nil {
		t.Fatal("Expected error for nil response")
	}
}
