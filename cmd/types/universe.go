// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package types

import (
	"cobalt/base"
	"cobalt/debug"
	"cobalt/src"
)

// Universe is the global scope containing an entire Cobalt program. It defines
// all built-in symbols and is the parent of all modules.
var Universe *Scope

// Types maps the Kinds of the built-in types to their respective Type. They
// are accessible by indexing using their Kind.
var Types [NBASIC]*Type

// PtrSize is the size of a pointer on the target architecture. It must be
// non-zero by the time [Init] is called, otherwise the program will abort.
var PtrSize int

func Init() {
	if PtrSize == 0 {
		base.Fatalf("types: PtrSize is unset")
	}

	Universe = NewScope(nil, src.NoPos, src.NoPos)
	initTypes()
	initConsts()
	initBuiltins()
}

func initTypes() {
	ttype := &Type{kind: TTYPE}
	sym := &Symbol{
		"type",
		src.NoPos,
		ttype,
		nil,
		nil,
		MakeType(ttype),
		symUsed | symConst | symStatic,
	}
	ttype.sym = sym

	Universe.Insert(sym)
	Types[TTYPE] = ttype

	const flags = symUsed | symConst | symStatic
	decl := func(kind Kind, name string) {
		debug.Assert(kind < NBASIC)

		typ := &Type{kind: kind}
		sym := &Symbol{name: name, typ: typ, extra: MakeType(typ), flags: flags}
		typ.sym = sym

		debug.Assert(Universe.Insert(sym) == nil, "duplicate declaration of builtin", name)
		Types[kind] = typ
	}

	decl(TVOID, "void")
	decl(TBOOL, "bool")
	decl(TINT8, "int8")
	decl(TINT16, "int16")
	decl(TINT32, "int32")
	decl(TINT64, "int64")
	decl(TINTPTR, "intptr")
	decl(TUINT8, "uint8")
	decl(TUINT16, "uint16")
	decl(TUINT32, "uint32")
	decl(TUINT64, "uint64")
	decl(TUINTPTR, "uintptr")
	decl(TFLOAT32, "float32")
	decl(TFLOAT64, "float64")
}

func initConsts() {
	const flags = symUsed | symConst | symStatic
	decl := func(kind Kind, name string, val Value) {
		debug.Assert(kind < NBASIC)
		debug.Assert(val != nil && val.Kind() != TUNDEF)

		sym := &Symbol{name: name, typ: Types[kind], extra: val, flags: flags}
		debug.Assert(Universe.Insert(sym) == nil, "duplicate declaration of builtin", name)
	}

	decl(TBOOL, "false", MakeBool(false))
	decl(TBOOL, "true", MakeBool(true))
}

func initBuiltins() {
	const flags = symUsed | symConst | symStatic | symBuiltin
	decl := func(builtin Builtin, name string) {
		sym := &Symbol{name: name, extra: builtin, flags: flags}
		debug.Assert(Universe.Insert(sym) == nil, "duplicate declaration of builtin", name)
	}

	decl(BuiltinTypeof, "typeof")
	decl(BuiltinSizeof, "sizeof")
}
