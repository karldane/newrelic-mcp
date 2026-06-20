package newrelic

import (
	"fmt"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type ListNotificationChannelsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListNotificationChannelsTool) Name() string        { return "list_notification_channels" }
func (t *ListNotificationChannelsTool) Description() string { return "List notification channels" }
func (t *ListNotificationChannelsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *ListNotificationChannelsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  aiNotifications {
			channels {
			  entities {
				id
				name
				type
			  }
			  totalCount
			}
		  }
		}
	  }
	}`, aid)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query notification channels: %w", err)
	}
	aiNotifs, _ := acct["aiNotifications"].(map[string]interface{})
	if aiNotifs == nil {
		return framework.TextResult("No notification channels found"), nil
	}
	channels, _ := aiNotifs["channels"].(map[string]interface{})
	if channels == nil {
		return framework.TextResult("No notification channels found"), nil
	}
	rawEntities, _ := channels["entities"].([]interface{})
	var items []map[string]interface{}
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}
	if len(items) == 0 {
		return framework.TextResult("No notification channels found"), nil
	}
	return framework.TextResult(formatResults(items)), nil
}
func (t *ListNotificationChannelsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}
func (t *ListNotificationChannelsTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entities": map[string]interface{}{
				"type":        "array",
				"description": "List of notification channels",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":   map[string]interface{}{"type": "string", "description": "Channel ID"},
						"name": map[string]interface{}{"type": "string", "description": "Channel name"},
						"type": map[string]interface{}{"type": "string", "description": "Channel type (e.g., SLACK, EMAIL)"},
					},
				},
			},
			"totalCount": map[string]interface{}{"type": "integer", "description": "Total number of notification channels"},
		},
	}
}

type ListDestinationsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListDestinationsTool) Name() string        { return "list_destinations" }
func (t *ListDestinationsTool) Description() string { return "List notification destinations" }
func (t *ListDestinationsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *ListDestinationsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  aiNotifications {
			destinations {
			  entities {
				id
				name
				type
				status
			  }
			  totalCount
			}
		  }
		}
	  }
	}`, aid)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query destinations: %w", err)
	}
	aiNotifs, _ := acct["aiNotifications"].(map[string]interface{})
	if aiNotifs == nil {
		return framework.TextResult("No destinations found"), nil
	}
	destinations, _ := aiNotifs["destinations"].(map[string]interface{})
	if destinations == nil {
		return framework.TextResult("No destinations found"), nil
	}
	rawEntities, _ := destinations["entities"].([]interface{})
	var items []map[string]interface{}
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}
	if len(items) == 0 {
		return framework.TextResult("No destinations found"), nil
	}
	return framework.TextResult(formatResults(items)), nil
}
func (t *ListDestinationsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}
func (t *ListDestinationsTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entities": map[string]interface{}{
				"type":        "array",
				"description": "List of notification destinations",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":     map[string]interface{}{"type": "string", "description": "Destination ID"},
						"name":   map[string]interface{}{"type": "string", "description": "Destination name"},
						"type":   map[string]interface{}{"type": "string", "description": "Destination type (e.g., SLACK, EMAIL)"},
						"status": map[string]interface{}{"type": "string", "description": "Destination status"},
					},
				},
			},
			"totalCount": map[string]interface{}{"type": "integer", "description": "Total number of destinations"},
		},
	}
}

type CreateSlackChannelTool struct {
	framework.BaseTool
	client *Client
}

func (t *CreateSlackChannelTool) Name() string        { return "create_slack_channel" }
func (t *CreateSlackChannelTool) Description() string { return "Create a Slack notification channel (requires --write-enabled)" }
func (t *CreateSlackChannelTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name":             map[string]interface{}{"type": "string", "description": "Channel name"},
			"destination_id":   map[string]interface{}{"type": "string", "description": "Destination ID for Slack"},
			"slack_channel_id": map[string]interface{}{"type": "string", "description": "Slack channel ID (e.g. C12345)"},
			"product":          map[string]interface{}{"type": "string", "description": "Product (e.g. ALERTS, ION, NR1)"},
			"account_id":       map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"name", "destination_id", "slack_channel_id"},
	}
}
func (t *CreateSlackChannelTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	name, _ := args["name"].(string)
	destID, _ := args["destination_id"].(string)
	slackChannelID, _ := args["slack_channel_id"].(string)
	if name == "" || destID == "" || slackChannelID == "" {
		return framework.TextResult(""), fmt.Errorf("name, destination_id, and slack_channel_id are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	product, _ := args["product"].(string)
	if product == "" {
		product = "ALERTS"
	}
	safeName := escapeString(name)
	gql := fmt.Sprintf(`mutation {
	  aiNotificationsCreateChannel(
		accountId: %s
		createChannelInput: {
		  name: "%s"
		  type: SLACK
		  product: %s
		  destinationId: "%s"
		  properties: [
			{ key: "channel_id", value: "%s" }
		  ]
		}
	  ) {
		channel {
		  id
		  name
		}
		errors {
		  type
		  description
		}
	  }
	}`, aid, safeName, product, destID, slackChannelID)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to create Slack channel: %w", err)
	}
	result, _ := data["aiNotificationsCreateChannel"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	errs, _ := result["errors"].([]interface{})
	if len(errs) > 0 {
		desc := "unknown"
		if errObj, ok := errs[0].(map[string]interface{}); ok {
			if d, ok2 := errObj["description"].(string); ok2 {
				desc = d
			}
		}
		return framework.TextResult(""), fmt.Errorf("API error: %s", desc)
	}
	return framework.TextResult(fmt.Sprintf("Created Slack channel: %s", name)), nil
}
func (t *CreateSlackChannelTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}
func (t *CreateSlackChannelTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"channel": map[string]interface{}{
				"type":        "object",
				"description": "Created Slack channel",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "string", "description": "Channel ID"},
					"name": map[string]interface{}{"type": "string", "description": "Channel name"},
				},
			},
			"errors": map[string]interface{}{
				"type":        "array",
				"description": "List of errors if creation failed",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":        map[string]interface{}{"type": "string", "description": "Error type"},
						"description": map[string]interface{}{"type": "string", "description": "Error description"},
					},
				},
			},
		},
	}
}

type CreateEmailChannelTool struct {
	framework.BaseTool
	client *Client
}

func (t *CreateEmailChannelTool) Name() string        { return "create_email_channel" }
func (t *CreateEmailChannelTool) Description() string { return "Create an email notification channel (requires --write-enabled)" }
func (t *CreateEmailChannelTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name":           map[string]interface{}{"type": "string", "description": "Channel name"},
			"destination_id": map[string]interface{}{"type": "string", "description": "Destination ID for email"},
			"product":        map[string]interface{}{"type": "string", "description": "Product (e.g. ALERTS, ION, NR1)"},
			"account_id":     map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"name", "destination_id"},
	}
}
func (t *CreateEmailChannelTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	name, _ := args["name"].(string)
	destID, _ := args["destination_id"].(string)
	if name == "" || destID == "" {
		return framework.TextResult(""), fmt.Errorf("name and destination_id are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	product, _ := args["product"].(string)
	if product == "" {
		product = "ALERTS"
	}
	safeName := escapeString(name)
	gql := fmt.Sprintf(`mutation {
	  aiNotificationsCreateChannel(
		accountId: %s
		createChannelInput: {
		  name: "%s"
		  type: EMAIL
		  product: %s
		  destinationId: "%s"
		}
	  ) {
		channel {
		  id
		  name
		}
		errors {
		  type
		  description
		}
	  }
	}`, aid, safeName, product, destID)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to create email channel: %w", err)
	}
	result, _ := data["aiNotificationsCreateChannel"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	errs, _ := result["errors"].([]interface{})
	if len(errs) > 0 {
		desc := "unknown"
		if errObj, ok := errs[0].(map[string]interface{}); ok {
			if d, ok2 := errObj["description"].(string); ok2 {
				desc = d
			}
		}
		return framework.TextResult(""), fmt.Errorf("API error: %s", desc)
	}
	return framework.TextResult(fmt.Sprintf("Created email channel: %s", name)), nil
}
func (t *CreateEmailChannelTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}
func (t *CreateEmailChannelTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"channel": map[string]interface{}{
				"type":        "object",
				"description": "Created email channel",
				"properties": map[string]interface{}{
					"id":   map[string]interface{}{"type": "string", "description": "Channel ID"},
					"name": map[string]interface{}{"type": "string", "description": "Channel name"},
				},
			},
			"errors": map[string]interface{}{
				"type":        "array",
				"description": "List of errors if creation failed",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":        map[string]interface{}{"type": "string", "description": "Error type"},
						"description": map[string]interface{}{"type": "string", "description": "Error description"},
					},
				},
			},
		},
	}
}

type DeleteNotificationChannelTool struct {
	framework.BaseTool
	client *Client
}

func (t *DeleteNotificationChannelTool) Name() string        { return "delete_notification_channel" }
func (t *DeleteNotificationChannelTool) Description() string { return "Delete a notification channel (requires --write-enabled)" }
func (t *DeleteNotificationChannelTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"channel_id": map[string]interface{}{"type": "string", "description": "Notification channel ID"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"channel_id"},
	}
}
func (t *DeleteNotificationChannelTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	channelID, _ := args["channel_id"].(string)
	if channelID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: channel_id")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`mutation {
	  aiNotificationsDeleteChannel(
		accountId: %s
		channelIds: ["%s"]
	  ) {
		ids
		errors {
		  type
		  description
		}
	  }
	}`, aid, channelID)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to delete notification channel: %w", err)
	}
	result, _ := data["aiNotificationsDeleteChannel"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	errs, _ := result["errors"].([]interface{})
	if len(errs) > 0 {
		desc := "unknown"
		if errObj, ok := errs[0].(map[string]interface{}); ok {
			if d, ok2 := errObj["description"].(string); ok2 {
				desc = d
			}
		}
		return framework.TextResult(""), fmt.Errorf("API error: %s", desc)
	}
	return framework.TextResult(fmt.Sprintf("Deleted notification channel: %s", channelID)), nil
}
func (t *DeleteNotificationChannelTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}
func (t *DeleteNotificationChannelTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"ids": map[string]interface{}{
				"type":        "array",
				"description": "List of deleted channel IDs",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"errors": map[string]interface{}{
				"type":        "array",
				"description": "List of errors if deletion failed",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type":        map[string]interface{}{"type": "string", "description": "Error type"},
						"description": map[string]interface{}{"type": "string", "description": "Error description"},
					},
				},
			},
		},
	}
}

