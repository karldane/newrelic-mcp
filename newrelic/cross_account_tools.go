package newrelic

import (
	"fmt"
	"strings"

	"github.com/karldane/mcp-framework/framework"
	"github.com/mark3labs/mcp-go/mcp"
)

type CrossAccountNRQLTool struct {
	framework.BaseTool
	client *Client
}

func (t *CrossAccountNRQLTool) Name() string        { return "cross_account_nrql" }
func (t *CrossAccountNRQLTool) Description() string { return "Run NRQL query across multiple accounts" }
func (t *CrossAccountNRQLTool) Schema() mcp.ToolInputSchema {
	return mcp.ToolInputSchema{
		Type: "object",
		Properties: map[string]interface{}{
			"query":       map[string]interface{}{"type": "string", "description": "NRQL query"},
			"account_ids": map[string]interface{}{"type": "string", "description": "Comma-separated account IDs"},
			"timeout":     map[string]interface{}{"type": "number", "description": "Query timeout in seconds (optional)"},
		},
		Required: []string{"query", "account_ids"},
	}
}
func (t *CrossAccountNRQLTool) Handle(ctx framework.CallContext, args map[string]interface{}) (framework.ToolResult, error) {
	query, _ := args["query"].(string)
	accountIDs, _ := args["account_ids"].(string)
	if query == "" || accountIDs == "" {
		return framework.TextResult(""), fmt.Errorf("query and account_ids are required")
	}
	parts := strings.Split(accountIDs, ",")
	var trimmed []string
	for _, p := range parts {
		trimmed = append(trimmed, strings.TrimSpace(p))
	}
	idsParam := strings.Join(trimmed, ",")

	timeout, _ := args["timeout"].(float64)
	if timeout <= 0 {
		timeout = 30
	}
	safeQuery := escapeString(query)
	gql := fmt.Sprintf(`{
	  actor {
		nrql(accounts: [%s], query: "%s", timeout: %.0f) {
		  results
		  metadata {
			timeWindow { begin end }
			facets
		  }
		}
	  }
	}`, idsParam, safeQuery, timeout)
	actor, err := t.client.actorGraphQuery(ctx, gql)
	if err != nil {
		return framework.TextResult(""), fmt.Errorf("cross-account NRQL query failed: %w", err)
	}
	nrqlResult, _ := actor["nrql"].(map[string]interface{})
	if nrqlResult == nil {
		return framework.TextResult("No results found"), nil
	}
	if errMsg, ok := nrqlResult["error"].(string); ok && errMsg != "" {
		return framework.TextResult(""), fmt.Errorf("NRQL error: %s", errMsg)
	}
	rawResults, _ := nrqlResult["results"].([]interface{})
	var results []map[string]interface{}
	for _, r := range rawResults {
		if m, ok := r.(map[string]interface{}); ok {
			results = append(results, m)
		}
	}
	if len(results) == 0 {
		return framework.TextResult("No results found"), nil
	}
	return framework.TextResult(formatResults(results)), nil
}
func (t *CrossAccountNRQLTool) EnforcerProfile(args map[string]interface{}) *framework.EnforcerProfile {
	return framework.NewEnforcerProfile(
		framework.WithRisk(framework.RiskMed),
		framework.WithImpact(framework.ImpactRead),
		framework.WithResourceCost(5),
		framework.WithPII(true),
	)
}
