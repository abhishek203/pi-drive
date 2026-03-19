#!/bin/bash
# Benchmark script for PiDrive codebase quality

set -e

cd "$(dirname "$0")"

# Clean any previous builds
rm -f /tmp/pidrive-bench

# Run staticcheck and count issues
STATICCHECK_ISSUES=0
if STATICCHECK_OUTPUT=$(~/go/bin/staticcheck ./... 2>&1); then
    STATICCHECK_ISSUES=0
else
    if [ -n "$STATICCHECK_OUTPUT" ]; then
        STATICCHECK_ISSUES=$(echo "$STATICCHECK_OUTPUT" | wc -l | tr -d ' ')
    fi
fi

# Run go vet and count issues
GOVET_ISSUES=0
if GOVET_OUTPUT=$(go vet ./... 2>&1); then
    GOVET_ISSUES=0
else
    if [ -n "$GOVET_OUTPUT" ]; then
        GOVET_ISSUES=$(echo "$GOVET_OUTPUT" | wc -l | tr -d ' ')
    fi
fi

# Check for gofmt issues
GOFMT_ISSUES=0
GOFMT_OUTPUT=$(gofmt -l $(find . -name "*.go" -not -path "./vendor/*") 2>&1 || true)
if [ -n "$GOFMT_OUTPUT" ]; then
    GOFMT_ISSUES=$(echo "$GOFMT_OUTPUT" | wc -l | tr -d ' ')
fi

# Count deadcode (unreachable functions) - informational
DEADCODE_COUNT=0
if DEADCODE_OUTPUT=$(/Users/abhisheke/go/bin/deadcode ./... 2>&1); then
    if [ -n "$DEADCODE_OUTPUT" ]; then
        DEADCODE_COUNT=$(echo "$DEADCODE_OUTPUT" | wc -l | tr -d ' ')
    fi
fi

# Total issues (primary metric - only staticcheck and go vet)
TOTAL_ISSUES=$((STATICCHECK_ISSUES + GOVET_ISSUES + GOFMT_ISSUES))

# Measure build time (with stripped binary for production-like size)
START_TIME=$(python3 -c 'import time; print(int(time.time() * 1000))')
go build -ldflags="-s -w" -o /tmp/pidrive-bench ./cmd/pidrive
END_TIME=$(python3 -c 'import time; print(int(time.time() * 1000))')
BUILD_MS=$((END_TIME - START_TIME))

# Get binary size in KB
BINARY_KB=$(du -k /tmp/pidrive-bench | cut -f1)

# Count lines of code
LOC=$(find . -name "*.go" -not -path "./vendor/*" -exec cat {} \; | wc -l | tr -d ' ')

# Output metrics
echo "METRIC issue_count=$TOTAL_ISSUES"
echo "METRIC binary_kb=$BINARY_KB"
echo "METRIC build_ms=$BUILD_MS"
echo "METRIC deadcode=$DEADCODE_COUNT"
echo "METRIC loc=$LOC"

# Cleanup
rm -f /tmp/pidrive-bench
