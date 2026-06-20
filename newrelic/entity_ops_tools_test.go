package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchEntitiesTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entitySearch": map[string]interface{}{
						"results": map[string]interface{}{
							"entities": []map[string]interface{}{
								{"guid": "e1", "name": "My App", "entityType": "APM_APPLICATION"},
								{"guid": "e2", "name": "My DB", "entityType": "INFRASTRUCTURE_HOST"},
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
	result, err := server.ExecuteTool(ctx, "search_entities", map[string]interface{}{
		"query": "name = 'My App'",
	})

	if err != nil {
		t.Fatalf("search_entities failed: %v", err)
	}
	if !contains(result.RawText, "My App") {
		t.Errorf("Expected 'My App' in result, got: %s", result.RawText)
	}
}

func TestSearchEntitiesToolNoResults(t *testing.T) {
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
	result, err := server.ExecuteTool(ctx, "search_entities", map[string]interface{}{
		"query": "name = 'Nothing'",
	})
	if err != nil {
		t.Fatalf("search_entities should not error on empty: %v", err)
	}
	if !contains(result.RawText, "No entities found") {
		t.Errorf("Expected 'No entities found', got: %s", result.RawText)
	}
}

func TestSearchEntitiesToolMissingQuery(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "search_entities", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing query")
	}
}

func TestDeleteEntityWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"entityDelete": map[string]interface{}{
					"deleted": []interface{}{"e1", "e2"},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "delete_entity", map[string]interface{}{
		"guids":      "e1,e2",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("delete_entity failed: %v", err)
	}
	if !contains(result.RawText, "Deleted entities") {
		t.Errorf("Expected 'Deleted entities' in result, got: %s", result.RawText)
	}
}

func TestDeleteEntityWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_entity", map[string]interface{}{
		"guids": "e1",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestDeleteEntityMissingGUIDs(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_entity", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing guids")
	}
}

func TestGetEntityTagsToolEntityNotFound(t *testing.T) {
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
	result, err := server.ExecuteTool(ctx, "get_entity_tags", map[string]interface{}{
		"entity_guid": "bad-guid",
	})
	if err != nil {
		t.Fatalf("get_entity_tags should not error on not found: %v", err)
	}
	if !contains(result.RawText, "not found") {
		t.Errorf("Expected 'not found' in result, got: %s", result.RawText)
	}
}
