package newrelic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListNotificationChannelsToolEmptyResponse(t *testing.T) {
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
	result, err := server.ExecuteTool(ctx, "list_notification_channels", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_notification_channels failed: %v", err)
	}
	if !contains(result.RawText, "No notification channels found") {
		t.Errorf("Expected 'No notification channels found', got: %s", result.RawText)
	}
}

func TestListNotificationChannelsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"aiNotifications": map[string]interface{}{
							"channels": map[string]interface{}{
								"entities": []map[string]interface{}{
									{"id": "c1", "name": "Slack Alerts"},
									{"id": "c2", "name": "Email Alerts"},
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
	result, err := server.ExecuteTool(ctx, "list_notification_channels", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("list_notification_channels failed: %v", err)
	}
	if !contains(result.RawText, "Slack Alerts") {
		t.Errorf("Expected 'Slack Alerts' in result, got: %s", result.RawText)
	}
	if !contains(result.RawText, "Email Alerts") {
		t.Errorf("Expected 'Email Alerts' in result, got: %s", result.RawText)
	}
}

func TestListNotificationChannelsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"aiNotifications": map[string]interface{}{
							"channels": map[string]interface{}{
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
	result, err := server.ExecuteTool(ctx, "list_notification_channels", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_notification_channels failed: %v", err)
	}
	if !contains(result.RawText, "No notification channels found") {
		t.Errorf("Expected 'No notification channels found', got: %s", result.RawText)
	}
}

func TestListDestinationsTool(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"aiNotifications": map[string]interface{}{
							"destinations": map[string]interface{}{
								"entities": []map[string]interface{}{
									{"id": "d1", "name": "Slack Workspace", "type": "SLACK"},
								},
								"totalCount": float64(1),
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
	result, err := server.ExecuteTool(ctx, "list_destinations", map[string]interface{}{
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("list_destinations failed: %v", err)
	}
	if !contains(result.RawText, "Slack Workspace") {
		t.Errorf("Expected 'Slack Workspace' in result, got: %s", result.RawText)
	}
}

func TestListDestinationsToolNoResults(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"actor": map[string]interface{}{
					"account": map[string]interface{}{
						"aiNotifications": map[string]interface{}{},
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
	result, err := server.ExecuteTool(ctx, "list_destinations", map[string]interface{}{
		"account_id": "12345",
	})
	if err != nil {
		t.Fatalf("list_destinations should not error on empty: %v", err)
	}
	if !contains(result.RawText, "No destinations found") {
		t.Errorf("Expected 'No destinations found', got: %s", result.RawText)
	}
}

func TestCreateSlackChannelWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiNotificationsCreateChannel": map[string]interface{}{
					"channel": map[string]interface{}{
						"id":   "chan-new",
						"name": "Slack Channel",
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_slack_channel", map[string]interface{}{
		"name":           "Slack Channel",
		"destination_id": "dest-1",
		"slack_channel_id": "C12345",
		"product":        "ALERTS",
		"account_id":     "12345",
	})

	if err != nil {
		t.Fatalf("create_slack_channel failed: %v", err)
	}
	if !contains(result.RawText, "Created Slack channel") {
		t.Errorf("Expected 'Created Slack channel' in result, got: %s", result.RawText)
	}
}

func TestCreateSlackChannelWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_slack_channel", map[string]interface{}{
		"name":           "Slack Channel",
		"destination_id": "dest-1",
		"slack_channel_id": "C12345",
		"product":        "ALERTS",
		"account_id":     "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestCreateSlackChannelMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_slack_channel", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing params")
	}
}

func TestCreateEmailChannelWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiNotificationsCreateChannel": map[string]interface{}{
					"channel": map[string]interface{}{
						"id":   "chan-email",
						"name": "Email Channel",
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "create_email_channel", map[string]interface{}{
		"name":           "Email Channel",
		"destination_id": "dest-2",
		"product":        "ALERTS",
		"account_id":     "12345",
	})

	if err != nil {
		t.Fatalf("create_email_channel failed: %v", err)
	}
	if !contains(result.RawText, "Created email channel") {
		t.Errorf("Expected 'Created email channel' in result, got: %s", result.RawText)
	}
}

func TestCreateEmailChannelWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_email_channel", map[string]interface{}{
		"name":           "Email Channel",
		"destination_id": "dest-2",
		"product":        "ALERTS",
		"account_id":     "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestCreateEmailChannelMissingParams(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_email_channel", map[string]interface{}{
		"name":       "Email Channel",
		"product":    "ALERTS",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing destination_id")
	}
}

func TestDeleteNotificationChannelWriteEnabled(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiNotificationsDeleteChannel": map[string]interface{}{
					"ids": []string{"chan-1"},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	result, err := server.ExecuteTool(ctx, "delete_notification_channel", map[string]interface{}{
		"channel_id": "chan-1",
		"account_id": "12345",
	})

	if err != nil {
		t.Fatalf("delete_notification_channel failed: %v", err)
	}
	if !contains(result.RawText, "Deleted notification channel") {
		t.Errorf("Expected 'Deleted notification channel' in result, got: %s", result.RawText)
	}
}

func TestDeleteNotificationChannelWriteDisabled(t *testing.T) {
	server := NewServer("test-key")
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_notification_channel", map[string]interface{}{
		"channel_id": "chan-1",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for write tool with writeEnabled=false")
	}
}

func TestCreateSlackChannelAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiNotificationsCreateChannel": map[string]interface{}{
					"channel": nil,
					"errors": []map[string]interface{}{
						{"type": "VALIDATION_ERROR", "description": "Invalid channel config"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_slack_channel", map[string]interface{}{
		"name":             "Slack Channel",
		"destination_id":   "dest-1",
		"slack_channel_id": "C12345",
		"product":          "ALERTS",
		"account_id":       "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API error response")
	}
}

func TestCreateEmailChannelAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiNotificationsCreateChannel": map[string]interface{}{
					"channel": nil,
					"errors": []map[string]interface{}{
						{"type": "VALIDATION_ERROR", "description": "Invalid email config"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "create_email_channel", map[string]interface{}{
		"name":           "Email Channel",
		"destination_id": "dest-2",
		"product":        "ALERTS",
		"account_id":     "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API error response")
	}
}

func TestDeleteNotificationChannelAPIError(t *testing.T) {
	mockNR := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"aiNotificationsDeleteChannel": map[string]interface{}{
					"ids": nil,
					"errors": []map[string]interface{}{
						{"type": "NOT_FOUND", "description": "Channel not found"},
					},
				},
			},
		})
	}))
	defer mockNR.Close()

	server := NewServerWithEndpoint("test-key", mockNR.URL, true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_notification_channel", map[string]interface{}{
		"channel_id": "bad-id",
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for API error response")
	}
}

func TestDeleteNotificationChannelMissingID(t *testing.T) {
	server := NewServer("test-key", true)
	ctx := context.Background()
	_, err := server.ExecuteTool(ctx, "delete_notification_channel", map[string]interface{}{
		"account_id": "12345",
	})
	if err == nil {
		t.Fatal("Expected error for missing channel_id")
	}
}
