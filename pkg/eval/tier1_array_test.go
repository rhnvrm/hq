package eval

import "testing"

// Array operation tests
// Tier 1 - Essential (90% of use cases)

var mapScenarios = ScenarioGroup{
	Name:        "map",
	Description: "map applies expression to each element",
	Scenarios: []Scenario{
		{
			Description: "map extract field",
			Document: huml(`
- name: "Alice"
  age: 30
- name: "Bob"
  age: 25
`),
			Expression: `map(.name)`,
			Expected:   []string{`["Alice", "Bob"]`},
		},
		{
			Description: "map transform values",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `map(. * 2)`,
			Expected:   []string{`[2, 4, 6]`},
		},
		{
			Description: "map with arithmetic",
			Document: huml(`
- price: 100
- price: 200
- price: 300
`),
			Expression: `map(.price * 1.1)`,
			Expected:   []string{`[110, 220, 330]`},
		},
		{
			Description: "map construct objects",
			Document: huml(`
- id: 1
  firstName: "Alice"
  lastName: "Smith"
- id: 2
  firstName: "Bob"
  lastName: "Jones"
`),
			Expression: `map({id: .id, fullName: .firstName + " " + .lastName})`,
			Expected:   []string{`[{"id": 1, "fullName": "Alice Smith"}, {"id": 2, "fullName": "Bob Jones"}]`},
		},
		{
			Description: "map empty array",
			Document:    `[]`,
			Expression:  `map(. * 2)`,
			Expected:    []string{`[]`},
		},
		{
			Description: "map with select",
			Document: huml(`
- 1
- 2
- 3
- 4
- 5
`),
			Expression: `map(select(. > 2))`,
			Expected:   []string{`[3, 4, 5]`},
		},
	},
}

var sortScenarios = ScenarioGroup{
	Name:        "sort",
	Description: "sort orders array elements",
	Scenarios: []Scenario{
		{
			Description: "sort numbers",
			Document: huml(`
- 3
- 1
- 4
- 1
- 5
`),
			Expression: `sort`,
			Expected:   []string{`[1, 1, 3, 4, 5]`},
		},
		{
			Description: "sort strings",
			Document: huml(`
- "banana"
- "apple"
- "cherry"
`),
			Expression: `sort`,
			Expected:   []string{`["apple", "banana", "cherry"]`},
		},
		{
			Description: "sort_by field",
			Document: huml(`
- name: "Carol"
  age: 35
- name: "Alice"
  age: 30
- name: "Bob"
  age: 25
`),
			Expression: `sort_by(.name)`,
			Expected:   []string{`[{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}, {"name": "Carol", "age": 35}]`},
		},
		{
			Description: "sort_by numeric field",
			Document: huml(`
- name: "Carol"
  age: 35
- name: "Alice"
  age: 30
- name: "Bob"
  age: 25
`),
			Expression: `sort_by(.age)`,
			Expected:   []string{`[{"name": "Bob", "age": 25}, {"name": "Alice", "age": 30}, {"name": "Carol", "age": 35}]`},
		},
		{
			Description: "sort descending with reverse",
			Document: huml(`
- 1
- 3
- 2
`),
			Expression: `sort | reverse`,
			Expected:   []string{`[3, 2, 1]`},
		},
		{
			Description: "sort empty array",
			Document:    `[]`,
			Expression:  `sort`,
			Expected:    []string{`[]`},
		},
	},
}

var uniqueScenarios = ScenarioGroup{
	Name:        "unique",
	Description: "unique removes duplicate values",
	Scenarios: []Scenario{
		{
			Description: "unique numbers",
			Document: huml(`
- 1
- 2
- 1
- 3
- 2
- 1
`),
			Expression: `unique`,
			Expected:   []string{`[1, 2, 3]`},
		},
		{
			Description: "unique strings",
			Document: huml(`
- "a"
- "b"
- "a"
- "c"
- "b"
`),
			Expression: `unique`,
			Expected:   []string{`["a", "b", "c"]`},
		},
		{
			Description: "unique_by field",
			Document: huml(`
- id: 1
  name: "Alice"
- id: 2
  name: "Bob"
- id: 1
  name: "Alice Copy"
`),
			Expression: `unique_by(.id)`,
			Expected:   []string{`[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`},
		},
		{
			Description: "unique empty array",
			Document:    `[]`,
			Expression:  `unique`,
			Expected:    []string{`[]`},
		},
		{
			Description: "unique already unique",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `unique`,
			Expected:   []string{`[1, 2, 3]`},
		},
	},
}

var groupByScenarios = ScenarioGroup{
	Name:        "group_by",
	Description: "group_by groups elements by expression",
	Scenarios: []Scenario{
		{
			Description: "group_by field",
			Document: huml(`
- category: "fruit"
  name: "apple"
- category: "vegetable"
  name: "carrot"
- category: "fruit"
  name: "banana"
- category: "vegetable"
  name: "broccoli"
`),
			Expression: `group_by(.category)`,
			Expected: []string{`[
  [{"category": "fruit", "name": "apple"}, {"category": "fruit", "name": "banana"}],
  [{"category": "vegetable", "name": "carrot"}, {"category": "vegetable", "name": "broccoli"}]
]`},
		},
		{
			Description: "group_by with map",
			Document: huml(`
- category: "A"
  value: 10
- category: "B"
  value: 20
- category: "A"
  value: 30
`),
			Expression: `group_by(.category) | map({category: .[0].category, total: map(.value) | add})`,
			Expected:   []string{`[{"category": "A", "total": 40}, {"category": "B", "total": 20}]`},
		},
		{
			Description: "group_by empty array",
			Document:    `[]`,
			Expression:  `group_by(.x)`,
			Expected:    []string{`[]`},
		},
	},
}

var reverseScenarios = ScenarioGroup{
	Name:        "reverse",
	Description: "reverse reverses array order",
	Scenarios: []Scenario{
		{
			Description: "reverse array",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `reverse`,
			Expected:   []string{`[3, 2, 1]`},
		},
		{
			Description: "reverse strings",
			Document: huml(`
- "a"
- "b"
- "c"
`),
			Expression: `reverse`,
			Expected:   []string{`["c", "b", "a"]`},
		},
		{
			Description: "reverse empty array",
			Document:    `[]`,
			Expression:  `reverse`,
			Expected:    []string{`[]`},
		},
		{
			Description: "reverse single element",
			Document: huml(`
- 42
`),
			Expression: `reverse`,
			Expected:   []string{`[42]`},
		},
		{
			Description: "reverse string",
			Document:    `"hello"`,
			Expression:  `reverse`,
			Expected:    []string{`"olleh"`},
		},
	},
}

var flattenScenarios = ScenarioGroup{
	Name:        "flatten",
	Description: "flatten flattens nested arrays",
	Scenarios: []Scenario{
		{
			Description: "flatten one level",
			Document: huml(`
- - 1
  - 2
- - 3
  - 4
`),
			Expression: `flatten`,
			Expected:   []string{`[1, 2, 3, 4]`},
		},
		{
			Description: "flatten two levels",
			Document: huml(`
- - - 1
    - 2
  - - 3
    - 4
`),
			Expression: `flatten(2)`,
			Expected:   []string{`[1, 2, 3, 4]`},
		},
		{
			Description: "flatten one level keeps nested",
			Document: huml(`
- - - 1
    - 2
  - - 3
    - 4
`),
			Expression: `flatten(1)`,
			Expected:   []string{`[[1, 2], [3, 4]]`},
		},
		{
			Description: "flatten empty array",
			Document:    `[]`,
			Expression:  `flatten`,
			Expected:    []string{`[]`},
		},
		{
			Description: "flatten already flat",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `flatten`,
			Expected:   []string{`[1, 2, 3]`},
		},
	},
}

var firstLastScenarios = ScenarioGroup{
	Name:        "first-last",
	Description: "first and last get boundary elements",
	Scenarios: []Scenario{
		{
			Description: "first element",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `first`,
			Expected:   []string{`1`},
		},
		{
			Description: "last element",
			Document: huml(`
- 1
- 2
- 3
`),
			Expression: `last`,
			Expected:   []string{`3`},
		},
		{
			Description: "first of iterator",
			Document: huml(`
- name: "Alice"
- name: "Bob"
- name: "Carol"
`),
			Expression: `first(.[].name)`,
			Expected:   []string{`"Alice"`},
		},
		{
			Description: "last of iterator",
			Document: huml(`
- name: "Alice"
- name: "Bob"
- name: "Carol"
`),
			Expression: `last(.[].name)`,
			Expected:   []string{`"Carol"`},
		},
		{
			Description:   "first of empty",
			Document:      `[]`,
			Expression:    `first`,
			ExpectedError: "cannot get first element of empty array",
		},
		{
			Description:   "last of empty",
			Document:      `[]`,
			Expression:    `last`,
			ExpectedError: "cannot get last element of empty array",
		},
	},
}

var minMaxScenarios = ScenarioGroup{
	Name:        "min-max",
	Description: "min and max find extreme values",
	Scenarios: []Scenario{
		{
			Description: "min of numbers",
			Document: huml(`
- 3
- 1
- 4
- 1
- 5
`),
			Expression: `min`,
			Expected:   []string{`1`},
		},
		{
			Description: "max of numbers",
			Document: huml(`
- 3
- 1
- 4
- 1
- 5
`),
			Expression: `max`,
			Expected:   []string{`5`},
		},
		{
			Description: "min of strings",
			Document: huml(`
- "banana"
- "apple"
- "cherry"
`),
			Expression: `min`,
			Expected:   []string{`"apple"`},
		},
		{
			Description: "max of strings",
			Document: huml(`
- "banana"
- "apple"
- "cherry"
`),
			Expression: `max`,
			Expected:   []string{`"cherry"`},
		},
		{
			Description: "min_by field",
			Document: huml(`
- name: "Alice"
  age: 30
- name: "Bob"
  age: 25
- name: "Carol"
  age: 35
`),
			Expression: `min_by(.age)`,
			Expected:   []string{`{"name": "Bob", "age": 25}`},
		},
		{
			Description: "max_by field",
			Document: huml(`
- name: "Alice"
  age: 30
- name: "Bob"
  age: 25
- name: "Carol"
  age: 35
`),
			Expression: `max_by(.age)`,
			Expected:   []string{`{"name": "Carol", "age": 35}`},
		},
		{
			Description: "min of empty",
			Document:    `[]`,
			Expression:  `min`,
			Expected:    []string{`null`},
		},
		{
			Description: "max of empty",
			Document:    `[]`,
			Expression:  `max`,
			Expected:    []string{`null`},
		},
	},
}

func TestMapScenarios(t *testing.T) {
	runScenarios(t, mapScenarios)
}

func TestSortScenarios(t *testing.T) {
	runScenarios(t, sortScenarios)
}

func TestUniqueScenarios(t *testing.T) {
	runScenarios(t, uniqueScenarios)
}

func TestGroupByScenarios(t *testing.T) {
	runScenarios(t, groupByScenarios)
}

func TestReverseScenarios(t *testing.T) {
	runScenarios(t, reverseScenarios)
}

func TestFlattenScenarios(t *testing.T) {
	runScenarios(t, flattenScenarios)
}

func TestFirstLastScenarios(t *testing.T) {
	runScenarios(t, firstLastScenarios)
}

func TestMinMaxScenarios(t *testing.T) {
	runScenarios(t, minMaxScenarios)
}
