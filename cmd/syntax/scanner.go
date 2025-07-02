// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type scanner struct {
	source

	// current token, valid after calling next()
	line, col uint
	lit       string // valid if tok is _Name, _Literal; may be malformed if bad is true
	tok       token
	kind      Literal  // valid if tok is _Literal
	op        Operator // valid if tok is _Operator, _Star, _AssignOp, or _IncOp
	prec      int      // valid if tok is _Operator, _Star, _AssignOp, or _IncOp
}

// errorf reports an error at the most recently read character position.
func (s *scanner) errorf(format string, args ...any) {
	s.error(fmt.Sprintf(format, args...))
}

// errorAtf reports an error at a byte column offset relative to the current token start.
func (s *scanner) errorAtf(offset int, format string, args ...any) {
	s.errorAt(s.at(s.line, s.col+uint(offset)), fmt.Sprintf(format, args...))
}

func (s *scanner) setLit(kind Literal) {
	s.tok = _Literal
	s.lit = string(s.segment())
	s.kind = kind
}

func (s *scanner) next() {
redo:
	// skip white space
	s.stop()
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.nextch()
	}

	// token start
	s.line, s.col = s.pos()
	s.start()
	if isLetter(s.ch) || s.ch >= utf8.RuneSelf && unicode.IsLetter(s.ch) {
		s.nextch()
		s.name()
		return
	}

	switch s.ch {
	case -1:
		s.tok = _EOF

	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		s.number(false)

	case '\'':
		s.char()

	// case '"':
	// 	s.string()

	case '(':
		s.nextch()
		s.tok = _Lparen

	case '[':
		s.nextch()
		s.tok = _Lbrack

	case '{':
		s.nextch()
		s.tok = _Lbrace

	case ',':
		s.nextch()
		s.tok = _Comma

	case ';':
		s.nextch()
		s.tok = _Semi

	case ')':
		s.nextch()
		s.tok = _Rparen

	case ']':
		s.nextch()
		s.tok = _Rbrack

	case '}':
		s.nextch()
		s.tok = _Rbrace

	case ':':
		s.nextch()
		s.tok = _Colon

	case '.':
		s.nextch()
		if isDecimal(s.ch) {
			s.number(true)
			break
		} else {
			s.tok = _Dot
		}

	case '+':
		s.nextch()
		s.op, s.prec = Add, precAdd
		if s.ch != '+' {
			goto assignop
		}
		s.nextch()
		s.tok = _Operator
		s.op = Inc

	case '-':
		s.nextch()
		s.op, s.prec = Sub, precAdd
		if s.ch != '-' {
			goto assignop
		}
		s.tok = _Operator
		s.op = Dec
		s.nextch()

	case '*':
		s.nextch()
		s.op, s.prec = Mul, precMul
		// don't goto assignop - want _Star token
		if s.ch == '=' {
			s.nextch()
			s.tok = _AssignOp
			break
		}
		s.tok = _Star

	case '/':
		s.nextch()
		if s.ch == '/' || s.ch == '*' {
			s.comment()
			goto redo
		}
		s.op, s.prec = Div, precMul
		goto assignop

	case '%':
		s.nextch()
		s.op, s.prec = Rem, precMul
		goto assignop

	case '&':
		s.nextch()
		if s.ch == '&' {
			s.nextch()
			s.op, s.prec = AndAnd, precAndAnd
			s.tok = _Operator
			break
		}
		s.op, s.prec = And, precMul
		goto assignop

	case '|':
		s.nextch()
		if s.ch == '|' {
			s.nextch()
			s.op, s.prec = OrOr, precOrOr
			s.tok = _Operator
			break
		}
		s.op, s.prec = Or, precAdd
		goto assignop

	case '^':
		s.nextch()
		s.op, s.prec = Xor, precAdd
		goto assignop

	case '<':
		s.nextch()
		if s.ch == '=' {
			s.nextch()
			s.op, s.prec = Leq, precCmp
			s.tok = _Operator
			break
		}
		if s.ch == '<' {
			s.nextch()
			s.op, s.prec = Shl, precMul
			goto assignop
		}
		s.op, s.prec = Lss, precCmp
		s.tok = _Operator

	case '>':
		s.nextch()
		if s.ch == '=' {
			s.nextch()
			s.op, s.prec = Geq, precCmp
			s.tok = _Operator
			break
		}
		if s.ch == '>' {
			s.nextch()
			s.op, s.prec = Shr, precMul
			goto assignop
		}
		s.op, s.prec = Gtr, precCmp
		s.tok = _Operator

	case '=':
		s.nextch()
		if s.ch == '=' {
			s.nextch()
			s.op, s.prec = Eql, precCmp
			s.tok = _Operator
			break
		}
		s.op = 0 // denotes a simple assignment
		s.tok = _Assign

	case '!':
		s.nextch()
		if s.ch == '=' {
			s.nextch()
			s.op, s.prec = Neq, precCmp
			s.tok = _Operator
			break
		}
		s.op, s.prec = LNot, 0
		s.tok = _Operator

	case '~':
		s.nextch()
		s.op, s.prec = Not, 0
		s.tok = _Operator

	case '?':
		s.nextch()
		s.tok = _Cond

	default:
		s.errorf("invalid character %#U", s.ch)
	}

	return

assignop:
	if s.ch == '=' {
		s.nextch()
		s.tok = _AssignOp
		return
	}
	s.tok = _Operator
}

func (s *scanner) name() {
	const maxlength = 100

	// accelerate common case (7bit ASCII)
	for isLetter(s.ch) || isDecimal(s.ch) {
		s.nextch()
	}

	// general case
	if s.ch >= utf8.RuneSelf {
		for s.atIdentChar() {
			s.nextch()
		}
	}

	// possibly a keyword
	lit := s.segment()
	if len(lit) >= 2 {
		if tok, ok := keywordMap[string(lit)]; ok {
			s.tok = tok
			return
		}
	}

	if len(lit) > maxlength {
		s.errorAt(s.at(s.line, s.col), "excessively long name")
	}

	s.lit = string(lit)
	s.tok = _Name
}

func (s *scanner) atIdentChar() bool {
	if unicode.IsLetter(s.ch) || unicode.IsDigit(s.ch) || s.ch == '_' {
		return true
	}
	if s.ch >= utf8.RuneSelf {
		s.errorf("invalid character %#U in identifier", s.ch)
	}
	return false
}

var keywordMap map[string]token

func init() {
	// populate keywordMap
	keywordMap = make(map[string]token, keywordLast-keywordFirst-1)
	for tok := keywordFirst + 1; tok < keywordLast; tok++ {
		keywordMap[tok.String()] = tok
	}
}

func lower(ch rune) rune     { return ('a' - 'A') | ch } // returns lower-case ch iff ch is ASCII letter
func isLetter(ch rune) bool  { return 'a' <= lower(ch) && lower(ch) <= 'z' || ch == '_' }
func isDecimal(ch rune) bool { return '0' <= ch && ch <= '9' }
func isHex(ch rune) bool     { return '0' <= ch && ch <= '9' || 'a' <= lower(ch) && lower(ch) <= 'f' }

func (s *scanner) digits(base int, invalid *int) (digsep int) {
	if base <= 10 {
		max := rune('0' + base)
		for isDecimal(s.ch) || s.ch == '_' {
			ds := 1
			if s.ch == '_' {
				ds = 2
			} else if s.ch >= max && *invalid < 0 {
				_, col := s.pos()
				*invalid = int(col - s.col) // record invalid rune index
			}
			digsep |= ds
			s.nextch()
		}
	} else {
		for isHex(s.ch) || s.ch == '_' {
			ds := 1
			if s.ch == '_' {
				ds = 2
			}
			digsep |= ds
			s.nextch()
		}
	}
	return
}

func (s *scanner) number(seenPoint bool) {
	const maxlength = 200

	kind := Int
	base := 10        // number base
	prefix := rune(0) // one of 0 (decimal), '0' (0-octal), 'x', 'o', or 'b'
	digsep := 0       // bit 0: digit present, bit 1: '_' present
	invalid := -1     // index of invalid digit in literal, or < 0

	// integer part
	if !seenPoint {
		if s.ch == '0' {
			s.nextch()
			switch lower(s.ch) {
			case 'x':
				s.nextch()
				base, prefix = 16, 'x'
			case 'o':
				s.nextch()
				base, prefix = 8, 'o'
			case 'b':
				s.nextch()
				base, prefix = 2, 'b'
			default:
				base, prefix = 8, '0'
				digsep = 1 // leading 0
			}
		}
		digsep |= s.digits(base, &invalid)
		if s.ch == '.' {
			if prefix != 0 {
				s.error("can only add decimal point to base-10 literals")
			}
			s.nextch()
			seenPoint = true
		}
	}

	// fractional part
	if seenPoint {
		kind = Float
		digsep |= s.digits(base, &invalid)
	}

	if digsep&1 == 0 {
		s.errorf("%s literal has no digits", baseName(base))
	}

	// exponent
	if lower(s.ch) == 'e' {
		if prefix != 0 {
			s.error("'e' exponent requires decimal mantissa")
		}
		s.nextch()
		kind = Float
		if s.ch == '+' || s.ch == '-' {
			s.nextch()
		}
		digsep = s.digits(10, nil) | digsep&2 // don't lose sep bit
		if digsep&1 == 0 {
			s.errorf("exponent has no digits")
		}
	}

	s.setLit(kind) // do this now so we can use s.lit below

	if kind == Int && invalid >= 0 {
		s.errorAtf(invalid, "invalid digit %q in %s literal", s.lit[invalid], baseName(base))
	}

	if digsep&2 != 0 {
		if i := invalidSep(s.lit); i >= 0 {
			s.errorAtf(i, "'_' must separate successive digits")
		}
	}

	if len(s.lit) > maxlength {
		s.errorAt(s.at(s.line, s.col), "excessively long number")
	}
}

func baseName(b int) string {
	switch b {
	case 2:
		return "binary"
	case 8:
		return "octal"
	case 10:
		return "decimal"
	case 16:
		return "hexadecimal"
	}
	panic("unreachable")
}

func invalidSep(x string) int {
	x1 := ' ' // prefix char, we only care if it's 'x'
	d := '.'  // digit, one of '_', '0' (a digit), or '.' (anything else)
	i := 0

	// a prefix counts as a digit
	if len(x) >= 2 && x[0] == '0' {
		x1 = lower(rune(x[1]))
		if x1 == 'x' || x1 == 'o' || x1 == 'b' {
			d = '0'
			i = 2
		}
	}

	// mantissa and exponent
	for ; i < len(x); i++ {
		p := d // previous digit
		d = rune(x[i])
		switch {
		case d == '_':
			if p != '0' {
				return i
			}
		case isDecimal(d) || x1 == 'x' && isHex(d):
			d = '0'
		default:
			if p == '_' {
				return i - 1
			}
			d = '.'
		}
	}
	if d == '_' {
		return len(x) - 1
	}

	return -1
}

func (s *scanner) char() {
	s.nextch()

	n := 0
loop:
	for ; ; n++ {
		switch s.ch {
		case '\'':
			if n == 0 {
				s.errorf("empty character literal or unescaped '")
			} else if n != 1 {
				s.errorAtf(0, "more than one character in character literal")
			}
			s.nextch()
			break loop
		case '\\':
			s.nextch()
			s.escape('\'')
			continue
		case '\n':
			s.errorf("newline in character literal")
		}
		if s.ch < 0 {
			s.errorAtf(0, "character literal not terminated")
		}
		s.nextch()
	}

	s.setLit(Char)
}

// func (s *scanner) string() {
// 	s.nextch()

// loop:
// 	for {
// 		switch s.ch {
// 		case '"':
// 			s.nextch()
// 			break loop

// 		case '\\':
// 			s.nextch()
// 			s.escape('"')
// 			continue

// 		case '\n':
// 			s.errorf("newline in string literal")
// 		}
// 		if s.ch < 0 {
// 			s.errorAtf(0, "string literal not terminated")
// 		}
// 		s.nextch()
// 	}

// 	s.setLit(String)
// }

func (s *scanner) comment() {
	ch := s.ch
	s.next()
	if ch == '/' {
		for s.ch >= 0 && s.ch != '\n' {
			s.nextch()
		}
	} else {
		// s.ch == '*'
		s.nextch()
		lev := 1
		for s.ch >= 0 && lev > 0 {
			switch s.ch {
			case '/':
				s.nextch()
				if s.ch == '*' {
					s.nextch()
					lev++
				}
			case '*':
				s.nextch()
				if s.ch == '/' {
					s.nextch()
					lev--
				}
			default:
				s.nextch()
			}
		}
		if lev > 0 {
			s.errorAtf(0, "comment not terminated")
		}
	}
}

func (s *scanner) escape(quote rune) {
	var n int
	var base, max uint32

	switch s.ch {
	case quote, 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\':
		s.nextch()
		return
	case '0', '1', '2', '3', '4', '5', '6', '7':
		n, base, max = 3, 8, 255
	case 'x':
		s.nextch()
		n, base, max = 2, 16, 255
	case 'u':
		s.nextch()
		n, base, max = 4, 16, unicode.MaxRune
	case 'U':
		s.nextch()
		n, base, max = 8, 16, unicode.MaxRune
	default:
		if s.ch < 0 {
			return
		}
		s.errorf("unknown escape")
	}

	var x uint32
	for i := n; i > 0; i-- {
		if s.ch < 0 {
			return
		}
		d := base
		if isDecimal(s.ch) {
			d = uint32(s.ch) - '0'
		} else if 'a' <= lower(s.ch) && lower(s.ch) <= 'f' {
			d = uint32(lower(s.ch)) - 'a' + 10
		}
		if d >= base {
			s.errorf("invalid character %q in %s escape", s.ch, baseName(int(base)))
		}
		// d < base
		x = x*base + d
		s.nextch()
	}

	if x > max && base == 8 {
		s.errorf("octal escape value %d > 255", x)
	}

	if x > max || 0xD800 <= x && x < 0xE000 /* surrogate range */ {
		s.errorf("escape is invalid Unicode code point %#U", x)
	}
}
