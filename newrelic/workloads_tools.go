package newrelic

import (
	"fmt"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type ListWorkloadsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListWorkloadsTool) Name() string        { return "list_workloads" }
func (t *ListWorkloadsTool) Description() string { return "List workloads for an account" }
func (t *ListWorkloadsTool) Title() string        { return "Workload List" }
func (t *ListWorkloadsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *ListWorkloadsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  workload {
			workloads {
			  guid
			  name
			  workloadStatus
			}
		  }
		}
	  }
	}`, aid)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query workloads: %w", err)
	}
	workload, _ := acct["workload"].(map[string]interface{})
	if workload == nil {
		return framework.TextResult("No workloads found"), nil
	}
	rawWorkloads, _ := workload["workloads"].([]interface{})
	var items []map[string]interface{}
	for _, w := range rawWorkloads {
		if m, ok := w.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}
	if len(items) == 0 {
		return framework.TextResult("No workloads found"), nil
	}
	return framework.TextResult(formatResults(items)), nil
}
func (t *ListWorkloadsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}
func (t *ListWorkloadsTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"workloads": map[string]interface{}{
				"type":        "array",
				"description": "List of workloads in the account",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"guid": map[string]interface{}{
							"type":        "string",
							"description": "Unique identifier for the workload",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Name of the workload",
						},
						"workloadStatus": map[string]interface{}{
							"type":        "string",
							"description": "Current status of the workload",
						},
					},
				},
			},
		},
	}
}

type GetWorkloadTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetWorkloadTool) Name() string        { return "get_workload" }
func (t *GetWorkloadTool) Description() string { return "Get a workload by GUID" }
func (t *GetWorkloadTool) Title() string        { return "Workload Details" }
func (t *GetWorkloadTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"workload_guid": map[string]interface{}{"type": "string", "description": "Workload GUID"},
		},
		Required: []string{"workload_guid"},
	}
}
func (t *GetWorkloadTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	workloadGUID, _ := args["workload_guid"].(string)
	if workloadGUID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: workload_guid")
	}
	gql := fmt.Sprintf(`{
	  actor {
		entity(guid: "%s") {
		  guid
		  name
		  ... on WorkloadEntity {
			workload {
			  status
			}
		  }
		}
	  }
	}`, workloadGUID)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query workload: %w", err)
	}
	entity, _ := actor["entity"].(map[string]interface{})
	if entity == nil {
		return framework.TextResult(fmt.Sprintf("Workload '%s' not found", workloadGUID)), nil
	}
	return framework.TextResult(formatSingleResult(entity)), nil
}
func (t *GetWorkloadTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}
func (t *GetWorkloadTool) OutputSchema() *mcp.ToolOutputSchema {
	return &mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"entity": map[string]interface{}{
				"type":        "object",
				"description": "Workload entity details",
				"properties": map[string]interface{}{
					"guid": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier for the workload entity",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the workload",
					},
					"workload": map[string]interface{}{
						"type":        "object",
						"description": "Workload-specific details",
						"properties": map[string]interface{}{
							"status": map[string]interface{}{
								"type":        "string",
								"description": "Current operational status of the workload",
							},
						},
					},
				},
			},
		},
	}
}
