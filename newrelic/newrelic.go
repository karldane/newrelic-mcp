package newrelic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	defaultNREndpoint   = "https://api.newrelic.com/graphql"
	defaultNREndpointEU = "https://api.eu.newrelic.com/graphql"
)

type Client struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
	accountID  string
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		endpoint:   defaultNREndpoint,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func NewClientWithRegion(apiKey, region string) *Client {
	endpoint := defaultNREndpoint
	if region == "eu" {
		endpoint = defaultNREndpointEU
	}
	return &Client{
		apiKey:     apiKey,
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func NewClientWithEndpoint(apiKey, endpoint string) *Client {
	client := NewClient(apiKey)
	client.endpoint = endpoint
	return client
}

func (c *Client) actorGraphQuery(ctx context.Context, gql string) (map[string]interface{}, error) {
	result, err := c.Query(ctx, gql, nil)
	if err != nil {
		return nil, err
	}
	data, _ := result["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("no data in response")
	}
	actor, _ := data["actor"].(map[string]interface{})
	if actor == nil {
		return nil, fmt.Errorf("no actor in response")
	}
	return actor, nil
}

func (c *Client) nerdGraphQuery(ctx context.Context, gql string) (map[string]interface{}, error) {
	result, err := c.Query(ctx, gql, nil)
	if err != nil {
		return nil, err
	}
	data, _ := result["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("no data in response")
	}
	actor, _ := data["actor"].(map[string]interface{})
	if actor == nil {
		return nil, fmt.Errorf("no actor in response")
	}
	acct, _ := actor["account"].(map[string]interface{})
	if acct == nil {
		return nil, fmt.Errorf("no account in response")
	}
	return acct, nil
}

func (c *Client) GetAccountID(ctx context.Context) (string, error) {
	if c.accountID != "" {
		return c.accountID, nil
	}
	query := `query { actor { accounts { id name } } }`
	result, err := c.Query(ctx, query, nil)
	if err != nil {
		return "", err
	}
	data, _ := result["data"].(map[string]interface{})
	actor, _ := data["actor"].(map[string]interface{})
	accounts, _ := actor["accounts"].([]interface{})
	if len(accounts) > 0 {
		account, _ := accounts[0].(map[string]interface{})
		id, _ := account["id"].(float64)
		c.accountID = fmt.Sprintf("%.0f", id)
	}
	if c.accountID == "" {
		return "", fmt.Errorf("no accounts found for API key")
	}
	return c.accountID, nil
}

// getOrDetectAccountID returns the provided accountID if non-empty,
// otherwise auto-detects it via the GraphQL API.
func (c *Client) getOrDetectAccountID(ctx context.Context, providedID string) (string, error) {
	if providedID != "" {
		return providedID, nil
	}
	return c.GetAccountID(ctx)
}

// executeNRQL sends an NRQL query via NerdGraph and returns the results array.
func (c *Client) executeNRQL(ctx context.Context, accountID, nrql string) ([]map[string]interface{}, error) {
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  nrql(query: %q, timeout: 30) {
			results
			metadata {
			  timeWindow { begin end }
			  facets
			}
		  }
		}
	  }
	}`, accountID, nrql)
	result, err := c.Query(ctx, gql, nil)
	if err != nil {
		return nil, fmt.Errorf("NRQL query failed: %w", err)
	}
	data, _ := result["data"].(map[string]interface{})
	if data == nil {
		return nil, fmt.Errorf("no data in response")
	}
	actor, _ := data["actor"].(map[string]interface{})
	if actor == nil {
		return nil, fmt.Errorf("no actor in response")
	}
	acct, _ := actor["account"].(map[string]interface{})
	if acct == nil {
		return nil, fmt.Errorf("no account in response")
	}
	nrqlResult, _ := acct["nrql"].(map[string]interface{})
	if nrqlResult == nil {
		return nil, fmt.Errorf("no nrql result in response")
	}
	if errMsg, ok := nrqlResult["error"].(string); ok && errMsg != "" {
		return nil, fmt.Errorf("NRQL error: %s", errMsg)
	}
	rawResults, _ := nrqlResult["results"].([]interface{})
	var results []map[string]interface{}
	for _, r := range rawResults {
		if m, ok := r.(map[string]interface{}); ok {
			results = append(results, m)
		}
	}
	return results, nil
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case map[string]interface{}, []interface{}:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatResults(results []map[string]interface{}) string {
	if len(results) == 0 {
		return "No results found"
	}
	var sb strings.Builder
	for i, r := range results {
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		keys := make([]string, 0, len(r))
		for k := range r {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, formatValue(r[k])))
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

func formatSingleResult(r map[string]interface{}) string {
	if len(r) == 0 {
		return ""
	}
	var sb strings.Builder
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s: %s\n", k, formatValue(r[k])))
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}) (map[string]interface{}, error) {
	requestBody := map[string]interface{}{"query": query, "variables": variables}
	jsonBody, _ := json.Marshal(requestBody)
	req, _ := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-Key", c.apiKey)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)
	return result, nil
}

type Server struct {
	*framework.Server
	client *Client
}

func NewServer(apiKey string, writeEnabled ...bool) *Server {
	enabled := false
	if len(writeEnabled) > 0 {
		enabled = writeEnabled[0]
	}
	return newServerWithClient(NewClient(apiKey), enabled)
}

func NewServerWithRegion(apiKey, region string, writeEnabled ...bool) *Server {
	enabled := false
	if len(writeEnabled) > 0 {
		enabled = writeEnabled[0]
	}
	return newServerWithClient(NewClientWithRegion(apiKey, region), enabled)
}

func NewServerWithEndpoint(apiKey, endpoint string, writeEnabled ...bool) *Server {
	enabled := false
	if len(writeEnabled) > 0 {
		enabled = writeEnabled[0]
	}
	return newServerWithClient(NewClientWithEndpoint(apiKey, endpoint), enabled)
}

func newServerWithClient(client *Client, writeEnabled bool) *Server {
	config := &framework.Config{
		Name:         "newrelic-mcp",
		Version:      "1.0.0",
		Instructions: "New Relic MCP Server with tools for querying data and managing alerts.",
	}
	s := &Server{Server: framework.NewServerWithConfig(config), client: client}
	s.SetWriteEnabled(writeEnabled)
	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	s.RegisterTool(&NRQLQueryTool{client: s.client})
	s.RegisterTool(&ListAlertsTool{client: s.client})
	s.RegisterTool(&GetAPMMetricsTool{client: s.client})
	s.RegisterTool(&SearchLogsTool{client: s.client})
	s.RegisterTool(&ListApplicationsTool{client: s.client})
	s.RegisterTool(&GetAlertConditionsTool{client: s.client})
	s.RegisterTool(&QueryTracesTool{client: s.client})
	s.RegisterTool(&GetApplicationMetricsTool{client: s.client})
	s.RegisterTool(&GetAlertViolationsTool{client: s.client})
	s.RegisterTool(&GetTransactionTracesTool{client: s.client})
	s.RegisterTool(&GetTraceDetailsTool{client: s.client})
	s.RegisterTool(&TailLogsTool{client: s.client})
	s.RegisterTool(&GetInfrastructureMetricsTool{client: s.client})
	s.RegisterTool(&ListDashboardsTool{client: s.client})
	s.RegisterTool(&GetDashboardDataTool{client: s.client})
	// Write tools - disabled by default
	s.RegisterTool(&AcknowledgeAlertViolationTool{client: s.client})
	s.RegisterTool(&CreateAlertConditionTool{client: s.client})
	s.RegisterTool(&AddDashboardWidgetTool{client: s.client})
}

type NRQLQueryTool struct {
	framework.BaseTool
	client *Client
}

func (t *NRQLQueryTool) Name() string        { return "nrql_query" }
func (t *NRQLQueryTool) Description() string { return "Execute NRQL queries" }
func (t *NRQLQueryTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"query": map[string]interface{}{"type": "string", "description": "NRQL query"},
		},
		Required: []string{"query"},
	}
}
func (t *NRQLQueryTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return framework.TextResult(""), fmt.Errorf("missing required parameter: query")
	}
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	results, err := t.client.executeNRQL(ctx, aid, query)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("NRQL query failed: %w", err)
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *NRQLQueryTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(4),
		framework.WithPII(true),
	)
}

type ListAlertsTool struct {
	framework.BaseTool
	client *Client
}

func (t *ListAlertsTool) Name() string                { return "list_alerts" }
func (t *ListAlertsTool) Description() string         { return "List alert policies" }
func (t *ListAlertsTool) Schema() mcp.ToolInputSchema { return mcp.ToolInputSchema{Type: "object"} }
func (t *ListAlertsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	gql := fmt.Sprintf(`{
	  actor {
		account(id: %s) {
		  alerts {
			policiesSearch {
			  policies {
				id
				name
				incidentPreference
			  }
			}
		  }
		}
	  }
	}`, aid)
	acct, err := t.client.nerdGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to query alerts: %w", err)
	}
	alertsMap, _ := acct["alerts"].(map[string]interface{})
	if alertsMap == nil {
		return framework.TextResult("No alert policies found"), nil
	}
	policiesSearch, _ := alertsMap["policiesSearch"].(map[string]interface{})
	if policiesSearch == nil {
		return framework.TextResult("No alert policies found"), nil
	}
	rawPolicies, _ := policiesSearch["policies"].([]interface{})
	var policies []map[string]interface{}
	for _, p := range rawPolicies {
		if m, ok := p.(map[string]interface{}); ok {
			policies = append(policies, m)
		}
	}
	if len(policies) == 0 {
		return framework.TextResult("No alert policies found"), nil
	}
	return framework.TextResult(formatResults(policies)), nil
}
func (t *ListAlertsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(2),
		framework.WithPII(false),
	)
}

type GetAPMMetricsTool struct {
	framework.BaseTool
	client *Client
}

func (t *GetAPMMetricsTool) Name() string                { return "get_apm_metrics" }
func (t *GetAPMMetricsTool) Description() string         { return "Get APM metrics" }
func (t *GetAPMMetricsTool) Schema() mcp.ToolInputSchema { return mcp.ToolInputSchema{Type: "object"} }
func (t *GetAPMMetricsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	appName, _ := args["app_name"].(string)
	duration, _ := args["duration"].(string)
	if duration == "" {
		duration = "1 hour"
	}
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	nrql := fmt.Sprintf("SELECT appName, duration, throughput, errorPercentage FROM APMApplication WHERE appName = '%s' SINCE %s", appName, duration)
	results, err := t.client.executeNRQL(ctx, aid, nrql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("APM metrics query failed: %w", err)
	}
	if len(results) == 0 {
		return framework.TextResult(fmt.Sprintf("No APM metrics found for %s", appName)), nil
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *GetAPMMetricsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskLow),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(3),
		framework.WithPII(false),
	)
}

type SearchLogsTool struct {
	framework.BaseTool
	client *Client
}

func (t *SearchLogsTool) Name() string                { return "search_logs" }
func (t *SearchLogsTool) Description() string         { return "Search logs" }
func (t *SearchLogsTool) Schema() mcp.ToolInputSchema { return mcp.ToolInputSchema{Type: "object"} }
func (t *SearchLogsTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	accountID, _ := args["account_id"].(string)
	queryVal, _ := args["query"].(string)
	duration, _ := args["duration"].(string)
	if duration == "" {
		duration = "30 minutes"
	}
	aid, err := t.client.getOrDetectAccountID(ctx, accountID)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("failed to get account ID: %w", err)
	}
	whereClause := ""
	if queryVal != "" {
		parsed, err := parseLogQuery(queryVal)
		if err != nil {
			return framework.TextResult(""), fmt.Errorf("failed to parse log query: %w", err)
		}
		if parsed != "" {
			whereClause = " WHERE " + parsed
		}
	}
	nrql := fmt.Sprintf("SELECT timestamp, message, level, service FROM Log SINCE %s%s LIMIT 100", duration, whereClause)
	results, err := t.client.executeNRQL(ctx, aid, nrql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("log search failed: %w", err)
	}
	if len(results) == 0 {
		return framework.TextResult("No log entries found"), nil
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *SearchLogsTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(5),
		framework.WithPII(true),
	)
}
