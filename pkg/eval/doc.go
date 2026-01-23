// Package eval provides the expression evaluation engine for hq.
//
// # Semantic Test Comparison
//
// The test harness performs semantic comparison of expected vs actual values using
// JSON, YAML, and HUML parsers. This means you can write Expected values in either format:
//
//	// These are equivalent in tests:
//	Expected: []string{`{"name": "Alice", "age": 30}`}     // JSON
//	Expected: []string{huml(`
//	name: "Alice"
//	age: 30
//	`)}                                                     // HUML
//
// The comparison handles:
//   - Format differences (JSON vs HUML vs YAML)
//   - Numeric type differences (int vs float for whole numbers)
//   - Whitespace and formatting differences
//
// # Test Organization
//
// Tests are organized by tier (matching the spec) and by operator/feature:
//
// ## Tier 1 - Essential (90% of use cases)
//
//   - tier1_identity_test.go: Identity (.), field access, array access, slicing, iterators
//   - tier1_pipe_test.go: Pipe (|), comma (,), parentheses
//   - tier1_select_test.go: select, comparison operators, boolean operators
//   - tier1_arithmetic_test.go: Arithmetic (+, -, *, /, %), add function
//   - tier1_construction_test.go: Object and array construction
//   - tier1_assignment_test.go: Assignment (=, |=, +=), delete
//   - tier1_functions_test.go: length, keys, has, type, default (//), empty
//   - tier1_array_test.go: map, sort, unique, group_by, reverse, flatten, first/last, min/max
//   - tier1_string_test.go: Case conversion, trimming, split/join, contains/startswith/endswith, interpolation
//
// ## Tier 2 - Important (next 8% of use cases)
//
//   - tier2_regex_test.go: test, match, capture, sub, gsub
//   - tier2_object_test.go: to_entries, from_entries, with_entries, map_values
//   - tier2_conditionals_test.go: if-then-else, variables, recursive descent, reduce
//   - tier2_path_test.go: path, getpath, setpath, delpaths, contains/inside
//   - tier2_error_test.go: try-catch, optional access (?), error function
//
// ## CLI Tests (cmd package)
//
//   - cli_test.go: End-to-end CLI testing
//
// # Running Tests
//
// Run all tests:
//
//	go test ./...
//
// Run with verbose output:
//
//	go test ./... -v
//
// Run specific tier:
//
//	go test ./pkg/eval -run Tier1
//
// # Test Harness
//
// The test harness uses table-driven tests with the Scenario struct.
// Each scenario specifies:
//   - Description: Human-readable test name
//   - Document: Input HUML document
//   - Expression: hq expression to evaluate
//   - Expected: Expected output(s)
//   - ExpectedError: Expected error message (for error cases)
//
// Tests are currently skipped with TODO messages until the expression
// evaluator is implemented. As implementation progresses, tests will
// start passing automatically.
package eval
