package gqcl

import (
	"fmt"
	"strconv"
	"unicode"
)

type tokenType int

const (
	tokEOF tokenType = iota
	tokLParen
	tokRParen
	tokLBrace
	tokRBrace
	tokLBracket
	tokRBracket
	tokDot
	tokColon
	tokComma
	tokAdd
	tokSub
	tokMul
	tokDiv
	tokMod
	tokEq
	tokNe
	tokGt
	tokLt
	tokGe
	tokLe
	tokIn
	tokAnd
	tokOr
	tokNot
	tokAt
	tokStr
	tokInt
	tokFloat
	tokBool
	tokNil
	tokId
)

type token struct {
	typ tokenType
	lit string
	pos int
}

func tokenize(input string) ([]token, error) {
	var tokens []token
	chars := []rune(input)
	for i := 0; i < len(chars); {
		c := chars[i]
		if unicode.IsSpace(c) {
			i++
			continue
		}

		// Comments
		if c == '/' && i+1 < len(chars) && chars[i+1] == '/' {
			for i < len(chars) && chars[i] != '\n' {
				i++
			}
			continue
		}

		switch c {
		case '(':
			tokens = append(tokens, token{typ: tokLParen, pos: i})
			i++
		case ')':
			tokens = append(tokens, token{typ: tokRParen, pos: i})
			i++
		case '{':
			tokens = append(tokens, token{typ: tokLBrace, pos: i})
			i++
		case '}':
			tokens = append(tokens, token{typ: tokRBrace, pos: i})
			i++
		case '[':
			tokens = append(tokens, token{typ: tokLBracket, pos: i})
			i++
		case ']':
			tokens = append(tokens, token{typ: tokRBracket, pos: i})
			i++
		case '.':
			tokens = append(tokens, token{typ: tokDot, pos: i})
			i++
		case ':':
			tokens = append(tokens, token{typ: tokColon, pos: i})
			i++
		case ',':
			tokens = append(tokens, token{typ: tokComma, pos: i})
			i++
		case '+':
			if i+1 < len(chars) && unicode.IsDigit(chars[i+1]) {
				t, next, err := parseNumber(chars, i)
				if err != nil {
					return nil, err
				}
				tokens = append(tokens, t)
				i = next
			} else {
				tokens = append(tokens, token{typ: tokAdd, pos: i})
				i++
			}
		case '-':
			if i+1 < len(chars) && unicode.IsDigit(chars[i+1]) {
				t, next, err := parseNumber(chars, i)
				if err != nil {
					return nil, err
				}
				tokens = append(tokens, t)
				i = next
			} else {
				tokens = append(tokens, token{typ: tokSub, pos: i})
				i++
			}
		case '*':
			tokens = append(tokens, token{typ: tokMul, pos: i})
			i++
		case '/':
			tokens = append(tokens, token{typ: tokDiv, pos: i})
			i++
		case '%':
			tokens = append(tokens, token{typ: tokMod, pos: i})
			i++
		case '@':
			tokens = append(tokens, token{typ: tokAt, pos: i})
			i++
		case '!':
			if i+1 < len(chars) && chars[i+1] == '=' {
				tokens = append(tokens, token{typ: tokNe, pos: i})
				i += 2
			} else {
				tokens = append(tokens, token{typ: tokNot, pos: i})
				i++
			}
		case '=':
			if i+1 < len(chars) && chars[i+1] == '=' {
				tokens = append(tokens, token{typ: tokEq, pos: i})
				i += 2
			} else {
				return nil, fmt.Errorf("unexpected '=' at %d", i)
			}
		case '&':
			if i+1 < len(chars) && chars[i+1] == '&' {
				tokens = append(tokens, token{typ: tokAnd, pos: i})
				i += 2
			} else {
				return nil, fmt.Errorf("expected '&&' at %d", i)
			}
		case '|':
			if i+1 < len(chars) && chars[i+1] == '|' {
				tokens = append(tokens, token{typ: tokOr, pos: i})
				i += 2
			} else {
				return nil, fmt.Errorf("expected '||' at %d", i)
			}
		case '>':
			if i+1 < len(chars) && chars[i+1] == '=' {
				tokens = append(tokens, token{typ: tokGe, pos: i})
				i += 2
			} else {
				tokens = append(tokens, token{typ: tokGt, pos: i})
				i++
			}
		case '<':
			if i+1 < len(chars) && chars[i+1] == '=' {
				tokens = append(tokens, token{typ: tokLe, pos: i})
				i += 2
			} else {
				tokens = append(tokens, token{typ: tokLt, pos: i})
				i++
			}
		case '\'', '"':
			t, next, err := parseString(chars, i)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, t)
			i = next
		default:
			if unicode.IsDigit(c) {
				t, next, err := parseNumber(chars, i)
				if err != nil {
					return nil, err
				}
				tokens = append(tokens, t)
				i = next
				continue
			}
			if unicode.IsLetter(c) || c == '_' {
				t, next := parseIdent(chars, i)
				tokens = append(tokens, t)
				i = next
				continue
			}
			return nil, fmt.Errorf("unexpected character %q at %d", c, i)
		}
	}
	tokens = append(tokens, token{typ: tokEOF, pos: len(input)})
	return tokens, nil
}

func parseString(chars []rune, start int) (token, int, error) {
	quote := chars[start]
	i := start + 1
	for i < len(chars) && chars[i] != quote {
		i++
	}
	if i >= len(chars) {
		return token{}, 0, fmt.Errorf("string not closed starting at %d", start)
	}
	lit := string(chars[start+1 : i])
	return token{typ: tokStr, lit: lit, pos: start}, i + 1, nil
}

func parseNumber(chars []rune, start int) (token, int, error) {
	i := start
	if chars[i] == '+' || chars[i] == '-' {
		i++
	}
	dotCount := 0
	for i < len(chars) {
		if unicode.IsDigit(chars[i]) {
			i++
			continue
		}
		if chars[i] == '.' {
			if dotCount > 0 || i+1 >= len(chars) || !unicode.IsDigit(chars[i+1]) {
				break
			}
			dotCount++
			i++
			continue
		}
		break
	}
	lit := string(chars[start:i])
	if dotCount > 0 {
		if _, err := strconv.ParseFloat(lit, 64); err != nil {
			return token{}, 0, err
		}
		return token{typ: tokFloat, lit: lit, pos: start}, i, nil
	}
	iv, err := strconv.ParseInt(lit, 10, 64)
	if err != nil {
		return token{}, 0, err
	}
	return token{typ: tokInt, lit: strconv.FormatInt(iv, 10), pos: start}, i, nil
}

func parseIdent(chars []rune, start int) (token, int) {
	i := start
	for i < len(chars) {
		c := chars[i]
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' || c == '-' {
			i++
			continue
		}
		break
	}
	lit := string(chars[start:i])
	switch lit {
	case "true", "false":
		return token{typ: tokBool, lit: lit, pos: start}, i
	case "nil":
		return token{typ: tokNil, lit: lit, pos: start}, i
	case "in":
		return token{typ: tokIn, lit: lit, pos: start}, i
	default:
		return token{typ: tokId, lit: lit, pos: start}, i
	}
}
