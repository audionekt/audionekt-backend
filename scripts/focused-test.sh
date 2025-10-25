#!/bin/bash

# Focused testing script for Music Producer Social Network
# Only tests packages that make sense to unit test

echo "🎯 FOCUSED TESTING - Testing only testable packages"
echo "=================================================="

# Define testable packages
TESTABLE_PACKAGES="./pkg/utils/ ./internal/service/ ./internal/models/ ./internal/config/ ./internal/middleware/"

echo ""
echo "📊 Running tests on testable packages:"
echo "  - pkg/utils/ (pure functions)"
echo "  - internal/service/ (business logic)" 
echo "  - internal/models/ (data structures)"
echo "  - internal/config/ (configuration)"
echo "  - internal/middleware/ (HTTP middleware)"
echo ""

# Run tests with coverage
echo "🧪 Running tests..."
go test -coverprofile=coverage.out $TESTABLE_PACKAGES

echo ""
echo "📈 Coverage Report:"
echo "==================="
go tool cover -func=coverage.out | tail -1

echo ""
echo "📋 Detailed Coverage by Package:"
echo "================================"
go tool cover -func=coverage.out | grep -E "\.go:" | awk -F: '{print $1}' | sort | uniq -c | sort -nr

echo ""
echo "🎯 Files with High Coverage (>80%):"
echo "==================================="
go tool cover -func=coverage.out | awk -F: '$3 > 80.0 {print $0}' | head -10

echo ""
echo "⚠️  Files with Low Coverage (<50%):"
echo "==================================="
go tool cover -func=coverage.out | awk -F: '$3 < 50.0 {print $0}' | head -5

echo ""
echo "📊 Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html
echo "✅ Coverage report generated: coverage.html"

echo ""
echo "🚀 Focused testing complete!"
echo "   We're only testing packages that make sense to unit test."
echo "   Integration tests should be separate for handlers, repositories, etc."
