#!/bin/bash
# Checks script: verify code compiles and is valid

set -e

cd "$(dirname "$0")"

echo "Checking code compilation..."
go build ./... || { echo "FAIL: Code does not compile"; exit 1; }

echo "Checking go mod tidy..."
go mod tidy || { echo "FAIL: go mod tidy failed"; exit 1; }

echo "All checks passed!"
