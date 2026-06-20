package newrelic

import (
	"fmt"
	"strings"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type ListWorkflowsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListWorkflowsTool) Name() string        { return "list_workflows" }
func (t *ListWorkflowsTool) Description() string { return "List workflows for an account" }
func (t *ListWorkflowsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *ListWorkflowsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  aiWorkflows {
			workflows {
			  entities {
				id
				name
				workflowEnabled
			  }
			  totalCount
			}
		  }
		}
	  }
	}`, aid)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query workflows: %w", err)
	}
	aiWorkflows, _ := acct["aiWorkflows"].(map[string]interface{})
	if aiWorkflows == nil {
		return framework.TextResult("No workflows found"), nil
	}
	workflows, _ := aiWorkflows["workflows"].(map[string]interface{})
	if workflows == nil {
		return framework.TextResult("No workflows found"), nil
	}
	rawEntities, _ := workflows["entities"].([]interface{})
	var items []map[string]interface{}
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}
	if len(items) == 0 {
		return framework.TextResult("No workflows found"), nil
	}
	return framework.TextResult(formatResults(items)), nil
}
func (t *ListWorkflowsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}

type GetWorkflowTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetWorkflowTool) Name() string        { return "get_workflow" }
func (t *GetWorkflowTool) Description() string { return "Get a workflow by ID" }
func (t *GetWorkflowTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"workflow_id": map[string]interface{}{"type": "string", "description": "Workflow ID"},
			"account_id":  map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"workflow_id"},
	}
}
func (t *GetWorkflowTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	workflowID, _ := args["workflow_id"].(string)
	if workflowID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: workflow_id")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  aiWorkflows {
			workflows(filters: {id: "%s"}) {
			  entities {
				id
				name
				workflowEnabled
			  }
			}
		  }
		}
	  }
	}`, aid, workflowID)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query workflow: %w", err)
	}
	aiWorkflows, _ := acct["aiWorkflows"].(map[string]interface{})
	if aiWorkflows == nil {
		return framework.TextResult(fmt.Sprintf("Workflow '%s' not found", workflowID)), nil
	}
	workflows, _ := aiWorkflows["workflows"].(map[string]interface{})
	if workflows == nil {
		return framework.TextResult(fmt.Sprintf("Workflow '%s' not found", workflowID)), nil
	}
	rawEntities, _ := workflows["entities"].([]interface{})
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			return framework.TextResult(formatSingleResult(m)), nil
		}
	}
	return framework.TextResult(fmt.Sprintf("Workflow '%s' not found", workflowID)), nil
}
func (t *GetWorkflowTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}

type CreateWorkflowTool struct {
	framework.BaseTool
	client *Client
}

func (t *CreateWorkflowTool) Name() string        { return "create_workflow" }
func (t *CreateWorkflowTool) Description() string { return "Create a workflow (requires --write-enabled)" }
func (t *CreateWorkflowTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name":              map[string]interface{}{"type": "string", "description": "Workflow name"},
			"channel_ids":       map[string]interface{}{"type": "string", "description": "Comma-separated notification channel IDs"},
			"enrichments":       map[string]interface{}{"type": "string", "description": "Comma-separated enrichment IDs (optional)"},
			"account_id":        map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"name"},
	}
}
func (t *CreateWorkflowTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	name, _ := args["name"].(string)
	if name == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: name")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}

	safeName := escapeString(name)

	channelIDs, _ := args["channel_ids"].(string)
	var channelPart string
	if channelIDs != "" {
		parts := strings.Split(channelIDs, ",")
		var quoted []string
		for _, p := range parts {
			quoted = append(quoted, fmt.Sprintf("%q", strings.TrimSpace(p)))
		}
		channelPart = fmt.Sprintf(", notificationChannelIds: [%s]", strings.Join(quoted, ","))
	}

	enrichments, _ := args["enrichments"].(string)
	var enrichPart string
	if enrichments != "" {
		parts := strings.Split(enrichments, ",")
		var quoted []string
		for _, p := range parts {
			quoted = append(quoted, fmt.Sprintf("%q", strings.TrimSpace(p)))
		}
		enrichPart = fmt.Sprintf(", enrichments: { nrql: { configurations: [{ enrichOn: WORKFLOW, queryIds: [%s] }] } }", strings.Join(quoted, ","))
	}

	gql := fmt.Sprintf(`mutation {
	  aiWorkflowsCreateWorkflow(
		accountId: %s
		createWorkflowData: {
		  name: "%s"
		  issuesFilter: {
			accountIds: [%s]
			policyIds: []
			processor: "NewRelic"
		    }
		  destinationConfigurations: [{
		    channelId: ""
		  }]%s%s
		}
	  ) {
		workflow {
		  id
		  name
		}
		errors {
		  type
		  description
		}
	  }
	}`, aid, safeName, aid, channelPart, enrichPart)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to create workflow: %w", err)
	}
	result, _ := data["aiWorkflowsCreateWorkflow"].(map[string]interface{})
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
	return framework.TextResult(fmt.Sprintf("Created workflow: %s", name)), nil
}
func (t *CreateWorkflowTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

type UpdateWorkflowTool struct {
	framework.BaseTool
	client *Client
}

func (t *UpdateWorkflowTool) Name() string        { return "update_workflow" }
func (t *UpdateWorkflowTool) Description() string { return "Update a workflow (requires --write-enabled)" }
func (t *UpdateWorkflowTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"workflow_id":  map[string]interface{}{"type": "string", "description": "Workflow ID"},
			"name":         map[string]interface{}{"type": "string", "description": "New workflow name"},
			"account_id":   map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"workflow_id", "name"},
	}
}
func (t *UpdateWorkflowTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	workflowID, _ := args["workflow_id"].(string)
	name, _ := args["name"].(string)
	if workflowID == "" || name == "" {
		return framework.TextResult(""), fmt.Errorf("workflow_id and name are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	safeName := escapeString(name)
	gql := fmt.Sprintf(`mutation {
	  aiWorkflowsUpdateWorkflow(
		accountId: %s
		id: "%s"
		updateWorkflowData: {
		  name: "%s"
		  destinationConfigurations: [{ channelId: "" }]
		}
	  ) {
		workflow {
		  id
		  name
		}
		errors {
		  type
		  description
		}
	  }
	}`, aid, workflowID, safeName)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to update workflow: %w", err)
	}
	result, _ := data["aiWorkflowsUpdateWorkflow"].(map[string]interface{})
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
	return framework.TextResult(fmt.Sprintf("Updated workflow: %s", name)), nil
}
func (t *UpdateWorkflowTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

type DeleteWorkflowTool struct {
	framework.BaseTool
	client *Client
}

func (t *DeleteWorkflowTool) Name() string        { return "delete_workflow" }
func (t *DeleteWorkflowTool) Description() string { return "Delete a workflow (requires --write-enabled)" }
func (t *DeleteWorkflowTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"workflow_id": map[string]interface{}{"type": "string", "description": "Workflow ID"},
			"account_id":  map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"workflow_id"},
	}
}
func (t *DeleteWorkflowTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	workflowID, _ := args["workflow_id"].(string)
	if workflowID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: workflow_id")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`mutation {
	  aiWorkflowsDeleteWorkflow(accountId: %s, id: "%s") {
		id
		errors {
		  type
		  description
		}
	  }
	}`, aid, workflowID)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to delete workflow: %w", err)
	}
	result, _ := data["aiWorkflowsDeleteWorkflow"].(map[string]interface{})
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
	return framework.TextResult(fmt.Sprintf("Deleted workflow: %s", workflowID)), nil
}
func (t *DeleteWorkflowTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}
