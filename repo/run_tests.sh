#!/usr/bin/env bash
set -euo pipefail

echo "============================================"
echo "  Agricultural Platform — Test Suite Runner"
echo "============================================"
echo ""

# Determine project root (script directory)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Check for Go toolchain
if ! command -v go &>/dev/null; then
    echo "ERROR: Go toolchain not found in PATH."
    echo ""
    echo "You can run tests inside Docker instead:"
    echo "  docker build --target builder -t agri-build-test ."
    echo "  docker run --rm agri-build-test sh -c 'cd /app && go test ./... -v -cover'"
    exit 1
fi

echo "Go version: $(go version)"
echo "Running all tests with coverage..."
echo ""

go test ./... -coverprofile=cov.out -cover -v
TEST_EXIT=$?

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

# Print coverage summary if the profile was generated
if [ -f cov.out ]; then
    echo ""
    echo "Coverage summary:"
    go tool cover -func=cov.out | tail -1
fi

exit $TEST_EXIT
