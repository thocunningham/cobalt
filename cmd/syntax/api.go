// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

// Package syntax implements the logic for syntactic analysis of source files.
package syntax

import (
	"cobalt/base"
	"cobalt/src"
	"io"
	"os"
)

// Error describes a syntax error that occurred at any point while scanning or
// parsing the source code. An Error is considered non-nil if it has a known
// position and a non-empty error message.
type Error struct {
	Pos src.Pos
	Msg string
}

func (e Error) Error() string {
	return e.Pos.String() + ": " + e.Msg
}

// Err returns e as an error, following the requirements for e to be
// considered non-nil.
func (e Error) Err() error {
	if e.Pos.Known() && e.Msg != "" {
		return e
	}
	return nil
}

// Parse parses the source code read from an io.Reader and the providded file
// name. If an error occurs during parsing, a nil [File] and a non-nil error is
// returned. This is to limit the chances of being able to type-check a
// malformed syntax tree.
//
// Parse panics if a nil io.Reader is provided.
func Parse(rd io.Reader, name string) (file *File, err error) {
	if rd == nil {
		panic("syntax: nil io.Reader provided")
	}

	defer base.CatchBailout(func(payload any) {
		file, err = nil, payload.(error)
	})

	var p parser
	p.init(rd, name)
	return p.file(), nil
}

// ParseFile is a wrapper for [Parse], using only a file name for parsing, it
// uses the OS's file system to get a reader to parse from.
func ParseFile(name string) (*File, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return Parse(file, name)
}
