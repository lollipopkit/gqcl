package main

import (
	"testing"

	"github.com/lollipopkit/gqcl"
)

func evalWithInput(t *testing.T, expr string, input string, force string) gqcl.Value {
	t.Helper()
	ctx, err := loadContext([]byte(input), force)
	if err != nil {
		t.Fatalf("load context: %v", err)
	}
	v, err := gqcl.Evaluate(expr, ctx)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	return v
}

func TestLoadContextJSONAutoDetection(t *testing.T) {
	v := evalWithInput(t, `@name == "test"`, `{"name": "test", "age": 25}`, "")
	if v.Kind != gqcl.KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestLoadContextYAMLDetection(t *testing.T) {
	input := "name: test\nage: 25"
	v := evalWithInput(t, `@name == "test"`, input, "")
	if v.Kind != gqcl.KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestLoadContextTOMLDetection(t *testing.T) {
	input := "name = \"test\"\nage = 25"
	v := evalWithInput(t, `@name == "test"`, input, "")
	if v.Kind != gqcl.KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestLoadContextExplicitFormats(t *testing.T) {
	yamlInput := "name: test\nage: 25"
	v := evalWithInput(t, `@name == "test"`, yamlInput, "yaml")
	if v.Kind != gqcl.KindBool || !v.Bool {
		t.Fatalf("expected yaml parse success, got %v", v)
	}

	tomlInput := "name = \"test\"\nage = 25"
	v = evalWithInput(t, `@name == "test"`, tomlInput, "toml")
	if v.Kind != gqcl.KindBool || !v.Bool {
		t.Fatalf("expected toml parse success, got %v", v)
	}
}

func TestLoadContextInvalidForcedFormat(t *testing.T) {
	if _, err := loadContext([]byte(`{"broken": `), "json"); err == nil {
		t.Fatalf("expected forced json to error on invalid input")
	}
	if _, err := loadContext([]byte("invalid: [\n  - yaml\n"), "yaml"); err == nil {
		t.Fatalf("expected forced yaml to error on invalid input")
	}
	if _, err := loadContext([]byte("key = value = 1"), "toml"); err == nil {
		t.Fatalf("expected forced toml to error on invalid input")
	}
}

func TestLoadContextNestedStructures(t *testing.T) {
	yamlInput := "req:\n  user:\n    role: admin\n    name: test"
	v := evalWithInput(t, `@req.user.role == "admin"`, yamlInput, "")
	if v.Kind != gqcl.KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestLoadContextArrays(t *testing.T) {
	tomlInput := "permissions = [\"read\", \"write\", \"execute\"]"
	v := evalWithInput(t, `@permissions.0 == "read"`, tomlInput, "")
	if v.Kind != gqcl.KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}
