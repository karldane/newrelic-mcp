package newrelic

import (
	"fmt"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type ListServiceLevelsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListServiceLevelsTool) Name() string        { return "list_service_levels" }
func (t *ListServiceLevelsTool) Description() string { return "List service levels for an entity" }
func (t *ListServiceLevelsTool) Title() string        { return "Service Levels" }
func (t *ListServiceLevelsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entity_guid": map[string]interface{}{"type": "string", "description": "Entity GUID"},
		},
		Required: []string{"entity_guid"},
	}
}
func (t *ListServiceLevelsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	entityGUID, _ := args["entity_guid"].(string)
	if entityGUID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: entity_guid")
	}
	gql := fmt.Sprintf(`{
	  actor {
		entity(guid: "%s") {
		  guid
		  ... on EntityInterfaceWithServiceLevel {
			serviceLevel {
			  indicators {
				id
				name
				description
				sli
			  }
			}
		  }
		}
	  }
	}`, entityGUID)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query service levels: %w", err)
	}
	entity, _ := actor["entity"].(map[string]interface{})
	if entity == nil {
		return framework.TextResult(fmt.Sprintf("Entity '%s' not found", entityGUID)), nil
	}
	sl, _ := entity["serviceLevel"].(map[string]interface{})
	if sl == nil {
		return framework.TextResult("No service levels found"), nil
	}
	rawIndicators, _ := sl["indicators"].([]interface{})
	var indicators []map[string]interface{}
	for _, i := range rawIndicators {
		if m, ok := i.(map[string]interface{}); ok {
			indicators = append(indicators, m)
		}
	}
	if len(indicators) == 0 {
		return framework.TextResult("No service levels found"), nil
	}
	return framework.TextResult(formatResults(indicators)), nil
}
func (t *ListServiceLevelsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}
func (t *ListServiceLevelsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"indicators": map[string]interface{}{
				"type":        "array",
				"description": "List of service level indicators for the entity",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Unique identifier for the service level indicator",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Name of the service level indicator",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Description of the service level indicator",
						},
						"sli": map[string]interface{}{
							"type":        "object",
							"description": "Service level indicator configuration and metrics",
						},
					},
				},
			},
		},
	}
	return &schema
}

type CreateServiceLevelTool struct {
	framework.BaseTool
	client *Client
}

func (t *CreateServiceLevelTool) Name() string        { return "create_service_level" }
func (t *CreateServiceLevelTool) Description() string { return "Create a service level indicator with SLO (requires --write-enabled)" }
func (t *CreateServiceLevelTool) Title() string        { return "Create Service Level" }
func (t *CreateServiceLevelTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entity_guid":  map[string]interface{}{"type": "string", "description": "Entity GUID"},
			"name":         map[string]interface{}{"type": "string", "description": "SLI name"},
			"description":  map[string]interface{}{"type": "string", "description": "SLI description (optional)"},
			"valid_events": map[string]interface{}{"type": "string", "description": "NRQL for valid events"},
			"good_events":  map[string]interface{}{"type": "string", "description": "NRQL for good events"},
			"target":       map[string]interface{}{"type": "number", "description": "SLO target percentage (e.g. 99.9)"},
			"time_window":  map[string]interface{}{"type": "string", "description": "WEEK or MONTH", "default": "WEEK"},
			"account_id":   map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"entity_guid", "name", "valid_events", "good_events", "target"},
	}
}
func (t *CreateServiceLevelTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	entityGUID, _ := args["entity_guid"].(string)
	name, _ := args["name"].(string)
	validEvents, _ := args["valid_events"].(string)
	goodEvents, _ := args["good_events"].(string)
	target, _ := args["target"].(float64)
	if entityGUID == "" || name == "" || validEvents == "" || goodEvents == "" || target <= 0 {
		return framework.TextResult(""), fmt.Errorf("entity_guid, name, valid_events, good_events, and target are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	desc, _ := args["description"].(string)
	timeWindow, _ := args["time_window"].(string)
	if timeWindow == "" {
		timeWindow = "WEEK"
	}
	safeName := escapeString(name)
	safeDesc := escapeString(desc)
	safeValid := escapeString(validEvents)
	safeGood := escapeString(goodEvents)
	gql := fmt.Sprintf(`mutation {
	  serviceLevelCreate(
		accountId: %s
		createSliInput: {
		  name: "%s"
		  description: "%s"
		  entityGuid: "%s"
		  events: {
			validEvents: { where: { query: "%s" } }
			goodEvents: { where: { query: "%s" } }
		  }
		  objective: { target: %v, timeWindow: %s }
		}
	  ) {
		indicator {
		  id
		  name
		}
		errors {
		  type
		  description
		}
	  }
	}`, aid, safeName, safeDesc, entityGUID, safeValid, safeGood, target, timeWindow)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to create service level: %w", err)
	}
	result, _ := data["serviceLevelCreate"].(map[string]interface{})
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
	return framework.TextResult(fmt.Sprintf("Created service level: %s", name)), nil
}
func (t *CreateServiceLevelTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}
func (t *CreateServiceLevelTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"indicator": map[string]interface{}{
				"type":        "object",
				"description": "The created service level indicator",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier for the created SLI",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the created SLI",
					},
				},
			},
			"errors": map[string]interface{}{
				"type":        "array",
				"description": "Any errors that occurred during creation",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type": map[string]interface{}{
							"type":        "string",
							"description": "Error type",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Error description",
						},
					},
				},
			},
		},
	}
	return &schema
}

type UpdateServiceLevelTool struct {
	framework.BaseTool
	client *Client
}

func (t *UpdateServiceLevelTool) Name() string        { return "update_service_level" }
func (t *UpdateServiceLevelTool) Description() string { return "Update a service level SLO target (requires --write-enabled)" }
func (t *UpdateServiceLevelTool) Title() string        { return "Update Service Level" }
func (t *UpdateServiceLevelTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"sli_id":      map[string]interface{}{"type": "string", "description": "SLI ID"},
			"target":      map[string]interface{}{"type": "number", "description": "New SLO target percentage"},
			"time_window": map[string]interface{}{"type": "string", "description": "WEEK or MONTH (optional)"},
			"account_id":  map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"sli_id", "target"},
	}
}
func (t *UpdateServiceLevelTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	sliID, _ := args["sli_id"].(string)
	target, _ := args["target"].(float64)
	if sliID == "" || target <= 0 {
		return framework.TextResult(""), fmt.Errorf("sli_id and target are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	timeWindow, _ := args["time_window"].(string)
	var twPart string
	if timeWindow != "" {
		twPart = fmt.Sprintf(", timeWindow: %s", timeWindow)
	}
	gql := fmt.Sprintf(`mutation {
	  serviceLevelUpdate(
		accountId: %s
		updateSliInput: {
		  sliId: "%s"
		  objective: { target: %v%s }
		}
	  ) {
		indicator {
		  id
		  name
		}
		errors {
		  type
		  description
		}
	  }
	}`, aid, sliID, target, twPart)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to update service level: %w", err)
	}
	result, _ := data["serviceLevelUpdate"].(map[string]interface{})
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
	return framework.TextResult(fmt.Sprintf("Updated service level: %s", sliID)), nil
}
func (t *UpdateServiceLevelTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}
func (t *UpdateServiceLevelTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"indicator": map[string]interface{}{
				"type":        "object",
				"description": "The updated service level indicator",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier for the updated SLI",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the updated SLI",
					},
				},
			},
			"errors": map[string]interface{}{
				"type":        "array",
				"description": "Any errors that occurred during update",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"type": map[string]interface{}{
							"type":        "string",
							"description": "Error type",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Error description",
						},
					},
				},
			},
		},
	}
	return &schema
}
