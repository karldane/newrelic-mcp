package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListWorkflowsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"aiWorkflows": map[string]interface{}{
							"workflows": map[string]interface{}{
								"entities": []map[string]interface{}{
									{"id": "w1", "name": "On-Call Workflow", "workflowEnabled": true},
									{"id": "w2", "name": "Slack Notify", "workflowEnabled": false},
								},
								"totalCount": float64(2),
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
	result, err := server.ExecuteTool(ctx, "list_workflows", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("list_workflows failed: %v", err)
	}
	if !contains(result.RawText, "On-Call Workflow") {
		t.Errorf("Expected 'On-Call Workflow' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "Slack Notify") {
		t.Errorf("Expected 'Slack Notify' in result, got: %s", result.RawText)
	}
}

func TestListWorkflowsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"aiWorkflows": map[string]interface{}{
							"workflows": map[string]interface{}{
								"entities": []interface{}{},
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
	result, err := server.ExecuteTool(ctx, "list_workflows", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_workflows failed: %v", err)
	}
	if !contains(result.RawText, "No workflows found") {
		t.Errorf("Expected 'No workflows found', got: %s", result.RawText)
	}
}

func TestGetWorkflowTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"aiWorkflows": map[string]interface{}{
							"workflows": map[string]interface{}{
								"entities": []map[string]interface{}{
									{"id": "w1", "name": "My Workflow", "workflowEnabled": true},
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
	result, err := server.ExecuteTool(ctx, "get_workflow", map[string]interface{}{
		"workflow_id": "w1",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("get_workflow failed: %v", err)
	}
	if !contains(result.RawText, "My Workflow") {
		t.Errorf("Expected 'My Workflow' in result, got: %s", result.RawText)
	}
}

func TestGetWorkflowToolNotFound(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"aiWorkflows": map[string]interface{}{
							"workflows": map[string]interface{}{
								"entities": []interface{}{},
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
	result, err := server.ExecuteTool(ctx, "get_workflow", map[string]interface{}{
		"workflow_id": "bad-id",
		"account_id":  "12345",
	})
	if err != nil {
		t.Fatalf("get_workflow should not error on not found: %v", err)
	}
	if !contains(result.RawText, "not found") {
		t.Errorf("Expected 'not found' in result, got: %s", result.RawText)
	}
}

func TestGetWorkflowToolMissingID(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_workflow", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing workflow_id")
	}
}

func TestCreateWorkflowWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiWorkflowsCreateWorkflow": map[string]interface{}{
					"workflow": map[string]interface{}{
						"id":   "w-new",
						"name": "API Workflow",
					},
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_workflow", map[string]interface{}{
		"name":        "API Workflow",
		"channel_ids": "chan-1",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("create_workflow failed: %v", err)
	}
	if !contains(result.RawText, "Created workflow") {
		t.Errorf("Expected 'Created workflow' in result, got: %s", result.RawText)
	}
}

func TestCreateWorkflowWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_workflow", map[string]interface{}{
		"name":        "API Workflow",
		"channel_ids": "chan-1",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestCreateWorkflowMissingName(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_workflow", map[string]interface{}{
		"channel_ids": "chan-1",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing name")
	}
}

func TestUpdateWorkflowWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiWorkflowsUpdateWorkflow": map[string]interface{}{
					"workflow": map[string]interface{}{
						"id":   "w1",
						"name": "Updated Workflow",
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "update_workflow", map[string]interface{}{
		"workflow_id": "w1",
		"name":        "Updated Workflow",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("update_workflow failed: %v", err)
	}
	if !contains(result.RawText, "Updated workflow") {
		t.Errorf("Expected 'Updated workflow' in result, got: %s", result.RawText)
	}
}

func TestUpdateWorkflowWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_workflow", map[string]interface{}{
		"workflow_id": "w1",
		"name":        "Updated Workflow",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestUpdateWorkflowMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_workflow", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing params")
	}
}

func TestDeleteWorkflowWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiWorkflowsDeleteWorkflow": map[string]interface{}{
					"id": "w1",
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "delete_workflow", map[string]interface{}{
		"workflow_id": "w1",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("delete_workflow failed: %v", err)
	}
	if !contains(result.RawText, "Deleted workflow") {
		t.Errorf("Expected 'Deleted workflow' in result, got: %s", result.RawText)
	}
}

func TestDeleteWorkflowWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_workflow", map[string]interface{}{
		"workflow_id": "w1",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestCreateWorkflowAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiWorkflowsCreateWorkflow": map[string]interface{}{
					"workflow": nil,
					"errors": []map[string]interface{}{
						{"type": "VALIDATION_ERROR", "description": "Invalid workflow config"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_workflow", map[string]interface{}{
		"name":        "Bad Workflow",
		"channel_ids": "chan-1",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API error response")
	}
}

func TestUpdateWorkflowAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiWorkflowsUpdateWorkflow": map[string]interface{}{
					"workflow": nil,
					"errors": []map[string]interface{}{
						{"type": "NOT_FOUND", "description": "Workflow not found"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_workflow", map[string]interface{}{
		"workflow_id": "bad-id",
		"name":        "Updated",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API error response")
	}
}

func TestDeleteWorkflowAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiWorkflowsDeleteWorkflow": map[string]interface{}{
					"id": nil,
					"errors": []map[string]interface{}{
						{"type": "NOT_FOUND", "description": "Workflow not found"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_workflow", map[string]interface{}{
		"workflow_id": "bad-id",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API error response")
	}
}

func TestCreateWorkflowUnexpectedResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiWorkflowsCreateWorkflow": nil,
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_workflow", map[string]interface{}{
		"name":        "Test",
		"channel_ids": "chan-1",
		"account_id":  "12345",
	})
	if err == nil {
		t.Fatal("Expected error for unexpected response")
	}
}

func TestDeleteWorkflowMissingID(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_workflow", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing workflow_id")
	}
}
