// Package eval provides the expression evaluation engine for hq.
package eval

// Scenario represents a single test case for expression evaluation.
// This struct is used both for testing and documentation generation.
type Scenario struct {
	// Description is a human-readable name for the test
	Description string

	// SubDescription provides additional context
	SubDescription string

	// Document is the input HUML document
	Document string

	// Document2 is an optional second document (for multi-doc operations)
	Document2 string

	// Expression is the hq expression to evaluate
	Expression string

	// Expected contains the expected output(s).
	// Each string is one output value.
	//
	// Values are compared SEMANTICALLY, not as strings. This means:
	//   - `{"a": 1}` equals `a: 1` (JSON vs HUML)
	//   - `[1, 2, 3]` equals `- 1\n- 2\n- 3` (inline vs multi-line)
	//   - `42` equals `42.0` (int vs float for whole numbers)
	//
	// You can use either JSON or HUML syntax in Expected values.
	// The test harness performs semantic comparison using JSON, YAML, and HUML parsers.
	Expected []string

	// ExpectedError is the expected error message (if any)
	ExpectedError string

	// SkipDoc excludes this scenario from documentation generation
	SkipDoc bool

	// EnvVars sets environment variables for the test
	EnvVars map[string]string

	// OutputFormat specifies expected output format (huml, json, yaml)
	OutputFormat string
}

// ScenarioGroup groups related scenarios for an operator or feature.
type ScenarioGroup struct {
	// Name is the operator or feature name (e.g., "select", "add")
	Name string

	// Description describes what this group tests
	Description string

	// Scenarios contains all test cases for this group
	Scenarios []Scenario
}
