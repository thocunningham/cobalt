// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package types

import (
	"cobalt/base"
	"cobalt/src"
)

//go:generate stringer -type Kind -trimprefix T type.go

// Kind describes the kind of a type.
type Kind uint8

const (
	TUNDEF Kind = iota

	TTYPE

	TVOID
	TBOOL
	TINT8
	TINT16
	TINT32
	TINT64
	TINTPTR
	TUINT8
	TUINT16
	TUINT32
	TUINT64
	TUINTPTR
	TFLOAT32
	TFLOAT64

	NBASIC

	TPOINTER
	TOPTION
	TARRAY
	TPROC
	TSTRUCT

	NTYPES
)

func (k Kind) IsBasic() bool    { return k != TUNDEF && k < NBASIC }
func (k Kind) IsCompound() bool { return k > NBASIC && k < NTYPES }
func (k Kind) IsSigned() bool   { return k >= TINT8 && k <= TINTPTR }
func (k Kind) IsUnsigned() bool { return k >= TUINT8 && k <= TUINTPTR }
func (k Kind) IsIntegral() bool { return k >= TINT8 && k <= TUINTPTR }
func (k Kind) IsFloat() bool    { return k == TFLOAT32 || k == TFLOAT64 }
func (k Kind) IsNumeric() bool  { return k >= TINT8 && k <= TFLOAT64 }

// Type represents a Cobalt type, which describes the set of permitted values
// and the in-memory representation of the type.
type Type struct {
	// this field contains additional, kind-specific fields for compound types.
	//  TPOINTER: *Pointer
	//  TOPTION: *Option
	//  TARRAY: *Array
	//  TPROC: *Signature
	//  TSTRUCT: *Struct
	extra any

	// only valid once align > 0
	width uint32
	align uint8

	kind Kind

	// if this type is a named type, decl points to the symbol declaring
	// this type. If so, decl.typ.Kind == TTYPE.
	sym *Symbol
}

// Kind returns the kind of t.
func (t *Type) Kind() Kind { return t.kind }

// Sym returns the symbol declaring t, if any.
func (t *Type) Sym() *Symbol { return t.sym }

// Pos returns the associated position with t, if any.
// This is the position where the type is declared.
func (t *Type) Pos() src.Pos {
	if t.sym != nil {
		return t.sym.pos
	}
	return src.NoPos
}

// Elem returns the element type for t, if possible.
// It returns a non-nil *Type for kinds TPOINTER, TOPTION or TARRAY.
func (t *Type) Elem() *Type {
	switch t.kind {
	case TPOINTER:
		return t.extra.(*Pointer).Elem
	case TOPTION:
		return t.extra.(*Option).Elem
	case TARRAY:
		return t.extra.(*Array).Elem
	}
	return nil
}

// Pointer contains additional Type fields for pointer types.
type Pointer struct {
	Elem  *Type
	Const bool
}

// Option contains additional Type fields for option types.
type Option struct {
	Elem  *Type
	Under *Type // underlying structure
}

// Array contains additional Type fields for array types.
type Array struct {
	Elem   *Type
	Length int32 // < 0 if unknown yet
}

// Signature contains additional Type fields for procedure types.
type Signature struct {
	Params []*Field
	Result *Type
}

// Struct contains additional Type fields for struct types.
type Struct struct {
	Fields []*Field
}

// Field is a field in a struct or a procedure parameter.
type Field struct {
	Name  string
	Type  *Type
	Const bool
}

func NewPointer(elem *Type, const_ bool) *Type {
	return &Type{
		extra: &Pointer{elem, const_},
		kind:  TPOINTER,
	}
}

func NewOption(elem *Type) *Type {
	return &Type{
		extra: &Option{elem, nil},
		kind:  TOPTION,
	}
}

func NewArray(elem *Type, length int32) *Type {
	if length < 0 {
		base.Fatalf("types: invalid array length %d", length)
	}
	return &Type{
		extra: &Array{elem, length},
		kind:  TARRAY,
	}
}

func NewSignature(params []*Field, result *Type) *Type {
	return &Type{
		extra: &Signature{params, result},
		kind:  TPROC,
	}
}

func NewStruct(fields []*Field) *Type {
	return &Type{
		extra: &Struct{fields},
		kind:  TSTRUCT,
	}
}
