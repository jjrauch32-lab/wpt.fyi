#!/bin/bash
# Run all Dockerfile tests - both shell and Go

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$REPO_ROOT"

echo "=========================================="
echo "  Running Dockerfile Test Suite"
echo "=========================================="
echo ""

echo "1. Running Shell-based validation tests..."
echo "-------------------------------------------"
"$SCRIPT_DIR/validate_dockerfile.sh"
SHELL_EXIT=$?
echo ""

echo "2. Running Go-based unit tests..."
echo "-------------------------------------------"
go test -tags=small -v ./tests/dockerfile/...
GO_EXIT=$?
echo ""

echo "=========================================="
echo "  Test Results Summary"
echo "=========================================="
if [ $SHELL_EXIT -eq 0 ] && [ $GO_EXIT -eq 0 ]; then
    echo "✅ All tests passed!"
    exit 0
else
    echo "❌ Some tests failed:"
    [ $SHELL_EXIT -ne 0 ] && echo "   - Shell tests failed"
    [ $GO_EXIT -ne 0 ] && echo "   - Go tests failed"
    echo ""
    echo "See above for details."
    exit 1
fi