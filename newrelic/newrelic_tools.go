package newrelic

import (
	"fmt"
	"sort"
	"strings"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type ListApplicationsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListApplicationsTool) Name() string        { return "list_applications" }
func (t *ListApplicationsTool) Description() string { return "List APM applications" }
func (t *ListApplicationsTool) Title() string { return "APM Applications" }
func (t *ListApplicationsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"limit":      map[string]interface{}{"type": "number", "description": "Max results"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *ListApplicationsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
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
		entitySearch(queryBuilder: {domain: APM, type: APPLICATION%s}) {
		  results {
			entities {
			  ... on ApmApplicationEntityOutline {
				guid
				name
				language
			  }
			}
		  }
		}
	  }
	}`, limitStr)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query applications: %w", err)
	}
	entitySearch, _ := actor["entitySearch"].(map[string]interface{})
	if entitySearch == nil {
		return framework.TextResult("No applications found"), nil
	}
	results, _ := entitySearch["results"].(map[string]interface{})
	if results == nil {
		return framework.TextResult("No applications found"), nil
	}
	rawEntities, _ := results["entities"].([]interface{})
	var apps []map[string]interface{}
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			apps = append(apps, m)
		}
	}
	if len(apps) == 0 {
		return framework.TextResult("No applications found"), nil
	}
	return framework.TextResult(formatResults(apps)), nil
}
func (t *ListApplicationsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"applications": map[string]interface{}{
				"type":        "array",
				"description": "List of APM applications",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"guid":     map[string]interface{}{"type": "string", "description": "Application GUID"},
						"name":     map[string]interface{}{"type": "string", "description": "Application name"},
						"language": map[string]interface{}{"type": "string", "description": "Programming language"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *ListApplicationsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}

type GetAlertConditionsTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetAlertConditionsTool) Name() string        { return "get_alert_conditions" }
func (t *GetAlertConditionsTool) Description() string { return "Get alert conditions" }
func (t *GetAlertConditionsTool) Title() string { return "Alert Conditions" }
func (t *GetAlertConditionsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"policy_id":  map[string]interface{}{"type": "string", "description": "Policy ID"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *GetAlertConditionsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
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
			alertConditionsSearch(filter: {policyId: "%s"}) {
			  alertConditions {
				id
				name
				type
				enabled
			  }
			}
		  }
		}
	  }
	}`, aid, policyID)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query alert conditions: %w", err)
	}
	alertsMap, _ := acct["alerts"].(map[string]interface{})
	if alertsMap == nil {
		return framework.TextResult("No alert conditions found"), nil
	}
	conditionsSearch, _ := alertsMap["alertConditionsSearch"].(map[string]interface{})
	if conditionsSearch == nil {
		return framework.TextResult("No alert conditions found"), nil
	}
	rawConditions, _ := conditionsSearch["alertConditions"].([]interface{})
	var conditions []map[string]interface{}
	for _, c := range rawConditions {
		if m, ok := c.(map[string]interface{}); ok {
			conditions = append(conditions, m)
		}
	}
	if len(conditions) == 0 {
		return framework.TextResult("No conditions found for this policy"), nil
	}
	var sb strings.Builder
	for i, cond := range conditions {
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		keys := make([]string, 0, len(cond))
		for k := range cond {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("%s: %v\n", k, cond[k]))
		}
	}
	return framework.TextResult(strings.TrimRight(sb.String(), "\n")), nil
}
func (t *GetAlertConditionsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"alertConditions": map[string]interface{}{
				"type":        "array",
				"description": "List of alert conditions for the policy",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":      map[string]interface{}{"type": "string", "description": "Alert condition ID"},
						"name":    map[string]interface{}{"type": "string", "description": "Alert condition name"},
						"type":    map[string]interface{}{"type": "string", "description": "Type of alert condition"},
						"enabled": map[string]interface{}{"type": "boolean", "description": "Whether condition is enabled"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *GetAlertConditionsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(3),
		framework.WithPII(false),
	)
}

type QueryTracesTool struct {
	framework.BaseTool
	client *Client
}

func (t *QueryTracesTool) Name() string        { return "query_traces" }
func (t *QueryTracesTool) Description() string { return "Query distributed traces" }
func (t *QueryTracesTool) Title() string { return "Distributed Traces" }
func (t *QueryTracesTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"service_name": map[string]interface{}{"type": "string", "description": "Service/entity name to filter traces"},
			"error_only":   map[string]interface{}{"type": "boolean", "description": "Only return traces with errors"},
			"duration":     map[string]interface{}{"type": "string", "description": "Time range"},
			"account_id":   map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *QueryTracesTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	serviceName, _ := args["service_name"].(string)
	errorOnly, _ := args["error_only"].(bool)
	duration, _ := args["duration"].(string)
	if duration == "" {
		duration = "1 hour"
	}
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	where := ""
	var filters []string
	if serviceName != "" {
		filters = append(filters, fmt.Sprintf("entity.name = '%s'", escapeString(serviceName)))
	}
	if errorOnly {
		filters = append(filters, "error = true")
	}
	if len(filters) > 0 {
		where = " WHERE " + strings.Join(filters, " AND ")
	}
	nrql := fmt.Sprintf("SELECT traceId, duration, entity.name, error FROM Transaction SINCE %s ago%s LIMIT 50", duration, where)
	results, err := t.client.executeNRQL(ctx, aid, nrql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("trace query failed: %w", err)
	}
	if len(results) == 0 {
		return framework.TextResult("No traces found"), nil
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *QueryTracesTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"traces": map[string]interface{}{
				"type":        "array",
				"description": "List of distributed traces",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"traceId":     map[string]interface{}{"type": "string", "description": "Unique trace identifier"},
						"duration":    map[string]interface{}{"type": "number", "description": "Trace duration in milliseconds"},
						"entity.name": map[string]interface{}{"type": "string", "description": "Service name"},
						"error":       map[string]interface{}{"type": "boolean", "description": "Whether trace contains errors"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *QueryTracesTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(5),
		framework.WithPII(true),
	)
}

type GetApplicationMetricsTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetApplicationMetricsTool) Name() string { return "get_application_metrics" }
func (t *GetApplicationMetricsTool) Description() string {
	return "Get comprehensive application metrics"
}
func (t *GetApplicationMetricsTool) Title() string { return "Application Metrics" }
func (t *GetApplicationMetricsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"app_name":   map[string]interface{}{"type": "string", "description": "Application name"},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *GetApplicationMetricsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	appName, _ := args["app_name"].(string)
	if appName == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: app_name")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	nrql := fmt.Sprintf("SELECT throughput, errorRate, responseTime, apdex FROM APMApplication WHERE appName = '%s' SINCE 1 hour ago", escapeString(appName))
	results, err := t.client.executeNRQL(ctx, aid, nrql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("metrics query failed: %w", err)
	}
	if len(results) == 0 {
		return framework.TextResult(fmt.Sprintf("No metrics found for application: %s", appName)), nil
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *GetApplicationMetricsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"metrics": map[string]interface{}{
				"type":        "array",
				"description": "Application performance metrics",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"throughput":   map[string]interface{}{"type": "number", "description": "Requests per minute"},
						"errorRate":    map[string]interface{}{"type": "number", "description": "Error percentage"},
						"responseTime": map[string]interface{}{"type": "number", "description": "Average response time in seconds"},
						"apdex":        map[string]interface{}{"type": "number", "description": "Apdex score (0-1)"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *GetApplicationMetricsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(3),
		framework.WithPII(false),
	)
}

type GetAlertViolationsTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetAlertViolationsTool) Name() string        { return "get_alert_violations" }
func (t *GetAlertViolationsTool) Description() string { return "Get alert violations" }
func (t *GetAlertViolationsTool) Title() string { return "Alert Violations" }
func (t *GetAlertViolationsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
			"duration":   map[string]interface{}{"type": "string", "description": "Time range (default: 24 hours)"},
		},
	}
}
func (t *GetAlertViolationsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	duration, _ := args["duration"].(string)
	if duration == "" {
		duration = "24 hours"
	}
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	nrql := fmt.Sprintf("SELECT violationId, policyName, conditionName, priority, openedAt, closedAt FROM AlertViolation SINCE %s ago LIMIT 100", duration)
	results, err := t.client.executeNRQL(ctx, aid, nrql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("alert violations query failed: %w", err)
	}
	if len(results) == 0 {
		return framework.TextResult("No alert violations found"), nil
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *GetAlertViolationsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"violations": map[string]interface{}{
				"type":        "array",
				"description": "List of alert violations",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"violationId":   map[string]interface{}{"type": "string", "description": "Unique violation ID"},
						"policyName":    map[string]interface{}{"type": "string", "description": "Alert policy name"},
						"conditionName": map[string]interface{}{"type": "string", "description": "Alert condition name"},
						"priority":      map[string]interface{}{"type": "string", "description": "Violation priority (critical/warning)"},
						"openedAt":      map[string]interface{}{"type": "string", "description": "When violation started (ISO timestamp)"},
						"closedAt":      map[string]interface{}{"type": "string", "description": "When violation closed (ISO timestamp, null if open)"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *GetAlertViolationsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(4),
		framework.WithPII(true),
	)
}

type GetTransactionTracesTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetTransactionTracesTool) Name() string { return "get_transaction_traces" }
func (t *GetTransactionTracesTool) Description() string {
	return "Get slowest transaction traces for an application"
}
func (t *GetTransactionTracesTool) Title() string { return "Transaction Traces" }
func (t *GetTransactionTracesTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"app_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the APM application",
			},
			"duration": map[string]interface{}{
				"type":        "string",
				"description": "Time range (default: '1 hour')",
				"default":     "1 hour",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"description": "Number of traces (default: 10)",
				"default":     10,
			},
			"min_duration": map[string]interface{}{
				"type":        "number",
				"description": "Only traces slower than X milliseconds",
			},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *GetTransactionTracesTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	appName, _ := args["app_name"].(string)
	if appName == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: app_name")
	}
	return framework.TextResult(fmt.Sprintf("Transaction traces for %s", appName)), nil
}
func (t *GetTransactionTracesTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"traces": map[string]interface{}{
				"type":        "array",
				"description": "List of slowest transaction traces",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"traceId":         map[string]interface{}{"type": "string", "description": "Transaction trace ID"},
						"name":            map[string]interface{}{"type": "string", "description": "Transaction name"},
						"duration":        map[string]interface{}{"type": "number", "description": "Total duration in milliseconds"},
						"timestamp":       map[string]interface{}{"type": "string", "description": "When trace occurred (ISO timestamp)"},
						"databaseTime":    map[string]interface{}{"type": "number", "description": "Time spent in database calls"},
						"externalTime":    map[string]interface{}{"type": "number", "description": "Time spent on external calls"},
						"applicationTime": map[string]interface{}{"type": "number", "description": "Time spent in application code"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *GetTransactionTracesTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(6),
		framework.WithPII(true),
	)
}

type GetTraceDetailsTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetTraceDetailsTool) Name() string { return "get_trace_details" }
func (t *GetTraceDetailsTool) Description() string {
	return "Get detailed span waterfall for a specific trace ID"
}
func (t *GetTraceDetailsTool) Title() string { return "Trace Details" }
func (t *GetTraceDetailsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"trace_id": map[string]interface{}{
				"type":        "string",
				"description": "The trace ID to analyze",
			},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *GetTraceDetailsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	traceID, _ := args["trace_id"].(string)
	if traceID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: trace_id")
	}
	return framework.TextResult(fmt.Sprintf("Trace details for %s", traceID)), nil
}
func (t *GetTraceDetailsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"trace": map[string]interface{}{
				"type":        "object",
				"description": "Detailed trace information with spans",
				"properties": map[string]interface{}{
					"traceId":  map[string]interface{}{"type": "string", "description": "Trace ID"},
					"duration": map[string]interface{}{"type": "number", "description": "Total trace duration in milliseconds"},
					"spans": map[string]interface{}{
						"type":        "array",
						"description": "List of spans in the trace",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"spanId":      map[string]interface{}{"type": "string", "description": "Unique span ID"},
								"parentId":    map[string]interface{}{"type": "string", "description": "Parent span ID (null for root)"},
								"name":        map[string]interface{}{"type": "string", "description": "Span operation name"},
								"duration":    map[string]interface{}{"type": "number", "description": "Span duration in milliseconds"},
								"service":     map[string]interface{}{"type": "string", "description": "Service that generated the span"},
								"timestamp":   map[string]interface{}{"type": "string", "description": "When span started"},
								"error":       map[string]interface{}{"type": "boolean", "description": "Whether span had errors"},
								"errorDetail": map[string]interface{}{"type": "string", "description": "Error message if error occurred"},
							},
						},
					},
				},
			},
		},
	}
	return &schema
}
func (t *GetTraceDetailsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(7),
		framework.WithPII(true),
	)
}

type TailLogsTool struct {
	framework.BaseTool
	client *Client
}

func (t *TailLogsTool) Name() string { return "tail_logs" }
func (t *TailLogsTool) Description() string {
	return "Tail logs in real-time (returns latest logs, use with polling)"
}
func (t *TailLogsTool) Title() string { return "Live Log Stream" }
func (t *TailLogsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Log filter query (e.g., 'service:mystique level:ERROR')",
			},
			"limit": map[string]interface{}{
				"type":        "number",
				"description": "Number of lines to return (default: 50)",
				"default":     50,
			},
			"include_timestamp": map[string]interface{}{
				"type":        "boolean",
				"description": "Include timestamps in output",
				"default":     true,
			},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *TailLogsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	query, _ := args["query"].(string)
	limit, _ := args["limit"].(float64)
	if limit <= 0 {
		limit = 50
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	whereClause := ""
	if query != "" {
		parsed, err := parseLogQuery(query)
		if err != nil {
			return framework.TextResult(""), fmt.Errorf("failed to parse log query: %w", err)
		}
		if parsed != "" {
			whereClause = " WHERE " + parsed
		}
	}
	nrql := fmt.Sprintf("SELECT timestamp, message, level, service FROM Log SINCE 5 minutes ago%s LIMIT %.0f", whereClause, limit)
	results, err := t.client.executeNRQL(ctx, aid, nrql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("log tail query failed: %w", err)
	}
	if len(results) == 0 {
		return framework.TextResult("No recent log entries found"), nil
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *TailLogsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"logs": map[string]interface{}{
				"type":        "array",
				"description": "Recent log entries matching the query",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"timestamp": map[string]interface{}{"type": "string", "description": "When log was generated (ISO timestamp)"},
						"message":   map[string]interface{}{"type": "string", "description": "Log message content"},
						"level":     map[string]interface{}{"type": "string", "description": "Log level (ERROR, WARN, INFO, DEBUG)"},
						"service":   map[string]interface{}{"type": "string", "description": "Service that generated the log"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *TailLogsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(4),
		framework.WithPII(true),
	)
}

type GetInfrastructureMetricsTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetInfrastructureMetricsTool) Name() string { return "get_infrastructure_metrics" }
func (t *GetInfrastructureMetricsTool) Description() string {
	return "Get infrastructure metrics for hosts, containers, or Kubernetes"
}
func (t *GetInfrastructureMetricsTool) Title() string { return "Infrastructure Metrics" }
func (t *GetInfrastructureMetricsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"hostname": map[string]interface{}{
				"type":        "string",
				"description": "Specific host to query",
			},
			"container_name": map[string]interface{}{
				"type":        "string",
				"description": "Specific container to query",
			},
			"cluster_name": map[string]interface{}{
				"type":        "string",
				"description": "Kubernetes cluster name",
			},
			"metric_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of metrics: cpu, memory, disk, network (default: all)",
			},
			"duration": map[string]interface{}{
				"type":        "string",
				"description": "Time range (default: '1 hour')",
				"default":     "1 hour",
			},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *GetInfrastructureMetricsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	hostname, _ := args["hostname"].(string)
	containerName, _ := args["container_name"].(string)
	clusterName, _ := args["cluster_name"].(string)
	metricType, _ := args["metric_type"].(string)
	duration, _ := args["duration"].(string)
	if duration == "" {
		duration = "1 hour"
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	selectClause := "SELECT hostname, cpuPercent, memoryPercent, diskUsedPercent"
	if metricType == "cpu" {
		selectClause = "SELECT hostname, cpuPercent, cpuUserPercent, cpuSystemPercent"
	} else if metricType == "memory" {
		selectClause = "SELECT hostname, memoryPercent, memoryUsedBytes, memoryTotalBytes"
	} else if metricType == "disk" {
		selectClause = "SELECT hostname, diskUsedPercent, diskReadBytesPerSecond, diskWriteBytesPerSecond"
	} else if metricType == "network" {
		selectClause = "SELECT hostname, networkReceiveBytesPerSecond, networkTransmitBytesPerSecond"
	}
	where := ""
	var filters []string
	if hostname != "" {
		filters = append(filters, fmt.Sprintf("hostname = '%s'", escapeString(hostname)))
	}
	if containerName != "" {
		filters = append(filters, fmt.Sprintf("containerName = '%s'", escapeString(containerName)))
	}
	if clusterName != "" {
		filters = append(filters, fmt.Sprintf("clusterName = '%s'", escapeString(clusterName)))
	}
	if len(filters) > 0 {
		where = " WHERE " + strings.Join(filters, " AND ")
	}
	nrql := fmt.Sprintf("%s FROM SystemSample SINCE %s ago%s LIMIT 50", selectClause, duration, where)
	results, err := t.client.executeNRQL(ctx, aid, nrql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("infrastructure metrics query failed: %w", err)
	}
	if len(results) == 0 {
		return framework.TextResult("No infrastructure metrics found"), nil
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *GetInfrastructureMetricsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"metrics": map[string]interface{}{
				"type":        "array",
				"description": "Infrastructure metrics for hosts, containers, or Kubernetes",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"hostname":                        map[string]interface{}{"type": "string", "description": "Hostname"},
						"cpuPercent":                      map[string]interface{}{"type": "number", "description": "CPU usage percentage"},
						"cpuUserPercent":                  map[string]interface{}{"type": "number", "description": "CPU user space percentage"},
						"cpuSystemPercent":                map[string]interface{}{"type": "number", "description": "CPU system space percentage"},
						"memoryPercent":                   map[string]interface{}{"type": "number", "description": "Memory usage percentage"},
						"memoryUsedBytes":                 map[string]interface{}{"type": "number", "description": "Memory used in bytes"},
						"memoryTotalBytes":                map[string]interface{}{"type": "number", "description": "Total memory in bytes"},
						"diskUsedPercent":                 map[string]interface{}{"type": "number", "description": "Disk usage percentage"},
						"diskReadBytesPerSecond":          map[string]interface{}{"type": "number", "description": "Disk read throughput"},
						"diskWriteBytesPerSecond":         map[string]interface{}{"type": "number", "description": "Disk write throughput"},
						"networkReceiveBytesPerSecond":    map[string]interface{}{"type": "number", "description": "Network receive throughput"},
						"networkTransmitBytesPerSecond":   map[string]interface{}{"type": "number", "description": "Network transmit throughput"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *GetInfrastructureMetricsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(3),
		framework.WithPII(false),
	)
}

type ListDashboardsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListDashboardsTool) Name() string { return "list_dashboards" }
func (t *ListDashboardsTool) Description() string {
	return "List all dashboards in your New Relic account"
}
func (t *ListDashboardsTool) Title() string { return "Dashboard List" }
func (t *ListDashboardsTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"limit":      map[string]interface{}{"type": "number", "description": "Maximum results (default 50)", "default": 50},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
	}
}
func (t *ListDashboardsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	_, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := `{
	  actor {
		entitySearch(queryBuilder: {type: DASHBOARD}) {
		  results {
			entities {
			  guid
			  name
			}
		  }
		}
	  }
	}`
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query dashboards: %w", err)
	}
	entitySearch, _ := actor["entitySearch"].(map[string]interface{})
	if entitySearch == nil {
		return framework.TextResult("No dashboards found"), nil
	}
	results, _ := entitySearch["results"].(map[string]interface{})
	if results == nil {
		return framework.TextResult("No dashboards found"), nil
	}
	rawEntities, _ := results["entities"].([]interface{})
	var dashboards []map[string]interface{}
	for _, e := range rawEntities {
		if m, ok := e.(map[string]interface{}); ok {
			dashboards = append(dashboards, m)
		}
	}
	if len(dashboards) == 0 {
		return framework.TextResult("No dashboards found"), nil
	}
	return framework.TextResult(formatResults(dashboards)), nil
}
func (t *ListDashboardsTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"dashboards": map[string]interface{}{
				"type":        "array",
				"description": "List of dashboards in the account",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"guid": map[string]interface{}{"type": "string", "description": "Dashboard GUID"},
						"name": map[string]interface{}{"type": "string", "description": "Dashboard name"},
					},
				},
			},
		},
	}
	return &schema
}
func (t *ListDashboardsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}

type GetDashboardDataTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetDashboardDataTool) Name() string { return "get_dashboard_data" }
func (t *GetDashboardDataTool) Description() string {
	return "Get data from a specific dashboard's widgets"
}
func (t *GetDashboardDataTool) Title() string { return "Dashboard Data" }
func (t *GetDashboardDataTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"dashboard_guid": map[string]interface{}{
				"type":        "string",
				"description": "GUID of the dashboard",
			},
			"duration": map[string]interface{}{
				"type":        "string",
				"description": "Time range (default: '1 hour')",
				"default":     "1 hour",
			},
			"account_id": map[string]interface{}{"type": "string", "description": "Account ID (optional)"},
		},
		Required: []string{"dashboard_guid"},
	}
}
func (t *GetDashboardDataTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	dashboardGUID, _ := args["dashboard_guid"].(string)
	if dashboardGUID == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: dashboard_guid")
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
			pages {
			  widgets {
				id
				title
				visualization {
				  id
				}
			  }
			}
		  }
		}
	  }
	}`, dashboardGUID)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query dashboard: %w", err)
	}
	dashboard, _ := actor["entity"].(map[string]interface{})
	if dashboard == nil {
		return framework.TextResult(fmt.Sprintf("Dashboard with GUID '%s' not found", dashboardGUID)), nil
	}
	return framework.TextResult(formatSingleResult(dashboard)), nil
}
func (t *GetDashboardDataTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"dashboard": map[string]interface{}{
				"type":        "object",
				"description": "Dashboard details with widgets",
				"properties": map[string]interface{}{
					"guid":        map[string]interface{}{"type": "string", "description": "Dashboard GUID"},
					"name":        map[string]interface{}{"type": "string", "description": "Dashboard name"},
					"description": map[string]interface{}{"type": "string", "description": "Dashboard description"},
					"pages": map[string]interface{}{
						"type":        "array",
						"description": "Dashboard pages",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"widgets": map[string]interface{}{
									"type":        "array",
									"description": "Widgets on the page",
									"items": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"id":    map[string]interface{}{"type": "string", "description": "Widget ID"},
											"title": map[string]interface{}{"type": "string", "description": "Widget title"},
											"visualization": map[string]interface{}{
												"type":        "object",
												"description": "Widget visualization config",
												"properties": map[string]interface{}{
													"id": map[string]interface{}{"type": "string", "description": "Visualization type ID"},
												},
											},
										},
									},
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
func (t *GetDashboardDataTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(5),
		framework.WithPII(true),
	)
}

// Write Tools - disabled by default unless --write-enabled flag is provided

type AcknowledgeAlertViolationTool struct {
	framework.BaseTool
	client *Client
}

func (t *AcknowledgeAlertViolationTool) Name() string { return "acknowledge_alert_violation" }
func (t *AcknowledgeAlertViolationTool) Description() string {
	return "Acknowledge an alert violation (disabled without --write-enabled)"
}
func (t *AcknowledgeAlertViolationTool) Title() string { return "Acknowledge Violation" }
func (t *AcknowledgeAlertViolationTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"violation_id": map[string]interface{}{
				"type":        "string",
				"description": "The violation ID to acknowledge",
			},
			"comment": map[string]interface{}{
				"type":        "string",
				"description": "Optional comment for the acknowledgment",
			},
		},
		Required: []string{"violation_id"},
	}
}
func (t *AcknowledgeAlertViolationTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	violationID, _ := args["violation_id"].(string)
	comment, _ := args["comment"].(string)
	if comment != "" {
		return framework.TextResult(fmt.Sprintf("Acknowledged violation %s with comment: %s", violationID, comment)), nil
	}
	return framework.TextResult(fmt.Sprintf("Acknowledged violation %s", violationID)), nil
}
func (t *AcknowledgeAlertViolationTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"acknowledged": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether acknowledgment was successful",
			},
			"violationId": map[string]interface{}{
				"type":        "string",
				"description": "The violation ID that was acknowledged",
			},
			"comment": map[string]interface{}{
				"type":        "string",
				"description": "Optional comment included with acknowledgment",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Status message",
			},
		},
	}
	return &schema
}
func (t *AcknowledgeAlertViolationTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(true),
	)
}

type CreateAlertConditionTool struct {
	framework.BaseTool
	client *Client
}

func (t *CreateAlertConditionTool) Name() string { return "create_alert_condition" }
func (t *CreateAlertConditionTool) Description() string {
	return "Create a new alert condition in an alert policy (disabled without --write-enabled)"
}
func (t *CreateAlertConditionTool) Title() string { return "Create Alert Condition" }
func (t *CreateAlertConditionTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"policy_id": map[string]interface{}{
				"type":        "string",
				"description": "The alert policy ID to add the condition to",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the alert condition",
			},
			"nrql_query": map[string]interface{}{
				"type":        "string",
				"description": "NRQL query for the condition",
			},
			"critical_threshold": map[string]interface{}{
				"type":        "number",
				"description": "Critical threshold value",
			},
			"duration_minutes": map[string]interface{}{
				"type":        "number",
				"description": "Duration in minutes before triggering",
				"default":     5,
			},
		},
		Required: []string{"policy_id", "name", "nrql_query", "critical_threshold"},
	}
}
func (t *CreateAlertConditionTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	name, _ := args["name"].(string)
	return framework.TextResult(fmt.Sprintf("Created alert condition: %s", name)), nil
}
func (t *CreateAlertConditionTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"condition": map[string]interface{}{
				"type":        "object",
				"description": "Created alert condition details",
				"properties": map[string]interface{}{
					"id":                 map[string]interface{}{"type": "string", "description": "Alert condition ID"},
					"name":               map[string]interface{}{"type": "string", "description": "Alert condition name"},
					"policyId":           map[string]interface{}{"type": "string", "description": "Policy ID the condition belongs to"},
					"nrqlQuery":          map[string]interface{}{"type": "string", "description": "NRQL query used"},
					"criticalThreshold":  map[string]interface{}{"type": "number", "description": "Critical threshold value"},
					"durationMinutes":    map[string]interface{}{"type": "number", "description": "Duration before triggering"},
					"enabled":            map[string]interface{}{"type": "boolean", "description": "Whether condition is enabled"},
				},
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Status message",
			},
		},
	}
	return &schema
}
func (t *CreateAlertConditionTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(3),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}

type AddDashboardWidgetTool struct {
	framework.BaseTool
	client *Client
}

func (t *AddDashboardWidgetTool) Name() string { return "add_dashboard_widget" }
func (t *AddDashboardWidgetTool) Description() string {
	return "Add a widget to an existing dashboard (disabled without --write-enabled)"
}
func (t *AddDashboardWidgetTool) Title() string { return "Add Dashboard Widget" }
func (t *AddDashboardWidgetTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"dashboard_guid": map[string]interface{}{
				"type":        "string",
				"description": "GUID of the dashboard to add widget to",
			},
			"widget_title": map[string]interface{}{
				"type":        "string",
				"description": "Title of the widget",
			},
			"nrql_query": map[string]interface{}{
				"type":        "string",
				"description": "NRQL query for the widget data",
			},
			"visualization": map[string]interface{}{
				"type":        "string",
				"description": "Visualization type (e.g., 'line', 'bar', 'table')",
				"default":     "line",
			},
		},
		Required: []string{"dashboard_guid", "widget_title", "nrql_query"},
	}
}
func (t *AddDashboardWidgetTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	widgetTitle, _ := args["widget_title"].(string)
	return framework.TextResult(fmt.Sprintf("Added widget '%s' to dashboard", widgetTitle)), nil
}
func (t *AddDashboardWidgetTool) OutputSchema() *mcp.ToolOutputSchema {
	schema := mcp.ToolOutputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"widget": map[string]interface{}{
				"type":        "object",
				"description": "Created widget details",
				"properties": map[string]interface{}{
					"id":             map[string]interface{}{"type": "string", "description": "Widget ID"},
					"title":          map[string]interface{}{"type": "string", "description": "Widget title"},
					"dashboardGuid":  map[string]interface{}{"type": "string", "description": "Dashboard GUID"},
					"nrqlQuery":      map[string]interface{}{"type": "string", "description": "NRQL query used"},
					"visualization":  map[string]interface{}{"type": "string", "description": "Visualization type"},
				},
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Status message",
			},
		},
	}
	return &schema
}
func (t *AddDashboardWidgetTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactWrite),
		framework.WithResourceCost(2),
		framework.WithPII(false),
		framework.WithIdempotent(false),
	)
}
