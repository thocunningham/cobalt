// Copyright (c) 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is derived from the source code of the Go compile
// tool ("src/cmd/compile/internal/syntax/source.go").
//
// There have been made slight changes to incorporate the use of the bail-out
// mechanism implemented in package base but for the rest remains untouched.
//
// Original source: https://github.com/golang/go/blob/master/src/cmd/compile/internal/syntax/source.go

// This file implements source, a buffered rune reader
// specialized for scanning source code: Reading
// ASCII characters, maintaining current (line, col)
// position information, and recording of the most
// recently read source segment are highly optimized.

package syntax

import (
	"cobalt/base"
	"cobalt/src"
	"io"
	"unicode/utf8"
)

// The source buffer is accessed using three indices b (begin),
// r (read), and e (end):
//
// - If b >= 0, it points to the beginning of a segment of most
//   recently read characters (typically a literal).
//
// - r points to the byte immediately following the most recently
//   read character ch, which starts at r-chw.
//
// - e points to the byte immediately following the last byte that
//   was read into the buffer.
//
// The buffer content is terminated at buf[e] with the sentinel
// character utf8.RuneSelf. This makes it possible to test for
// the common case of ASCII characters with a single 'if' (see
// nextch method).
//
//                +------ content in use -------+
//                v                             v
// buf [...read...|...segment...|ch|...unread...|s|...free...]
//                ^             ^  ^            ^
//                |             |  |            |
//                b         r-chw  r            e
//
// Invariant: -1 <= b < r <= e < len(buf) && buf[e] == sentinel

type source struct {
	in   io.Reader
	file string

	buf       []byte // source buffer
	ioerr     error  // pending I/O error, or nil
	b, r, e   int    // buffer indices (see comment above)
	line, col uint   // source position of ch (0-based)
	ch        rune   // most recently read character
	chw       int    // width of ch
}

const sentinel = utf8.RuneSelf

func (s *source) init(in io.Reader, file string) {
	s.in = in
	s.file = file

	s.buf = make([]byte, nextSize(0))
	s.buf[0] = sentinel
	s.ioerr = nil
	s.b, s.r, s.e = -1, 0, 0
	s.line, s.col = 0, 0
	s.ch = ' '
	s.chw = 0
}

// starting points for line and column numbers
const linebase = 1
const colbase = 1

// pos returns the (line, col) source position of s.ch.
func (s *source) pos() (line, col uint) {
	return linebase + s.line, colbase + s.col
}

// error reports the error msg at source position s.pos().
func (s *source) error(msg string) {
	line, col := s.pos()
	base.Bailout(Error{src.MakePos(s.file, line, col), msg})
}

// start starts a new active source segment (including s.ch).
// As long as stop has not been called, the active segment's
// bytes (excluding s.ch) may be retrieved by calling segment.
func (s *source) start()          { s.b = s.r - s.chw }
func (s *source) stop()           { s.b = -1 }
func (s *source) segment() []byte { return s.buf[s.b : s.r-s.chw] }

func (s *source) nextch() {
redo:
	s.col += uint(s.chw)
	if s.ch == '\n' {
		s.line++
		s.col = 0
	}

	// fast common case: at least one ASCII character
	if s.ch = rune(s.buf[s.r]); s.ch < sentinel {
		s.r++
		s.chw = 1
		if s.ch == 0 {
			s.error("invalid NUL character")
			goto redo
		}
		return
	}

	// slower general case: add more bytes to buffer if we don't have a full rune
	for s.e-s.r < utf8.UTFMax && !utf8.FullRune(s.buf[s.r:s.e]) && s.ioerr == nil {
		s.fill()
	}

	// EOF
	if s.r == s.e {
		if s.ioerr != io.EOF {
			// ensure we never start with a '/' (e.g., rooted path) in the error message
			s.error("I/O error: " + s.ioerr.Error())
			s.ioerr = nil
		}
		s.ch = -1
		s.chw = 0
		return
	}

	s.ch, s.chw = utf8.DecodeRune(s.buf[s.r:s.e])
	s.r += s.chw

	if s.ch == utf8.RuneError && s.chw == 1 {
		s.error("invalid UTF-8 encoding")
		goto redo
	}

	// BOM's are only allowed as the first character in a file
	const BOM = 0xfeff
	if s.ch == BOM {
		if s.line > 0 || s.col > 0 {
			s.error("invalid BOM in the middle of the file")
		}
		goto redo
	}
}

// fill reads more source bytes into s.buf.
// It returns with at least one more byte in the buffer, or with s.ioerr != nil.
func (s *source) fill() {
	// determine content to preserve
	b := s.r
	if s.b >= 0 {
		b = s.b
		s.b = 0 // after buffer has grown or content has been moved down
	}
	content := s.buf[b:s.e]

	// grow buffer or move content down
	if len(content)*2 > len(s.buf) {
		s.buf = make([]byte, nextSize(len(s.buf)))
		copy(s.buf, content)
	} else if b > 0 {
		copy(s.buf, content)
	}
	s.r -= b
	s.e -= b

	// read more data: try a limited number of times
	for range 10 {
		var n int
		n, s.ioerr = s.in.Read(s.buf[s.e : len(s.buf)-1]) // -1 to leave space for sentinel
		if n < 0 {
			panic("negative read") // incorrect underlying io.Reader implementation
		}
		if n > 0 || s.ioerr != nil {
			s.e += n
			s.buf[s.e] = sentinel
			return
		}
		// n == 0
	}

	s.buf[s.e] = sentinel
	s.ioerr = io.ErrNoProgress
}

// nextSize returns the next bigger size for a buffer of a given size.
func nextSize(size int) int {
	const min = 4 << 10 // 4K: minimum buffer size
	const max = 1 << 20 // 1M: maximum buffer size which is still doubled
	if size < min {
		return min
	}
	if size <= max {
		return size << 1
	}
	return size + max
}
