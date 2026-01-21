package eval

import "testing"

// Regular expression tests
// Tier 2 - Important (next 8% of use cases)

var testRegexScenarios = ScenarioGroup{
	Name:        "test",
	Description: "test checks if string matches regex",
	Scenarios: []Scenario{
		{
			Description: "test simple match",
			Document:    `"hello world"`,
			Expression:  `test("world")`,
			Expected:    []string{`true`},
		},
		{
			Description: "test no match",
			Document:    `"hello world"`,
			Expression:  `test("foo")`,
			Expected:    []string{`false`},
		},
		{
			Description: "test with anchors",
			Document:    `"hello"`,
			Expression:  `test("^hello$")`,
			Expected:    []string{`true`},
		},
		{
			Description: "test email pattern",
			Document:    `"alice@example.com"`,
			Expression:  `test("@example\\.com$")`,
			Expected:    []string{`true`},
		},
		{
			Description: "test digit pattern",
			Document:    `"abc123def"`,
			Expression:  `test("\\d+")`,
			Expected:    []string{`true`},
		},
		{
			Description: "test case insensitive",
			Document:    `"Hello World"`,
			Expression:  `test("(?i)hello")`,
			Expected:    []string{`true`},
		},
		{
			Description: "test in select",
			Document: huml(`
- "admin_alice"
- "user_bob"
- "admin_carol"
`),
			Expression: `[.[] | select(test("^admin_"))]`,
			Expected:   []string{`["admin_alice", "admin_carol"]`},
		},
	},
}

var matchRegexScenarios = ScenarioGroup{
	Name:        "match",
	Description: "match returns match details",
	Scenarios: []Scenario{
		{
			Description: "match returns match object",
			Document:    `"test 123 hello"`,
			Expression:  `match("\\d+")`,
			Expected:    []string{`{"offset": 5, "length": 3, "string": "123", "captures": []}`},
		},
		{
			Description: "match with captures",
			Document:    `"2024-01-15"`,
			Expression:  `match("(\\d{4})-(\\d{2})-(\\d{2})")`,
			Expected: []string{`{
  "offset": 0,
  "length": 10,
  "string": "2024-01-15",
  "captures": [
    {"offset": 0, "length": 4, "string": "2024", "name": null},
    {"offset": 5, "length": 2, "string": "01", "name": null},
    {"offset": 8, "length": 2, "string": "15", "name": null}
  ]
}`},
		},
		{
			Description: "match extract string",
			Document:    `"price: $42.50"`,
			Expression:  `match("\\$[\\d.]+") | .string`,
			Expected:    []string{`"$42.50"`},
		},
		{
			Description: "match no match returns null",
			Document:    `"hello"`,
			Expression:  `match("\\d+")`,
			Expected:    []string{`null`},
		},
	},
}

var captureRegexScenarios = ScenarioGroup{
	Name:        "capture",
	Description: "capture returns named groups as object",
	Scenarios: []Scenario{
		{
			Description: "capture named groups",
			Document:    `"2024-01-15"`,
			Expression:  `capture("(?<year>\\d{4})-(?<month>\\d{2})-(?<day>\\d{2})")`,
			Expected:    []string{`{"year": "2024", "month": "01", "day": "15"}`},
		},
		{
			Description: "capture log line",
			Document:    `"ERROR: Connection failed"`,
			Expression:  `capture("(?<level>\\w+): (?<message>.*)")`,
			Expected:    []string{`{"level": "ERROR", "message": "Connection failed"}`},
		},
		{
			Description: "capture no match returns null",
			Document:    `"hello"`,
			Expression:  `capture("(?<num>\\d+)")`,
			Expected:    []string{`null`},
		},
		{
			Description: "capture in array",
			Document: huml(`
- "INFO: Starting"
- "ERROR: Failed"
- "WARN: Slow"
`),
			Expression: `[.[] | capture("(?<level>\\w+): (?<msg>.*)")]`,
			Expected:   []string{`[{"level": "INFO", "msg": "Starting"}, {"level": "ERROR", "msg": "Failed"}, {"level": "WARN", "msg": "Slow"}]`},
		},
	},
}

var substituteRegexScenarios = ScenarioGroup{
	Name:        "substitute",
	Description: "sub and gsub replace patterns",
	Scenarios: []Scenario{
		{
			Description: "sub replaces first match",
			Document:    `"hello hello hello"`,
			Expression:  `sub("hello"; "hi")`,
			Expected:    []string{`"hi hello hello"`},
		},
		{
			Description: "gsub replaces all matches",
			Document:    `"hello hello hello"`,
			Expression:  `gsub("hello"; "hi")`,
			Expected:    []string{`"hi hi hi"`},
		},
		{
			Description: "gsub with regex",
			Document:    `"a1b2c3"`,
			Expression:  `gsub("\\d"; "X")`,
			Expected:    []string{`"aXbXcX"`},
		},
		{
			Description: "gsub normalize whitespace",
			Document:    `"hello   world  foo"`,
			Expression:  `gsub("\\s+"; " ")`,
			Expected:    []string{`"hello world foo"`},
		},
		{
			Description: "sub no match unchanged",
			Document:    `"hello"`,
			Expression:  `sub("xyz"; "abc")`,
			Expected:    []string{`"hello"`},
		},
		{
			Description: "gsub remove pattern",
			Document:    `"a1b2c3"`,
			Expression:  `gsub("\\d"; "")`,
			Expected:    []string{`"abc"`},
		},
		{
			Description: "gsub with backreference",
			Document:    `"hello world"`,
			Expression:  `gsub("(\\w+)"; "[\\1]")`,
			Expected:    []string{`"[hello] [world]"`},
		},
	},
}

func TestTestRegexScenarios(t *testing.T) {
	runScenarios(t, testRegexScenarios)
}

func TestMatchRegexScenarios(t *testing.T) {
	runScenarios(t, matchRegexScenarios)
}

func TestCaptureRegexScenarios(t *testing.T) {
	runScenarios(t, captureRegexScenarios)
}

func TestSubstituteRegexScenarios(t *testing.T) {
	runScenarios(t, substituteRegexScenarios)
}
