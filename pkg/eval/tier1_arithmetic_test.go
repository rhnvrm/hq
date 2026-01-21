package eval

import "testing"

// Arithmetic tests
// Tier 1 - Essential (90% of use cases)

var arithmeticScenarios = ScenarioGroup{
	Name:        "arithmetic",
	Description: "Arithmetic operators (+, -, *, /, %)",
	Scenarios: []Scenario{
		// Addition
		{
			Description: "add integers",
			Document:    `null`,
			Expression:  `2 + 3`,
			Expected:    []string{`5`},
		},
		{
			Description: "add floats",
			Document:    `null`,
			Expression:  `2.5 + 3.5`,
			Expected:    []string{`6`},
		},
		{
			Description: "add mixed",
			Document:    `null`,
			Expression:  `2 + 3.5`,
			Expected:    []string{`5.5`},
		},
		{
			Description: "add from document",
			Document: huml(`
a: 10
b: 20
`),
			Expression: `.a + .b`,
			Expected:   []string{`30`},
		},
		{
			Description: "add strings (concatenation)",
			Document:    `null`,
			Expression:  `"hello" + " " + "world"`,
			Expected:    []string{`"hello world"`},
		},
		{
			Description: "add arrays (concatenation)",
			Document:    `null`,
			Expression:  `[1, 2] + [3, 4]`,
			Expected:    []string{`[1, 2, 3, 4]`},
		},
		{
			Description: "add objects (merge)",
			Document:    `null`,
			Expression:  `{"a": 1} + {"b": 2}`,
			Expected:    []string{`{"a": 1, "b": 2}`},
		},
		{
			Description: "add null is identity",
			Document:    `null`,
			Expression:  `5 + null`,
			Expected:    []string{`5`},
		},

		// Subtraction
		{
			Description: "subtract integers",
			Document:    `null`,
			Expression:  `10 - 3`,
			Expected:    []string{`7`},
		},
		{
			Description: "subtract floats",
			Document:    `null`,
			Expression:  `10.5 - 3.2`,
			Expected:    []string{`7.3`},
		},
		{
			Description: "subtract from document",
			Document: huml(`
total: 100
discount: 15
`),
			Expression: `.total - .discount`,
			Expected:   []string{`85`},
		},
		{
			Description: "subtract arrays (remove elements)",
			Document:    `null`,
			Expression:  `[1, 2, 3, 4] - [2, 4]`,
			Expected:    []string{`[1, 3]`},
		},

		// Multiplication
		{
			Description: "multiply integers",
			Document:    `null`,
			Expression:  `4 * 5`,
			Expected:    []string{`20`},
		},
		{
			Description: "multiply floats",
			Document:    `null`,
			Expression:  `2.5 * 4`,
			Expected:    []string{`10`},
		},
		{
			Description: "multiply from document",
			Document: huml(`
quantity: 5
price: 10.50
`),
			Expression: `.quantity * .price`,
			Expected:   []string{`52.5`},
		},
		{
			Description: "multiply string (repeat)",
			Document:    `null`,
			Expression:  `"ab" * 3`,
			Expected:    []string{`"ababab"`},
		},
		{
			Description: "multiply objects (deep merge)",
			Document:    `null`,
			Expression:  `{"a": {"x": 1}} * {"a": {"y": 2}}`,
			Expected:    []string{`{"a": {"x": 1, "y": 2}}`},
		},

		// Division
		{
			Description: "divide integers",
			Document:    `null`,
			Expression:  `20 / 4`,
			Expected:    []string{`5`},
		},
		{
			Description: "divide with remainder",
			Document:    `null`,
			Expression:  `7 / 2`,
			Expected:    []string{`3.5`},
		},
		{
			Description: "divide floats",
			Document:    `null`,
			Expression:  `10.0 / 4.0`,
			Expected:    []string{`2.5`},
		},
		{
			Description: "divide from document",
			Document: huml(`
total: 100
count: 4
`),
			Expression: `.total / .count`,
			Expected:   []string{`25`},
		},
		{
			Description:   "divide by zero is error",
			Document:      `null`,
			Expression:    `10 / 0`,
			ExpectedError: "division by zero",
		},

		// Modulo
		{
			Description: "modulo",
			Document:    `null`,
			Expression:  `17 % 5`,
			Expected:    []string{`2`},
		},
		{
			Description: "modulo even check",
			Document:    `10`,
			Expression:  `. % 2`,
			Expected:    []string{`0`},
		},
		{
			Description: "modulo odd check",
			Document:    `11`,
			Expression:  `. % 2`,
			Expected:    []string{`1`},
		},

		// Operator precedence
		{
			Description: "precedence: multiply before add",
			Document:    `null`,
			Expression:  `2 + 3 * 4`,
			Expected:    []string{`14`},
		},
		{
			Description: "precedence: divide before subtract",
			Document:    `null`,
			Expression:  `10 - 6 / 2`,
			Expected:    []string{`7`},
		},
		{
			Description: "parentheses override precedence",
			Document:    `null`,
			Expression:  `(2 + 3) * 4`,
			Expected:    []string{`20`},
		},

		// Negative numbers
		{
			Description: "negative number",
			Document:    `null`,
			Expression:  `-5`,
			Expected:    []string{`-5`},
		},
		{
			Description: "negate value",
			Document:    `10`,
			Expression:  `-.`,
			Expected:    []string{`-10`},
		},
		{
			Description: "subtract negative",
			Document:    `null`,
			Expression:  `10 - -5`,
			Expected:    []string{`15`},
		},
	},
}

var addFunctionScenarios = ScenarioGroup{
	Name:        "add",
	Description: "add function sums array elements",
	Scenarios: []Scenario{
		{
			Description: "add numbers",
			Document: huml(`
- 1
- 2
- 3
- 4
`),
			Expression: `add`,
			Expected:   []string{`10`},
		},
		{
			Description: "add with pipe",
			Document: huml(`
values:
  - 10
  - 20
  - 30
`),
			Expression: `.values | add`,
			Expected:   []string{`60`},
		},
		{
			Description: "add strings",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `add`,
			Expected:   []string{`"abc"`},
		},
		{
			Description: "add arrays (flatten one level)",
			Document: huml(`
- - 1
  - 2
- - 3
  - 4
`),
			Expression: `add`,
			Expected:   []string{`[1, 2, 3, 4]`},
		},
		{
			Description: "add empty array",
			Document:    `[]`,
			Expression:  `add`,
			Expected:    []string{`null`},
		},
		{
			Description: "add single element",
			Document: huml(`
- 42
`),
			Expression: `add`,
			Expected:   []string{`42`},
		},
	},
}

func TestArithmeticScenarios(t *testing.T) {
	runScenarios(t, arithmeticScenarios)
}

func TestAddFunctionScenarios(t *testing.T) {
	runScenarios(t, addFunctionScenarios)
}
