#!/usr/bin/env bash
# Generate detailed test progress report for hq development
set -euo pipefail

cd "$(dirname "$0")/.."

echo "=============================================="
echo "       hq Development Progress Report"
echo "=============================================="
echo ""
echo "Generated: $(date)"
echo ""

# Run tests and capture output
output=$(go test ./... -v 2>&1)

# Overall stats
total=$(echo "$output" | grep -c "=== RUN" || true)
passed=$(echo "$output" | grep -c "--- PASS" || true)
skipped=$(echo "$output" | grep -c "--- SKIP" || true)
failed=$(echo "$output" | grep -c "--- FAIL" || true)

echo "## Overall Progress"
echo ""
echo "| Status   | Count | Percentage |"
echo "|----------|-------|------------|"
printf "| Passed   | %5d | %5.1f%%     |\n" "$passed" "$(echo "scale=1; $passed * 100 / $total" | bc)"
printf "| Skipped  | %5d | %5.1f%%     |\n" "$skipped" "$(echo "scale=1; $skipped * 100 / $total" | bc)"
printf "| Failed   | %5d | %5.1f%%     |\n" "$failed" "$(echo "scale=1; $failed * 100 / $total" | bc)"
echo "|----------|-------|------------|"
printf "| **Total**| %5d | 100.0%%     |\n" "$total"
echo ""

# Progress bar
if [ "$total" -gt 0 ]; then
    pct=$((passed * 100 / total))
    bar_width=50
    filled=$((pct * bar_width / 100))
    empty=$((bar_width - filled))
    printf "Progress: ["
    printf "%${filled}s" | tr ' ' '█'
    printf "%${empty}s" | tr ' ' '░'
    printf "] %d%%\n" "$pct"
fi
echo ""

# Per-file breakdown
echo "## Progress by Test File"
echo ""
echo "| File | Passed | Skipped | Total |"
echo "|------|--------|---------|-------|"

for file in pkg/eval/tier*_test.go cmd/cli_test.go; do
    if [ -f "$file" ]; then
        basename=$(basename "$file" .go)
        file_output=$(go test ./$(dirname "$file") -run "$(grep -o 'func Test[A-Za-z]*' "$file" | sed 's/func //' | tr '\n' '|' | sed 's/|$//')" -v 2>&1 || true)
        file_passed=$(echo "$file_output" | grep -c "--- PASS" || true)
        file_skipped=$(echo "$file_output" | grep -c "--- SKIP" || true)
        file_total=$((file_passed + file_skipped))
        printf "| %-30s | %6d | %7d | %5d |\n" "$basename" "$file_passed" "$file_skipped" "$file_total"
    fi
done

echo ""

# Show any failures
if [ "$failed" -gt 0 ]; then
    echo "## Failures"
    echo ""
    echo '```'
    echo "$output" | grep -A 5 "--- FAIL" || true
    echo '```'
    echo ""
fi

# Implementation checklist
echo "## Implementation Checklist"
echo ""
echo "### Tier 1 - Essential (target: 90% use cases)"
echo ""

tier1_features=(
    "Identity (.)"
    "Field access (.foo)"
    "Array access (.[n])"
    "Slicing (.[n:m])"
    "Iterator (.[])"
    "Pipe (|)"
    "Comma (,)"
    "Parentheses"
    "select"
    "Comparison (==, !=, <, >, <=, >=)"
    "Boolean (and, or, not)"
    "Arithmetic (+, -, *, /, %)"
    "add function"
    "Object construction ({...})"
    "Array construction ([...])"
    "Assignment (=, |=, +=)"
    "del"
    "length, keys, has, type"
    "Default (//)"
    "empty"
    "map, sort, unique, group_by"
    "reverse, flatten, first, last"
    "min, max"
    "String operations"
    "String interpolation"
)

for feature in "${tier1_features[@]}"; do
    echo "- [ ] $feature"
done

echo ""
echo "### Tier 2 - Important (target: next 8%)"
echo ""

tier2_features=(
    "Regex (test, match, capture, sub, gsub)"
    "to_entries, from_entries, with_entries"
    "map_values"
    "if-then-else"
    "Variables (as \$var)"
    "Recursive descent (..)"
    "reduce"
    "Path operations"
    "contains, inside"
    "try-catch"
    "Optional access (?)"
    "error function"
)

for feature in "${tier2_features[@]}"; do
    echo "- [ ] $feature"
done

echo ""
echo "---"
echo "Run \`just progress\` for quick stats or \`./scripts/progress-report.sh\` for this full report."
