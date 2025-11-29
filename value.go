package gqcl

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Kind identifies the concrete type stored in a Value.
type Kind int

const (
	KindNil Kind = iota
	KindStr
	KindInt
	KindFloat
	KindBool
	KindList
	KindMap
)

// Value is the runtime data representation used by the evaluator.
type Value struct {
	Kind  Kind
	Str   string
	Int   int64
	Float float64
	Bool  bool
	List  []Value
	Map   map[string]Value
}

// Nil is a convenient sentinel for nil values.
var Nil = Value{Kind: KindNil}

// Access traverses maps and lists using the provided field.
func (v Value) Access(field Value) (Value, bool) {
	switch v.Kind {
	case KindMap:
		switch field.Kind {
		case KindStr:
			val, ok := v.Map[field.Str]
			return val, ok
		case KindInt:
			key := strconv.FormatInt(field.Int, 10)
			val, ok := v.Map[key]
			return val, ok
		case KindFloat:
			key := strconv.FormatFloat(field.Float, 'f', -1, 64)
			val, ok := v.Map[key]
			return val, ok
		case KindBool:
			key := strconv.FormatBool(field.Bool)
			val, ok := v.Map[key]
			return val, ok
		default:
			return Value{}, false
		}
	case KindList:
		if field.Kind == KindInt && field.Int >= 0 && int(field.Int) < len(v.List) {
			return v.List[field.Int], true
		}
	}
	return Value{}, false
}

func (v Value) IsNumber() bool {
	return v.Kind == KindInt || v.Kind == KindFloat
}

// Equal performs deep equality between two Values.
func Equal(a, b Value) bool {
	if a.Kind != b.Kind {
		// Allow int/float comparison by value
		if a.IsNumber() && b.IsNumber() {
			return toFloat(a) == toFloat(b)
		}
		return false
	}

	switch a.Kind {
	case KindNil:
		return true
	case KindStr:
		return a.Str == b.Str
	case KindInt:
		return a.Int == b.Int
	case KindFloat:
		return a.Float == b.Float
	case KindBool:
		return a.Bool == b.Bool
	case KindList:
		if len(a.List) != len(b.List) {
			return false
		}
		for i := range a.List {
			if !Equal(a.List[i], b.List[i]) {
				return false
			}
		}
		return true
	case KindMap:
		if len(a.Map) != len(b.Map) {
			return false
		}
		for k, va := range a.Map {
			vb, ok := b.Map[k]
			if !ok || !Equal(va, vb) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func compare(a, b Value) (int, bool) {
	if a.Kind == KindStr && b.Kind == KindStr {
		switch {
		case a.Str < b.Str:
			return -1, true
		case a.Str > b.Str:
			return 1, true
		default:
			return 0, true
		}
	}

	if a.IsNumber() && b.IsNumber() {
		af, bf := toFloat(a), toFloat(b)
		switch {
		case af < bf:
			return -1, true
		case af > bf:
			return 1, true
		default:
			return 0, true
		}
	}
	return 0, false
}

func toFloat(v Value) float64 {
	if v.Kind == KindInt {
		return float64(v.Int)
	}
	return v.Float
}

// Add implements + semantics.
func Add(a, b Value) (Value, error) {
	switch {
	case a.Kind == KindStr && b.Kind == KindStr:
		return Value{Kind: KindStr, Str: a.Str + b.Str}, nil
	case a.IsNumber() && b.IsNumber():
		if a.Kind == KindFloat || b.Kind == KindFloat {
			return Value{Kind: KindFloat, Float: toFloat(a) + toFloat(b)}, nil
		}
		return Value{Kind: KindInt, Int: a.Int + b.Int}, nil
	default:
		return Value{}, fmt.Errorf("invalid op: %s + %s", a.String(), b.String())
	}
}

// Sub implements - semantics.
func Sub(a, b Value) (Value, error) {
	if a.IsNumber() && b.IsNumber() {
		if a.Kind == KindFloat || b.Kind == KindFloat {
			return Value{Kind: KindFloat, Float: toFloat(a) - toFloat(b)}, nil
		}
		return Value{Kind: KindInt, Int: a.Int - b.Int}, nil
	}
	return Value{}, fmt.Errorf("invalid op: %v - %v", a.Kind, b.Kind)
}

// Mul implements * semantics.
func Mul(a, b Value) (Value, error) {
	if a.IsNumber() && b.IsNumber() {
		if a.Kind == KindFloat || b.Kind == KindFloat {
			return Value{Kind: KindFloat, Float: toFloat(a) * toFloat(b)}, nil
		}
		return Value{Kind: KindInt, Int: a.Int * b.Int}, nil
	}
	return Value{}, fmt.Errorf("invalid op: %v * %v", a.Kind, b.Kind)
}

// Div implements / semantics with semantic arithmetic (int division can become float).
func Div(a, b Value) (Value, error) {
	if a.IsNumber() && b.IsNumber() {
		af, bf := toFloat(a), toFloat(b)
		if bf == 0 {
			return Value{}, fmt.Errorf("division by zero")
		}
		res := af / bf
		if a.Kind == KindInt && b.Kind == KindInt && res == float64(int64(res)) {
			return Value{Kind: KindInt, Int: int64(res)}, nil
		}
		return Value{Kind: KindFloat, Float: res}, nil
	}
	return Value{}, fmt.Errorf("invalid op: %v / %v", a.Kind, b.Kind)
}

// Mod implements % semantics.
func Mod(a, b Value) (Value, error) {
	if a.IsNumber() && b.IsNumber() {
		if (b.Kind == KindInt && b.Int == 0) || (b.Kind == KindFloat && b.Float == 0) {
			return Value{}, fmt.Errorf("mod by zero")
		}
		if a.Kind == KindInt && b.Kind == KindInt {
			return Value{Kind: KindInt, Int: a.Int % b.Int}, nil
		}
		return Value{Kind: KindFloat, Float: math.Mod(toFloat(a), toFloat(b))}, nil
	}
	return Value{}, fmt.Errorf("invalid op: %v %% %v", a.Kind, b.Kind)
}

// Contains implements the `in` operator semantics.
func Contains(l, r Value) (bool, error) {
	switch {
	case l.Kind == KindStr && r.Kind == KindStr:
		return strings.Contains(r.Str, l.Str), nil
	case r.Kind == KindList:
		if l.Kind == KindList {
			for _, item := range l.List {
				found := false
				for _, candidate := range r.List {
					if Equal(item, candidate) {
						found = true
						break
					}
				}
				if !found {
					return false, nil
				}
			}
			return true, nil
		}
		for _, item := range r.List {
			if Equal(l, item) {
				return true, nil
			}
		}
		return false, nil
	case r.Kind == KindMap:
		var key string
		switch l.Kind {
		case KindStr:
			key = l.Str
		case KindInt:
			key = strconv.FormatInt(l.Int, 10)
		case KindFloat:
			key = strconv.FormatFloat(l.Float, 'f', -1, 64)
		case KindBool:
			key = strconv.FormatBool(l.Bool)
		default:
			return false, nil
		}
		_, ok := r.Map[key]
		return ok, nil
	default:
		return false, fmt.Errorf("invalid op: %v in %v", l.Kind, r.Kind)
	}
}

// FromInterface converts arbitrary Go values (e.g., decoded JSON/YAML/TOML) into Value.
func FromInterface(v any) Value {
	switch t := v.(type) {
	case nil:
		return Nil
	case Value:
		return t
	case string:
		return Value{Kind: KindStr, Str: t}
	case []byte:
		return Value{Kind: KindStr, Str: string(t)}
	case bool:
		return Value{Kind: KindBool, Bool: t}
	case int:
		return Value{Kind: KindInt, Int: int64(t)}
	case int64:
		return Value{Kind: KindInt, Int: t}
	case int32:
		return Value{Kind: KindInt, Int: int64(t)}
	case float32:
		return Value{Kind: KindFloat, Float: float64(t)}
	case float64:
		return Value{Kind: KindFloat, Float: t}
	case []any:
		out := make([]Value, len(t))
		for i, v := range t {
			out[i] = FromInterface(v)
		}
		return Value{Kind: KindList, List: out}
	case map[string]any:
		out := make(map[string]Value, len(t))
		for k, v := range t {
			out[k] = FromInterface(v)
		}
		return Value{Kind: KindMap, Map: out}
	case map[any]any:
		out := make(map[string]Value, len(t))
		for k, v := range t {
			out[fmt.Sprint(k)] = FromInterface(v)
		}
		return Value{Kind: KindMap, Map: out}
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			out := make([]Value, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				out[i] = FromInterface(rv.Index(i).Interface())
			}
			return Value{Kind: KindList, List: out}
		case reflect.Map:
			out := make(map[string]Value, rv.Len())
			for _, key := range rv.MapKeys() {
				out[fmt.Sprint(key.Interface())] = FromInterface(rv.MapIndex(key).Interface())
			}
			return Value{Kind: KindMap, Map: out}
		}
		return Value{Kind: KindStr, Str: fmt.Sprint(v)}
	}
}

// String returns a human-friendly form similar to the Rust Display impl.
func (v Value) String() string {
	switch v.Kind {
	case KindNil:
		return "nil"
	case KindStr:
		return v.Str
	case KindInt:
		return strconv.FormatInt(v.Int, 10)
	case KindFloat:
		return strconv.FormatFloat(v.Float, 'f', -1, 64)
	case KindBool:
		return strconv.FormatBool(v.Bool)
	case KindList:
		parts := make([]string, len(v.List))
		for i, el := range v.List {
			parts[i] = el.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case KindMap:
		parts := make([]string, 0, len(v.Map))
		for k, vv := range v.Map {
			parts = append(parts, fmt.Sprintf("%s:%s", k, vv.String()))
		}
		return "{" + strings.Join(parts, ",") + "}"
	default:
		return "?"
	}
}
