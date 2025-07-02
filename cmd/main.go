// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package main

import (
	"cobalt/base"
	"cobalt/syntax"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: co <file.co>")
		os.Exit(1)
	}

	file, err := syntax.ParseFile(os.Args[1])
	if err != nil {
		base.Errorf("%v", err)
	}

	_ = file
}
