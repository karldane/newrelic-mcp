package newrelic

import (
	"fmt"
	"strings"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type ListSyntheticMonitorsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListSyntheticMonitorsTool) Name() string        { return "list_synthetic_monitors" }
func (t *ListSyntheticMonitorsTool) Description() string { return "List synthetic monitors" }
func (t *ListSyntheticMonitorsTool) Title() string        { return "Synthetic Monitors" }
func (t *ListSyntheticMonitorsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"limit":      map[string]interface{}{"type": "number", "description": "Max results"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *ListSyntheticMonitorsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	limit, _ := args["limit"].(float64)
	limitStr := ""
	if limit > 0 {
		limitStr = fmt.Sprintf(", limit: %.0f", limit)
	}
	gql := fmt.Sprintf(`{
	  actor {
		entitySearch(queryBuilder: {domain: SYNTH, type: MONITOR%s}) {
		  results {
			entities {
			  ... on SyntheticMonitorEntityOutline {
				guid
				name
				accountId
				monitorType
			  }
			}
		  }
		}
	  }
	}`, limitStr)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query synthetic monitors: %w", err)
	}
	entitySearch, _ := actor["entitySearch"].(map[string]interface{})
	if entitySearch == nil {
		return framework.TextResult("No synthetic monitors found"), nil
	}
	results, _ := entitySearch["results"].(map[string]interface{})
	if results == nil {
		return framework.TextResult("No synthetic monitors found"), nil
	}
	rawEntities, _ := results["entities"].([]interface{})
	var monitors []map[string]interface{}
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			monitors = append(monitors, m)
		}
	}
	if len(monitors) == 0 {
		return framework.TextResult("No synthetic monitors found"), nil
	}
	return framework.TextResult(formatResults(monitors)), nil
}
func (t *ListSyntheticMonitorsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}
func (t *ListSyntheticMonitorsTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entities": map[string]interface{}{
				"type":        "array",
				"description": "List of synthetic monitors",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"guid":        map[string]interface{}{"type": "string", "description": "Unique identifier for the monitor"},
						"name":        map[string]interface{}{"type": "string", "description": "Monitor name"},
						"accountId":   map[string]interface{}{"type": "string", "description": "Account ID the monitor belongs to"},
						"monitorType": map[string]interface{}{"type": "string", "description": "Type of monitor (e.g., SIMPLE, BROWSER, SCRIPT_API, SCRIPT_BROWSER)"},
					},
				},
			},
		},
	}
}

type GetSyntheticMonitorTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetSyntheticMonitorTool) Name() string        { return "get_synthetic_monitor" }
func (t *GetSyntheticMonitorTool) Description() string { return "Get a synthetic monitor by GUID" }
func (t *GetSyntheticMonitorTool) Title() string        { return "Monitor Details" }
func (t *GetSyntheticMonitorTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"monitor_guid": map[string]interface{}{"type": "string", "description": "GUID of the monitor"},
			"account_id":   map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"monitor_guid"},
	}
}
func (t *GetSyntheticMonitorTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	monitorGUID, _ := args["monitor_guid"].(string)
	if monitorGUID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: monitor_guid")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		entity(guid: "%s") {
		  ... on SyntheticMonitorEntityOutline {
			guid
			name
			accountId
			monitorType
		  }
		}
	  }
	}`, monitorGUID)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query monitor: %w", err)
	}
	entity, _ := actor["entity"].(map[string]interface{})
	if entity == nil {
		return framework.TextResult(fmt.Sprintf("Monitor with GUID '%s' not found", monitorGUID)), nil
	}
	return framework.TextResult(formatSingleResult(entity)), nil
}
func (t *GetSyntheticMonitorTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}
func (t *GetSyntheticMonitorTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"guid":        map[string]interface{}{"type": "string", "description": "Unique identifier for the monitor"},
			"name":        map[string]interface{}{"type": "string", "description": "Monitor name"},
			"accountId":   map[string]interface{}{"type": "string", "description": "Account ID the monitor belongs to"},
			"monitorType": map[string]interface{}{"type": "string", "description": "Type of monitor (e.g., SIMPLE, BROWSER, SCRIPT_API, SCRIPT_BROWSER)"},
		},
	}
}

type ListPrivateLocationsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListPrivateLocationsTool) Name() string        { return "list_private_locations" }
func (t *ListPrivateLocationsTool) Description() string { return "List private locations" }
func (t *ListPrivateLocationsTool) Title() string        { return "Private Locations" }
func (t *ListPrivateLocationsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"limit":      map[string]interface{}{"type": "number", "description": "Max results"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *ListPrivateLocationsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := `{
	  actor {
		entitySearch(queryBuilder: {domain: SYNTH, type: PRIVATE_LOCATION}) {
		  results {
			entities {
			  accountId
			  guid
			  name
			}
		  }
		}
	  }
	}`
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query private locations: %w", err)
	}
	entitySearch, _ := actor["entitySearch"].(map[string]interface{})
	if entitySearch == nil {
		return framework.TextResult("No private locations found"), nil
	}
	results, _ := entitySearch["results"].(map[string]interface{})
	if results == nil {
		return framework.TextResult("No private locations found"), nil
	}
	rawEntities, _ := results["entities"].([]interface{})
	var locations []map[string]interface{}
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			locations = append(locations, m)
		}
	}
	if len(locations) == 0 {
		return framework.TextResult("No private locations found"), nil
	}
	return framework.TextResult(formatResults(locations)), nil
}
func (t *ListPrivateLocationsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}
func (t *ListPrivateLocationsTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entities": map[string]interface{}{
				"type":        "array",
				"description": "List of private locations for synthetic monitoring",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"accountId": map[string]interface{}{"type": "string", "description": "Account ID the location belongs to"},
						"guid":      map[string]interface{}{"type": "string", "description": "Unique identifier for the private location"},
						"name":      map[string]interface{}{"type": "string", "description": "Private location name"},
					},
				},
			},
		},
	}
}

type CreatePingMonitorTool struct {
	framework.BaseTool
	client *Client
}

func (t *CreatePingMonitorTool) Name() string        { return "create_ping_monitor" }
func (t *CreatePingMonitorTool) Description() string { return "Create a ping monitor (requires --write-enabled)" }
func (t *CreatePingMonitorTool) Title() string        { return "Create Ping Monitor" }
func (t *CreatePingMonitorTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name":      map[string]interface{}{"type": "string", "description": "Monitor name"},
			"uri":       map[string]interface{}{"type": "string", "description": "URL to monitor"},
			"locations": map[string]interface{}{"type": "string", "description": "Comma-separated public locations (e.g. US_EAST_1,US_WEST_1)"},
			"period":    map[string]interface{}{"type": "string", "description": "Run period: EVERY_MINUTE, EVERY_5_MINUTES, EVERY_10_MINUTES, EVERY_15_MINUTES, EVERY_30_MINUTES, EVERY_HOUR, EVERY_6_HOURS, EVERY_12_HOURS, EVERY_DAY", "default": "EVERY_5_MINUTES"},
			"status":    map[string]interface{}{"type": "string", "description": "ENABLED or DISABLED", "default": "ENABLED"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"name", "uri"},
	}
}
func (t *CreatePingMonitorTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	name, _ := args["name"].(string)
	uri, _ := args["uri"].(string)
	if name == "" || uri == "" {
		return framework.TextResult(""), fmt.Errorf("name and uri are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	locations, _ := args["locations"].(string)
	locArray := `["US_EAST_1", "US_WEST_1"]`
	if locations != "" {
		parts := strings.Split(locations, ",")
		var quoted []string
		for _, p := range parts {
			quoted = append(quoted, fmt.Sprintf("%q", strings.TrimSpace(p)))
		}
		locArray = "[" + strings.Join(quoted, ",") + "]"
	}
	period, _ := args["period"].(string)
	if period == "" {
		period = "EVERY_5_MINUTES"
	}
	status, _ := args["status"].(string)
	if status == "" {
		status = "ENABLED"
	}
	uriEscaped := escapeString(uri)
	gql := fmt.Sprintf(`mutation {
	  syntheticsCreateSimpleMonitor(
		accountId: %s
		monitor: {
		  name: "%s"
		  uri: "%s"
		  locations: { public: %s }
		  period: %s
		  status: %s
		}
	  ) {
		errors {
		  type
		  description
		}
	  }
	}`, aid, escapeString(name), uriEscaped, locArray, period, status)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to create monitor: %w", err)
	}
	result, _ := data["syntheticsCreateSimpleMonitor"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	errs, _ := result["errors"].([]interface{})
	if len(errs) > 0 {
		if errObj, ok := errs[0].(map[string]interface{}); ok {
			desc, _ := errObj["description"].(string)
			return framework.TextResult(""), fmt.Errorf("API error: %s", desc)
		}
		return framework.TextResult(""), fmt.Errorf("API error creating monitor")
	}
	return framework.TextResult(fmt.Sprintf("Created ping monitor: %s", name)), nil
}
func (t *CreatePingMonitorTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}
func (t *CreatePingMonitorTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
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
			"message": map[string]interface{}{"type": "string", "description": "Success message with created monitor name"},
		},
	}
}

type DeleteSyntheticMonitorTool struct {
	framework.BaseTool
	client *Client
}

func (t *DeleteSyntheticMonitorTool) Name() string        { return "delete_synthetic_monitor" }
func (t *DeleteSyntheticMonitorTool) Description() string { return "Delete a synthetic monitor (requires --write-enabled)" }
func (t *DeleteSyntheticMonitorTool) Title() string        { return "Delete Monitor" }
func (t *DeleteSyntheticMonitorTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"monitor_guid": map[string]interface{}{"type": "string", "description": "GUID of the monitor to delete"},
			"account_id":   map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"monitor_guid"},
	}
}
func (t *DeleteSyntheticMonitorTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	monitorGUID, _ := args["monitor_guid"].(string)
	if monitorGUID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: monitor_guid")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`mutation {
	  syntheticsDeleteMonitor(guid: "%s") {
		deletedGuid
	  }
	}`, monitorGUID)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to delete monitor: %w", err)
	}
	result, _ := data["syntheticsDeleteMonitor"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	return framework.TextResult(fmt.Sprintf("Deleted monitor: %s", monitorGUID)), nil
}
func (t *DeleteSyntheticMonitorTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}
func (t *DeleteSyntheticMonitorTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"deletedGuid": map[string]interface{}{"type": "string", "description": "GUID of the deleted monitor"},
			"message":     map[string]interface{}{"type": "string", "description": "Confirmation message with deleted monitor GUID"},
		},
	}
}
