// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package syntax

import (
	"cobalt/debug"
	"cobalt/src"
)

const trace = debug.Enabled && false // for if we want parser tracing

type parser struct{ scanner }

func (p *parser) got(tok token) bool {
	if p.tok == tok {
		p.next()
		return true
	}
	return false
}

func (p *parser) want(tok token) src.Pos {
	if p.tok != tok {
		p.error("expected " + tok.String())
	}
	pos := p.pos()
	p.next()
	return pos
}

func (p *parser) semi() {
	if p.tok != _Semi {
		p.error("expected semicolon")
	}
	p.next()
}

func (p *parser) pos() src.Pos {
	return p.at(p.line, p.col)
}

// errorAt reports an error at the specified position and bails out.

// error reports an error at the current token position and bails out.
func (p *parser) error(msg string) {
	p.errorAt(p.pos(), msg)
}

// ----------------------------------------------------------------------------
// Source file(s)

func (p *parser) file() *File {
	if trace {
		defer debug.Trace()()
	}

	p.next() // read first token

	f := new(File)
	f.pos = p.pos()

	for p.tok != _EOF {
		f.DeclList = append(f.DeclList, p.decl(true))
	}

	// p.tok == _EOF
	f.EOF = p.pos()
	return f
}

// ----------------------------------------------------------------------------
// Declarations

func (p *parser) decl(global bool) Decl {
	if trace {
		defer debug.Trace()()
	}

	switch p.tok {
	case _Const:
		return p.constDecl()

	case _Var:
		return p.varDecl()
	}

	p.error("expected a declaration")
	return nil // unreachable
}

func (p *parser) constDecl() *ConstDecl {
	if trace {
		defer debug.Trace()()
	}

	d := new(ConstDecl)
	d.pos = p.want(_Const)

	d.NameList = p.nameList()
	d.Type = p.annotationOrNil()
	d.Values = p.initialization(_Const)

	p.semi()
	return d
}

func (p *parser) varDecl() *VarDecl {
	if trace {
		defer debug.Trace()()
	}

	d := new(VarDecl)
	d.pos = p.want(_Var)

	d.NameList = p.nameList()
	d.Type = p.annotationOrNil()

	if d.Type == nil {
		// no type annotation, so we require an initialization expression
		d.Values = p.initialization(_Var)
	} else if p.got(_Assign) {
		// initialization is optional, so we check for assign
		d.Values = p.exprList()
	}

	p.semi()
	return d
}

func (p *parser) initialization(tok token) Expr {
	if trace {
		defer debug.Trace()()
	}

	if !p.got(_Assign) {
		msg := "expected an initialization"
		if tok == _Var {
			msg += " or type annotation"
		}
		p.error(msg)
	}

	return p.exprList()
}

func (p *parser) annotationOrNil() Expr {
	if trace {
		defer debug.Trace()()
	}

	if p.got(_Colon) {
		return p.type_()
	}

	return nil
}

// ----------------------------------------------------------------------------
// Statements

func (p *parser) stmt() Stmt {
	if trace {
		defer debug.Trace()()
	}

	// skip semicolons (empty statements)
	for p.tok == _Semi {
		p.next()
	}

	// common occurrence, so we give it a fast path
	if p.tok == _Name {
		return p.simpleStmt()
	}

	switch p.tok {
	case _Const, _Var:
		return p.declStmt()

	case _Lbrace:
		return p.blockStmt()

	case _Return:
		return p.returnStmt()

	default:
		return p.simpleStmt()
	}
}

func (p *parser) simpleStmt() Stmt {
	if trace {
		defer debug.Trace()()
	}

	lhs := p.exprList()

	if _, ok := lhs.(*ListExpr); ok {
		if p.got(_Assign) {
			return p.assign(lhs, 0, p.exprList())
		}

		// with multiple lhs expressions, only allowed is "="
		p.error("expected \"=\" or comma")
	}

	// singular lhs
	switch p.tok {
	case _AssignOp:
		op := p.op
		p.next()
		return p.assign(lhs, op, p.expr())

	case _Assign:
		p.next()
		return p.assign(lhs, 0, p.expr())

	default:
		// expression statement so p.tok should be semicolon
		p.semi()

		s := new(ExprStmt)
		s.pos = lhs.Pos()
		s.X = lhs
		return s
	}
}

func (p *parser) assign(lhs Expr, op Operator, rhs Expr) *AssignStmt {
	p.semi() // we expect a semicolon at end of statement

	a := new(AssignStmt)
	a.pos = lhs.Pos()
	a.Lhs = lhs
	a.Op = op
	a.Rhs = rhs
	return a
}

func (p *parser) declStmt() *DeclStmt {
	if trace {
		defer debug.Trace()()
	}

	s := new(DeclStmt)
	s.pos = p.pos()
	s.D = p.decl(false)

	return s
}

func (p *parser) blockStmt() *BlockStmt {
	if trace {
		defer debug.Trace()()
	}

	s := new(BlockStmt)
	s.pos = p.want(_Lbrace)

	for p.tok != _EOF && p.tok != _Rbrace {
		s.StmtList = append(s.StmtList, p.stmt())
	}
	p.want(_Rbrace)

	// a semicolon is not required after a block statement
	return s
}

func (p *parser) returnStmt() *ReturnStmt {
	if trace {
		defer debug.Trace()()
	}

	s := new(ReturnStmt)
	s.pos = p.want(_Return)

	if p.tok != _Semi {
		s.Result = p.expr() // no multi-value returns
	}

	p.semi()
	return s
}

// ----------------------------------------------------------------------------
// Expressions

func (p *parser) expr() Expr {
	if trace {
		defer debug.Trace()()
	}

	x := p.binaryExpr(nil, 0)

	if p.got(_Cond) {
		t := new(TernaryExpr)
		t.pos = x.Pos()
		t.Cond = x

		t.A = p.expr()
		p.want(_Colon)
		t.B = p.expr()

		x = t
	}

	return x
}

func (p *parser) exprList() Expr {
	if trace {
		defer debug.Trace()()
	}

	x := p.expr()
	if p.got(_Comma) {
		list := []Expr{x, p.expr()}
		for p.got(_Comma) {
			list = append(list, p.expr())
		}
		t := new(ListExpr)
		t.pos = x.Pos()
		t.List = list
		x = t
	}
	return x
}

func (p *parser) binaryExpr(x Expr, prec int) Expr {
	if trace {
		defer debug.Trace()()
	}

	if x == nil {
		x = p.unaryExpr()
	}

	for (p.tok == _Operator || p.tok == _Star) && p.prec > prec {
		t := new(Operation)
		t.pos = p.pos()
		t.Op = p.op
		tprec := p.prec
		p.next()
		t.Lhs = x
		t.Rhs = p.binaryExpr(nil, tprec)
		x = t
	}

	return x
}

func (p *parser) unaryExpr() Expr {
	if trace {
		defer debug.Trace()()
	}

	var x Expr
	if p.tok == _Operator {
		x = p.prefixUnary()
	} else {
		x = p.primaryExpr()
	}

	x = p.postfixUnary(x)

	return x
}

func (p *parser) prefixUnary() Expr {
	// contrary to postfixUnary, the logic here is not in a loop. With the
	// allowed operators, it is useless/ugly/unnecessary to put it a unary
	// expression.
	debug.Assert(p.tok == _Operator)

	switch p.op {
	case Add, Sub, And, Not, LNot, Inc, Dec:
		x := new(Operation)
		x.pos = p.pos()
		x.Op = p.op
		p.next()

		x.Rhs = p.unaryExpr()
		return x
	}

	p.error("expected a unary expression")
	return nil // unreachable
}

func (p *parser) postfixUnary(x Expr) Expr {
	// we have the logic in a loop because postfix unary expressions "can" be
	// chained together. e.g. `x.*.*` for a double dereference.
	for p.tok == _Operator {
		switch p.op {
		case Inc, Dec, Deref:
			t := new(Operation)
			t.pos = p.pos()
			t.Op = p.op
			p.next()

			t.Lhs = x
			x = t
		}
	}

	// no default case, as this could be the lhs of a binary expression.

	return x
}

func (p *parser) primaryExpr() Expr {
	if trace {
		defer debug.Trace()()
	}

	x := p.atomExpr()
	for {
		switch p.tok {
		case _Lparen:
			x = p.callExpr(x)

		case _Lbrack:
			x = p.indexExpr(x)

		default:
			return x
		}
	}
}

func (p *parser) atomExpr() Expr {
	x := p.atomExprOrNil()
	if x == nil {
		p.error("expected an expression")
	}
	return x
}

func (p *parser) atomExprOrNil() Expr {
	if trace {
		defer debug.Trace()()
	}

	switch p.tok {
	case _Name:
		return p.name()

	case _Literal:
		x := new(LiteralExpr)
		x.pos = p.pos()
		x.Value, x.Kind = p.lit, p.kind
		p.next()
		return x

	case _Lparen:
		pos := p.pos()
		p.next()
		x := p.expr()
		p.want(_Rparen)

		if t := p.atomExprOrNil(); t != nil {
			c := new(CastExpr)
			c.pos = pos
			c.Type, c.X = x, t
			x = c
		}

		return x

	case _Proc:
		typ := p.procType()
		if p.tok == _Lbrace {
			x := new(ProcExpr)
			x.pos = typ.pos
			x.Type = typ
			x.Body = p.blockStmt()
			return x
		}
		return typ

	default:
		return p.typeOrNil()
	}
}

func (p *parser) callExpr(x Expr) *CallExpr {
	if trace {
		defer debug.Trace()()
	}

	t := new(CallExpr)
	t.pos = p.pos()
	t.Proc = x

	p.want(_Lparen)
	if p.got(_Rparen) {
		return t
	}

	list := []Expr{p.expr()}
	for p.got(_Comma) {
		list = append(list, p.expr())
	}
	p.want(_Rparen)

	t.ArgList = list
	return t
}

func (p *parser) indexExpr(x Expr) *IndexExpr {
	if trace {
		defer debug.Trace()()
	}

	t := new(IndexExpr)
	t.pos = p.pos()
	t.X = x

	p.want(_Lbrack)
	t.Index = p.expr()
	p.want(_Rbrack)

	return t
}

func (p *parser) name() *Name {
	if p.tok != _Name {
		p.error("expected a name")
	}

	n := new(Name)
	n.Value, n.pos = p.lit, p.pos()
	p.next()
	return n
}

func (p *parser) nameList() []*Name {
	if trace {
		defer debug.Trace()()
	}

	list := []*Name{p.name()}
	for p.got(_Comma) {
		list = append(list, p.name())
	}
	return list
}

// ----------------------------------------------------------------------------
// Types

func (p *parser) type_() Expr {
	typ := p.typeOrNil()
	if typ == nil {
		p.error("expected a type")
	}
	return typ
}

func (p *parser) typeOrNil() Expr {
	if trace {
		defer debug.Trace()()
	}

	switch p.tok {
	case _Name:
		return p.name()

	case _Star:
		x := new(PointerType)
		x.pos = p.pos()
		p.next()
		x.Const = p.got(_Const)
		x.Elem = p.type_()
		return x

	case _Cond:
		x := new(OptionType)
		x.pos = p.pos()
		p.next()
		x.Elem = p.type_()
		return x

	case _Lbrack:
		x := new(ArrayType)
		x.pos = p.pos()
		p.next()
		x.Len = p.expr()
		p.want(_Rbrack)
		x.Elem = p.type_()
		return x

	case _Proc:
		return p.procType()
	}

	return nil
}

func (p *parser) procType() *ProcType {
	if trace {
		defer debug.Trace()()
	}

	typ := new(ProcType)
	typ.pos = p.want(_Proc)

	typ.ParamList = p.paramList()
	typ.Result = p.typeOrNil()

	return typ
}

func (p *parser) paramList() []*Field {
	if trace {
		defer debug.Trace()()
	}

	pos := p.want(_Lparen)
	if p.got(_Rparen) {
		return nil
	}

	var list []*Field
	var named, unnamed bool
	for p.tok != _EOF && p.tok != _Rparen {
		f, isNamed := p.field()
		list = append(list, f)

		named = named || isNamed
		unnamed = unnamed || !isNamed

		if !p.got(_Comma) && p.tok != _Rparen {
			p.error("expected a comma or \")\"")
		}
	}
	p.want(_Rparen)

	if named && unnamed {
		p.errorAt(pos, "got mixed named and unnamed parameters")
	}

	return list
}

func (p *parser) field() (f *Field, named bool) {
	if trace {
		defer debug.Trace()()
	}

	f = new(Field)
	f.pos = p.pos()
	f.Const = p.got(_Const)

	x := p.type_()
	if name, ok := x.(*Name); ok {
		// we potentially have a named field
		if typ := p.annotationOrNil(); typ == nil {
			// no type annotation, so the Name is a type name
			f.Type = name
		} else {
			// type annotation, so the Name is the field name
			named = true
			f.Name, f.Type = name, typ
		}
	} else {
		// no chance of being a named field
		f.Type = x
	}

	return
}
