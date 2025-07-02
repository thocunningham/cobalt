// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package syntax

import "cobalt/src"

// Node is a syntax node, used to represent the syntax structure of a source
// file. Each node implements this interface.
type Node interface {
	// Pos returns a position associated with the Node. For nodes such as names
	// or basic literals this is the start of the node, but for others such as
	// a binary expression, it is a different position not accessible via its
	// children.
	Pos() src.Pos

	sNode() // prohibits external implementations
}

type node struct{ pos src.Pos }

func (n *node) Pos() src.Pos { return n.pos }
func (*node) sNode()         {}

// ----------------------------------------------------------------------------
// Files

// File is a node representing the entirety of a source file.
type File struct {
	DeclList []Decl
	EOF      src.Pos
	node     // position of first non-comment token in file
}

// ----------------------------------------------------------------------------
// Declarations

type (
	// Decl is a collection of nodes that denote the declaration of a symbol in
	// some kind or another.
	Decl interface {
		Node
		sDecl()
	}

	// ConstDecl is a constant declaration.
	ConstDecl struct {
		NameList []*Name
		Type     Expr // nil means no type annotation
		Values   Expr // always non-nil
		decl          // position of "const"
	}

	// VarDecl is a variable declaration.
	VarDecl struct {
		NameList []*Name
		Type     Expr // nil means no type annotation
		Values   Expr // nil means no init expression
		decl          // position of "var"
	}
)

// decl ensures that all declaration nodes implement both Node and Decl.
type decl struct{ node }

func (*decl) sDecl() {}

// ----------------------------------------------------------------------------
// Expressions

type (
	// Expr is a collection of nodes that denote an evaluable expression.
	Expr interface {
		Node
		sExpr()
	}

	// Name is a name referencing a symbol.
	Name struct {
		Value string
		expr  // position of name
	}

	// BasicLit is a simple literal composed of a single token.
	LiteralExpr struct {
		Value string
		Kind  Literal
		expr  // position of literal
	}

	// ProcLit is a complete procedure literal with type and body.
	ProcExpr struct {
		Type *ProcType
		Body *BlockStmt
		expr // position of Type field
	}

	// Operation is a unary or binary expression.
	Operation struct {
		Lhs, Rhs Expr
		Op       Operator
		expr     // position of Op field
	}

	// TernaryExpr is a ternary expression.
	TernaryExpr struct {
		Cond Expr
		A, B Expr
		expr // position of "?"
	}

	// CallExpr is a call to a procedure.
	CallExpr struct {
		Proc    Expr
		ArgList []Expr
		expr    // position of "("
	}

	CastExpr struct {
		Type Expr
		X    Expr
		expr // position of "("
	}

	// ListExpr is a list of expressions.
	ListExpr struct {
		List []Expr
		expr // position of List[0]
	}

	// PointerType is a pointer type.
	PointerType struct {
		Const bool // "const" present?
		Elem  Expr
		expr  // position of "*"
	}

	// OptionType is an optional type.
	OptionType struct {
		Elem Expr
		expr // position of "?"
	}

	// ArrayType is a fixed-length array type.
	ArrayType struct {
		Len  Expr
		Elem Expr
		expr // position of "["
	}

	// ProcType is a procedure type.
	ProcType struct {
		ParamList []*Field
		Result    Expr // can be nil
		expr           // position of "proc"
	}

	// Field is a possibly named type field in a struct or procedure type.
	Field struct {
		Name  *Name // can be nil
		Type  Expr
		Const bool
		node  // position Name field
	}
)

type expr struct{ node }

func (*expr) sExpr() {}

// ----------------------------------------------------------------------------
// Statements

type (
	// Stmt represents an operation to be carried out in a procedure.
	Stmt interface {
		Node
		sStmt()
	}

	// BlockStmt is a sequence of statements enclosed in braces.
	BlockStmt struct {
		StmtList []Stmt
		Closing  src.Pos // position of "}"
		stmt             // position of "{"
	}

	// ExprStmt is an expression as a statement.
	ExprStmt struct {
		X    Expr
		stmt // position of X field
	}

	// DeclStmt is a declaration as a statement.
	DeclStmt struct {
		D    Decl
		stmt // position of D field
	}

	// AssignStmt is an assignment.
	AssignStmt struct {
		Op       Operator
		Lhs, Rhs Expr
		stmt     // position of Op field
	}

	// ReturnStmt is a procedure return statement.
	ReturnStmt struct {
		Result Expr
		stmt   // position of "return"
	}
)

type stmt struct{ node }

func (*stmt) sStmt() {}

// ----------------------------------------------------------------------------
// Utilities

// UnpackList unpacks an Expr or ListExpr into a slice of Expr.
// If x is nil, the result is also nil.
func UnpackList(x Expr) []Expr {
	switch x := x.(type) {
	case nil:
		return nil
	case *ListExpr:
		return x.List
	default:
		return []Expr{x}
	}
}
