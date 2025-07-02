// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package types

import "cobalt/src"

// Symbol represents a named symbol in a Cobalt program. Along with the name,
// it stores the position, type, scope and more details concerning the symbol.
type Symbol struct {
	name string
	pos  src.Pos
	typ  *Type

	// declaration environment
	// scope *Scope
	// mod   *Module

	// this field stores additional symbol data, depending on the symbol's flags.
	// This list is from highest priority to lowest, meaning that the highest set
	// symbol flag controls what is stored in here.
	//  symBuiltin: Builtin
	//  symStatic:  Value
	extra interface{}

	flags uint32
}

const (
	symUsed    = 1 << iota // symbol is in fact used
	symConst               // symbol is immutable after init
	symStatic              // symbol has a static (init) value
	symBuiltin             // symbol is a built-in procedure
)
