#!/bin/bash

# Custom coverage script that excludes untestable files
# Usage: ./scripts/coverage.sh

echo "Running tests with coverage (excluding untestable files)..."

# Run tests and generate coverage
go test -coverprofile=coverage.out ./...

# Filter out untestable files from coverage report
echo ""
echo "Coverage Report (excluding untestable files):"
echo "============================================="

# Get total coverage excluding untestable files
go tool cover -func=coverage.out | grep -v -E "(cmd/api/main\.go|cmd/seeder/main\.go|docs/docs\.go)" | tail -1

echo ""
echo "Detailed coverage by package:"
echo "============================="

# Show coverage by package, excluding untestable files
go tool cover -func=coverage.out | grep -v -E "(cmd/api/main\.go|cmd/seeder/main\.go|docs/docs\.go)" | grep -E "\.go:" | awk -F: '{print $1}' | sort | uniq -c | sort -nr

echo ""
echo "Files with low coverage (< 50%):"
echo "==============================="
go tool cover -func=coverage.out | grep -v -E "(cmd/api/main\.go|cmd/seeder/main\.go|docs/docs\.go)" | awk -F: '$3 < 50.0 {print $0}' | head -10

echo ""
echo "Files with high coverage (> 80%):"
echo "================================="
go tool cover -func=coverage.out | grep -v -E "(cmd/api/main\.go|cmd/seeder/main\.go|docs/docs\.go)" | awk -F: '$3 > 80.0 {print $0}' | head -10

echo ""
echo "Coverage HTML report generated: coverage.html"
go tool cover -html=coverage.out -o coverage.html

