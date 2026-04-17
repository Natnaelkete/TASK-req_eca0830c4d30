# audit_report-2-fix_check

## Scope
- Static re-check of items listed in `.tmp/audit_report-2.md`.
- No project runtime, Docker, or test execution was performed.

## Overall Result
- Overall conclusion: **Partial Pass**
- Summary:
  - All **High** issues from `audit_report-2` are resolved by static evidence.
  - **Medium** issues are improved; some are only partially resolved.

## Issue-by-Issue Fix Check

### 1) High: Public registration allowed self-assigned privileged roles
- Previous status: High / Fail
- Current status: **Resolved**
- Evidence:
  - Public registration now only allows `researcher`/`viewer`: `repo/pkg/services/auth_service.go:71`
  - Explicit guard rejects disallowed roles: `repo/pkg/services/auth_service.go:119`, `repo/pkg/services/auth_service.go:120`
  - Dedicated error handling path in handler: `repo/pkg/handlers/auth_handler.go:35`
  - Regression tests for allowed/denied roles: `repo/pkg/services/auth_service_test.go:96`, `repo/pkg/services/auth_service_test.go:104`
- Re-check conclusion: role-escalation path from public registration is closed statically.

### 2) High: Order message read-status endpoint lacked order-level authorization (IDOR risk)
- Previous status: High / Fail
- Current status: **Resolved**
- Evidence:
  - Handler now requires and validates `order_id` and passes role/user context: `repo/pkg/handlers/conversation_handler.go:146`, `repo/pkg/handlers/conversation_handler.go:160`
  - Service now enforces order access before update: `repo/pkg/services/conversation_service.go:220`, `repo/pkg/services/conversation_service.go:221`
  - Update query now scoped by `order_id`: `repo/pkg/services/conversation_service.go:227`
- Re-check conclusion: cross-order message read update by raw message ID is prevented statically.

### 3) High: Migration execution path not statically guaranteed
- Previous status: High / Partial Pass
- Current status: **Resolved (Static)**
- Evidence:
  - Startup now executes SQL migration bootstrap: `repo/cmd/server/main.go:37`
  - Config includes migration directory setting: `repo/internal/config/config.go:46`
  - Migration runner records applied SQL files and executes unapplied ones: `repo/pkg/models/migrate.go:32`
  - Runner unit coverage exists for parser/no-op behavior: `repo/pkg/models/migrate_test.go:12`, `repo/pkg/models/migrate_test.go:49`
- Re-check conclusion: deterministic migration execution is wired into startup path.
- Boundary note: runtime DB compatibility of every SQL statement remains manual verification.

### 4) Medium: Test suite skewed toward structure/constants, weak on critical behavioral/security paths
- Previous status: Medium / Partial Pass
- Current status: **Partially Resolved (Still Open)**
- Evidence of improvement:
  - Added migration/config security tests: `repo/pkg/models/migrate_test.go:12`, `repo/internal/config/config_test.go:82`
  - Added indicator service/handler test files: `repo/pkg/services/indicator_service_test.go:1`, `repo/pkg/handlers/indicator_handler_test.go:1`
- Evidence it remains open:
  - Large proportion of tests are still shape/default/error-string style (examples): `repo/pkg/services/analysis_service_test.go:16`, `repo/pkg/services/monitoring_data_service_test.go:20`, `repo/pkg/services/result_service_test.go:14`
  - Limited DB-backed authorization workflow verification for core risk flows.
- Re-check conclusion: quality improved but this medium issue is not fully closed.

### 5) Medium: No dedicated tests for indicator version service/handlers
- Previous status: Medium / Partial Pass
- Current status: **Partially Resolved**
- Evidence:
  - Dedicated files now exist: `repo/pkg/services/indicator_service_test.go:1`, `repo/pkg/handlers/indicator_handler_test.go:1`
  - Handler validation path for diff summary is tested: `repo/pkg/handlers/indicator_handler_test.go:53`
- Remaining gap:
  - Tests still do not validate persistence-level version increment/audit chain behavior in `IndicatorService`.
- Re-check conclusion: structural gap fixed; behavioral depth still pending.

### 6) Medium: Default JWT secret remained development-grade
- Previous status: Medium / Partial Pass
- Current status: **Resolved**
- Evidence:
  - Production-like env now rejects default secret/key: `repo/internal/config/config.go:58`, `repo/internal/config/config.go:62`
  - Tests verify rejection/acceptance behavior: `repo/internal/config/config_test.go:82`, `repo/internal/config/config_test.go:96`, `repo/internal/config/config_test.go:109`
- Re-check conclusion: insecure default is now environment-gated and blocked outside dev/test.

## Final Re-check Verdict
- **Partial Pass**
- Why:
  - All previous High findings are fixed by static evidence.
  - Medium findings are improved; however, test-depth quality and indicator version behavioral coverage remain only partially addressed.

## Notes
- This is a static-only determination.
- Runtime correctness/performance of migration SQL and background jobs remains **Manual Verification Required**.