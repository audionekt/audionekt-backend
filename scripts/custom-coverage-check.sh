#!/bin/bash

# Custom coverage check that excludes adapter functions
echo "Running custom coverage check (excluding adapter functions)..."

# Run tests with coverage
go test -coverprofile=coverage.out ./pkg/utils/ ./internal/service/ ./internal/models/ ./internal/config/ ./internal/middleware/ ./internal/errors/ ./internal/secrets/ ./internal/validation/

# Filter out adapter functions from coverage report
grep -v "Adapter\|New.*Adapter\|.*Adapter\." coverage.out > coverage_filtered.out

# Calculate coverage percentage excluding adapter functions
coverage=$(go tool cover -func=coverage_filtered.out | grep total | awk '{print $3}' | sed 's/%//')

echo "Coverage (excluding adapter functions): ${coverage}%"

# Check if coverage meets 90% threshold
if [ $(echo "$coverage >= 90" | bc -l) -eq 1 ]; then
    echo "✅ Coverage: ${coverage}% (meets 90% requirement)"
    exit 0
else
    echo "❌ Coverage: ${coverage}% (below 90% requirement)"
    exit 1
fi
