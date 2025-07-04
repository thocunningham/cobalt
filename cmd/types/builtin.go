// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package types

// Builtin is a built-in procedure.
type Builtin uint

const (
	_ Builtin = iota

	BuiltinTypeof
	BuiltinSizeof
)
