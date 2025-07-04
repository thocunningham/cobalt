// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package types

import "cobalt/src"

// Scope maintains a nested collection of symbols.
type Scope struct {
	parent   *Scope
	elems    map[string]*Symbol
	pos, end src.Pos
}

func NewScope(parent *Scope, pos, end src.Pos) *Scope {
	return &Scope{parent, nil, pos, end}
}

func (s *Scope) Parent() *Scope { return s.parent }
func (s *Scope) Pos() src.Pos   { return s.pos }
func (s *Scope) End() src.Pos   { return s.end }
func (s *Scope) Len() int       { return len(s.elems) }

func (s *Scope) Lookup(name string) *Symbol {
	return s.elems[name]
}

func (s *Scope) LookupParent(name string) (*Scope, *Symbol) {
	for ; s != nil; s = s.parent {
		if sym := s.Lookup(name); sym != nil {
			return s, sym
		}
	}
	return nil, nil
}

func (s *Scope) Insert(sym *Symbol) (alt *Symbol) {
	if alt = s.Lookup(sym.name); alt == nil {
		m := s.elems
		if m == nil {
			m = make(map[string]*Symbol)
			s.elems = m
		}
		m[sym.name] = sym
		if sym.scope == nil {
			sym.scope = s
		}
	}
	return
}

func (s *Scope) Contains(pos src.Pos) bool {
	return s.pos.Known() && s.end.Known() && !pos.Before(s.pos) && !pos.After(s.end)
}
