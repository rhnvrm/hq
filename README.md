# hq - HUML Query Processor

A lightweight, portable command-line HUML processor with jq-compatible syntax.

```bash
# Extract a field
hq '.name' config.huml

# Filter an array
hq '.users[] | select(.active == true)' data.huml

# Transform data
hq '.items | map({id, name})' data.huml

# Output as JSON
hq -o json '.' config.huml
```

## Status

**407 tests passing (100%)** - Core expression engine fully implemented.

```bash
go test ./pkg/eval/...
# ok  github.com/huml-lang/hq/pkg/eval  0.029s
```

## Features

### Tier 1 - Essential (90% of use cases)

| Feature | Syntax | Status |
|---------|--------|--------|
| Identity | `.` | Done |
| Field access | `.foo`, `.["key"]` | Done |
| Array access | `.[n]`, `.[-1]` | Done |
| Slicing | `.[2:5]`, `.[:-1]` | Done |
| Iterator | `.[]` | Done |
| Pipe | `\|` | Done |
| Comma | `,` | Done |
| Parentheses | `()` | Done |
| Comparison | `==`, `!=`, `<`, `>`, `<=`, `>=` | Done |
| Boolean | `and`, `or`, `not` | Done |
| Arithmetic | `+`, `-`, `*`, `/`, `%` | Done |
| String ops | `split`, `join`, `ltrimstr`, `rtrimstr`, `contains`, `startswith`, `endswith`, `ascii_downcase`, `ascii_upcase` | Done |
| Array functions | `length`, `first`, `last`, `nth`, `reverse`, `flatten`, `sort`, `sort_by`, `unique`, `unique_by`, `group_by`, `min`, `max`, `min_by`, `max_by`, `add` | Done |
| Object functions | `keys`, `keys_unsorted`, `values`, `has`, `in`, `to_entries`, `from_entries`, `with_entries` | Done |
| Type functions | `type`, `isnull`, `isboolean`, `isnumber`, `isstring`, `isarray`, `isobject`, `arrays`, `objects`, `strings`, `numbers`, `booleans`, `nulls`, `scalars`, `iterables` | Done |
| Conversion | `tostring`, `tonumber` | Done |
| Construction | `{...}`, `[...]` | Done |
| Default | `//` | Done |
| Select | `select(expr)` | Done |
| Map | `map(expr)`, `map_values(expr)` | Done |
| Assignment | `=`, `\|=`, `+=`, `-=`, `*=` | Done |
| Delete | `del(path)` | Done |

### Tier 2 - Important (next 8% of use cases)

| Feature | Syntax | Status |
|---------|--------|--------|
| Conditionals | `if-then-else-end` | Done |
| Variables | `.x as $v \| ...` | Done |
| Destructuring | `.point as {x: $x, y: $y} \| ...` | Done |
| Dynamic access | `.[$var]` | Done |
| Reduce | `reduce .[] as $x (init; update)` | Done |
| Try-catch | `try expr catch handler` | Done |
| Optional | `expr?` | Done |
| Recursive descent | `..` | Done |
| Error | `error(msg)` | Done |
| Empty | `empty` | Done |
| Regex | `test`, `match`, `capture`, `sub`, `gsub` | Done |
| Path operations | `path`, `paths`, `getpath`, `setpath`, `delpaths` | Done |
| String interpolation | `"Hello, \(.name)!"` | Done |
| Range | `range(n)`, `range(from;to)` | Done |
| Limit | `limit(n; expr)`, `first(expr)` | Done |
| Any/All | `any`, `any(cond)`, `all`, `all(cond)` | Done |
| Index/Indices | `index(val)`, `rindex(val)`, `indices(val)` | Done |
| Inside | `inside(obj)` | Done |
| Env | `env`, `$ENV` | Done |

## Installation

```bash
# From source
go install github.com/rhnvrm/hq@latest
```

## Usage

```bash
# Basic usage
hq [options] <expression> [file...]

# Read from stdin
cat config.huml | hq '.database.host'

# Read from file
hq '.users[] | .name' users.huml

# Output formats
hq -o json '.' config.huml    # JSON output
hq -o yaml '.' config.huml    # YAML output
hq -r '.name' config.huml     # Raw string (no quotes)

# In-place editing
hq -i '.version = "2.0.0"' config.huml

# Null input (for generating data)
hq -n '{name: "test", values: [1,2,3]}'
```

## Examples

### Basic Navigation

```bash
# Get nested field
echo 'user: {name: "Alice", age: 30}' | hq '.user.name'
# "Alice"

# Get array element
echo '- a\n- b\n- c' | hq '.[1]'
# "b"

# Iterate array
echo '- 1\n- 2\n- 3' | hq '.[]'
# 1
# 2
# 3
```

### Filtering and Selection

```bash
# Filter with select
echo '[{name: "Alice", age: 30}, {name: "Bob", age: 25}]' | hq '.[] | select(.age > 26)'
# {name: "Alice", age: 30}

# Map transformation
echo '[1, 2, 3]' | hq 'map(. * 2)'
# [2, 4, 6]

# Group by field
echo '[{type: "a", val: 1}, {type: "b", val: 2}, {type: "a", val: 3}]' | hq 'group_by(.type)'
# [[{type: "a", val: 1}, {type: "a", val: 3}], [{type: "b", val: 2}]]
```

### Variables and Reduce

```bash
# Variable binding
echo '{x: 10, y: 20}' | hq '.x as $a | .y as $b | $a + $b'
# 30

# Destructuring
echo '{point: {x: 10, y: 20}}' | hq '.point as {x: $x, y: $y} | $x + $y'
# 30

# Reduce to sum
echo '[1, 2, 3, 4, 5]' | hq 'reduce .[] as $x (0; . + $x)'
# 15

# Reduce to build object
echo '[{key: "a", val: 1}, {key: "b", val: 2}]' | hq 'reduce .[] as $i ({}; .[$i.key] = $i.val)'
# {a: 1, b: 2}
```

### String Operations

```bash
# String interpolation
echo '{name: "World"}' | hq '"Hello, \(.name)!"'
# "Hello, World!"

# Regex matching
echo '"test@example.com"' | hq 'test("@")'
# true

# Substitution
echo '"hello world"' | hq 'gsub("o"; "0")'
# "hell0 w0rld"
```

### Error Handling

```bash
# Try-catch
echo '{}' | hq 'try .foo.bar catch "not found"'
# null

# Optional operator
echo '[{a: 1}, {b: 2}]' | hq '[.[] | .a?]'
# [1]

# Default value
echo 'null' | hq '. // "default"'
# "default"
```

## Development

### Prerequisites

- Go 1.21+

### Quick Start

```bash
# Clone
git clone https://github.com/rhnvrm/hq.git
cd hq

# Run tests
go test ./pkg/...

# Build
go build -o hq ./cmd/hq
```

### Project Structure

```
hq/
├── cmd/hq/              # CLI entry point
├── pkg/
│   ├── eval/            # Expression evaluation engine
│   │   ├── evaluator.go # Main evaluator
│   │   ├── functions.go # Built-in functions
│   │   └── *_test.go    # 407 test scenarios
│   ├── parser/          # Expression parser
│   │   ├── ast.go       # AST node definitions
│   │   ├── parser.go    # Recursive descent parser
│   │   └── tokens.go    # Lexer tokens
│   └── types/           # Data types
└── docs/                # Documentation
```

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Expression syntax | jq-compatible | User familiarity, easier adoption |
| Map iteration | Sorted keys | Go maps are unordered; alphabetical sort ensures deterministic output |
| Null propagation | jq-style | `.foo.bar` on `{foo: null}` returns `null`, not error |
| Alternative operator | `false` is falsy | `false // "default"` returns `"default"` (matches jq) |

## Related Projects

- [go-huml](https://github.com/huml-lang/go-huml) - Go library for HUML parsing/encoding
- [jq](https://github.com/jqlang/jq) - JSON processor (inspiration)
- [yq](https://github.com/mikefarah/yq) - YAML/JSON/XML processor

## License

MIT

## Contributing

Contributions welcome! The comprehensive test suite makes it easy to verify changes:

```bash
# Run all tests
go test ./pkg/eval/... -v

# Run specific test
go test ./pkg/eval/... -v -run TestReduceScenarios
```
