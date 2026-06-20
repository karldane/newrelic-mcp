package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListSyntheticMonitorsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []map[string]interface{}{
								{
									"guid":        "g1",
									"name":        "Ping Monitor",
									"accountId":   float64(12345),
									"monitorType": "SIMPLE",
								},
								{
									"guid":        "g2",
									"name":        "API Monitor",
									"accountId":   float64(12345),
									"monitorType": "SCRIPTED_API",
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
	result, err := server.ExecuteTool(ctx, "list_synthetic_monitors", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("list_synthetic_monitors failed: %v", err)
	}
	if !contains(result.RawText, "Ping Monitor") {
		t.Errorf("Expected 'Ping Monitor' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "API Monitor") {
		t.Errorf("Expected 'API Monitor' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "SIMPLE") {
		t.Errorf("Expected 'SIMPLE' in result, got: %s", result.RawText)
	}
}

func TestListSyntheticMonitorsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []interface{}{},
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
	result, err := server.ExecuteTool(ctx, "list_synthetic_monitors", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("list_synthetic_monitors failed: %v", err)
	}
	if !contains(result.RawText, "No synthetic monitors found") {
		t.Errorf("Expected 'No synthetic monitors found', got: %s", result.RawText)
	}
}

func TestGetSyntheticMonitorTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": map[string]interface{}{
						"guid":        "g1",
						"name":        "My Monitor",
						"accountId":   float64(12345),
						"monitorType": "SIMPLE",
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
	result, err := server.ExecuteTool(ctx, "get_synthetic_monitor", map[string]interface{}{
		"monitor_guid": "g1",
		"account_id":   "12345",
	})

	if err != nil {
		t.Fatalf("get_synthetic_monitor failed: %v", err)
	}
	if !contains(result.RawText, "My Monitor") {
		t.Errorf("Expected 'My Monitor' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "SIMPLE") {
		t.Errorf("Expected 'SIMPLE' in result, got: %s", result.RawText)
	}
}

func TestGetSyntheticMonitorToolNotFound(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": nil,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "get_synthetic_monitor", map[string]interface{}{
		"monitor_guid": "bad-guid",
		"account_id":   "12345",
	})

	if err != nil {
		t.Fatalf("get_synthetic_monitor should not error on not found: %v", err)
	}
	if !contains(result.RawText, "not found") {
		t.Errorf("Expected 'not found' in result, got: %s", result.RawText)
	}
}

func TestGetSyntheticMonitorToolMissingGUID(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_synthetic_monitor", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing monitor_guid")
	}
}

func TestListPrivateLocationsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []map[string]interface{}{
								{
									"guid":      "pl1",
									"name":      "My Private Location",
									"accountId": float64(12345),
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
	result, err := server.ExecuteTool(ctx, "list_private_locations", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("list_private_locations failed: %v", err)
	}
	if !contains(result.RawText, "My Private Location") {
		t.Errorf("Expected 'My Private Location' in result, got: %s", result.RawText)
	}
}

func TestCreatePingMonitorAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"syntheticsCreateSimpleMonitor": map[string]interface{}{
					"errors": []map[string]interface{}{
						{"type": "VALIDATION_ERROR", "description": "Invalid location"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_ping_monitor", map[string]interface{}{
		"name":       "Test Monitor",
		"uri":        "https://example.com",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API failure")
	}
	if !contains(err.Error(), "Invalid location") {
		t.Errorf("Expected 'Invalid location', got: %v", err)
	}
}

func TestCreatePingMonitorDefaultLocationsArray(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"syntheticsCreateSimpleMonitor": map[string]interface{}{
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_ping_monitor", map[string]interface{}{
		"name":       "Test Monitor",
		"uri":        "https://example.com",
		"account_id": "12345",
		// No locations specified - should use defaults
	})
	if err != nil {
		t.Fatalf("create_ping_monitor failed: %v", err)
	}
	if !contains(result.RawText, "Created ping monitor") {
		t.Errorf("Expected 'Created ping monitor' in result, got: %s", result.RawText)
	}
}

func TestCreatePingMonitorMultipleLocations(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"syntheticsCreateSimpleMonitor": map[string]interface{}{
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_ping_monitor", map[string]interface{}{
		"name":       "Test Monitor",
		"uri":        "https://example.com",
		"locations":  "US_EAST_1, US_WEST_1, EU_WEST_1",
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("create_ping_monitor failed: %v", err)
	}
	if !contains(result.RawText, "Created ping monitor") {
		t.Errorf("Expected 'Created ping monitor' in result, got: %s", result.RawText)
	}
}

func TestCreatePingMonitorWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"syntheticsCreateSimpleMonitor": map[string]interface{}{
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_ping_monitor", map[string]interface{}{
		"name":       "Test Monitor",
		"uri":        "https://example.com",
		"period":     "EVERY_5_MINUTES",
		"status":     "ENABLED",
		"locations":  "US_EAST_1",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("create_ping_monitor failed: %v", err)
	}
	if !contains(result.RawText, "Created ping monitor") {
		t.Errorf("Expected 'Created ping monitor' in result, got: %s", result.RawText)
	}
}

func TestCreatePingMonitorWriteDisabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_ping_monitor", map[string]interface{}{
		"name":       "Test Monitor",
		"uri":        "https://example.com",
		"account_id": "12345",
	})

	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
	expected := "write tools are disabled"
	if !contains(err.Error(), expected) {
		t.Errorf("Expected error containing '%s', got: %v", expected, err)
	}
}

func TestCreatePingMonitorMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_ping_monitor", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing name/uri")
	}
}

func TestListPrivateLocationsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []interface{}{},
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
	result, err := server.ExecuteTool(ctx, "list_private_locations", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_private_locations failed: %v", err)
	}
	if !contains(result.RawText, "No private locations found") {
		t.Errorf("Expected 'No private locations found', got: %s", result.RawText)
	}
}

func TestListSyntheticMonitorsEmptyEntitySearch(t *testing.T) {
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
	result, err := server.ExecuteTool(ctx, "list_synthetic_monitors", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_synthetic_monitors failed: %v", err)
	}
	if !contains(result.RawText, "No synthetic monitors found") {
		t.Errorf("Expected 'No synthetic monitors found', got: %s", result.RawText)
	}
}

func TestListSyntheticMonitorsWithLimit(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []map[string]interface{}{
								{
									"guid":        "g1",
									"name":        "Monitor 1",
									"accountId":   float64(12345),
									"monitorType": "SIMPLE",
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
	result, err := server.ExecuteTool(ctx, "list_synthetic_monitors", map[string]interface{}{
		"account_id": "12345",
		"limit":      float64(10),
	})
	if err != nil {
		t.Fatalf("list_synthetic_monitors failed: %v", err)
	}
	if !contains(result.RawText, "Monitor 1") {
		t.Errorf("Expected 'Monitor 1' in result, got: %s", result.RawText)
	}
}

func TestDeleteSyntheticMonitorWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"syntheticsDeleteMonitor": map[string]interface{}{
					"deletedGuid": "g1",
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "delete_synthetic_monitor", map[string]interface{}{
		"monitor_guid": "g1",
		"account_id":   "12345",
	})

	if err != nil {
		t.Fatalf("delete_synthetic_monitor failed: %v", err)
	}
	if !contains(result.RawText, "Deleted monitor") {
		t.Errorf("Expected 'Deleted monitor' in result, got: %s", result.RawText)
	}
}

func TestDeleteSyntheticMonitorWriteDisabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_synthetic_monitor", map[string]interface{}{
		"monitor_guid": "g1",
		"account_id":   "12345",
	})

	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestDeleteSyntheticMonitorAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"syntheticsDeleteMonitor": nil,
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_synthetic_monitor", map[string]interface{}{
		"monitor_guid": "g1",
		"account_id":   "12345",
	})
	if err == nil {
		t.Fatal("Expected error for nil response")
	}
}

func TestCreatePingMonitorUnexpectedResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"someOtherField": "value",
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_ping_monitor", map[string]interface{}{
		"name":       "Test",
		"uri":        "https://example.com",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for unexpected response")
	}
}

func TestDeleteSyntheticMonitorMissingGUID(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_synthetic_monitor", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing monitor_guid")
	}
}
