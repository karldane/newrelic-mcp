package newrelic

import (
	"fmt"
	"strings"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type GetEntityTagsTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetEntityTagsTool) Name() string        { return "get_entity_tags" }
func (t *GetEntityTagsTool) Description() string { return "Get all tags for an entity" }
func (t *GetEntityTagsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entity_guid": map[string]interface{}{"type": "string", "description": "Entity GUID"},
		},
		Required: []string{"entity_guid"},
	}
}
func (t *GetEntityTagsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	entityGUID, _ := args["entity_guid"].(string)
	if entityGUID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: entity_guid")
	}
	gql := fmt.Sprintf(`{
	  actor {
		entity(guid: "%s") {
		  guid
		  tags {
			key
			values
		  }
		}
	  }
	}`, entityGUID)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query entity tags: %w", err)
	}
	entity, _ := actor["entity"].(map[string]interface{})
	if entity == nil {
		return framework.TextResult(fmt.Sprintf("Entity '%s' not found", entityGUID)), nil
	}
	rawTags, _ := entity["tags"].([]interface{})
	var tags []map[string]interface{}
	for _, t := range rawTags {
		if m, ok := t.(map[string]interface{}); ok {
			tags = append(tags, m)
		}
	}
	if len(tags) == 0 {
		return framework.TextResult("No tags found"), nil
	}
	return framework.TextResult(formatResults(tags)), nil
}
func (t *GetEntityTagsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(1),
		framework.WithPII(false),
	)
}

type AddEntityTagsTool struct {
	framework.BaseTool
	client *Client
}

func (t *AddEntityTagsTool) Name() string        { return "add_entity_tags" }
func (t *AddEntityTagsTool) Description() string { return "Add tags to an entity (requires --write-enabled)" }
func (t *AddEntityTagsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entity_guid": map[string]interface{}{"type": "string", "description": "Entity GUID"},
			"tags":        map[string]interface{}{"type": "string", "description": "Comma-separated key=value pairs (e.g. env=prod,team=sre)"},
			"account_id":  map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"entity_guid", "tags"},
	}
}
func (t *AddEntityTagsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	entityGUID, _ := args["entity_guid"].(string)
	tagsStr, _ := args["tags"].(string)
	if entityGUID == "" || tagsStr == "" {
		return framework.TextResult(""), fmt.Errorf("entity_guid and tags are required")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	tagPairs := strings.Split(tagsStr, ",")
	var tagInputs []string
	for _, pair := range tagPairs {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := ""
		if len(parts) > 1 {
			val = strings.TrimSpace(parts[1])
		}
		tagInputs = append(tagInputs, fmt.Sprintf(`{ key: "%s", values: ["%s"] }`, escapeString(key), escapeString(val)))
	}
	tagsParam := "[" + strings.Join(tagInputs, ",") + "]"
	gql := fmt.Sprintf(`mutation {
	  taggingAddTagsToEntity(
		guid: "%s"
		tags: %s
	  ) {
		errors {
		  type
		  description
		}
	  }
	}`, entityGUID, tagsParam)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to add tags: %w", err)
	}
	result, _ := data["taggingAddTagsToEntity"].(map[string]interface{})
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
	return framework.TextResult(fmt.Sprintf("Added tags to %s", entityGUID)), nil
}
func (t *AddEntityTagsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

type RemoveEntityTagsTool struct {
	framework.BaseTool
	client *Client
}

func (t *RemoveEntityTagsTool) Name() string        { return "remove_entity_tags" }
func (t *RemoveEntityTagsTool) Description() string { return "Remove tags from an entity (requires --write-enabled)" }
func (t *RemoveEntityTagsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entity_guid": map[string]interface{}{"type": "string", "description": "Entity GUID"},
			"tag_keys":    map[string]interface{}{"type": "string", "description": "Comma-separated tag keys to remove"},
			"account_id":  map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"entity_guid", "tag_keys"},
	}
}
func (t *RemoveEntityTagsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	entityGUID, _ := args["entity_guid"].(string)
	tagKeys, _ := args["tag_keys"].(string)
	if entityGUID == "" || tagKeys == "" {
		return framework.TextResult(""), fmt.Errorf("entity_guid and tag_keys are required")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	parts := strings.Split(tagKeys, ",")
	var quoted []string
	for _, p := range parts {
		quoted = append(quoted, fmt.Sprintf("%q", strings.TrimSpace(p)))
	}
	keysParam := "[" + strings.Join(quoted, ",") + "]"
	gql := fmt.Sprintf(`mutation {
	  taggingDeleteTagFromEntity(
		guid: "%s"
		tagKeys: %s
	  ) {
		errors {
		  type
		  description
		}
	  }
	}`, entityGUID, keysParam)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to remove tags: %w", err)
	}
	result, _ := data["taggingDeleteTagFromEntity"].(map[string]interface{})
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
	return framework.TextResult(fmt.Sprintf("Removed tags from %s", entityGUID)), nil
}
func (t *RemoveEntityTagsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}

type ReplaceEntityTagsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ReplaceEntityTagsTool) Name() string        { return "replace_entity_tags" }
func (t *ReplaceEntityTagsTool) Description() string { return "Replace all tags on an entity (requires --write-enabled)" }
func (t *ReplaceEntityTagsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entity_guid": map[string]interface{}{"type": "string", "description": "Entity GUID"},
			"tags":        map[string]interface{}{"type": "string", "description": "Comma-separated key=value pairs (e.g. env=prod,team=sre)"},
			"account_id":  map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"entity_guid", "tags"},
	}
}
func (t *ReplaceEntityTagsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	entityGUID, _ := args["entity_guid"].(string)
	tagsStr, _ := args["tags"].(string)
	if entityGUID == "" || tagsStr == "" {
		return framework.TextResult(""), fmt.Errorf("entity_guid and tags are required")
	}
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	tagPairs := strings.Split(tagsStr, ",")
	var tagInputs []string
	for _, pair := range tagPairs {
		pair = strings.TrimSpace(pair)
		parts := strings.SplitN(pair, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := ""
		if len(parts) > 1 {
			val = strings.TrimSpace(parts[1])
		}
		tagInputs = append(tagInputs, fmt.Sprintf(`{ key: "%s", values: ["%s"] }`, escapeString(key), escapeString(val)))
	}
	tagsParam := "[" + strings.Join(tagInputs, ",") + "]"
	gql := fmt.Sprintf(`mutation {
	  taggingReplaceTagsOnEntity(
		guid: "%s"
		tags: %s
	  ) {
		errors {
		  type
		  description
		}
	  }
	}`, entityGUID, tagsParam)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to replace tags: %w", err)
	}
	result, _ := data["taggingReplaceTagsOnEntity"].(map[string]interface{})
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
	return framework.TextResult(fmt.Sprintf("Replaced tags on %s", entityGUID)), nil
}
func (t *ReplaceEntityTagsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}
