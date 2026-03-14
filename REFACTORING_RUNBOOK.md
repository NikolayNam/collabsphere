# Refactoring Runbook

This runbook captures the guardrails and baseline workflow for the backend refactor.

## Baseline Metrics

1. Start the stack with metrics enabled:
   - `APPLICATION_METRICS_ENABLED=true`
2. Collect an initial snapshot:
   - `make baseline-metrics`
3. Keep snapshot artifacts under `logs/` for before/after comparison.

Recommended KPIs:
- `collabsphere_http_request_duration_seconds` p50/p95/p99
- `collabsphere_http_requests_total` by route and status
- `go_memstats_alloc_bytes`
- DB pool pressure from `sql.DB` stats logs

## Runtime Hardening Defaults

- `APPLICATION_TRUST_PROXY_HEADERS=false` by default (safe for direct deployments).
- `COLLAB_WS_ALLOW_QUERY_ACCESS_TOKEN=false` by default to avoid token leakage in query strings.
- Rate limiter now uses instance-scoped state per router construction.

## DB Pool Tuning

Environment variables:
- `POSTGRES_MAX_OPEN_CONNS` (default: `30`)
- `POSTGRES_MAX_IDLE_CONNS` (default: `15`)
- `POSTGRES_CONN_MAX_LIFETIME` (default: `30m`)
- `POSTGRES_CONN_MAX_IDLE_TIME` (default: `5m`)

Apply conservative values first, then tune with production load telemetry.

## Safety Checklist For Subsequent Refactor Steps

- Keep public API contracts stable while changing internals.
- Add tests before touching cross-cutting middleware and access paths.
- For large service files, extract behavior into helper modules in small slices.
- Validate performance changes against baseline snapshots before rollout.
