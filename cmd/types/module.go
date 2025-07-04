// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package types

import (
	"cobalt/base"
	"cobalt/src"
)

var modmap map[string]*Module

// A Module defines a named scope that groups symbols together.
type Module struct {
	name, path string
	scope      *Scope
}

func NewModule(name, path string) *Module {
	if mod := modmap[path]; mod != nil {
		if name != "" && name != mod.name {
			base.Fatalf("conflicting module names %s and %s for path %q", name, mod.name, path)
		}
		return mod
	}

	mod := new(Module)
	mod.path = path
	mod.name = name
	mod.scope = NewScope(nil, src.NoPos, src.NoPos) // TODO: implement universe scope
	modmap[path] = mod

	return mod
}

func (mod *Module) Name() string                     { return mod.name }
func (mod *Module) Path() string                     { return mod.path }
func (mod *Module) Lookup(name string) *Symbol       { return mod.scope.Lookup(name) }
func (mod *Module) Insert(sym *Symbol) (alt *Symbol) { return mod.scope.Insert(sym) }
