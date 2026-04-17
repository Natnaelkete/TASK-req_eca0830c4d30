#!/usr/bin/env bash
# Seed the demo users documented in README.md into a running API instance.
#
# Usage:
#   ./scripts/seed_demo_users.sh [API_BASE_URL]
# Default API_BASE_URL is http://localhost:8080.
#
# Non-privileged roles (researcher, viewer) are created through the public
# registration endpoint. Elevated roles (admin, reviewer, customer_service)
# require an already-provisioned admin to call the admin-only user creation
# path; this script prints a hint for those and is idempotent when the user
# already exists (409).

set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"

register_public() {
  local username="$1" email="$2" password="$3" role="$4"
  printf '→ registering %s (role=%s)\n' "$username" "$role"
  curl -sS -o /dev/null -w '  HTTP %{http_code}\n' \
    -X POST "$BASE_URL/v1/auth/register" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"$username\",\"email\":\"$email\",\"password\":\"$password\",\"role\":\"$role\"}" \
    || true
}

register_public demo_researcher demo_researcher@example.com Research1234 researcher
register_public demo_viewer      demo_viewer@example.com     Viewer1234   viewer

cat <<HINT

Admin / reviewer / customer_service accounts cannot be self-registered.
Have an existing admin call the admin-only user creation endpoint to seed:

  demo_admin     / Admin1234    / admin
  demo_reviewer  / Review1234   / reviewer
  demo_cs        / Support1234  / customer_service

HINT
