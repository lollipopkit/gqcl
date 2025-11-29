package gqcl

import "testing"

func TestBasicCompare(t *testing.T) {
	ctx := FromInterface(map[string]any{"name": "alice"})
	val, err := Evaluate(`@name == "alice"`, ctx)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if val.Kind != KindBool || !val.Bool {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestListAccess(t *testing.T) {
	ctx := FromInterface(map[string]any{"arr": []any{"a", "b"}})
	val, err := Evaluate(`@arr.1`, ctx)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if val.Kind != KindStr || val.Str != "b" {
		t.Fatalf("expected 'b', got %v", val)
	}
}

func TestInOperator(t *testing.T) {
	ctx := FromInterface(map[string]any{"roles": []any{"admin", "user"}})
	val, err := Evaluate(`"admin" in @roles`, ctx)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if val.Kind != KindBool || !val.Bool {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestArithmetic(t *testing.T) {
	ctx := Nil
	val, err := Evaluate(`1 + 2 * 3`, ctx)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if val.Kind != KindInt || val.Int != 7 {
		t.Fatalf("expected 7, got %v", val)
	}
}

func TestMapLiteralAccess(t *testing.T) {
	ctx := Nil
	val, err := Evaluate(`{"a": 1}.a`, ctx)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if val.Kind != KindInt || val.Int != 1 {
		t.Fatalf("expected 1, got %v", val)
	}
}

func TestNestedListFieldAccess(t *testing.T) {
	ctx := FromInterface(map[string]any{"files": []any{map[string]any{"size": 10}}})
	val, err := Evaluate(`@files.0.size`, ctx)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	if val.Kind != KindInt || val.Int != 10 {
		t.Fatalf("expected 10, got %v", val)
	}
}
