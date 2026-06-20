package newrelic

import (
	"fmt"
	"strings"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type GetAlertPolicyTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetAlertPolicyTool) Name() string        { return "get_alert_policy" }
func (t *GetAlertPolicyTool) Description() string { return "Get an alert policy by ID" }
func (t *GetAlertPolicyTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"policy_id":  map[string]interface{}{"type": "string", "description": "Alert policy ID"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"policy_id"},
	}
}
func (t *GetAlertPolicyTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	policyID, _ := args["policy_id"].(string)
	if policyID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: policy_id")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  alerts {
			policy(id: "%s") {
			  id
			  name
			  incidentPreference
			}
		  }
		}
	  }
	}`, aid, policyID)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query alert policy: %w", err)
	}
	alertsMap, _ := acct["alerts"].(map[string]interface{})
	if alertsMap == nil {
		return framework.TextResult("No alert policy found"), nil
	}
	policy, _ := alertsMap["policy"].(map[string]interface{})
	if policy == nil {
		return framework.TextResult("No alert policy found"), nil
	}
	return framework.TextResult(formatSingleResult(policy)), nil
}
func (t *GetAlertPolicyTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}

func (t *GetAlertPolicyTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"policy": map[string]interface{}{
				"type":        "object",
				"description": "Alert policy details",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Policy ID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Policy name",
					},
					"incidentPreference": map[string]interface{}{
						"type":        "string",
						"description": "Incident rollup preference (PER_POLICY, PER_CONDITION, or PER_CONDITION_AND_TARGET)",
					},
				},
			},
		},
	}
	return &schema
}

type CreateAlertPolicyTool struct {
	framework.BaseTool
	client *Client
}

func (t *CreateAlertPolicyTool) Name() string        { return "create_alert_policy" }
func (t *CreateAlertPolicyTool) Description() string { return "Create an alert policy (requires --write-enabled)" }
func (t *CreateAlertPolicyTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"name":                map[string]interface{}{"type": "string", "description": "Policy name"},
			"incident_preference": map[string]interface{}{"type": "string", "description": "PER_POLICY, PER_CONDITION, or PER_CONDITION_AND_TARGET", "default": "PER_CONDITION"},
			"account_id":          map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"name"},
	}
}
func (t *CreateAlertPolicyTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	name, _ := args["name"].(string)
	if name == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: name")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	incidentPref, _ := args["incident_preference"].(string)
	if incidentPref == "" {
		incidentPref = "PER_CONDITION"
	}
	gql := fmt.Sprintf(`mutation {
	  alertsPolicyCreate(
		accountId: %s
		policy: { name: "%s", incidentPreference: %s }
	  ) {
		id
		name
		incidentPreference
	  }
	}`, aid, escapeString(name), incidentPref)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to create alert policy: %w", err)
	}
	result, _ := data["alertsPolicyCreate"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	return framework.TextResult(fmt.Sprintf("Created alert policy: %s (ID: %v)", name, result["id"])), nil
}
func (t *CreateAlertPolicyTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

func (t *CreateAlertPolicyTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"alertsPolicyCreate": map[string]interface{}{
				"type":        "object",
				"description": "Created alert policy details",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Policy ID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Policy name",
					},
					"incidentPreference": map[string]interface{}{
						"type":        "string",
						"description": "Incident rollup preference",
					},
				},
			},
		},
	}
	return &schema
}

type UpdateAlertPolicyTool struct {
	framework.BaseTool
	client *Client
}

func (t *UpdateAlertPolicyTool) Name() string        { return "update_alert_policy" }
func (t *UpdateAlertPolicyTool) Description() string { return "Update an alert policy (requires --write-enabled)" }
func (t *UpdateAlertPolicyTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"policy_id":           map[string]interface{}{"type": "string", "description": "Alert policy ID"},
			"name":                map[string]interface{}{"type": "string", "description": "New policy name"},
			"incident_preference": map[string]interface{}{"type": "string", "description": "PER_POLICY, PER_CONDITION, or PER_CONDITION_AND_TARGET"},
			"account_id":          map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"policy_id", "name"},
	}
}
func (t *UpdateAlertPolicyTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	policyID, _ := args["policy_id"].(string)
	name, _ := args["name"].(string)
	if policyID == "" || name == "" {
		return framework.TextResult(""), fmt.Errorf("policy_id and name are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	safeName := escapeString(name)
	incidentPref, _ := args["incident_preference"].(string)
	var prefPart string
	if incidentPref != "" {
		prefPart = fmt.Sprintf(", incidentPreference: %s", incidentPref)
	}
	gql := fmt.Sprintf(`mutation {
	  alertsPolicyUpdate(
		accountId: %s
		id: "%s"
		policy: { name: "%s"%s }
	  ) {
		id
		name
		incidentPreference
	  }
	}`, aid, policyID, safeName, prefPart)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to update alert policy: %w", err)
	}
	result, _ := data["alertsPolicyUpdate"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	return framework.TextResult(fmt.Sprintf("Updated alert policy: %s", name)), nil
}
func (t *UpdateAlertPolicyTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

func (t *UpdateAlertPolicyTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"alertsPolicyUpdate": map[string]interface{}{
				"type":        "object",
				"description": "Updated alert policy details",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Policy ID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Policy name",
					},
					"incidentPreference": map[string]interface{}{
						"type":        "string",
						"description": "Incident rollup preference",
					},
				},
			},
		},
	}
	return &schema
}

type DeleteAlertPolicyTool struct {
	framework.BaseTool
	client *Client
}

func (t *DeleteAlertPolicyTool) Name() string        { return "delete_alert_policy" }
func (t *DeleteAlertPolicyTool) Description() string { return "Delete an alert policy (requires --write-enabled)" }
func (t *DeleteAlertPolicyTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"policy_id":  map[string]interface{}{"type": "string", "description": "Alert policy ID"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"policy_id"},
	}
}
func (t *DeleteAlertPolicyTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	policyID, _ := args["policy_id"].(string)
	if policyID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: policy_id")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`mutation {
	  alertsPolicyDelete(accountId: %s, id: "%s") {
		id
	  }
	}`, aid, policyID)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to delete alert policy: %w", err)
	}
	result, _ := data["alertsPolicyDelete"].(map[string]interface{})
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	return framework.TextResult(fmt.Sprintf("Deleted alert policy: %s", policyID)), nil
}
func (t *DeleteAlertPolicyTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskHigh),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}

func (t *DeleteAlertPolicyTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"alertsPolicyDelete": map[string]interface{}{
				"type":        "object",
				"description": "Deleted alert policy confirmation",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the deleted policy",
					},
				},
			},
		},
	}
	return &schema
}

type ListNRQLAlertConditionsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListNRQLAlertConditionsTool) Name() string        { return "list_nrql_alert_conditions" }
func (t *ListNRQLAlertConditionsTool) Description() string { return "List NRQL alert conditions for a policy" }
func (t *ListNRQLAlertConditionsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"policy_id":  map[string]interface{}{"type": "string", "description": "Alert policy ID"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"policy_id"},
	}
}
func (t *ListNRQLAlertConditionsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	policyID, _ := args["policy_id"].(string)
	if policyID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: policy_id")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  alerts {
			nrqlConditionsSearch(filter: { policyId: "%s" }) {
			  nrqlConditions {
				id
				name
				enabled
				nrql { query }
				critical { thresholdDuration duration }
			  }
			}
		  }
		}
	  }
	}`, aid, policyID)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query NRQL conditions: %w", err)
	}
	alertsMap, _ := acct["alerts"].(map[string]interface{})
	if alertsMap == nil {
		return framework.TextResult("No alert conditions found"), nil
	}
	searchResult, _ := alertsMap["nrqlConditionsSearch"].(map[string]interface{})
	if searchResult == nil {
		return framework.TextResult("No alert conditions found"), nil
	}
	rawConditions, _ := searchResult["nrqlConditions"].([]interface{})
	if len(rawConditions) == 0 {
		return framework.TextResult("No alert conditions found"), nil
	}
	var conditions []map[string]interface{}
	for _, c := range rawConditions {
		if m, ok := c.(map[string]interface{}); ok {
			conditions = append(conditions, m)
		}
	}
	if len(conditions) == 0 {
		return framework.TextResult("No alert conditions found"), nil
	}
	return framework.TextResult(formatResults(conditions)), nil
}
func (t *ListNRQLAlertConditionsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}

func (t *ListNRQLAlertConditionsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"nrqlConditions": map[string]interface{}{
				"type":        "array",
				"description": "List of NRQL alert conditions",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Condition ID",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Condition name",
						},
						"enabled": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether the condition is enabled",
						},
						"nrql": map[string]interface{}{
							"type":        "object",
							"description": "NRQL query details",
							"properties": map[string]interface{}{
								"query": map[string]interface{}{
									"type":        "string",
									"description": "NRQL query string",
								},
							},
						},
						"critical": map[string]interface{}{
							"type":        "object",
							"description": "Critical threshold settings",
							"properties": map[string]interface{}{
								"thresholdDuration": map[string]interface{}{
									"type":        "number",
									"description": "Threshold duration in seconds",
								},
								"duration": map[string]interface{}{
									"type":        "number",
									"description": "Duration value",
								},
							},
						},
					},
				},
			},
		},
	}
	return &schema
}

// Make existing create_alert_condition tool real by updating its Handle method.
// We can't redefine the type, so we create a replacement version to register instead.
type RealCreateAlertConditionTool struct {
	framework.BaseTool
	client *Client
}

func (t *RealCreateAlertConditionTool) Name() string        { return "create_alert_condition" }
func (t *RealCreateAlertConditionTool) Description() string { return "Create a new alert condition in an alert policy (requires --write-enabled)" }
func (t *RealCreateAlertConditionTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"policy_id":  map[string]interface{}{"type": "string", "description": "The alert policy ID"},
			"name":       map[string]interface{}{"type": "string", "description": "Name of the alert condition"},
			"nrql_query": map[string]interface{}{"type": "string", "description": "NRQL query for the condition"},
			"critical_threshold": map[string]interface{}{
				"type":        "number",
				"description": "Critical threshold value",
			},
			"warning_threshold": map[string]interface{}{
				"type":        "number",
				"description": "Warning threshold value (optional)",
			},
			"duration_minutes": map[string]interface{}{
				"type":        "number",
				"description": "Duration in minutes (1, 2, 5, 10, 15, 30, 60, 120)",
				"default":     5,
			},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"policy_id", "name", "nrql_query", "critical_threshold"},
	}
}
func (t *RealCreateAlertConditionTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	policyID, _ := args["policy_id"].(string)
	name, _ := args["name"].(string)
	nrqlQuery, _ := args["nrql_query"].(string)
	if policyID == "" || name == "" || nrqlQuery == "" {
		return framework.TextResult(""), fmt.Errorf("policy_id, name, and nrql_query are required")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	criticalThreshold, _ := args["critical_threshold"].(float64)
	warningThreshold, _ := args["warning_threshold"].(float64)
	durationMin, _ := args["duration_minutes"].(float64)
	if durationMin <= 0 {
		durationMin = 5
	}
	safeName := escapeString(name)
	safeNRQL := escapeString(nrqlQuery)
	var warningPart string
	if warningThreshold > 0 {
		warningPart = fmt.Sprintf(`, warning: { threshold: %v, thresholdOccurrences: ALL }`, warningThreshold)
	}
	gql := fmt.Sprintf(`mutation {
	  alertsNrqlConditionCreate(
		accountId: %s
		policyId: "%s"
		condition: {
		  name: "%s"
		  nrql: { query: "%s" }
		  critical: { threshold: %v, thresholdOccurrences: ALL }
		  %s
		  runbookUrl: ""
		  enabled: true
		}
	  ) {
		id
		name
		enabled
	  }
	}`, aid, policyID, safeName, safeNRQL, criticalThreshold, warningPart)
	data, err := t.client.rawGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to create alert condition: %w", err)
	}
	// Try both mutation naming conventions
	result, _ := data["alertsNrqlConditionCreate"].(map[string]interface{})
	if result == nil {
		result, _ = data["alertsNrqlConditionUpdate"].(map[string]interface{})
	}
	if result == nil {
		return framework.TextResult(""), fmt.Errorf("unexpected response format")
	}
	if id, ok := result["id"].(string); ok && id != "" {
		return framework.TextResult(fmt.Sprintf("Created alert condition: %s (ID: %s)", name, id)), nil
	}
	return framework.TextResult(fmt.Sprintf("Created alert condition: %s", name)), nil
}
func (t *RealCreateAlertConditionTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

func (t *RealCreateAlertConditionTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"alertsNrqlConditionCreate": map[string]interface{}{
				"type":        "object",
				"description": "Created alert condition details",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Condition ID",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Condition name",
					},
					"enabled": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the condition is enabled",
					},
				},
			},
		},
	}
	return &schema
}

func getAlertPolicyErrorMessages(result map[string]interface{}) string {
	errs, _ := result["errors"].([]interface{})
	if len(errs) == 0 {
		return ""
	}
	var msgs []string
	for _, e := range errs {
		if errObj, ok := e.(map[string]interface{}); ok {
			desc, _ := errObj["description"].(string)
			if desc != "" {
				msgs = append(msgs, desc)
			}
		}
	}
	return strings.Join(msgs, "; ")
}
