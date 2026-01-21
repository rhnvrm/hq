package eval

import "testing"

// Error handling tests
// Tier 2 - Important (next 8% of use cases)

var tryCatchScenarios = ScenarioGroup{
	Name:        "try-catch",
	Description: "try-catch for error handling",
	Scenarios: []Scenario{
		{
			Description: "try success passes through",
			Document: huml(`
a:
  b: 42
`),
			Expression: `try .a.b`,
			Expected:   []string{`42`},
		},
		{
			Description: "try catch missing path",
			Document: huml(`
a: 1
`),
			Expression: `try .x.y.z catch "not found"`,
			Expected:   []string{`"not found"`},
		},
		{
			Description: "try catch division by zero",
			Document: huml(`
a: 10
b: 0
`),
			Expression: `try (.a / .b) catch "division error"`,
			Expected:   []string{`"division error"`},
		},
		{
			Description: "try catch empty",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `.[] | try (if . == 2 then error("skip") else . end) catch empty`,
			Expected:   []string{`1`, `3`},
		},
		{
			Description: "try without catch returns empty on error",
			Document: huml(`
a: 1
`),
			Expression: `[try .x.y.z]`,
			Expected:   []string{`[]`},
		},
		{
			Description: "nested try",
			Document: huml(`
a: 1
`),
			Expression: `try (try .x.y catch .z.w) catch "fallback"`,
			Expected:   []string{`"fallback"`},
		},
	},
}

var optionalAccessScenarios = ScenarioGroup{
	Name:        "optional-access",
	Description: "Optional access operator (?) suppresses errors",
	Scenarios: []Scenario{
		{
			Description: "optional field exists",
			Document: huml(`
a:
  b: 42
`),
			Expression: `.a.b?`,
			Expected:   []string{`42`},
		},
		{
			Description: "optional field missing",
			Document: huml(`
a: 1
`),
			Expression: `.x.y.z?`,
			Expected:   []string{}, // no output, no error
		},
		{
			Description: "optional on non-object",
			Document:    `42`,
			Expression:  `.foo?`,
			Expected:    []string{}, // no output, no error
		},
		{
			Description: "optional iterator",
			Document: huml(`
items: null
`),
			Expression: `.items[]?`,
			Expected:   []string{}, // no error on null
		},
		{
			Description: "optional in select",
			Document: huml(`
- name: "Alice"
  email: "alice@example.com"
- name: "Bob"
- name: "Carol"
  email: "carol@example.com"
`),
			Expression: `[.[] | .email?]`,
			Expected:   []string{`["alice@example.com", "carol@example.com"]`},
		},
		{
			Description: "optional with default",
			Document: huml(`
a: 1
`),
			Expression: `.x.y.z? // "default"`,
			Expected:   []string{`"default"`},
		},
	},
}

var errorFunctionScenarios = ScenarioGroup{
	Name:        "error",
	Description: "error function raises errors",
	Scenarios: []Scenario{
		{
			Description:   "error with message",
			Document:      `null`,
			Expression:    `error("something went wrong")`,
			ExpectedError: "something went wrong",
		},
		{
			Description:   "error in conditional",
			Document:      `-5`,
			Expression:    `if . < 0 then error("negative value") else . end`,
			ExpectedError: "negative value",
		},
		{
			Description: "error caught by try",
			Document:    `null`,
			Expression:  `try error("fail") catch "caught"`,
			Expected:    []string{`"caught"`},
		},
		{
			Description: "conditional error",
			Document: huml(`
count: -1
`),
			Expression:    `if .count < 0 then error("count must be non-negative") else .count end`,
			ExpectedError: "count must be non-negative",
		},
	},
}

func TestTryCatchScenarios(t *testing.T) {
	runScenarios(t, tryCatchScenarios)
}

func TestOptionalAccessScenarios(t *testing.T) {
	runScenarios(t, optionalAccessScenarios)
}

func TestErrorFunctionScenarios(t *testing.T) {
	runScenarios(t, errorFunctionScenarios)
}
