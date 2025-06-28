// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

// Package src implements the management and creation of source code positions.
// It maintains a collection of source files and allows for compact encodings
// of source code positions.
package src

import (
	"fmt"
	"sync"
)

// A Pos is an absolute position of a byte a source file. It encodes the file,
// line number, and column number. A zero Pos is a ready-to-use Pos, but is
// considered "unknown". A Pos is considered known once it has an associated
// source file.
//
// A Pos is intentionally lightweight, such that it can be used without any
// concern for memory use.
type Pos struct {
	index uint32
	lico  uint32
}

// NoPos is the zero value for Pos and is to be used for representing invalid
// or absent soure code positions.
var NoPos Pos = Pos{0, 0}

// Known reports whether p is considered a known position.
func (p Pos) Known() bool {
	return p.index != 0
}

// MakePos creates a new Pos value with the provided file name, line-, and
// column numbers. There is a hard limit for line- and column numbers, defined
// by LineMax and ColMax respectively.
func MakePos(filename string, line uint, col uint) Pos {
	return Pos{
		index: insert(filename),
		lico:  lico(line, col),
	}
}

// Before reports whether p appears before q in the source code.
// It also reports false if either p or q are unknown or are from different
// source files.
func (p Pos) Before(q Pos) bool {
	return p.index != 0 && p.index == q.index && p.lico < q.lico
}

// After reports whether p appears after q in the source code.
// It also reports false if either p or q are unknown or are from different
// source files.
func (p Pos) After(q Pos) bool {
	return p.index != 0 && p.index == q.index && p.lico > q.lico
}

// Filename returns the file name for p. If p has no source file, Filename
// returns an empty string.
func (p Pos) Filename() string {
	return lookup(p.index)
}

// Line returns the line number for p. A zero line number indicates an unknown
// or invalid line number.
func (p Pos) Line() uint {
	return uint(p.lico >> colbits)
}

// Col returns the column number for p. A zero column number indicates an
// unknown or invalid column number.
func (p Pos) Col() uint {
	return uint(p.lico & ColMax)
}

// String returns a string representation of p. If p has no associated source
// file, String returns "<unknown position>".
func (p Pos) String() string {
	if p.index == 0 {
		return "<unknown position>"
	}
	if p.Line() == 0 {
		return lookup(p.index) // file
	}
	if p.Col() == 0 {
		return fmt.Sprintf("%s:%d", lookup(p.index), p.Line()) // file:line
	}
	return fmt.Sprintf("%s:%d:%d", lookup(p.index), p.Line(), p.Col()) // file:line:col
}

// ----------------------------------------------------------------------------
// Internal Details

const (
	linebits, colbits = 20, 12 // should add up to 32

	// LineMax is the maximum line number representable by a Pos value.
	// Line numbers greater than LineMax are reset to LineMax.
	LineMax = 1<<linebits - 1

	// ColMax is the maximum column number representable by a Pos value.
	// Column numbers greater than ColMax are reset to ColMax.
	ColMax = 1<<colbits - 1
)

var _ = [1]struct{}{}[linebits+colbits-32] // linebits + colbits == 32

// lico is a line- and column number combination encoded as a single uint32.
// The above constants linebits and colbits define the distribution of the bits.
//
// The line bits are placed at the more signifcant side of the int. This allows
// us to compare a lico with the equality and/or comparison operators (==, !=, <, <=, >, >=).
func lico(line, col uint) uint32 {
	line, col = min(line, LineMax), min(col, ColMax)
	return uint32(line)<<colbits | uint32(col)
}

var (
	namelist = make([]string, 0)       // index -> filename
	indexmap = make(map[string]uint32) // filename -> index
	mu       sync.RWMutex              // protects namelist and indexmap
)

// insert inserts the provided file name into the global file table
// and returns the corresponding index. If the file name is already
// present, it returns the associated index.
func insert(filename string) (index uint32) {
	if filename == "" {
		return 0 // don't insert empty file names
	}

	mu.Lock()
	defer mu.Unlock()

	if index = indexmap[filename]; index == 0 {
		index = uint32(len(indexmap) + 1)
		indexmap[filename] = index
		namelist = append(namelist, filename)
	}

	return
}

// lookup looks up the provided index into the global file table
// and returns the associated string. If the index is not present
// in the table, lookup returns an empty string.
func lookup(index uint32) string {
	index -= 1 // adjust for zero index

	mu.RLock()
	defer mu.RUnlock()
	if index < uint32(len(namelist)) {
		return namelist[index]
	}

	return ""
}
