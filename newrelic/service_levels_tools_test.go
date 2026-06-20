package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListServiceLevelsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": map[string]interface{}{
						"guid": "ENTITY-1",
						"serviceLevel": map[string]interface{}{
							"indicators": []map[string]interface{}{
								{"id": "sli-1", "name": "Error Budget", "sli": "valid/total"},
								{"id": "sli-2", "name": "Latency SLO", "sli": "fast/slow"},
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
	result, err := server.ExecuteTool(ctx, "list_service_levels", map[string]interface{}{
		"entity_guid": "ENTITY-1",
	})

	if err != nil {
		t.Fatalf("list_service_levels failed: %v", err)
	}
	if !contains(result.RawText, "Error Budget") {
		t.Errorf("Expected 'Error Budget' in result, got: %s", result.RawText)
	}
}

func TestListServiceLevelsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": map[string]interface{}{
						"guid":         "ENTITY-1",
						"serviceLevel": nil,
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
	result, err := server.ExecuteTool(ctx, "list_service_levels", map[string]interface{}{
		"entity_guid": "ENTITY-1",
	})
	if err != nil {
		t.Fatalf("list_service_levels failed: %v", err)
	}
	if !contains(result.RawText, "No service levels found") {
		t.Errorf("Expected 'No service levels found', got: %s", result.RawText)
	}
}

func TestListServiceLevelsToolMissingGUID(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "list_service_levels", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing entity_guid")
	}
}

func TestCreateServiceLevelWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"serviceLevelCreate": map[string]interface{}{
					"indicator": map[string]interface{}{
						"id":   "sli-new",
						"name": "Test SLO",
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_service_level", map[string]interface{}{
		"entity_guid":  "ENTITY-1",
		"name":         "Test SLO",
		"valid_events": "SELECT count(*) FROM Transaction",
		"good_events":  "SELECT count(*) FROM Transaction WHERE httpResponseCode = '200'",
		"target":       float64(99),
		"time_window":  "WEEK",
		"account_id":   "12345",
	})

	if err != nil {
		t.Fatalf("create_service_level failed: %v", err)
	}
	if !contains(result.RawText, "Created service level") {
		t.Errorf("Expected 'Created service level' in result, got: %s", result.RawText)
	}
}

func TestCreateServiceLevelWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_service_level", map[string]interface{}{
		"entity_guid":  "ENTITY-1",
		"name":         "Test SLO",
		"valid_events": "SELECT count(*) FROM Transaction",
		"good_events":  "SELECT count(*) FROM Transaction WHERE httpResponseCode = '200'",
		"account_id":   "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestCreateServiceLevelMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_service_level", map[string]interface{}{
		"entity_guid": "ENTITY-1",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing params")
	}
}

func TestUpdateServiceLevelWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"serviceLevelUpdate": map[string]interface{}{
					"indicator": map[string]interface{}{
						"id": "sli-1",
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "update_service_level", map[string]interface{}{
		"sli_id":      "sli-1",
		"target":      float64(99.9),
		"time_window": "WEEK",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("update_service_level failed: %v", err)
	}
	if !contains(result.RawText, "Updated service level") {
		t.Errorf("Expected 'Updated service level' in result, got: %s", result.RawText)
	}
}

func TestUpdateServiceLevelWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_service_level", map[string]interface{}{
		"sli_id": "sli-1",
		"target": float64(99),
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestUpdateServiceLevelMissingID(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_service_level", map[string]interface{}{
		"target": float64(99),
	})
	if err == nil {
		t.Fatal("Expected error for missing sli_id")
	}
}
