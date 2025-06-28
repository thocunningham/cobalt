// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

// This file implements a bail-out mechanism to allow the program to safely
// and easily "escape" from a deeply nested section of code without any
// complicated control flow. This mechanism works by utilizing Go's built-in
// panic() and recover() functions to safely return to a caller function.

package base

import "fmt"

type bailout struct{ payload any }

// When a bail-out is not caught, the program will most likely panic.
// To provide some level of debugging information we add a String method
// so that an appropriate message is printed when the program panics.
func (bail bailout) String() string {
	return fmt.Sprintf("uncaught call to base.Bailout(%#v)", bail.payload)
}

// Bailout bails the program out of a nested section of code. Useful for
// scenarios where a chain of return statements is not a good option.
// A call to Bailout must always be accompanied by a deferred call to
// CatchBailout in any of the caller's callers. If not, the program will panic.
// Bailout allows a payload to be sent back to the catching code to handle.
//
// Example:
//
//	func main() {
//	  defer base.CatchBailout(func(payload any) {
//	    fmt.Fprintln(os.Stderr, payload)
//	    os.Exit(1)
//	  })
//	  base.Bailout("example message")
//	}
func Bailout(payload any) {
	panic(bailout{payload})
}

// CatchBailout catches a bail-out and restores regular execution for the
// caller's caller. A call to CatchBailout must therefore be deferred to have
// any effect.
//
// A handler may be provided to accept any incoming bail-out payload sent
// via a call to Bailout.
//
// Example:
//
//	defer CatchBailout(func(payload any) {
//	  fmt.Fprintln(os.Stderr, payload)
//	  os.Exit(1)
//	})
func CatchBailout(handler func(payload any)) {
	if e := recover(); e != nil {
		if bail, ok := e.(bailout); ok {
			if handler != nil {
				handler(bail.payload)
			}
			return
		}
		panic(e) // not a bail-out
	}
}
