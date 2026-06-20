# New Relic MCP Server

A Go-native MCP (Model Context Protocol) server for the New Relic observability platform, providing comprehensive access to APM, infrastructure, logs, traces, alerts, dashboards, synthetics, workflows, notifications, service levels, tags, entities, and workloads — all through a single GraphQL endpoint.

## Overview

- **55 registered MCP tools** covering the full NerdGraph API surface
- **Safety-first design**: every tool carries `EnforcerProfile` metadata (risk, impact, resource cost, PII flag, idempotency) for automated policy enforcement via MCP Bridge
- **Write protection**: all mutation tools gated behind a `-write-enabled` flag; disabled by default
- **Multi-region**: US and EU data center support (`NEWRELIC_REGION=us|eu`)
- **TDD-built**: 80%+ test coverage across all tools

## API Architecture

All tools communicate with a single NerdGraph endpoint:

| Region | Endpoint |
|--------|----------|
| US (default) | `https://api.newrelic.com/graphql` |
| EU | `https://api.eu.newrelic.com/graphql` |

Internal query helpers determine the response-scope:

| Helper | Returns | Used For |
|--------|---------|----------|
| `actorGraphQuery` | `data.actor.*` | Entity queries, cross-account NRQL |
| `nerdGraphQuery` | `data.actor.account.*` | Account-scoped queries (alerts, workflows, etc.) |
| `rawGraphQuery` | `data.<mutationName>` | All mutations |
| `executeNRQL` | Parsed NRQL results | NRQL-based tools |

## Tools

### NRQL Querying

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `nrql_query` | Execute arbitrary NRQL queries | `actor.account.nrql` | Med | Read |

### APM & Applications

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `list_applications` | List all APM-monitored applications | `actor.entitySearch` (domain:APM, type:APPLICATION) | Low | Read |
| `get_apm_metrics` | Get APM performance metrics | `actor.account.nrql` (FROM APMApplication) | Low | Read |
| `get_application_metrics` | Get detailed metrics for a specific app | `actor.account.nrql` (FROM APMApplication) | Low | Read |
| `query_traces` | Query distributed traces | `actor.account.nrql` (FROM Transaction) | Low | Read |
| `get_transaction_traces` | Get transaction trace details | NRQL via executeNRQL | Low | Read |
| `get_trace_details` | Get specific trace details | NRQL via executeNRQL | Low | Read |

### Alert Policies & Conditions

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `list_alerts` | List all alert policies | `actor.account.alerts.policiesSearch` | Low | Read |
| `get_alert_policy` | Get a single alert policy by ID | `actor.account.alerts.policy` | Low | Read |
| `get_alert_conditions` | Get conditions for a policy | `actor.account.alerts.alertConditionsSearch` | Low | Read |
| `get_alert_violations` | Get active alert violations | `actor.account.nrql` (NRQL alert query) | Low | Read |
| `list_nrql_alert_conditions` | List NRQL conditions for a policy | `actor.account.alerts.nrqlConditionsSearch` | Low | Read |
| `create_alert_policy` | Create a new alert policy | `alertsPolicyCreate` | Low | Write |
| `update_alert_policy` | Update an alert policy | `alertsPolicyUpdate` | Low | Write |
| `delete_alert_policy` | Delete an alert policy | `alertsPolicyDelete` | High | Write |
| `create_alert_condition` | Create an NRQL alert condition | `alertsNrqlConditionCreate` | Med | Write |
| `acknowledge_alert_violation` | Acknowledge an alert violation | `alertsViolationAcknowledge` | Med | Write |

### Synthetic Monitoring

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `list_synthetic_monitors` | List all synthetic monitors | `actor.entitySearch` (domain:SYNTH, type:MONITOR) | Low | Read |
| `get_synthetic_monitor` | Get a synthetic monitor by GUID | `actor.entity` (SyntheticMonitorEntityOutline) | Low | Read |
| `list_private_locations` | List private locations | `actor.entitySearch` (domain:SYNTH, type:PRIVATE_LOCATION) | Low | Read |
| `create_ping_monitor` | Create a simple ping monitor | `syntheticsCreateSimpleMonitor` | Low | Write |
| `delete_synthetic_monitor` | Delete a synthetic monitor | `syntheticsDeleteMonitor` | Med | Write |

### Dashboards

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `list_dashboards` | List all dashboards | `actor.entitySearch` (domain:DASH, type:DASHBOARD) | Low | Read |
| `get_dashboard` | Get full dashboard configuration by GUID | `actor.entity` (DashboardEntity) | Low | Read |
| `get_dashboard_data` | Get dashboard widget data | `actor.entity` (DashboardEntity) | Low | Read |
| `create_dashboard` | Create a new dashboard | `dashboardCreate` | Low | Write |
| `update_dashboard` | Update an existing dashboard | `dashboardUpdate` | Low | Write |
| `delete_dashboard` | Delete a dashboard | `dashboardDelete` | Med | Write |
| `add_dashboard_widget` | Add a widget to a dashboard | `dashboardUpdate` | Med | Write |

### Logs

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `search_logs` | Search logs with Lucene syntax | `actor.account.nrql` (FROM Log) | Med | Read |
| `tail_logs` | Stream recent log entries | `actor.account.nrql` (FROM Log) | Med | Read |

### Infrastructure

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `get_infrastructure_metrics` | Get host/container metrics | `actor.account.nrql` (FROM SystemSample) | Low | Read |

### Alert Workflows (aiWorkflows)

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `list_workflows` | List alert notification workflows | `actor.account.aiWorkflows.workflows` | Low | Read |
| `get_workflow` | Get a workflow by ID | `actor.account.aiWorkflows.workflows` (filter) | Low | Read |
| `create_workflow` | Create a new alert workflow | `aiWorkflowsCreateWorkflow` | Low | Write |
| `update_workflow` | Update a workflow | `aiWorkflowsUpdateWorkflow` | Low | Write |
| `delete_workflow` | Delete a workflow | `aiWorkflowsDeleteWorkflow` | High | Write |

### Notification Channels & Destinations (aiNotifications)

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `list_notification_channels` | List all notification channels | `actor.account.aiNotifications.channels` | Low | Read |
| `list_destinations` | List notification destinations | `actor.account.aiNotifications.destinations` | Low | Read |
| `create_slack_channel` | Create a Slack notification channel | `aiNotificationsCreateChannel` (type:SLACK) | Low | Write |
| `create_email_channel` | Create an email notification channel | `aiNotificationsCreateChannel` (type:EMAIL) | Low | Write |
| `delete_notification_channel` | Delete a notification channel | `aiNotificationsDeleteChannel` | High | Write |

### Service Levels

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `list_service_levels` | List SLIs for an entity | `actor.entity` (... EntityInterfaceWithServiceLevel) | Low | Read |
| `create_service_level` | Create an SLI with SLO | `serviceLevelCreate` | Med | Write |
| `update_service_level` | Update an SLI's SLO target | `serviceLevelUpdate` | Low | Write |

### Tag Management (tagging)

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `get_entity_tags` | Get all tags for an entity | `actor.entity.tags` | Low | Read |
| `add_entity_tags` | Add tags to an entity | `taggingAddTagsToEntity` | Low | Write |
| `remove_entity_tags` | Remove tags from an entity | `taggingDeleteTagFromEntity` | Med | Write |
| `replace_entity_tags` | Replace all tags on an entity | `taggingReplaceTagsOnEntity` | Med | Write |

### Entity Operations

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `search_entities` | Advanced entity search | `actor.entitySearch` (queryBuilder) | Low | Read |
| `delete_entity` | Delete entities by GUID | `entityDelete` | High | Write |

### Cross-Account NRQL

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `cross_account_nrql` | Run NRQL across multiple accounts | `actor.nrql` (multi-account) | Med | Read |

### Workloads

| Tool | Description | GraphQL Operation | Risk | Impact |
|------|-------------|-------------------|------|--------|
| `list_workloads` | List all workloads for an account | `actor.account.workload.workloads` | Low | Read |
| `get_workload` | Get a workload by GUID | `actor.entity` (WorkloadEntity) | Low | Read |

## NerdGraph Reference

All tools access the [New Relic NerdGraph API](https://docs.newrelic.com/docs/apis/nerdgraph/get-started/introduction-new-relic-nerdgraph/). The following NerdGraph query/mutation names are used:

| GraphQL Operation | Documentation |
|-------------------|---------------|
| `actor.entity` | [Entity](https://docs.newrelic.com/docs/apis/nerdgraph/examples/use-nerdgraph-manage-entities/) |
| `actor.entitySearch` | [Entity Search](https://docs.newrelic.com/docs/apis/nerdgraph/examples/use-nerdgraph-entity-search/) |
| `actor.account.nrql` | [NRQL](https://docs.newrelic.com/docs/apis/nerdgraph/examples/use-nerdgraph-query-data-nrql/) |
| `alertsPolicyCreate / Update / Delete` | [Alert Policies](https://docs.newrelic.com/docs/alerts-applied-intelligence/new-relic-alerts/alert-policies/manage-alert-policies-using-nerdgraph/) |
| `alertsNrqlConditionCreate` | [NRQL Alert Conditions](https://docs.newrelic.com/docs/alerts-applied-intelligence/new-relic-alerts/alert-conditions/create-nrql-alert-conditions-using-nerdgraph/) |
| `syntheticsCreateSimpleMonitor` | [Synthetic Monitoring](https://docs.newrelic.com/docs/synthetics/synthetic-monitoring/using-monitors/create-simple-browser-monitor-nerdgraph/) |
| `dashboardCreate / Update / Delete` | [Dashboards](https://docs.newrelic.com/docs/query-your-data/explore-query-data/dashboards/manage-dashboards-nerdgraph/) |
| `aiWorkflowsCreate / Update / DeleteWorkflow` | [Workflows](https://docs.newrelic.com/docs/alerts-applied-intelligence/applied-intelligence/incident-workflows/use-nerdgraph-manage-workflows/) |
| `aiNotificationsCreateChannel / DeleteChannel` | [Notification Channels](https://docs.newrelic.com/docs/alerts-applied-intelligence/notifications/manage-notification-channels-using-nerdgraph/) |
| `serviceLevelCreate / Update` | [Service Levels](https://docs.newrelic.com/docs/service-level-management/slm-api-create-sli/) |
| `taggingAddTagsToEntity / DeleteTagFromEntity / ReplaceTagsOnEntity` | [Tags](https://docs.newrelic.com/docs/apis/nerdgraph/examples/use-nerdgraph-manage-tags/) |
| `entityDelete` | [Entity Deletion](https://docs.newrelic.com/docs/apis/nerdgraph/examples/use-nerdgraph-delete-entities/) |

## Installation

### Prerequisites

- Go 1.22 or later
- New Relic account with API access

### Building from Source

```bash
git clone https://github.com/karldane/newrelic-mcp.git
cd newrelic-mcp
make
```

#### Build Options

```bash
make              # Download deps and build (default)
make deps         # Download dependencies only
make build        # Build binary only (assumes deps exist)
make build-all    # Build for Linux, macOS, and Windows
make test         # Run tests
make clean        # Remove build artifacts
make install      # Install to GOPATH/bin
make help         # Show all options
```

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `NEWRELIC_API_KEY` | New Relic API key (User or Admin) | Yes | - |
| `NEWRELIC_REGION` | Data center region (`us` or `eu`) | No | `us` |

### Obtaining an API Key

1. Log in to your New Relic account
2. Go to **API Keys** in the account settings
3. Create a **User Key** (for read-only) or **Admin Key** (for write operations)
4. Copy the key and set it as `NEWRELIC_API_KEY`

## Usage

### Basic Usage (Read-Only)

```bash
export NEWRELIC_API_KEY="your-api-key"
./newrelic-mcp
```

### Enable Write Operations

Write operations require the `-write-enabled` flag:

```bash
export NEWRELIC_API_KEY="your-api-key"
./newrelic-mcp -write-enabled
```

Write tools are disabled by default. When `-write-enabled` is set, all mutations (create, update, delete, acknowledge) become available. Each write tool carries an `EnforcerProfile` with `ImpactWrite` so MCP Bridge operators can enforce policies independently.

### Tools List

Query the full tool list via MCP:

```json
{
  "tool": "tools/list"
}
```

## Safety Features

### EnforcerProfile Metadata

Every tool declares its safety profile:

| Field | Description | Values |
|-------|-------------|--------|
| `RiskLevel` | Potential impact of misuse | low, med, high, critical |
| `ImpactScope` | What the tool affects | read, write |
| `ResourceCost` | API call weight (1-10) | integer |
| `PIIExposure` | Whether tool returns sensitive data | true / false |
| `Idempotent` | Whether retries are safe | true / false |

This metadata enables automated policy enforcement by MCP Bridge without hardcoding tool lists.

### Risk Classification

- **Low Risk**: Read-only queries, log searches, metric retrieval, listing operations
- **Medium Risk**: Creating non-destructive resources (alert conditions, service levels, dashboards, channels), acknowledging alerts
- **High Risk**: Deleting resources (policies, entities, dashboards, workflows, channels)

### Write Protection

Write tools are disabled by default and require:
1. Valid New Relic API key with appropriate permissions
2. `-write-enabled` command-line flag

## Examples

### Query Application Performance

```json
{
  "tool": "nrql_query",
  "arguments": {
    "query": "SELECT average(duration) FROM Transaction WHERE appName = 'MyApp' TIMESERIES SINCE 1 hour ago"
  }
}
```

### List Active Alerts

```json
{
  "tool": "get_alert_violations",
  "arguments": {}
}
```

### Create a Ping Monitor

```json
{
  "tool": "create_ping_monitor",
  "arguments": {
    "name": "My Monitor",
    "uri": "https://example.com/health",
    "locations": "US_EAST_1,US_WEST_1",
    "period": "EVERY_5_MINUTES"
  }
}
```

### Cross-Account NRQL Query

```json
{
  "tool": "cross_account_nrql",
  "arguments": {
    "query": "SELECT count(*) FROM Transaction",
    "account_ids": "1234567,7654321"
  }
}
```

## Architecture

The server wraps the New Relic GraphQL client and registers each tool through the MCP Framework. The tools map 1:1 to NerdGraph operations — no REST API calls are made.

```
Client (MCP Host)
    │
    ▼
newrelic-mcp (MCP Server)
    │
    ├── actorGraphQuery  →  actor.*  (entities, entitySearch, cross-account NRQL)
    ├── nerdGraphQuery   →  actor.account.*  (alerts, workflows, notifications, workloads)
    ├── rawGraphQuery    →  mutation results
    └── executeNRQL      →  actor.account.nrql  (NRQL query tools)
         │
         ▼
New Relic NerdGraph (api.newrelic.com/graphql)
```

## Testing

```bash
go test ./newrelic -v
```

Tests cover:
- Happy path (mock NerdGraph responses)
- Empty/no-results responses
- Missing required parameters
- Write-enabled / write-disabled gating
- API error responses
- Unexpected response format

## License

This project is licensed under the Functional Source License, Version 1.1, ALv2 Future License.

Copyright 2026 Karl Dane

See LICENSE file for full terms.

## References

- [MCP Framework](https://github.com/karldane/mcp-framework) — Base framework with EnforcerProfile support
- [New Relic NerdGraph API](https://docs.newrelic.com/docs/apis/nerdgraph/get-started/introduction-new-relic-nerdgraph/)
- [NRQL Reference](https://docs.newrelic.com/docs/query-your-data/nrql-new-relic-query-language/get-started/introduction-nrql-new-relics-query-language/)
- [New Relic API Explorer](https://api.newrelic.com/graphiql) — Interactive GraphQL playground

## Tool Inventory

The server registers **55 tools** total:

- **6 NRQL / APM / Trace tools**: nrql_query, list_applications, get_apm_metrics, get_application_metrics, query_traces, get_transaction_traces, get_trace_details
- **10 Alert tools**: list_alerts, get_alert_policy, get_alert_conditions, get_alert_violations, list_nrql_alert_conditions, create_alert_policy, update_alert_policy, delete_alert_policy, create_alert_condition, acknowledge_alert_violation
- **5 Synthetic Monitoring tools**: list_synthetic_monitors, get_synthetic_monitor, list_private_locations, create_ping_monitor, delete_synthetic_monitor
- **7 Dashboard tools**: list_dashboards, get_dashboard, get_dashboard_data, create_dashboard, update_dashboard, delete_dashboard, add_dashboard_widget
- **2 Log tools**: search_logs, tail_logs
- **1 Infrastructure tool**: get_infrastructure_metrics
- **5 Workflow tools**: list_workflows, get_workflow, create_workflow, update_workflow, delete_workflow
- **5 Notification tools**: list_notification_channels, list_destinations, create_slack_channel, create_email_channel, delete_notification_channel
- **3 Service Level tools**: list_service_levels, create_service_level, update_service_level
- **4 Tag tools**: get_entity_tags, add_entity_tags, remove_entity_tags, replace_entity_tags
- **2 Entity tools**: search_entities, delete_entity
- **1 Cross-account NRQL tool**: cross_account_nrql
- **2 Workload tools**: list_workloads, get_workload
