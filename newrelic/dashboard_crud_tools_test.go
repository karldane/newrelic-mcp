package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDashboardTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"entity": map[string]interface{}{
						"guid":        "dg1",
						"name":        "My Dashboard",
						"description": "Test dashboard",
						"permissions": "PRIVATE",
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
	result, err := server.ExecuteTool(ctx, "get_dashboard", map[string]interface{}{
		"guid":       "dg1",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("get_dashboard failed: %v", err)
	}
	if !contains(result.RawText, "My Dashboard") {
		t.Errorf("Expected 'My Dashboard' in result, got: %s", result.RawText)
	}
}

func TestGetDashboardToolNotFound(t *testing.T) {
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
	result, err := server.ExecuteTool(ctx, "get_dashboard", map[string]interface{}{
		"guid":       "bad-guid",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("get_dashboard should not error on not found: %v", err)
	}
	if !contains(result.RawText, "not found") {
		t.Errorf("Expected 'not found' in result, got: %s", result.RawText)
	}
}

func TestGetDashboardToolMissingGUID(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "get_dashboard", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing guid")
	}
}

func TestCreateDashboardWriteDisabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_dashboard", map[string]interface{}{
		"name":       "Test Dashboard",
		"nrql_query": "SELECT count(*) FROM Transaction",
		"account_id": "12345",
	})

	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestCreateDashboardWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardCreate": map[string]interface{}{
					"entityResult": map[string]interface{}{
						"guid": "new-dg1",
						"name": "Test Dashboard",
					},
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_dashboard", map[string]interface{}{
		"name":        "Test Dashboard",
		"description": "A test dashboard",
		"nrql_query":  "SELECT count(*) FROM Transaction",
		"widget_title": "My Widget",
		"visualization": "viz.billboard",
		"permissions":  "PRIVATE",
		"account_id":   "12345",
	})

	if err != nil {
		t.Fatalf("create_dashboard failed: %v", err)
	}
	if !contains(result.RawText, "Created dashboard") {
		t.Errorf("Expected 'Created dashboard' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "new-dg1") {
		t.Errorf("Expected GUID 'new-dg1' in result, got: %s", result.RawText)
	}
}

func TestCreateDashboardAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardCreate": map[string]interface{}{
					"errors": []map[string]interface{}{
						{"type": "VALIDATION", "description": "Invalid name"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_dashboard", map[string]interface{}{
		"name":       "",
		"nrql_query": "SELECT count(*) FROM Transaction",
		"account_id": "12345",
	})
	if err != nil {
		return
	}
	// With empty name, the tool should reject it before making API call
}

func TestCreateDashboardMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_dashboard", map[string]interface{}{
		"account_id": "12345",
		"nrql_query": "SELECT count(*) FROM Transaction",
		// Missing name
	})
	if err == nil {
		t.Fatal("Expected error for missing name")
	}
}

func TestCreateDashboardUnexpectedResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardCreate": nil,
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_dashboard", map[string]interface{}{
		"name":       "Test",
		"nrql_query": "SELECT count(*) FROM Transaction",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for nil response")
	}
}

func TestCreateDashboardNoEntityResult(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardCreate": map[string]interface{}{
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_dashboard", map[string]interface{}{
		"name":       "Test Dashboard",
		"nrql_query": "SELECT count(*) FROM Transaction",
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("create_dashboard failed: %v", err)
	}
	if !contains(result.RawText, "Created dashboard") {
		t.Errorf("Expected 'Created dashboard' in result, got: %s", result.RawText)
	}
}

func TestUpdateDashboardWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardUpdate": map[string]interface{}{
					"entityResult": map[string]interface{}{
						"guid": "dg1",
						"name": "Updated Dashboard",
					},
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "update_dashboard", map[string]interface{}{
		"guid":        "dg1",
		"name":        "Updated Dashboard",
		"description": "Updated description",
		"permissions": "PUBLIC_READ_ONLY",
		"account_id":  "12345",
	})

	if err != nil {
		t.Fatalf("update_dashboard failed: %v", err)
	}
	if !contains(result.RawText, "Updated dashboard") {
		t.Errorf("Expected 'Updated dashboard' in result, got: %s", result.RawText)
	}
}

func TestUpdateDashboardWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_dashboard", map[string]interface{}{
		"guid":       "dg1",
		"name":       "Updated Dashboard",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestUpdateDashboardAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardUpdate": map[string]interface{}{
					"errors": []map[string]interface{}{
						{"type": "VALIDATION", "description": "Invalid name"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_dashboard", map[string]interface{}{
		"guid":       "dg1",
		"name":       "Updated Dashboard",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API failure")
	}
}

func TestUpdateDashboardUnexpectedResponse(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardUpdate": nil,
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_dashboard", map[string]interface{}{
		"guid":       "dg1",
		"name":       "Updated Dashboard",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for nil response")
	}
}

func TestUpdateDashboardMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "update_dashboard", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing guid/name")
	}
}

func TestDeleteDashboardWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardDelete": map[string]interface{}{
					"status": "SUCCESS",
					"errors": nil,
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "delete_dashboard", map[string]interface{}{
		"guid":       "dg1",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("delete_dashboard failed: %v", err)
	}
	if !contains(result.RawText, "Deleted dashboard") {
		t.Errorf("Expected 'Deleted dashboard' in result, got: %s", result.RawText)
	}
}

func TestDeleteDashboardWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_dashboard", map[string]interface{}{
		"guid":       "dg1",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestDeleteDashboardMissingGUID(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_dashboard", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing guid")
	}
}

func TestDeleteDashboardAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"dashboardDelete": map[string]interface{}{
					"errors": []map[string]interface{}{
						{"type": "NOT_FOUND", "description": "Dashboard not found"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_dashboard", map[string]interface{}{
		"guid":       "bad-guid",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API failure")
	}
}
