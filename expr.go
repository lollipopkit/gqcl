package gqcl

import (
	"fmt"
)

type nodeKind int

const (
	nodeVal nodeKind = iota
	nodeBin
	nodeUnary
	nodeAnd
	nodeOr
	nodeAt
	nodeAccess
	nodeList
	nodeMap
	nodeParen
)

type Expr struct {
	Kind     nodeKind
	Val      Value
	Left     *Expr
	Right    *Expr
	Op       tokenType
	Children []*Expr
	Pairs    [][2]*Expr
}

func (e *Expr) Eval(ctx Value) (Value, error) {
	switch e.Kind {
	case nodeVal:
		return e.Val, nil
	case nodeUnary:
		v, err := e.Left.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		if e.Op == tokNot {
			if v.Kind != KindBool {
				return Value{}, fmt.Errorf("invalid operand: !%s", v.String())
			}
			return Value{Kind: KindBool, Bool: !v.Bool}, nil
		}
		return Value{}, fmt.Errorf("unknown unary op")
	case nodeBin:
		l, err := e.Left.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		r, err := e.Right.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		switch e.Op {
		case tokAdd:
			return Add(l, r)
		case tokSub:
			return Sub(l, r)
		case tokMul:
			return Mul(l, r)
		case tokDiv:
			return Div(l, r)
		case tokMod:
			return Mod(l, r)
		case tokEq:
			return Value{Kind: KindBool, Bool: Equal(l, r)}, nil
		case tokNe:
			return Value{Kind: KindBool, Bool: !Equal(l, r)}, nil
		case tokGt:
			cmp, ok := compare(l, r)
			if !ok {
				return Value{}, fmt.Errorf("invalid op: %v > %v", l.Kind, r.Kind)
			}
			return Value{Kind: KindBool, Bool: cmp > 0}, nil
		case tokLt:
			cmp, ok := compare(l, r)
			if !ok {
				return Value{}, fmt.Errorf("invalid op: %v < %v", l.Kind, r.Kind)
			}
			return Value{Kind: KindBool, Bool: cmp < 0}, nil
		case tokGe:
			cmp, ok := compare(l, r)
			if !ok {
				return Value{}, fmt.Errorf("invalid op: %v >= %v", l.Kind, r.Kind)
			}
			return Value{Kind: KindBool, Bool: cmp >= 0}, nil
		case tokLe:
			cmp, ok := compare(l, r)
			if !ok {
				return Value{}, fmt.Errorf("invalid op: %v <= %v", l.Kind, r.Kind)
			}
			return Value{Kind: KindBool, Bool: cmp <= 0}, nil
		case tokIn:
			res, err := Contains(l, r)
			if err != nil {
				return Value{}, err
			}
			return Value{Kind: KindBool, Bool: res}, nil
		default:
			return Value{}, fmt.Errorf("unknown op")
		}
	case nodeAnd:
		l, err := e.Left.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		if l.Kind != KindBool {
			return Value{}, fmt.Errorf("invalid op: && expects bool")
		}
		if !l.Bool {
			return Value{Kind: KindBool, Bool: false}, nil
		}
		r, err := e.Right.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		if r.Kind != KindBool {
			return Value{}, fmt.Errorf("invalid op: && expects bool")
		}
		return Value{Kind: KindBool, Bool: r.Bool}, nil
	case nodeOr:
		l, err := e.Left.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		if l.Kind != KindBool {
			return Value{}, fmt.Errorf("invalid op: || expects bool")
		}
		if l.Bool {
			return Value{Kind: KindBool, Bool: true}, nil
		}
		r, err := e.Right.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		if r.Kind != KindBool {
			return Value{}, fmt.Errorf("invalid op: || expects bool")
		}
		return Value{Kind: KindBool, Bool: r.Bool}, nil
	case nodeAt:
		cur := ctx
		for _, p := range e.Children {
			field, err := p.Eval(ctx)
			if err != nil {
				return Value{}, err
			}
			next, ok := cur.Access(field)
			if !ok {
				return Nil, nil
			}
			cur = next
		}
		return cur, nil
	case nodeAccess:
		base, err := e.Left.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		field, err := e.Right.Eval(ctx)
		if err != nil {
			return Value{}, err
		}
		next, ok := base.Access(field)
		if !ok {
			return Nil, nil
		}
		return next, nil
	case nodeList:
		out := make([]Value, len(e.Children))
		for i, child := range e.Children {
			v, err := child.Eval(ctx)
			if err != nil {
				return Value{}, err
			}
			out[i] = v
		}
		return Value{Kind: KindList, List: out}, nil
	case nodeMap:
		out := make(map[string]Value, len(e.Pairs))
		for _, kv := range e.Pairs {
			kVal, err := kv[0].Eval(ctx)
			if err != nil {
				return Value{}, err
			}
			vVal, err := kv[1].Eval(ctx)
			if err != nil {
				return Value{}, err
			}
			var keyStr string
			switch kVal.Kind {
			case KindStr:
				keyStr = kVal.Str
			case KindInt:
				keyStr = fmt.Sprintf("%d", kVal.Int)
			case KindFloat:
				keyStr = fmt.Sprintf("%g", kVal.Float)
			case KindBool:
				keyStr = fmt.Sprintf("%t", kVal.Bool)
			default:
				return Value{}, fmt.Errorf("map key must be primitive, got %v", kVal.Kind)
			}
			out[keyStr] = vVal
		}
		return Value{Kind: KindMap, Map: out}, nil
	case nodeParen:
		return e.Left.Eval(ctx)
	default:
		return Value{}, fmt.Errorf("unknown expr kind")
	}
}

// RequestedCtx returns a set of top-level context names used via @.
func (e *Expr) RequestedCtx() map[string]struct{} {
	out := map[string]struct{}{}
	e.collectCtx(out)
	return out
}

func (e *Expr) collectCtx(out map[string]struct{}) {
	if e == nil {
		return
	}
	switch e.Kind {
	case nodeAt:
		if len(e.Children) > 0 {
			first := e.Children[0]
			if first.Kind == nodeVal && first.Val.Kind == KindStr {
				out[first.Val.Str] = struct{}{}
			} else {
				first.collectCtx(out)
			}
			for _, child := range e.Children[1:] {
				if child.Kind == nodeVal {
					continue
				}
				child.collectCtx(out)
			}
		}
	case nodeAccess, nodeBin, nodeAnd, nodeOr, nodeUnary, nodeParen:
		if e.Left != nil {
			e.Left.collectCtx(out)
		}
		if e.Right != nil {
			e.Right.collectCtx(out)
		}
	case nodeList:
		for _, c := range e.Children {
			c.collectCtx(out)
		}
	case nodeMap:
		for _, kv := range e.Pairs {
			kv[0].collectCtx(out)
			kv[1].collectCtx(out)
		}
	}
}
