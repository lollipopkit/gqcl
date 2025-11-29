package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	toml "github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"gqcl"
)

func main() {
	jsonFlag := flag.Bool("json", false, "force json input")
	yamlFlag := flag.Bool("yaml", false, "force yaml input")
	tomlFlag := flag.Bool("toml", false, "force toml input")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: qcl [--json|--yaml|--toml] <expr>")
		os.Exit(1)
	}

	exprStr := flag.Arg(0)
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read stdin: %v\n", err)
		os.Exit(1)
	}

	format := ""
	switch {
	case *jsonFlag:
		format = "json"
	case *yamlFlag:
		format = "yaml"
	case *tomlFlag:
		format = "toml"
	}

	ctx, err := loadContext(data, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse context: %v\n", err)
		os.Exit(1)
	}

	expr, err := gqcl.Parse(exprStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse expr: %v\n", err)
		os.Exit(1)
	}

	res, err := expr.Eval(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "eval: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(res.String())
}

func loadContext(data []byte, force string) (gqcl.Value, error) {
	var v any

	decode := func(fn func([]byte, any) error) (gqcl.Value, error) {
		if err := fn(data, &v); err != nil {
			return gqcl.Nil, err
		}
		return gqcl.FromInterface(v), nil
	}

	switch force {
	case "json":
		val, err := decode(json.Unmarshal)
		if err != nil {
			return gqcl.Nil, fmt.Errorf("invalid json")
		}
		return val, nil
	case "yaml":
		val, err := decode(yaml.Unmarshal)
		if err != nil {
			return gqcl.Nil, fmt.Errorf("invalid yaml")
		}
		return val, nil
	case "toml":
		val, err := decode(toml.Unmarshal)
		if err != nil {
			return gqcl.Nil, fmt.Errorf("invalid toml")
		}
		return val, nil
	}

	if val, err := decode(json.Unmarshal); err == nil {
		return val, nil
	}
	if val, err := decode(toml.Unmarshal); err == nil {
		return val, nil
	}
	if val, err := decode(yaml.Unmarshal); err == nil {
		return val, nil
	}

	return gqcl.Nil, fmt.Errorf("unable to detect input format")
}
