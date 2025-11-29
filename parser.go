package gqcl

import (
	"fmt"
	"strconv"
)

type parser struct {
	tokens []token
	pos    int
}

// Parse builds an AST from expression string.
func Parse(src string) (*Expr, error) {
	toks, err := tokenize(src)
	if err != nil {
		return nil, err
	}
	p := parser{tokens: toks}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.peek().typ != tokEOF {
		return nil, fmt.Errorf("unexpected tokens at end")
	}
	return expr, nil
}

func (p *parser) parseExpr() (*Expr, error) {
	return p.parseOr()
}

func (p *parser) parseOr() (*Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokOr {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &Expr{Kind: nodeOr, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (*Expr, error) {
	left, err := p.parseCmp()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokAnd {
		p.next()
		right, err := p.parseCmp()
		if err != nil {
			return nil, err
		}
		left = &Expr{Kind: nodeAnd, Left: left, Right: right}
	}
	return left, nil
}

func (p *parser) parseCmp() (*Expr, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}
	for {
		tok := p.peek().typ
		switch tok {
		case tokEq, tokNe, tokGt, tokLt, tokGe, tokLe, tokIn:
			p.next()
			right, err := p.parseAddSub()
			if err != nil {
				return nil, err
			}
			left = &Expr{Kind: nodeBin, Op: tok, Left: left, Right: right}
		default:
			return left, nil
		}
	}
}

func (p *parser) parseAddSub() (*Expr, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}
	for {
		tok := p.peek().typ
		if tok != tokAdd && tok != tokSub {
			return left, nil
		}
		p.next()
		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}
		left = &Expr{Kind: nodeBin, Op: tok, Left: left, Right: right}
	}
}

func (p *parser) parseMulDiv() (*Expr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		tok := p.peek().typ
		if tok != tokMul && tok != tokDiv && tok != tokMod {
			return left, nil
		}
		p.next()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &Expr{Kind: nodeBin, Op: tok, Left: left, Right: right}
	}
}

func (p *parser) parseUnary() (*Expr, error) {
	tok := p.peek()
	if tok.typ == tokNot {
		p.next()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &Expr{Kind: nodeUnary, Op: tokNot, Left: expr}, nil
	}
	return p.parsePostfix()
}

func (p *parser) parsePostfix() (*Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for p.peek().typ == tokDot {
		p.next()
		field, err := p.parseFieldAccessor()
		if err != nil {
			return nil, err
		}
		if expr.Kind == nodeAt {
			expr.Children = append(expr.Children, field)
		} else {
			expr = &Expr{Kind: nodeAccess, Left: expr, Right: field}
		}
	}
	return expr, nil
}

func (p *parser) parseFieldAccessor() (*Expr, error) {
	tok := p.peek()
	switch tok.typ {
	case tokId:
		p.next()
		return &Expr{Kind: nodeVal, Val: Value{Kind: KindStr, Str: tok.lit}}, nil
	case tokInt:
		p.next()
		val, _ := parseInt(tok.lit)
		return &Expr{Kind: nodeVal, Val: Value{Kind: KindInt, Int: val}}, nil
	case tokStr:
		p.next()
		return &Expr{Kind: nodeVal, Val: Value{Kind: KindStr, Str: tok.lit}}, nil
	default:
		return nil, fmt.Errorf("invalid field accessor near %s", tok.lit)
	}
}

func (p *parser) parsePrimary() (*Expr, error) {
	tok := p.peek()
	switch tok.typ {
	case tokNil:
		p.next()
		return &Expr{Kind: nodeVal, Val: Nil}, nil
	case tokBool:
		p.next()
		return &Expr{Kind: nodeVal, Val: Value{Kind: KindBool, Bool: tok.lit == "true"}}, nil
	case tokInt:
		p.next()
		v, _ := parseInt(tok.lit)
		return &Expr{Kind: nodeVal, Val: Value{Kind: KindInt, Int: v}}, nil
	case tokFloat:
		p.next()
		f, _ := parseFloat(tok.lit)
		return &Expr{Kind: nodeVal, Val: Value{Kind: KindFloat, Float: f}}, nil
	case tokStr:
		p.next()
		return &Expr{Kind: nodeVal, Val: Value{Kind: KindStr, Str: tok.lit}}, nil
	case tokAt:
		return p.parseAt()
	case tokLBracket:
		return p.parseList()
	case tokLBrace:
		return p.parseMap()
	case tokLParen:
		p.next()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().typ != tokRParen {
			return nil, fmt.Errorf("expecting ')' at %d", p.peek().pos)
		}
		p.next()
		return &Expr{Kind: nodeParen, Left: expr}, nil
	case tokId:
		p.next()
		return &Expr{Kind: nodeVal, Val: Value{Kind: KindStr, Str: tok.lit}}, nil
	default:
		return nil, fmt.Errorf("unexpected token %v", tok.typ)
	}
}

func (p *parser) parseAt() (*Expr, error) {
	if p.peek().typ != tokAt {
		return nil, fmt.Errorf("expect '@'")
	}
	p.next()
	first, err := p.parseFieldAccessor()
	if err != nil {
		return nil, err
	}
	paths := []*Expr{first}
	for p.peek().typ == tokDot {
		p.next()
		f, err := p.parseFieldAccessor()
		if err != nil {
			return nil, err
		}
		paths = append(paths, f)
	}
	return &Expr{Kind: nodeAt, Children: paths}, nil
}

func (p *parser) parseList() (*Expr, error) {
	if err := p.expect(tokLBracket); err != nil {
		return nil, err
	}
	var items []*Expr
	if p.peek().typ == tokRBracket {
		p.next()
		return &Expr{Kind: nodeList, Children: items}, nil
	}
	for {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		items = append(items, expr)
		if p.peek().typ == tokComma {
			p.next()
			if p.peek().typ == tokRBracket {
				break
			}
			continue
		}
		break
	}
	if err := p.expect(tokRBracket); err != nil {
		return nil, err
	}
	return &Expr{Kind: nodeList, Children: items}, nil
}

func (p *parser) parseMap() (*Expr, error) {
	if err := p.expect(tokLBrace); err != nil {
		return nil, err
	}
	var pairs [][2]*Expr
	if p.peek().typ == tokRBrace {
		p.next()
		return &Expr{Kind: nodeMap, Pairs: pairs}, nil
	}
	for {
		key, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().typ != tokColon {
			return nil, fmt.Errorf("expect ':' in map")
		}
		p.next()
		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, [2]*Expr{key, val})
		if p.peek().typ == tokComma {
			p.next()
			if p.peek().typ == tokRBrace {
				break
			}
			continue
		}
		break
	}
	if err := p.expect(tokRBrace); err != nil {
		return nil, err
	}
	return &Expr{Kind: nodeMap, Pairs: pairs}, nil
}

func (p *parser) expect(t tokenType) error {
	if p.peek().typ != t {
		return fmt.Errorf("expect %v, got %v", t, p.peek().typ)
	}
	p.next()
	return nil
}

func (p *parser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokEOF}
	}
	return p.tokens[p.pos]
}

func (p *parser) next() token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
