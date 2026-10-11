package calc

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"unicode"
)

type tokenType int

const (
	tokenNumber tokenType = iota
	tokenPlus
	tokenMinus
	tokenStar
	tokenSlash
	tokenPercent
	tokenCaret
	tokenLParen
	tokenRParen
	tokenEOF
)

type token struct {
	typ tokenType
	val float64
}

var operators = map[byte]tokenType{
	'+': tokenPlus,
	'-': tokenMinus,
	'*': tokenStar,
	'/': tokenSlash,
	'%': tokenPercent,
	'^': tokenCaret,
	'(': tokenLParen,
	')': tokenRParen,
}

func tokenize(input string) ([]token, error) {
	var tokens []token
	i := 0
	for i < len(input) {
		ch := input[i]
		if unicode.IsSpace(rune(ch)) {
			i++
			continue
		}
		if typ, ok := operators[ch]; ok {
			tokens = append(tokens, token{typ: typ})
			i++
			continue
		}
		if !isNumberChar(ch) {
			return nil, fmt.Errorf("unexpected character '%c' at position %d", ch, i)
		}
		val, end, err := readNumber(input, i)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token{typ: tokenNumber, val: val})
		i = end
	}
	tokens = append(tokens, token{typ: tokenEOF})
	return tokens, nil
}

func isNumberChar(ch byte) bool {
	return ch == '.' || (ch >= '0' && ch <= '9')
}

func readNumber(input string, start int) (float64, int, error) {
	i := start
	dotSeen := false
	for i < len(input) && isNumberChar(input[i]) {
		if input[i] == '.' {
			if dotSeen {
				return 0, 0, fmt.Errorf("invalid number at position %d", start)
			}
			dotSeen = true
		}
		i++
	}
	var val float64
	if _, err := fmt.Sscanf(input[start:i], "%f", &val); err != nil {
		return 0, 0, fmt.Errorf("invalid number at position %d", start)
	}
	return val, i, nil
}

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) peek() token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return token{typ: tokenEOF}
}

func (p *parser) next() {
	p.pos++
}

func (p *parser) parseExpr() (float64, error) {
	return p.parseAddSub()
}

func (p *parser) parseAddSub() (float64, error) {
	left, err := p.parseMulDivMod()
	if err != nil {
		return 0, err
	}
	for {
		op := p.peek().typ
		if op != tokenPlus && op != tokenMinus {
			return left, nil
		}
		p.next()
		right, err := p.parseMulDivMod()
		if err != nil {
			return 0, err
		}
		if op == tokenPlus {
			left += right
		} else {
			left -= right
		}
	}
}

func (p *parser) parseMulDivMod() (float64, error) {
	left, err := p.parsePower()
	if err != nil {
		return 0, err
	}
	for {
		op := p.peek().typ
		if op != tokenStar && op != tokenSlash && op != tokenPercent {
			return left, nil
		}
		p.next()
		right, err := p.parsePower()
		if err != nil {
			return 0, err
		}
		left, err = applyMulDivMod(op, left, right)
		if err != nil {
			return 0, err
		}
	}
}

func applyMulDivMod(op tokenType, left, right float64) (float64, error) {
	switch op {
	case tokenSlash:
		if right == 0 {
			return 0, errors.New("division by zero")
		}
		return left / right, nil
	case tokenPercent:
		if right == 0 {
			return 0, errors.New("modulo by zero")
		}
		return math.Mod(left, right), nil
	default:
		return left * right, nil
	}
}

func (p *parser) parsePower() (float64, error) {
	base, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	if p.peek().typ == tokenCaret {
		p.next()
		exp, err := p.parsePower()
		if err != nil {
			return 0, err
		}
		return math.Pow(base, exp), nil
	}
	return base, nil
}

func (p *parser) parseUnary() (float64, error) {
	if p.peek().typ == tokenMinus {
		p.next()
		val, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return -val, nil
	}
	if p.peek().typ == tokenPlus {
		p.next()
		return p.parseUnary()
	}
	return p.parsePrimary()
}

func (p *parser) parsePrimary() (float64, error) {
	t := p.peek()
	if t.typ == tokenNumber {
		p.next()
		return t.val, nil
	}
	if t.typ == tokenLParen {
		p.next()
		val, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if p.peek().typ != tokenRParen {
			return 0, errors.New("expected ')'")
		}
		p.next()
		return val, nil
	}
	return 0, errors.New("unexpected token")
}

func evaluate(input string) (float64, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, errors.New("empty expression")
	}
	tokens, err := tokenize(input)
	if err != nil {
		return 0, err
	}
	p := &parser{tokens: tokens}
	result, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	if p.peek().typ != tokenEOF {
		return 0, errors.New("unexpected trailing characters")
	}
	return result, nil
}

func looksLikeExpression(query string) bool {
	for _, ch := range query {
		if strings.ContainsRune("+-*/%^().", ch) {
			return true
		}
		if unicode.IsDigit(ch) {
			continue
		}
		if unicode.IsSpace(ch) {
			continue
		}
		return false
	}
	return false
}

func formatResult(val float64) string {
	if val == math.Trunc(val) && !math.IsInf(val, 0) && !math.IsNaN(val) {
		return fmt.Sprintf("%.0f", val)
	}
	s := fmt.Sprintf("%g", val)
	return s
}
