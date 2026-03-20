# New Relic MCP Server

A Go-native MCP (Model Context Protocol) server for New Relic observability platform, providing comprehensive access to APM metrics, logs, alerts, and infrastructure monitoring with self-reporting safety metadata.

## Overview

This server integrates with the New Relic observability platform to provide:

- **Full Stack Observability**: Access to APM, infrastructure, logs, and traces
- **NRQL Query Support**: Execute custom NRQL queries for advanced analytics
- **Alert Management**: View and acknowledge alert violations
- **Safety First**: All tools include `EnforcerProfile` metadata for automated policy enforcement
- **Multi-Region Support**: Works with both US and EU data centers

## Tools

### APM & Applications (Read-Only)

| Tool | Description | Risk | Impact | Approval |
|------|-------------|------|--------|----------|
| `nrql_query` | Execute custom NRQL queries | Med | Read | No |
| `list_applications` | List all APM-monitored applications | Low | Read | No |
| `get_apm_metrics` | Get application performance metrics | Low | Read | No |
| `get_application_metrics` | Get detailed metrics for specific app | Low | Read | No |
| `query_traces` | Query distributed traces | Low | Read | No |
| `get_transaction_traces` | Get detailed transaction traces | Low | Read | No |
| `get_trace_details` | Get specific trace details | Low | Read | No |

### Alerts (Read)

| Tool | Description | Risk | Impact | Approval |
|------|-------------|------|--------|----------|
| `list_alerts` | List all alert policies and conditions | Low | Read | No |
| `get_alert_conditions` | Get conditions for a policy | Low | Read | No |
| `get_alert_violations` | Get active alert violations | Low | Read | No |

### Logs (Read-Only)

| Tool | Description | Risk | Impact | Approval |
|------|-------------|------|--------|----------|
| `search_logs` | Search logs with Lucene syntax | Low | Read | No |
| `tail_logs` | Stream recent log entries | Low | Read | No |

### Infrastructure (Read-Only)

| Tool | Description | Risk | Impact | Approval |
|------|-------------|------|--------|----------|
| `get_infrastructure_metrics` | Get host/container metrics | Low | Read | No |

### Dashboards (Read-Only)

| Tool | Description | Risk | Impact | Approval |
|------|-------------|------|--------|----------|
| `list_dashboards` | List all dashboards | Low | Read | No |
| `get_dashboard_data` | Get dashboard widget data | Low | Read | No |

### Write Operations (Requires -write-enabled flag)

| Tool | Description | Risk | Impact | Approval |
|------|-------------|------|--------|----------|
| `acknowledge_alert_violation` | Acknowledge an alert | Med | Write | No |
| `create_alert_condition` | Create new alert condition | High | Write | Yes |
| `add_dashboard_widget` | Add widget to dashboard | High | Write | Yes |

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

This will automatically download dependencies and build a stripped binary.

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

### Download Binary

Pre-built binaries are available in the releases section.

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `NEWRELIC_API_KEY` | New Relic API key (User or Admin) | Yes | - |
| `NEWRELIC_REGION` | Data center region (`us` or `eu`) | No | `us` |

### Obtaining an API Key

1. Log in to your New Relic account
2. Go to **API Keys** in the account settings
3. Create a new **User Key** (for read-only) or **Admin Key** (for write operations)
4. Copy the key and set it as `NEWRELIC_API_KEY`

### Region Configuration

- **US Region**: Default, no configuration needed
- **EU Region**: Set `NEWRELIC_REGION=eu`

```bash
export NEWRELIC_API_KEY="NRAK-..."
export NEWRELIC_REGION="eu"  # For EU data center
```

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

### Self-Reporting Mode (No API Key)

The server requires an API key to start. However, if the key is invalid, tools will be registered but return errors on execution.

## Safety Features

### EnforcerProfile Metadata

Every tool reports its safety characteristics via `EnforcerProfile`:

```go
type EnforcerProfile struct {
    RiskLevel    RiskLevel   // low, med, high, critical
    ImpactScope  ImpactScope // read, write, delete, admin
    ResourceCost int         // 1-10 (API call weight)
    PIIExposure  bool        // Returns sensitive data?
    Idempotent   bool        // Safe to retry?
    ApprovalReq  bool        // Requires human approval?
}
```

This metadata enables automated policy enforcement by the MCP Bridge.

### Risk Classification

- **Low Risk**: Read-only queries, log searches, metric retrieval
- **Medium Risk**: Acknowledging alerts (non-destructive state change)
- **High Risk**: Creating alerts, modifying dashboards (infrastructure changes)

### Write Protection

Write tools are disabled by default and require:
1. Valid New Relic API key with appropriate permissions
2. `-write-enabled` command-line flag

## NRQL Queries

The `nrql_query` tool accepts any valid NRQL query:

```sql
-- Application performance
SELECT average(duration) FROM Transaction TIMESERIES SINCE 1 hour ago

-- Error analysis
SELECT count(*) FROM TransactionError FACET errorMessage SINCE 24 hours ago

-- Custom events
SELECT * FROM MyCustomEvent SINCE 30 minutes ago

-- Infrastructure metrics
SELECT average(cpuPercent) FROM SystemSample TIMESERIES FACET hostname
```

## Architecture

### API Client

- Uses New Relic GraphQL API (NerdGraph) and REST APIs
- Supports both US and EU data centers
- Automatic rate limiting and retry logic
- Connection pooling for efficient resource usage

### Data Handling

- NRQL query results are formatted as JSON
- Large result sets are automatically paginated
- Sensitive data is not logged

## Testing

Run the test suite:

```bash
go test ./newrelic -v
```

Tests cover:
- EnforcerProfile metadata accuracy
- API client initialization
- Tool execution with mock responses
- Error handling

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

### Search Logs

```json
{
  "tool": "search_logs",
  "arguments": {
    "query": "error AND service:payment-service",
    "time_range": "30 MINUTES"
  }
}
```

## License

This project is licensed under the Functional Source License, Version 1.1, ALv2 Future License.

Copyright 2026 Karl Dane

See LICENSE file for full terms.

## References

- [MCP Framework](https://github.com/karldane/mcp-framework) - Base framework with EnforcerProfile support
- [New Relic API Documentation](https://docs.newrelic.com/docs/apis/intro-apis/introduction-new-relic-apis/)
- [NRQL Reference](https://docs.newrelic.com/docs/query-your-data/nrql-new-relic-query-language/get-started/introduction-nrql-new-relics-query-language/)
