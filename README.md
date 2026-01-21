# hq - HUML Query Processor

A lightweight command-line HUML processor with jq-compatible syntax.

```bash
# Extract a field
hq '.name' config.huml

# Filter an array
hq '.users[] | select(.active)' data.huml

# Transform data
hq '.items | map({id, name})' data.huml
```

## Installation

```bash
go install github.com/rhnvrm/hq@latest
```

## Usage

```bash
hq [options] <expression> [file...]

# Read from stdin
cat config.huml | hq '.database.host'

# Output formats
hq -o json '.' config.huml    # JSON output
hq -o yaml '.' config.huml    # YAML output
hq -r '.name' config.huml     # Raw string (no quotes)

# Null input (for generating data)
hq -n '{name: "test", values: [1,2,3]}'
```

## Features

Full jq-compatible expression language including:

- **Navigation**: `.`, `.foo`, `.[]`, `.[n]`, `.[n:m]`, `..`
- **Operators**: `|`, `,`, `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, `and`, `or`, `not`
- **Conditionals**: `if-then-else`, `//`, `try-catch`, `?`
- **Variables**: `.x as $v | ...`, destructuring `{x: $x, y: $y}`
- **Functions**: `select`, `map`, `sort`, `unique`, `group_by`, `keys`, `values`, `length`, `type`, `has`, `in`, `contains`, `split`, `join`, `test`, `match`, `sub`, `gsub`, and more
- **Construction**: `{...}`, `[...]`, string interpolation `"Hello \(.name)"`
- **Assignment**: `=`, `|=`, `+=`, `-=`, `*=`, `del()`
- **Advanced**: `reduce`, `path`, `getpath`, `setpath`, `to_entries`, `from_entries`, `with_entries`

See [docs/WALKTHROUGH.md](docs/WALKTHROUGH.md) for comprehensive examples.

## Examples

```bash
# Sum numbers
echo '[1, 2, 3, 4, 5]' | hq 'add'
# 15

# Filter and transform
echo '[{"name":"Alice","age":30},{"name":"Bob","age":25}]' | hq '[.[] | select(.age > 26) | .name]'
# ["Alice"]

# Build object with reduce
echo '[{"k":"a","v":1},{"k":"b","v":2}]' | hq 'reduce .[] as $x ({}; .[$x.k] = $x.v)'
# {"a":1,"b":2}

# String interpolation
echo '{"name":"World"}' | hq '"Hello, \(.name)!"'
# "Hello, World!"

# Destructuring
echo '{"point":{"x":10,"y":20}}' | hq '.point as {x:$x,y:$y} | $x + $y'
# 30
```

## Related Projects

- [go-huml](https://github.com/huml-lang/go-huml) - Go library for HUML
- [jq](https://github.com/jqlang/jq) - JSON processor (inspiration)

## License

MIT
