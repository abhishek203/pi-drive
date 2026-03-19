#!/bin/bash
# Benchmark script for PiDrive codebase quality

set -e

cd "$(dirname "$0")"

# Clean any previous builds
rm -f /tmp/pidrive-bench

# Run staticcheck and count issues
STATICCHECK_OUTPUT=$(~/go/bin/staticcheck ./... 2>&1 || true)
if [ -n "$STATICCHECK_OUTPUT" ]; then
    STATICCHECK_ISSUES=$(echo "$STATICCHECK_OUTPUT" | wc -l | tr -d ' ')
else
    STATICCHECK_ISSUES=0
fi

# Run go vet and count issues
GOVET_OUTPUT=$(go vet ./... 2>&1 || true)
if [ -n "$GOVET_OUTPUT" ]; then
    GOVET_ISSUES=$(echo "$GOVET_OUTPUT" | wc -l | tr -d ' ')
else
    GOVET_ISSUES=0
fi

# Total issues
TOTAL_ISSUES=$((STATICCHECK_ISSUES + GOVET_ISSUES))

# Measure build time using python for better precision
START_TIME=$(python3 -c 'import time; print(int(time.time() * 1000))')
go build -o /tmp/pidrive-bench ./cmd/pidrive
END_TIME=$(python3 -c 'import time; print(int(time.time() * 1000))')
BUILD_MS=$((END_TIME - START_TIME))

# Get binary size in KB
BINARY_KB=$(du -k /tmp/pidrive-bench | cut -f1)

# Output metrics
echo "METRIC issue_count=$TOTAL_ISSUES"
echo "METRIC binary_kb=$BINARY_KB"
echo "METRIC build_ms=$BUILD_MS"

# Cleanup
rm -f /tmp/pidrive-bench
