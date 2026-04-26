package calc

import (
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

func tokenize(input string) ([]token, error) {
	var tokens []token
	i := 0
	for i < len(input) {
		ch := input[i]
		if unicode.IsSpace(rune(ch)) {
			i++
			continue
		}
		switch ch {
		case '+':
			tokens = append(tokens, token{typ: tokenPlus})
		case '-':
			tokens = append(tokens, token{typ: tokenMinus})
		case '*':
			tokens = append(tokens, token{typ: tokenStar})
		case '/':
			tokens = append(tokens, token{typ: tokenSlash})
		case '%':
			tokens = append(tokens, token{typ: tokenPercent})
		case '^':
			tokens = append(tokens, token{typ: tokenCaret})
		case '(':
			tokens = append(tokens, token{typ: tokenLParen})
		case ')':
			tokens = append(tokens, token{typ: tokenRParen})
		default:
			if ch == '.' || (ch >= '0' && ch <= '9') {
				start := i
				dotSeen := false
				for i < len(input) && (input[i] == '.' || (input[i] >= '0' && input[i] <= '9')) {
					if input[i] == '.' {
						if dotSeen {
							return nil, fmt.Errorf("invalid number at position %d", start)
						}
						dotSeen = true
					}
					i++
				}
				var val float64
				_, err := fmt.Sscanf(input[start:i], "%f", &val)
				if err != nil {
					return nil, fmt.Errorf("invalid number at position %d", start)
				}
				tokens = append(tokens, token{typ: tokenNumber, val: val})
				continue
			} else {
				return nil, fmt.Errorf("unexpected character '%c' at position %d", ch, i)
			}
		}
		i++
	}
	tokens = append(tokens, token{typ: tokenEOF})
	return tokens, nil
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

func (p *parser) next() token {
	t := p.peek()
	p.pos++
	return t
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
		t := p.peek()
		if t.typ == tokenPlus {
			p.next()
			right, err := p.parseMulDivMod()
			if err != nil {
				return 0, err
			}
			left = left + right
		} else if t.typ == tokenMinus {
			p.next()
			right, err := p.parseMulDivMod()
			if err != nil {
				return 0, err
			}
			left = left - right
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parseMulDivMod() (float64, error) {
	left, err := p.parsePower()
	if err != nil {
		return 0, err
	}
	for {
		t := p.peek()
		if t.typ == tokenStar {
			p.next()
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			left = left * right
		} else if t.typ == tokenSlash {
			p.next()
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left = left / right
		} else if t.typ == tokenPercent {
			p.next()
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			left = math.Mod(left, right)
		} else {
			break
		}
	}
	return left, nil
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
			return 0, fmt.Errorf("expected ')'")
		}
		p.next()
		return val, nil
	}
	return 0, fmt.Errorf("unexpected token")
}

func evaluate(input string) (float64, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("empty expression")
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
		return 0, fmt.Errorf("unexpected trailing characters")
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
