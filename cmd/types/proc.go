// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package types

import (
	"cobalt/src"
	"cobalt/syntax"
)

var procmap map[*syntax.ProcExpr]*Proc

// Proc represents a singular procedure, with its own type and body.
type Proc struct {
	pos src.Pos // position of "proc"
	typ *Type

	body   *Scope
	params []*Symbol // parameters, in order
	code   *syntax.BlockStmt

	flags uint32
}

const (
	procNoreturn = 1 << iota
	procConst
	procPure
)

func NewProc(typ *Type, params []*Symbol, parent *Scope, node *syntax.ProcExpr) *Proc {
	if proc, ok := procmap[node]; ok {
		return proc
	}

	proc := new(Proc)
	proc.pos = node.Pos()
	proc.typ = typ
	proc.body = NewScope(parent, node.Body.Pos(), node.Body.Closing)
	proc.params = params
	proc.code = node.Body

	procmap[node] = proc
	return proc
}
