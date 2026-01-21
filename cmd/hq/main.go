// hq is a lightweight command-line HUML processor with jq-compatible syntax.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	huml "github.com/huml-lang/go-huml"
	"github.com/rhnvrm/hq/pkg/eval"
	"gopkg.in/yaml.v3"
)

// Version information set by GoReleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "hq: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	// Parse flags
	var (
		rawOutput    bool
		nullInput    bool
		compactJSON  bool
		outputFormat = "huml" // huml, json, yaml
		expression   string
		inputFiles   []string
	)

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-r", "--raw-output":
			rawOutput = true
		case "-n", "--null-input":
			nullInput = true
		case "-c", "--compact-output":
			compactJSON = true
		case "-o", "--output":
			if i+1 >= len(args) {
				return fmt.Errorf("missing argument for %s", arg)
			}
			i++
			outputFormat = args[i]
		case "-h", "--help":
			printHelp(stdout)
			return nil
		case "-V", "--version":
			fmt.Fprintf(stdout, "hq %s (%s) built %s\n", version, commit, date)
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown flag: %s", arg)
			}
			if expression == "" {
				expression = arg
			} else {
				inputFiles = append(inputFiles, arg)
			}
		}
	}

	if expression == "" {
		return fmt.Errorf("no expression provided\nUsage: hq [flags] EXPRESSION [FILE...]")
	}

	// Get input
	var input any
	if nullInput {
		input = nil
	} else if len(inputFiles) > 0 {
		// Read from file(s)
		for _, file := range inputFiles {
			data, err := os.ReadFile(file)
			if err != nil {
				return fmt.Errorf("reading %s: %w", file, err)
			}
			var v any
			if err := parseInput(data, &v); err != nil {
				return fmt.Errorf("parsing %s: %w", file, err)
			}
			input = v // For now, just use the last file
		}
	} else {
		// Read from stdin
		data, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		if len(data) > 0 {
			if err := parseInput(data, &input); err != nil {
				return fmt.Errorf("parsing stdin: %w", err)
			}
		}
	}

	// Evaluate expression
	results, err := eval.Evaluate(expression, input)
	if err != nil {
		return fmt.Errorf("evaluation error: %w", err)
	}

	// Output results
	for i, result := range results {
		if i > 0 {
			fmt.Fprintln(stdout)
		}
		if err := outputValue(stdout, result, outputFormat, rawOutput, compactJSON); err != nil {
			return err
		}
	}

	return nil
}

// parseInput tries to parse input as HUML, JSON, or YAML
func parseInput(data []byte, v *any) error {
	text := strings.TrimSpace(string(data))

	// Try HUML first (native format for hq)
	if err := huml.Unmarshal([]byte(text), v); err == nil {
		return nil
	}

	// Try JSON (common for piping)
	if err := json.Unmarshal([]byte(text), v); err == nil {
		return nil
	}

	// Try YAML as fallback
	if err := yaml.Unmarshal([]byte(text), v); err == nil {
		return nil
	}

	return fmt.Errorf("could not parse as HUML, JSON, or YAML")
}

// outputValue formats and writes a single result
func outputValue(w io.Writer, v any, format string, raw, compact bool) error {
	// Handle raw string output
	if raw {
		if s, ok := v.(string); ok {
			fmt.Fprintln(w, s)
			return nil
		}
	}

	switch format {
	case "json":
		var data []byte
		var err error
		if compact {
			data, err = json.Marshal(v)
		} else {
			data, err = json.MarshalIndent(v, "", "  ")
		}
		if err != nil {
			return err
		}
		fmt.Fprintln(w, string(data))

	case "yaml":
		data, err := yaml.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Fprint(w, string(data))

	default: // huml or default
		// Use go-huml for proper HUML output
		data, err := huml.Marshal(v)
		if err != nil {
			// Fallback to simple output for types huml can't handle
			return outputSimple(w, v)
		}
		fmt.Fprint(w, string(data))
	}

	return nil
}

// outputSimple outputs a simple scalar value
func outputSimple(w io.Writer, v any) error {
	switch val := v.(type) {
	case nil:
		fmt.Fprintln(w, "null")
	case bool:
		fmt.Fprintln(w, val)
	case float64:
		if val == float64(int64(val)) {
			fmt.Fprintln(w, int64(val))
		} else {
			fmt.Fprintln(w, val)
		}
	case string:
		fmt.Fprintf(w, "%q\n", val)
	default:
		fmt.Fprintf(w, "%v\n", val)
	}
	return nil
}

func printHelp(w io.Writer) {
	help := `hq - a lightweight HUML processor with jq-compatible syntax

Usage:
  hq [flags] EXPRESSION [FILE...]

Flags:
  -r, --raw-output     Output raw strings without quotes
  -n, --null-input     Use null as input (don't read stdin)
  -c, --compact-output Compact JSON output (no pretty-printing)
  -o, --output FORMAT  Output format: huml (default), json, yaml
  -h, --help           Show this help message
  -V, --version        Show version

Examples:
  # Get a field from JSON/YAML
  echo '{"name": "Alice"}' | hq '.name'

  # Filter array elements
  echo '[1,2,3,4,5]' | hq '.[] | select(. > 2)'

  # Transform data
  echo '{"users": [{"name": "Alice"}, {"name": "Bob"}]}' | hq '.users[].name'

  # Arithmetic
  hq -n '1 + 2 * 3'

  # Output as JSON
  echo 'name: Alice' | hq -o json '.'
`
	fmt.Fprint(w, help)
}
