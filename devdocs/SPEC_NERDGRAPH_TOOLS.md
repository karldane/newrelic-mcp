# NerdGraph MCP Tools — Phase 2 Specification

## Overview

Extend the newrelic-mcp backend with new tools covering all major NerdGraph API capabilities not yet exposed. Organized in 3 tiers by priority.

## Conventions

- **Read tools**: follow existing `ListApplicationsTool` pattern (struct + 5 methods). Use `actorGraphQuery` for entity-based queries, `nerdGraphQuery` for `actor.account.*` queries.
- **Write tools**: gated behind `writeEnabled` flag. Use `nerdGraphMutation` helper if needed, or direct `client.Query()`.
- **Tests**: `httptest.NewServer` mock pattern matching the NerdGraph response shape.
- **EnforcerProfile**: use `RiskLow`/`RiskMed`, `ImpactRead`/`ImpactWrite`, `WithIdempotent(true)` for writes that are safe to retry.
- **No comments in production code** — tests may include comments.
- **Coverage target**: 80% minimum.

---

## Tier 1 — Core Monitoring

### A) Synthetic Monitoring

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 1 | `list_synthetic_monitors` | List all synthetic monitors | `account_id`, `limit`, `type` | `actor.account.synthetics.monitors` query | read |
| 2 | `get_synthetic_monitor` | Get a synthetic monitor details | `monitor_guid`, `account_id` | `actor.entity(guid).name, monitorType, etc.` via DashboardEntity-style query | read |
| 3 | `create_ping_monitor` | Create a simple ping monitor | `name`, `url`, `period`, `locations`, `account_id` | `syntheticsCreateMonitor` mutation (PING type) | write |
| 4 | `create_simple_browser_monitor` | Create a simple browser monitor | `name`, `url`, `period`, `locations`, `account_id` | `syntheticsCreateMonitor` mutation (BROWSER type) | write |
| 5 | `create_scripted_api_monitor` | Create a scripted API monitor | `name`, `script`, `period`, `locations`, `account_id` | `syntheticsCreateMonitor` mutation (SCRIPTED_API type) | write |
| 6 | `create_scripted_browser_monitor` | Create a scripted browser monitor | `name`, `script`, `period`, `locations`, `account_id` | `syntheticsCreateMonitor` mutation (SCRIPTED_BROWSER type) | write |
| 7 | `update_synthetic_monitor` | Update an existing monitor | `monitor_guid`, `name`, `url`, `period`, `locations`, `account_id` | `syntheticsUpdateMonitor` mutation | write |
| 8 | `delete_synthetic_monitor` | Delete a synthetic monitor | `monitor_guid`, `account_id` | `syntheticsDeleteMonitor` mutation | write |
| 9 | `list_private_locations` | List private locations | `account_id`, `limit` | `actor.account.synthetics.privateLocations` query | read |
| 10 | `create_private_location` | Create a private location | `name`, `description`, `account_id` | `syntheticsCreatePrivateLocation` mutation | write |

### B) Dashboard CRUD

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 11 | `create_dashboard` | Create a new dashboard | `name`, `description`, `permissions`, `pages`, `account_id` | `dashboardCreate` mutation | write |
| 12 | `update_dashboard` | Update an existing dashboard (full replacement) | `guid`, `name`, `description`, `permissions`, `pages`, `account_id` | `dashboardUpdate` mutation | write |
| 13 | `delete_dashboard` | Delete a dashboard (logical delete) | `guid` | `dashboardDelete` mutation | write |
| 14 | `get_dashboard` | Get full dashboard config by GUID | `guid` | `actor.entity(guid) { ... on DashboardEntity { ... } }` query | read |

### C) Alert Policies CRUD + NRQL Conditions

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 15 | `create_alert_policy` | Create a new alert policy | `name`, `incident_preference`, `account_id` | `alertsPolicyCreate` mutation | write |
| 16 | `update_alert_policy` | Update an alert policy | `policy_id`, `name`, `incident_preference`, `account_id` | `alertsPolicyUpdate` mutation | write |
| 17 | `delete_alert_policy` | Delete an alert policy | `policy_id`, `account_id` | `alertsPolicyDelete` mutation | write |
| 18 | `get_alert_policy` | Get a single alert policy by ID | `policy_id`, `account_id` | `actor.account.alerts.policy(id)` query | read |
| 19 | `list_nrql_alert_conditions` | List NRQL alert conditions for a policy | `policy_id`, `account_id` | `actor.account.alerts.nrqlConditionsSearch(filter: {policyId})` query | read |
| 20 | `update_alert_condition` | Update an NRQL alert condition | `condition_id`, `name`, `enabled`, `account_id` | `alertsNrqlConditionUpdate` mutation | write |
| 21 | `delete_alert_condition` | Delete an NRQL alert condition | `condition_id`, `policy_id`, `account_id` | `alertsNrqlConditionDelete` mutation | write |
| 22 | `mute_alert_condition` | Mute/unmute an alert condition | `condition_id`, `muted`, `account_id` | `alertsNrqlConditionMute` mutation | write |

---

## Tier 2 — Alerting & Notification Infrastructure

### D) Alert Workflows

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 23 | `list_workflows` | List alert notification workflows | `account_id`, `limit` | `actor.account.aiWorkflows.workflows` query | read |
| 24 | `get_workflow` | Get a workflow by ID | `workflow_id`, `account_id` | `actor.account.aiWorkflows.workflows(filters: {id})` query | read |
| 25 | `create_workflow` | Create a new alert workflow | `name`, `channel_ids`, `issue_filter`, `account_id` | `aiWorkflowsCreateWorkflow` mutation | write |
| 26 | `update_workflow` | Update a workflow | `workflow_id`, `name`, `account_id` | `aiWorkflowsUpdateWorkflow` mutation | write |
| 27 | `delete_workflow` | Delete a workflow | `workflow_id`, `account_id` | `aiWorkflowsDeleteWorkflow` mutation | write |
| 28 | `test_workflow` | Test a workflow configuration | `channel_id`, `filter_type`, `account_id` | `aiWorkflowsTestWorkflow` mutation | write |

### E) Notification Channels & Destinations

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 29 | `list_notification_channels` | List notification channels | `account_id`, `limit` | `actor.account.aiNotifications.channels` query | read |
| 30 | `create_slack_channel` | Create a Slack notification channel | `name`, `destination_id`, `channel_id`, `product`, `account_id` | `aiNotificationsCreateChannel` mutation | write |
| 31 | `create_email_channel` | Create an email notification channel | `name`, `destination_id`, `product`, `account_id` | `aiNotificationsCreateChannel` mutation (EMAIL type) | write |
| 32 | `create_webhook_channel` | Create a webhook notification channel | `name`, `destination_id`, `payload`, `product`, `account_id` | `aiNotificationsCreateChannel` mutation (WEBHOOK type) | write |
| 33 | `delete_notification_channel` | Delete a notification channel | `channel_id`, `account_id` | `aiNotificationsDeleteChannel` mutation | write |
| 34 | `list_destinations` | List notification destinations | `account_id`, `limit` | `actor.account.aiNotifications.destinations` query | read |

---

## Tier 3 — Specialized & Infrastructure

### F) Tag Management

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 35 | `get_entity_tags` | Get all tags for an entity | `entity_guid` | `actor.entity(guid).tags` query | read |
| 36 | `add_entity_tags` | Add tags to an entity | `entity_guid`, `tags` (key-value map) | `taggingAddTagsToEntity` mutation | write |
| 37 | `remove_entity_tags` | Remove tags from an entity | `entity_guid`, `tag_keys` | `taggingDeleteTagFromEntity` mutation | write |
| 38 | `replace_entity_tags` | Replace all tags on an entity | `entity_guid`, `tags` (key-value map) | `taggingReplaceTagsOnEntity` mutation | write |

### G) Service Level Management

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 39 | `list_service_levels` | List SLIs for an entity | `entity_guid` | `actor.entity(guid).serviceLevel.indicators` query | read |
| 40 | `create_service_level` | Create an SLI with SLO | `entity_guid`, `name`, `description`, `valid_events`, `good_events`, `target`, `time_window`, `account_id` | `serviceLevelCreate` mutation | write |
| 41 | `update_service_level` | Update an SLI's SLO | `sli_id`, `target`, `time_window` | `serviceLevelUpdate` mutation | write |

### H) Entity Operations

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 42 | `delete_entity` | Delete entities by GUID | `guids` (list) | `entityDelete` mutation | write |
| 43 | `search_entities` | Advanced entity search with freeform query | `query`, `limit`, `cursor` | `actor.entitySearch(query)` query | read |

### I) Cross-Account NRQL

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 44 | `cross_account_nrql` | Run NRQL query across multiple accounts | `query`, `account_ids` (list), `timeout` | `actor.nrql(accounts: [...], query: "...")` query | read |

### J) Workloads

| # | Tool | Description | Parameters | GraphQL Pattern | Type |
|---|---|---|---|---|---|
| 45 | `list_workloads` | List all workloads | `account_id`, `limit` | `actor.account.workloads` query | read |
| 46 | `get_workload` | Get a workload by GUID | `workload_guid` | `actor.entity(guid) { ... on WorkloadEntity { ... } }` query | read |

---

## Implementation Plan

### Phase 1 — Synthetics (first pass: read-only + one write type)
1. `list_synthetic_monitors` — query `actor.account.synthetics.monitors`
2. `get_synthetic_monitor` — query `actor.entity(guid)` with `SyntheticMonitorEntity` fragment
3. `create_ping_monitor` — `syntheticsCreateMonitor` mutation
4. `list_private_locations` — query `actor.account.synthetics.privateLocations`

### Phase 2 — Dashboard CRUD
1. `get_dashboard` — consolidate existing get_dashboard_data, expose full config
2. `create_dashboard` — `dashboardCreate` mutation
3. `update_dashboard` — `dashboardUpdate` mutation
4. `delete_dashboard` — `dashboardDelete` mutation

### Phase 3 — Alert Policies CRUD + NRQL conditions
1. `get_alert_policy` — `actor.account.alerts.policy(id)`
2. `create_alert_policy` — `alertsPolicyCreate` mutation
3. `update_alert_policy` — `alertsPolicyUpdate` mutation
4. `delete_alert_policy` — `alertsPolicyDelete` mutation
5. Extend existing `create_alert_condition` from stub to real
6. `list_nrql_alert_conditions` — `actor.account.alerts.nrqlConditionsSearch`

### Phase 4 — Workflows + Notification Channels
1. `list_workflows` + `get_workflow`
2. `list_notification_channels` + `list_destinations`
3. `create_slack_channel` + `create_email_channel`
4. `create_workflow`

### Phase 5 — Tier 3 (all remaining)

## Testing Strategy

- Each tool gets a mock test for: happy path, empty results, missing required params, error response
- Write tools get paired tests: write-disabled shows error, write-enabled succeeds
- Use table-driven tests where multiple variants exist
- Verify NRQL escaping in test captures (like `TestGetApplicationMetricsEscapesAppName`)
