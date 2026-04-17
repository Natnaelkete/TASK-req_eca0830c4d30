# Delivery Architecture Audit 2 (Static Re-check)

## Test Execution Attempt

- Command attempted: `go test ./... -cover -v` (in `repo/`)
- Result: `/usr/bin/bash: line 1: go: command not found`
- Conclusion: **Cannot Confirm Statistically** for test pass/fail in this environment.

## Re-check Verdict

- **Overall conclusion:** **Partial Pass**
- You fixed multiple previously reported issues (notably object-level checks for list/get in tasks/results/devices/monitoring and deterministic partition DDL), but there are still material High issues.

## Confirmed Fixes (Previously Reported)

1) **Monitoring partition bootstrap now deterministic**
- **Status:** Fixed
- **Evidence:** `repo/migrations/002_indicator_versions_and_partitioning.sql:74`

2) **Task list/get object-level scope added**
- **Status:** Fixed
- **Evidence:** `repo/pkg/services/task_service.go:282`, `repo/pkg/services/task_service.go:333`, `repo/pkg/handlers/task_handler.go:57`

3) **Result list/get object-level scope added**
- **Status:** Fixed
- **Evidence:** `repo/pkg/services/result_service.go:292`, `repo/pkg/services/result_service.go:339`, `repo/pkg/handlers/result_handler.go:47`

4) **Device list/get ownership checks added**
- **Status:** Fixed
- **Evidence:** `repo/pkg/services/device_service.go:116`, `repo/pkg/services/device_service.go:167`, `repo/pkg/handlers/device_handler.go:48`

5) **Monitoring data list/get user scope added**
- **Status:** Fixed
- **Evidence:** `repo/pkg/services/monitoring_data_service.go:409`, `repo/pkg/services/monitoring_data_service.go:454`, `repo/pkg/handlers/monitoring_data_handler.go:50`

6) **Overdue status scope now includes submitted/under_review**
- **Status:** Fixed
- **Evidence:** `repo/pkg/services/task_service.go:219`

## Remaining Material Issues

1) **Severity: High**
- **Title:** Monitoring analytics/export endpoints are not user-scoped (possible cross-user data exposure)
- **Conclusion:** Fail
- **Evidence:**
  - Routes allow all authenticated users: `repo/cmd/server/main.go:186`, `repo/cmd/server/main.go:187`, `repo/cmd/server/main.go:188`, `repo/cmd/server/main.go:189`, `repo/cmd/server/main.go:190`
  - Handler methods do not pass `user_id`/`role` into these operations: `repo/pkg/handlers/monitoring_data_handler.go:105`, `repo/pkg/handlers/monitoring_data_handler.go:122`, `repo/pkg/handlers/monitoring_data_handler.go:139`, `repo/pkg/handlers/monitoring_data_handler.go:153`, `repo/pkg/handlers/monitoring_data_handler.go:176`
  - Service methods have no ownership filter inputs for these operations: `repo/pkg/services/monitoring_data_service.go:150`, `repo/pkg/services/monitoring_data_service.go:218`, `repo/pkg/services/monitoring_data_service.go:271`, `repo/pkg/services/monitoring_data_service.go:477`
- **Impact:** Users may access aggregated/real-time/exported data outside their authorized plot/device scope.
- **Minimum actionable fix:** Add scoped params (`user_id`, `role`) to aggregate/curve/trends/export paths and enforce plot ownership filters in service.

2) **Severity: High**
- **Title:** Task submit/review/complete flows lack object-level authorization checks
- **Conclusion:** Fail
- **Evidence:**
  - Handler forwards only task ID for submit/complete (no user/role guard at service level): `repo/pkg/handlers/task_handler.go:153`, `repo/pkg/handlers/task_handler.go:200`
  - Service submit/complete methods do not validate requester ownership/assignment: `repo/pkg/services/task_service.go:138`, `repo/pkg/services/task_service.go:191`
  - Review sets reviewer but does not enforce pre-assigned reviewer/admin ownership policy: `repo/pkg/services/task_service.go:165`, `repo/pkg/services/task_service.go:179`
- **Impact:** Any user with route-level role access can transition tasks they do not own.
- **Minimum actionable fix:** Pass `user_id`/`role` into submit/review/complete and enforce allowed actor rules per task ownership/reviewer assignment.

## Notes

- This is a static-only judgment; no runtime behavior was inferred from docs.
- Test execution is blocked by missing Go toolchain in current environment.
