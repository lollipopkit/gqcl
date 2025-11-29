package gqcl

import "testing"

func TestValueAccessMapAndList(t *testing.T) {
	root := FromInterface(map[string]any{
		"user": map[string]any{"age": 30},
		"0":    "zero",
	})

	user, ok := root.Access(Value{Kind: KindStr, Str: "user"})
	if !ok || user.Kind != KindMap {
		t.Fatalf("expected user map, got %v", user)
	}
	age, ok := user.Access(Value{Kind: KindStr, Str: "age"})
	if !ok || age.Kind != KindInt || age.Int != 30 {
		t.Fatalf("expected age 30, got %v", age)
	}

	zero, ok := root.Access(Value{Kind: KindInt, Int: 0})
	if !ok || zero.Kind != KindStr || zero.Str != "zero" {
		t.Fatalf("expected zero string, got %v", zero)
	}
}

func TestValueAccessOutOfBounds(t *testing.T) {
	list := FromInterface([]any{1, 2})

	if _, ok := list.Access(Value{Kind: KindInt, Int: -1}); ok {
		t.Fatalf("expected negative index to be out of bounds")
	}
	if _, ok := list.Access(Value{Kind: KindInt, Int: 5}); ok {
		t.Fatalf("expected positive OOB index to be out of bounds")
	}
}

func TestContainsListSubset(t *testing.T) {
	l := FromInterface([]any{1, 2})
	r := FromInterface([]any{1, 2, 3})

	ok, err := Contains(l, r)
	if err != nil {
		t.Fatalf("contains error: %v", err)
	}
	if !ok {
		t.Fatalf("expected left list to be contained in right list")
	}
}

func TestModMixedTypes(t *testing.T) {
	res, err := Mod(Value{Kind: KindInt, Int: 7}, Value{Kind: KindFloat, Float: 2.5})
	if err != nil {
		t.Fatalf("mod error: %v", err)
	}
	if res.Kind != KindFloat || res.Float != 2 {
		t.Fatalf("expected 2, got %v", res)
	}

	res, err = Mod(Value{Kind: KindFloat, Float: 7.5}, Value{Kind: KindInt, Int: 2})
	if err != nil {
		t.Fatalf("mod error: %v", err)
	}
	if res.Kind != KindFloat || res.Float != 1.5 {
		t.Fatalf("expected 1.5, got %v", res)
	}
}

func TestFromInterfaceConversions(t *testing.T) {
	mixed := map[any]any{
		"name": "alice",
		1:      2,
		true:   []any{"x", 3},
	}

	v := FromInterface(mixed)
	if v.Kind != KindMap {
		t.Fatalf("expected map, got %v", v)
	}

	if got, ok := v.Access(Value{Kind: KindStr, Str: "name"}); !ok || got.Str != "alice" {
		t.Fatalf("expected name alice, got %v", got)
	}
	if got, ok := v.Access(Value{Kind: KindStr, Str: "1"}); !ok || got.Int != 2 {
		t.Fatalf("expected numeric key as string with value 2, got %v", got)
	}

	nested, ok := v.Access(Value{Kind: KindStr, Str: "true"})
	if !ok || nested.Kind != KindList || len(nested.List) != 2 {
		t.Fatalf("expected nested list, got %v", nested)
	}
	if nested.List[0].Str != "x" || nested.List[1].Int != 3 {
		t.Fatalf("unexpected nested list contents: %v", nested.List)
	}
}
