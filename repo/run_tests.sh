#!/usr/bin/env bash
# Docker-first test runner. The default path requires only Docker so contributors
# without a local Go toolchain can still run the full suite. If a local Go is
# available and USE_LOCAL_GO=1 is set, tests run in-process instead.
set -euo pipefail

echo "============================================"
echo "  Agricultural Platform — Test Suite Runner"
echo "============================================"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

run_in_docker() {
    echo "Running tests inside Docker (no local Go required)..."
    docker build --target builder -t agri-build-test -f Dockerfile .
    docker run --rm -v "$SCRIPT_DIR":/app -w /app agri-build-test \
        sh -c 'go test ./... -coverprofile=cov.out -cover -v'
}

run_local() {
    echo "Go version: $(go version)"
    echo "Running all tests with coverage..."
    echo ""
    go test ./... -coverprofile=cov.out -cover -v
}

TEST_EXIT=0
if [ "${USE_LOCAL_GO:-0}" = "1" ] && command -v go &>/dev/null; then
    run_local || TEST_EXIT=$?
elif command -v docker &>/dev/null; then
    run_in_docker || TEST_EXIT=$?
elif command -v go &>/dev/null; then
    echo "Docker not found; falling back to local Go toolchain."
    run_local || TEST_EXIT=$?
else
    echo "ERROR: neither Docker nor a local Go toolchain was found."
    echo "Install Docker (recommended) or Go 1.22+, then rerun this script."
    exit 1
fi

echo ""
if [ $TEST_EXIT -eq 0 ]; then
    echo "============================================"
    echo "  ALL TESTS PASSED"
    echo "============================================"
else
    echo "============================================"
    echo "  SOME TESTS FAILED (exit code: $TEST_EXIT)"
    echo "============================================"
fi

# Print coverage summary if the profile was generated and go is available locally.
if [ -f cov.out ] && command -v go &>/dev/null; then
    echo ""
    echo "Coverage summary:"
    go tool cover -func=cov.out | tail -1
fi

exit $TEST_EXIT
