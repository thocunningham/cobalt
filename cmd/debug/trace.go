// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package debug

import (
	"cobalt/base"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
)

var (
	traceIndent []byte
	traceOutput io.Writer = os.Stdout
	traceLock   sync.Mutex
)

const traceTab = ". "

// Trace performs a function call trace.
//
// Usage pattern:
//
//	defer debug.Trace()()
func Trace() func() {
	if !Enabled {
		return func() {}
	}

	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		base.Fatalf("could not get caller information")
	}

	traceLock.Lock()
	fmt.Fprintf(traceOutput, " %s%s() {\n", traceIndent, runtime.FuncForPC(pc).Name())
	traceIndent = append(traceIndent, traceTab...)
	traceLock.Unlock()

	return untrace
}

func untrace() {
	traceIndent = traceIndent[:len(traceIndent)-len(traceTab)]
	fmt.Fprintf(traceOutput, " %s}\n", traceIndent)
}

// TraceOutput sets the tracing output to the provided writer and returns the
// previous one. If w == nil, then the output writer will remain unchanged.
func TraceOuput(w io.Writer) (old io.Writer) {
	old = traceOutput
	if w != nil {
		traceOutput = w
	}
	return
}
