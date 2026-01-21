# hq - HUML Query Processor
# justfile for development commands

# Default recipe - show help
default:
    @just --list

# ============================================================================
# Testing
# ============================================================================

# Run all tests
test:
    go test ./...

# Run all tests with verbose output
test-verbose:
    go test ./... -v

# Run tests and show only failures
test-failures:
    go test ./... -v 2>&1 | grep -E "(FAIL|Error|panic)" || echo "No failures!"

# Run Tier 1 (Essential) tests only
test-tier1:
    go test ./pkg/eval -run "Identity|FieldAccess|ArrayAccess|Slice|Iterator|Pipe|Comma|Parentheses|Select|Comparison|Boolean|Arithmetic|AddFunction|ObjectConstruction|ArrayConstruction|Assignment|Update|AddAssign|Delete|Length|Keys|Has|Type|Default|Empty|Map|Sort|Unique|GroupBy|Reverse|Flatten|FirstLast|MinMax|StringCase|StringTrim|SplitJoin|StringCheck|StringInterpolation" -v

# Run Tier 2 (Important) tests only
test-tier2:
    go test ./pkg/eval -run "TestRegex|ToEntries|FromEntries|WithEntries|MapValues|Conditional|Variable|RecursiveDescent|Reduce|Path|Getpath|Setpath|Delpaths|ContainsInside|TryCatch|OptionalAccess|ErrorFunction" -v

# Run CLI integration tests
test-cli:
    go test ./cmd -v

# Run a specific test by name pattern
test-pattern PATTERN:
    go test ./pkg/eval -run "{{PATTERN}}" -v

# ============================================================================
# Progress Tracking
# ============================================================================

# Show test progress (passing vs skipped)
progress:
    #!/usr/bin/env bash
    set -euo pipefail
    
    output=$(go test ./... -v 2>&1)
    
    # Count only subtests (those with / in the name) for accurate progress
    # Parent tests always pass if subtests skip
    total_subtests=$(echo "$output" | grep -c "=== RUN.*/" || true)
    passed_subtests=$(echo "$output" | grep "PASS:.*/" | grep -c "/" || true)
    skipped_subtests=$(echo "$output" | grep "SKIP:.*/" | grep -c "/" || true)
    failed_subtests=$(echo "$output" | grep "FAIL:.*/" | grep -c "/" || true)
    
    # Separate harness tests from expression tests
    harness_tests=$(echo "$output" | grep "PASS:.*/" | grep -c "TestCompare\|TestNormalize" || true)
    expr_passed=$((passed_subtests - harness_tests))
    
    echo "================================"
    echo "    hq Test Progress Report"
    echo "================================"
    echo ""
    echo "Expression Tests:"
    echo "  Total:     $skipped_subtests + $expr_passed = $((skipped_subtests + expr_passed))"
    echo "  Passing:   $expr_passed"
    echo "  Pending:   $skipped_subtests"
    echo "  Failed:    $failed_subtests"
    echo ""
    echo "Harness Tests: $harness_tests passing"
    echo ""
    
    expr_total=$((skipped_subtests + expr_passed))
    if [ "$expr_total" -gt 0 ]; then
        pct=$((expr_passed * 100 / expr_total))
        echo "Expression Implementation: $pct%"
        
        # Progress bar
        bar_width=40
        filled=$((pct * bar_width / 100))
        empty=$((bar_width - filled))
        printf "["
        printf "%${filled}s" | tr ' ' '#'
        printf "%${empty}s" | tr ' ' '-'
        printf "] $pct%%\n"
    fi
    echo ""

# Show progress for a specific tier
progress-tier1:
    #!/usr/bin/env bash
    echo "=== Tier 1 Progress ==="
    go test ./pkg/eval -run "Identity|FieldAccess|ArrayAccess|Slice|Iterator|Pipe|Comma|Parentheses|Select|Comparison|Boolean|Arithmetic" -v 2>&1 | grep -E "(PASS|SKIP|FAIL)" | sort | uniq -c

# List all test scenario groups
list-tests:
    @grep -h "^var.*Scenarios = ScenarioGroup" pkg/eval/*_test.go | sed 's/var \(.*\)Scenarios.*/\1/'

# ============================================================================
# Building
# ============================================================================

# Build hq binary
build:
    go build -o bin/hq ./cmd/hq

# Build with version info
build-release VERSION:
    go build -ldflags="-X main.version={{VERSION}}" -o bin/hq ./cmd/hq

# Install hq to GOPATH/bin
install:
    go install ./cmd/hq

# Clean build artifacts
clean:
    rm -rf bin/
    go clean

# ============================================================================
# Development
# ============================================================================

# Run go mod tidy
tidy:
    go mod tidy

# Format all Go code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run ./...

# Run linter with auto-fix
lint-fix:
    golangci-lint run --fix ./...

# Check for issues without fixing
check: fmt lint test

# ============================================================================
# Documentation
# ============================================================================

# Generate test documentation from scenarios
docs:
    @echo "TODO: Generate docs from test scenarios"

# Show the hq spec
spec:
    @cat ../../users/rhnvrm/plans/2026-01/21-13-59-hq-spec.md

# ============================================================================
# Utilities
# ============================================================================

# Watch tests and re-run on changes (requires entr)
watch:
    find . -name "*.go" | entr -c just test

# Watch specific test pattern
watch-pattern PATTERN:
    find . -name "*.go" | entr -c just test-pattern {{PATTERN}}

# Open test coverage in browser
coverage:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out

# Count lines of code
loc:
    @echo "Source code:"
    @find . -name "*.go" ! -name "*_test.go" | xargs wc -l | tail -1
    @echo "Test code:"
    @find . -name "*_test.go" | xargs wc -l | tail -1

# Show project structure
tree:
    @find . -type f -name "*.go" -o -name "*.huml" -o -name "justfile" | grep -v ".git" | sort
