// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package base

import (
	"fmt"
	"os"
)

// Exit causes the current program to exit with the given status code.
//
// Use one of the following exit codes:
//   - 0: No errors occurred.
//   - 1: A source code error occurred.
//   - 2: An internal compiler error occurred.
func Exit(code int) {
	os.Exit(code)
}

// Fatalf reports an internal error and exits.
// If Debug is true, Fatalf panics to force a runtime info dump.
func Fatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "internal error: "+format+"\n", a...)
	Exit(2)
}
