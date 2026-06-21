package newrelic

import (
	"fmt"
	"strings"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type SearchEntitiesTool struct {
	framework.BaseTool
	client *Client
}

func (t *SearchEntitiesTool) Name() string        { return "search_entities" }
func (t *SearchEntitiesTool) Description() string { return "Search entities with freeform NRQL query" }
func (t *SearchEntitiesTool) Title() string        { return "Entity Search" }
func (t *SearchEntitiesTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"query": map[string]interface{}{"type": "string", "description": "Entity search query (e.g. name = 'My App')"},
			"limit": map[string]interface{}{"type": "number", "description": "Max results (optional)"},
		},
		Required: []string{"query"},
	}
}
func (t *SearchEntitiesTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: query")
	}
	limit, _ := args["limit"].(float64)
	limitStr := ""
	if limit > 0 {
		limitStr = fmt.Sprintf(", limit: %.0f", limit)
	}
	safeQuery := escapeString(query)
	gql := fmt.Sprintf(`{
	  actor {
		entitySearch(queryBuilder: {query: "%s"%s}) {
		  results {
			entities {
			  guid
			  name
			  entityType
			  reporting
			}
		  }
		}
	  }
	}`, safeQuery, limitStr)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to search entities: %w", err)
	}
	entitySearch, _ := actor["entitySearch"].(map[string]interface{})
	if entitySearch == nil {
		return framework.TextResult("No entities found"), nil
	}
	results, _ := entitySearch["results"].(map[string]interface{})
	if results == nil {
		return framework.TextResult("No entities found"), nil
	}
	rawEntities, _ := results["entities"].([]interface{})
	var entities []map[string]interface{}
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			entities = append(entities, m)
		}
	}
	if len(entities) == 0 {
		return framework.TextResult("No entities found"), nil
	}
	return framework.TextResult(formatResults(entities)), nil
}
func (t *SearchEntitiesTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(3),
		framework.WithPII(false),
	)
}
func (t *SearchEntitiesTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entities": map[string]interface{}{
				"type":        "array",
				"description": "List of entities matching the search query",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"guid": map[string]interface{}{
							"type":        "string",
							"description": "Unique identifier for the entity",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Name of the entity",
						},
						"entityType": map[string]interface{}{
							"type":        "string",
							"description": "Type of the entity (e.g., APPLICATION, HOST)",
						},
						"reporting": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether the entity is currently reporting data",
						},
					},
				},
			},
		},
	}
	return &schema
}

type DeleteEntityTool struct {
	framework.BaseTool
	client *Client
}

func (t *DeleteEntityTool) Name() string        { return "delete_entity" }
func (t *DeleteEntityTool) Description() string { return "Delete entities by GUID (requires --write-enabled)" }
func (t *DeleteEntityTool) Title() string        { return "Delete Entity" }
func (t *DeleteEntityTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"guids":      map[string]interface{}{"type": "string", "description": "Comma-separated entity GUIDs to delete"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"guids"},
	}
}
func (t *DeleteEntityTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	guids, _ := args["guids"].(string)
	if guids == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: guids")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	parts := strings.Split(guids, ",")
	var quoted []string
	for _, p := range parts {
		quoted = append(quoted, fmt.Sprintf("%q", strings.TrimSpace(p)))
	}
	guidsParam := "[" + strings.Join(quoted, ",") + "]"
	gql := fmt.Sprintf(`mutation {
	  entityDelete(guids: %s) {
		deleted
	  }
	}`, guidsParam)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to delete entities: %w", err)
	}
	result, _ := data["entityDelete"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	return framework.TextResult(fmt.Sprintf("Deleted entities: %s", guids)), nil
}
func (t *DeleteEntityTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}
func (t *DeleteEntityTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"deleted": map[string]interface{}{
				"type":        "array",
				"description": "List of entity GUIDs that were successfully deleted",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Confirmation message indicating which entities were deleted",
			},
		},
	}
	return &schema
}
