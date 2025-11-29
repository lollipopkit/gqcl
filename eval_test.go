package gqcl

import (
	"testing"
)

func mustEval(t *testing.T, expr string, ctx Value) Value {
	t.Helper()
	v, err := Evaluate(expr, ctx)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	return v
}

func TestShortCircuitAnd(t *testing.T) {
	v := mustEval(t, `false && (1 / 0 == 0)`, Nil)
	if v.Kind != KindBool || v.Bool {
		t.Fatalf("expected false, got %v", v)
	}
}

func TestShortCircuitOr(t *testing.T) {
	v := mustEval(t, `true || (1 / 0 == 0)`, Nil)
	if v.Kind != KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestNumericComparisonCrossType(t *testing.T) {
	if v := mustEval(t, `1.5 > 1`, Nil); v.Kind != KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
	if v := mustEval(t, `2 == 2.0`, Nil); v.Kind != KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestContainsList(t *testing.T) {
	if v := mustEval(t, `2 in [1,2,3]`, Nil); v.Kind != KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
	if v := mustEval(t, `[1,2] in [1,2,3]`, Nil); v.Kind != KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestContainsMapKey(t *testing.T) {
	ctx := FromInterface(map[string]any{"record": map[string]any{"owner": "alice"}})
	if v := mustEval(t, `"owner" in @record`, ctx); v.Kind != KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestNilEquality(t *testing.T) {
	if v := mustEval(t, `@missing == nil`, FromInterface(map[string]any{})); v.Kind != KindBool || !v.Bool {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestRequestedCtx(t *testing.T) {
	expr, err := Parse(`@req.user.role == "admin" || @record.id == 1`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	ctx := expr.RequestedCtx()
	if _, ok := ctx["req"]; !ok {
		t.Fatalf("missing req in requested ctx: %v", ctx)
	}
	if _, ok := ctx["record"]; !ok {
		t.Fatalf("missing record in requested ctx: %v", ctx)
	}
}

func TestListOutOfBoundsReturnsNil(t *testing.T) {
	ctx := FromInterface(map[string]any{"arr": []any{1, 2}})
	v := mustEval(t, `@arr.5`, ctx)
	if v.Kind != KindNil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestTrailingCommaLiterals(t *testing.T) {
	if v := mustEval(t, `[1,2,]`, Nil); v.Kind != KindList || len(v.List) != 2 {
		t.Fatalf("expected list len 2, got %v", v)
	}
	if v := mustEval(t, `{ "a": 1, "b": 2, }`, Nil); v.Kind != KindMap || len(v.Map) != 2 {
		t.Fatalf("expected map len 2, got %v", v)
	}
}
