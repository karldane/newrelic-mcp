package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEntityTagsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": map[string]interface{}{
						"guid": "ENTITY-1",
						"tags": []map[string]interface{}{
							{"key": "env", "values": []string{"production"}},
							{"key": "team", "values": []string{"platform"}},
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
	result, err := server.ExecuteTool(ctx, "get_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
	})

	if err != nil {
		t.Fatalf("get_entity_tags failed: %v", err)
	}
	if !contains(result.RawText, "env") {
		t.Errorf("Expected 'env' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "team") {
		t.Errorf("Expected 'team' in result, got: %s", result.RawText)
	}
}

func TestGetEntityTagsToolNoTags(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": map[string]interface{}{
						"guid": "ENTITY-1",
						"tags": []interface{}{},
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
	result, err := server.ExecuteTool(ctx, "get_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
	})
	if err != nil {
		t.Fatalf("get_entity_tags failed: %v", err)
	}
	if !contains(result.RawText, "No tags found") {
		t.Errorf("Expected 'No tags found', got: %s", result.RawText)
	}
}

func TestGetEntityTagsToolMissingGUID(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_entity_tags", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error for missing entity_guid")
	}
}

func TestAddEntityTagsWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"taggingAddTagsToEntity": map[string]interface{}{
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "add_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
		"tags":        "env=staging,team=infra",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("add_entity_tags failed: %v", err)
	}
	if !contains(result.RawText, "Added tags") {
		t.Errorf("Expected 'Added tags' in result, got: %s", result.RawText)
	}
}

func TestAddEntityTagsWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "add_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
		"tags":        "env=staging",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestAddEntityTagsMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "add_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
	})
	if err == nil {
		t.Fatal("Expected error for missing tags param")
	}
}

func TestRemoveEntityTagsWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"taggingDeleteTagFromEntity": map[string]interface{}{
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "remove_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
		"tag_keys":    "env,team",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("remove_entity_tags failed: %v", err)
	}
	if !contains(result.RawText, "Removed tags") {
		t.Errorf("Expected 'Removed tags' in result, got: %s", result.RawText)
	}
}

func TestRemoveEntityTagsWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "remove_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
		"tag_keys":    "env",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestRemoveEntityTagsMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "remove_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
	})
	if err == nil {
		t.Fatal("Expected error for missing tag_keys")
	}
}

func TestReplaceEntityTagsWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"taggingReplaceTagsOnEntity": map[string]interface{}{
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "replace_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
		"tags":        "env=prod,team=sre",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("replace_entity_tags failed: %v", err)
	}
	if !contains(result.RawText, "Replaced tags") {
		t.Errorf("Expected 'Replaced tags' in result, got: %s", result.RawText)
	}
}

func TestReplaceEntityTagsWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "replace_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
		"tags":        "env=prod",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestReplaceEntityTagsMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "replace_entity_tags", map[string]interface{}{
		"entity_guid": "ENTITY-1",
	})
	if err == nil {
		t.Fatal("Expected error for missing tags param")
	}
}
