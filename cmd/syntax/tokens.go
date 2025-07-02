// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

// This file declares all lexical tokens of which a Cobalt source file may be
// composed. These are scanned and passed on to the parser.

package syntax

type token uint8

//go:generate stringer -type token -linecomment tokens.go

const (
	_    token = iota
	_EOF       // EOF

	// names and literals
	_Name    // name
	_Literal // literal

	// operators and operations, _Operator not including "*".
	_Operator // op
	_AssignOp // op=
	_Assign   // =
	_Star     // *

	// delimiters
	_Lparen // (
	_Lbrack // [
	_Lbrace // {
	_Rparen // )
	_Rbrack // ]
	_Rbrace // }
	_Comma  // ,
	_Semi   // ;
	_Colon  // :
	_Dot    // .
	_Cond   // ?

	// keywords, more will be added over time.
	keywordFirst //
	_Const       // const
	_Proc        // proc
	_Return      // return
	_Var         // var
	keywordLast  //
)

//go:generate stringer -type Literal tokens.go

// LitKind represents the kind of a basic literal.
type Literal uint8

const (
	Int Literal = iota
	Float
	Char
	String
)

// An Operator represents an operator used in expressions.
type Operator uint8

//go:generate stringer -type Operator -linecomment tokens.go

const (
	_ Operator = iota

	// unary operators
	Not   // ~
	LNot  // !
	Inc   // ++
	Dec   // --
	Deref // .*

	// binary operators, highest precedence first
	// precOrOr
	OrOr // ||

	// precAndAnd
	AndAnd // &&

	// precCmp
	Eql // ==
	Neq // !=
	Lss // <
	Leq // <=
	Gtr // >
	Geq // >=

	// precAdd
	Add // +
	Sub // -
	Or  // |
	Xor // ^

	// precMul
	Mul // *
	Div // /
	Rem // %
	And // &
	Shl // <<
	Shr // >>

	OperatorMax
)

// Operator precedences
const (
	_ = iota
	precOrOr
	precAndAnd
	precCmp
	precAdd
	precMul
)
