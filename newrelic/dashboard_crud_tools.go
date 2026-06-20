package newrelic

import (
	"fmt"
	"strings"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type GetDashboardTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetDashboardTool) Name() string        { return "get_dashboard" }
func (t *GetDashboardTool) Description() string { return "Get full dashboard configuration by GUID" }
func (t *GetDashboardTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"guid":       map[string]interface{}{"type": "string", "description": "Dashboard GUID"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"guid"},
	}
}
func (t *GetDashboardTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	guid, _ := args["guid"].(string)
	if guid == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: guid")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		entity(guid: "%s") {
		  ... on DashboardEntity {
			guid
			name
			description
			permissions
			pages {
			  guid
			  name
			  widgets {
				id
				title
				visualization { id }
			  }
			}
		  }
		}
	  }
	}`, guid)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query dashboard: %w", err)
	}
	entity, _ := actor["entity"].(map[string]interface{})
	if entity == nil {
		return framework.TextResult(fmt.Sprintf("Dashboard with GUID '%s' not found", guid)), nil
	}
	return framework.TextResult(formatSingleResult(entity)), nil
}
func (t *GetDashboardTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}

type CreateDashboardTool struct {
	framework.BaseTool
	client *Client
}

func (t *CreateDashboardTool) Name() string        { return "create_dashboard" }
func (t *CreateDashboardTool) Description() string { return "Create a new dashboard (requires --write-enabled)" }
func (t *CreateDashboardTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name":          map[string]interface{}{"type": "string", "description": "Dashboard name"},
			"description":   map[string]interface{}{"type": "string", "description": "Dashboard description"},
			"permissions":   map[string]interface{}{"type": "string", "description": "PUBLIC_READ_WRITE, PUBLIC_READ_ONLY, or PRIVATE", "default": "PRIVATE"},
			"nrql_query":    map[string]interface{}{"type": "string", "description": "NRQL query for the first widget"},
			"widget_title":  map[string]interface{}{"type": "string", "description": "Title for the first widget", "default": "Query"},
			"visualization": map[string]interface{}{"type": "string", "description": "viz.billboard, viz.line, viz.table, etc.", "default": "viz.billboard"},
			"account_id":    map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"name", "nrql_query"},
	}
}
func (t *CreateDashboardTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	name, _ := args["name"].(string)
	nrqlQuery, _ := args["nrql_query"].(string)
	if name == "" || nrqlQuery == "" {
		return framework.TextResult(""), fmt.Errorf("name and nrql_query are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	desc, _ := args["description"].(string)
	permissions, _ := args["permissions"].(string)
	if permissions == "" {
		permissions = "PRIVATE"
	}
	widgetTitle, _ := args["widget_title"].(string)
	if widgetTitle == "" {
		widgetTitle = "Query"
	}
	vis, _ := args["visualization"].(string)
	if vis == "" {
		vis = "viz.billboard"
	}
	safeName := escapeString(name)
	safeDesc := escapeString(desc)
	safeWidgetTitle := escapeString(widgetTitle)
	safeNRQL := escapeString(nrqlQuery)
	gql := fmt.Sprintf(`mutation {
	  dashboardCreate(
		accountId: %s
		dashboard: {
		  name: "%s"
		  description: "%s"
		  permissions: %s
		  pages: [{
			name: "Page 1"
			widgets: [{
			  visualization: { id: "%s" }
			  layout: { column: 1, row: 1, height: 3, width: 4 }
			  title: "%s"
			  rawConfiguration: {
				nrqlQueries: [{ accountIds: [%s], query: "%s" }]
			  }
			}]
		  }]
		}
	  ) {
		entityResult {
		  guid
		  name
		}
		errors { type description }
	  }
	}`, aid, safeName, safeDesc, permissions, vis, safeWidgetTitle, aid, safeNRQL)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to create dashboard: %w", err)
	}
	result, _ := data["dashboardCreate"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	errs, _ := result["errors"].([]interface{})
	if len(errs) > 0 {
		if errObj, ok := errs[0].(map[string]interface{}); ok {
			desc, _ := errObj["description"].(string)
			return framework.TextResult(""), fmt.Errorf("API error: %s", desc)
		}
		return framework.TextResult(""), fmt.Errorf("API error creating dashboard")
	}
	entityResult, _ := result["entityResult"].(map[string]interface{})
	if entityResult != nil {
		if g, ok := entityResult["guid"].(string); ok {
			return framework.TextResult(fmt.Sprintf("Created dashboard '%s' with GUID: %s", name, g)), nil
		}
	}
	return framework.TextResult(fmt.Sprintf("Created dashboard: %s", name)), nil
}
func (t *CreateDashboardTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(4),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

type UpdateDashboardTool struct {
	framework.BaseTool
	client *Client
}

func (t *UpdateDashboardTool) Name() string        { return "update_dashboard" }
func (t *UpdateDashboardTool) Description() string { return "Update a dashboard (requires --write-enabled)" }
func (t *UpdateDashboardTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"guid":        map[string]interface{}{"type": "string", "description": "Dashboard GUID"},
			"name":        map[string]interface{}{"type": "string", "description": "New dashboard name"},
			"description": map[string]interface{}{"type": "string", "description": "New dashboard description"},
			"permissions": map[string]interface{}{"type": "string", "description": "PUBLIC_READ_WRITE, PUBLIC_READ_ONLY, or PRIVATE"},
			"account_id":  map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"guid", "name"},
	}
}
func (t *UpdateDashboardTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	guid, _ := args["guid"].(string)
	name, _ := args["name"].(string)
	if guid == "" || name == "" {
		return framework.TextResult(""), fmt.Errorf("guid and name are required")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	desc, _ := args["description"].(string)
	perms, _ := args["permissions"].(string)
	if perms == "" {
		perms = "PRIVATE"
	}
	safeName := escapeString(name)
	safeDesc := escapeString(desc)
	gql := fmt.Sprintf(`mutation {
	  dashboardUpdate(
		guid: "%s"
		dashboard: {
		  name: "%s"
		  description: "%s"
		  permissions: %s
		  pages: [{ name: "Page 1" }]
		}
	  ) {
		entityResult {
		  guid
		  name
		}
		errors { type description }
	  }
	}`, guid, safeName, safeDesc, perms)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to update dashboard: %w", err)
	}
	result, _ := data["dashboardUpdate"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	errs, _ := result["errors"].([]interface{})
	if len(errs) > 0 {
		if errObj, ok := errs[0].(map[string]interface{}); ok {
			desc, _ := errObj["description"].(string)
			return framework.TextResult(""), fmt.Errorf("API error: %s", desc)
		}
		return framework.TextResult(""), fmt.Errorf("API error updating dashboard")
	}
	return framework.TextResult(fmt.Sprintf("Updated dashboard: %s", name)), nil
}
func (t *UpdateDashboardTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

type DeleteDashboardTool struct {
	framework.BaseTool
	client *Client
}

func (t *DeleteDashboardTool) Name() string        { return "delete_dashboard" }
func (t *DeleteDashboardTool) Description() string { return "Delete a dashboard (requires --write-enabled)" }
func (t *DeleteDashboardTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"guid":       map[string]interface{}{"type": "string", "description": "Dashboard GUID"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"guid"},
	}
}
func (t *DeleteDashboardTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	guid, _ := args["guid"].(string)
	if guid == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: guid")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`mutation {
	  dashboardDelete(guid: "%s") {
		status
		errors { type description }
	  }
	}`, guid)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to delete dashboard: %w", err)
	}
	result, _ := data["dashboardDelete"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	errs, _ := result["errors"].([]interface{})
	if len(errs) > 0 {
		var msgs []string
		for _, e := range errs {
			if errObj, ok := e.(map[string]interface{}); ok {
				desc, _ := errObj["description"].(string)
				msgs = append(msgs, desc)
			}
		}
		return framework.TextResult(""), fmt.Errorf("API error: %s", strings.Join(msgs, "; "))
	}
	return framework.TextResult(fmt.Sprintf("Deleted dashboard: %s", guid)), nil
}
func (t *DeleteDashboardTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}
