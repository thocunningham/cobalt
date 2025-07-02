// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package debug

import (
	"cobalt/base"
	"fmt"
)

// Assert checks the given condition and raises an internal error if the
// condition evaluates to false. Arguments may be provided for additional
// context in the error message.
func Assert(cond bool, args ...any) {
	if Enabled && !cond {
		assertfail(args) // outlined slow path
	}
}

func assertfail(args []any) {
	msg := "assertion failed"
	if len(args) > 0 {
		msg += ": " + fmt.Sprint(args...)
	}
	base.Fatalf(msg)
}
