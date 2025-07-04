// Copyright (c) 2025 Thomas Cunningham. All rights reserved.
// Use of this source code is governed by an MIT license that
// can be found in the LICENSE file.

package types

import (
	"cobalt/syntax"
	"math"
	"strconv"
)

// Value is a value that is representable in a Cobalt program. It is to be used
// for representing and evaluating static values. [Undefined] is to be used for
// unknown/undefined values, not nil.
//
// For arithmetic operations, the result's Kind is promoted to a higher precision
// if it does not fit within the original Kind. Signed and unsigned will always
// remain the same, but just with a higher precision. Operations involving an
// integral type with a floating-point type return a floating-point type.
type Value interface {
	Kind() Kind
	String() string

	// Unary performs the unary operation on v with the provided operator. If the
	// value is of an incompatible kind, an undefined values is returned.
	Unary(syntax.Operator) Value

	// Binary performs the binary operation on v and w with the provided operator.
	// If the value is of an incompatible kind, an undefined values is returned.
	// All binary operators can be used, including comparitive operators are
	// supported.
	Binary(syntax.Operator, Value) Value

	// Convert attempts to convert v to the desired Kind. If this is not possible,
	// Undefined is returned.
	Convert(Kind) Value
}

// Undefined is the value to be used to represent undefined values.
var Undefined Value = undefValue{}

// undefValue is an undefined value
type undefValue struct{}

func (undefValue) Kind() Kind                          { return TUNDEF }
func (undefValue) String() string                      { return "<undefined>" }
func (undefValue) Unary(syntax.Operator) Value         { return Undefined }
func (undefValue) Binary(syntax.Operator, Value) Value { return Undefined }
func (undefValue) Convert(Kind) Value                  { return Undefined }

// typeValue is a type as a value
type typeValue struct{ t *Type }

// MakeType returns a Value with the provided type.
// If t is nil, it returns Undefined.
func MakeType(t *Type) Value {
	if t == nil {
		return Undefined
	}
	return typeValue{t}
}

func (typeValue) Kind() Kind                          { return TTYPE }
func (typeValue) String() string                      { return "<type>" } // TODO: implement type strings
func (typeValue) Unary(syntax.Operator) Value         { return Undefined }
func (typeValue) Binary(syntax.Operator, Value) Value { return Undefined }
func (v typeValue) Convert(to Kind) Value {
	if to == v.Kind() {
		return v
	}
	return Undefined
}

// boolValue is a boolean as a value
type boolValue struct{ b bool }

// MakeBool returns a Value with the provided boolean.
func MakeBool(b bool) Value {
	return boolValue{b}
}

func (boolValue) Kind() Kind {
	return TBOOL
}

func (v boolValue) String() string {
	return strconv.FormatBool(v.b)
}

func (v boolValue) Unary(op syntax.Operator) Value {
	if op == syntax.LNot {
		return MakeBool(!v.b)
	}
	return Undefined
}

func (v boolValue) Binary(op syntax.Operator, w Value) Value {
	if w, ok := w.(boolValue); ok {
		switch op {
		case syntax.OrOr:
			return MakeBool(v.b || w.b)

		case syntax.AndAnd:
			return MakeBool(v.b && w.b)

		case syntax.Eql:
			return MakeBool(v.b == w.b)

		case syntax.Neq:
			return MakeBool(v.b != w.b)
		}
	}

	return Undefined
}

func (v boolValue) Convert(to Kind) Value {
	if to == v.Kind() {
		return v
	}
	return Undefined
}

// intValue is a signed integral value
type intValue struct {
	x    int64
	bits int // 8, 16, 32 or 64
}

// MakeInt returns a signed integer Value with the provided integer.
//
// It defaults to a 32-bit integer, but uses a 64-bit integer if x does not fit
// in just 32 bits.
func MakeInt(x int64) Value {
	if x < -1<<31 || x > 1<<31-1 {
		return intValue{x, 64}
	} else {
		return intValue{x, 32}
	}
}

func (v intValue) Kind() Kind {
	switch v.bits {
	case 8:
		return TINT8
	case 16:
		return TINT16
	case 32:
		return TINT32
	case 64:
		return TINT64
	}
	panic("unreachable")
}

func (v intValue) String() string {
	return strconv.FormatInt(v.x, 10)
}

func (v intValue) Unary(op syntax.Operator) Value {
	switch op {
	case syntax.Not: // ~v
		v.x = ^v.x
	case syntax.Inc: // ++v or v++
		v.x += 1
	case syntax.Dec: // --v or v--
		v.x -= 1
	case syntax.Add: // +v
		// no-op
	case syntax.Sub: // -v
		v.x = -v.x
	}

	return MakeInt(v.x)
}

func (v intValue) Binary(op syntax.Operator, w Value) Value {
	switch op {
	case syntax.Eql:
		switch w := w.(type) {
		case intValue:
			return MakeBool(v.x == w.x)
		case uintValue:
			return MakeBool(v.x >= 0 && uint64(v.x) == w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanInt64(w.x) {
				return MakeBool(v.x == int64(w.x))
			}
			return MakeBool(float64(v.x) == w.x)
		}

	case syntax.Neq:
		switch w := w.(type) {
		case intValue:
			return MakeBool(v.x != w.x)
		case uintValue:
			return MakeBool(v.x < 0 || uint64(v.x) != w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanInt64(w.x) {
				return MakeBool(v.x != int64(w.x))
			}
			return MakeBool(float64(v.x) != w.x)
		}

	case syntax.Lss:
		switch w := w.(type) {
		case intValue:
			return MakeBool(v.x < w.x)
		case uintValue:
			return MakeBool(v.x < 0 || uint64(v.x) < w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanInt64(w.x) {
				return MakeBool(v.x < int64(w.x))
			}
			return MakeBool(float64(v.x) < w.x)
		}

	case syntax.Leq:
		switch w := w.(type) {
		case intValue:
			return MakeBool(v.x <= w.x)
		case uintValue:
			return MakeBool(v.x < 0 && uint64(v.x) <= w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanInt64(w.x) {
				return MakeBool(v.x <= int64(w.x))
			}
			return MakeBool(float64(v.x) <= w.x)
		}

	case syntax.Gtr:
		switch w := w.(type) {
		case intValue:
			return MakeBool(v.x > w.x)
		case uintValue:
			return MakeBool(v.x >= 0 && uint64(v.x) > w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanInt64(w.x) {
				return MakeBool(v.x > int64(w.x))
			}
			return MakeBool(float64(v.x) > w.x)
		}

	case syntax.Geq:
		switch w := w.(type) {
		case intValue:
			return MakeBool(v.x >= w.x)
		case uintValue:
			return MakeBool(v.x >= 0 && uint64(v.x) >= w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanInt64(w.x) {
				return MakeBool(v.x >= int64(w.x))
			}
			return MakeBool(float64(v.x) >= w.x)
		}

	case syntax.Add:
		switch w := w.(type) {
		case intValue:
			return MakeInt(v.x + w.x)
		case uintValue:
			return MakeInt(v.x + int64(w.x))
		case floatValue:
			return MakeFloat(float64(v.x) + w.x)
		}

	case syntax.Sub:
		switch w := w.(type) {
		case intValue:
			return MakeInt(v.x - w.x)
		case uintValue:
			return MakeInt(v.x - int64(w.x))
		case floatValue:
			return MakeFloat(float64(v.x) - w.x)
		}

	case syntax.Or:
		switch w := w.(type) {
		case intValue:
			return MakeInt(v.x | w.x)
		case uintValue:
			return MakeInt(v.x | int64(w.x))
		}

	case syntax.Xor:
		switch w := w.(type) {
		case intValue:
			return MakeInt(v.x ^ int64(w.x))
		case uintValue:
			return MakeInt(v.x ^ int64(w.x))
		}

	case syntax.Mul:
		switch w := w.(type) {
		case intValue:
			return MakeInt(v.x * int64(w.x))
		case uintValue:
			return MakeInt(v.x * int64(w.x))
		case floatValue:
			return MakeFloat(float64(v.x) * w.x)
		}

	case syntax.Div:
		switch w := w.(type) {
		case intValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeInt(v.x / w.x)
		case uintValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeInt(v.x / int64(w.x))
		case floatValue:
			if w.x == 0.0 {
				return Undefined
			}
			return MakeFloat(float64(v.x) / w.x)
		}

	case syntax.Rem:
		switch w := w.(type) {
		case intValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeInt(v.x % w.x)
		case uintValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeInt(v.x % int64(w.x))
		}

	case syntax.And:
		switch w := w.(type) {
		case intValue:
			return MakeInt(v.x & w.x)
		case uintValue:
			return MakeInt(v.x & int64(w.x))
		}

	case syntax.Shl:
		switch w := w.(type) {
		case intValue:
			if w.x < 0 {
				return Undefined
			}
			return MakeInt(v.x << w.x)
		case uintValue:
			return MakeInt(v.x << w.x)
		}

	case syntax.Shr:
		switch w := w.(type) {
		case intValue:
			if w.x < 0 {
				return Undefined
			}
			return MakeInt(v.x >> w.x)
		case uintValue:
			return MakeInt(v.x >> w.x)
		}
	}

	return Undefined
}

func (v intValue) Convert(to Kind) Value {
	if to == v.Kind() {
		return v
	}

	if to.IsSigned() {
		if n := kindbits(to); n > v.bits {
			return intValue{sext(v.x, v.bits), n}
		} else {
			return intValue{sext(v.x, n), n}
		}
	}

	if to.IsUnsigned() {
		if n := kindbits(to); n > v.bits {
			return uintValue{uint64(sext(v.x, v.bits)), n}
		} else {
			return uintValue{zext(uint64(v.x), n), n}
		}
	}

	if to.IsFloat() {
		if n := kindbits(to); n == 32 {
			return floatValue{float64(float32(v.x)), n}
		} else {
			return floatValue{float64(v.x), n}
		}
	}

	return Undefined
}

// uintValue is an unsigned integral value
type uintValue struct {
	x    uint64
	bits int // 8, 16, 32 or 64
}

// MakeInt returns a unsigned integer Value with the provided integer.
//
// It defaults to a 32-bit integer, but uses a 64-bit integer if x does not fit
// in just 32 bits.
func MakeUint(x uint64) Value {
	if x > 1<<32-1 {
		return uintValue{x, 64}
	} else {
		return uintValue{x, 32}
	}
}

func (v uintValue) Kind() Kind {
	switch v.bits {
	case 8:
		return TUINT8
	case 16:
		return TUINT16
	case 32:
		return TUINT32
	case 64:
		return TUINT64
	}
	panic("unreachable")
}

func (v uintValue) String() string {
	return strconv.FormatUint(v.x, 10)
}

func (v uintValue) Unary(op syntax.Operator) Value {
	switch op {
	case syntax.Not: // ~v
		v.x = ^v.x
	case syntax.Inc: // ++v or v++
		v.x += 1
	case syntax.Dec: // --v or v--
		v.x -= 1
	case syntax.Add: // +v
		// no-op
	case syntax.Sub: // -v
		v.x = -v.x
	}

	return MakeUint(v.x)
}

func (v uintValue) Binary(op syntax.Operator, w Value) Value {
	switch op {
	case syntax.Eql:
		switch w := w.(type) {
		case intValue: // unsigned == signed
			return MakeBool(w.x >= 0 && v.x == uint64(w.x))
		case uintValue:
			return MakeBool(v.x == w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanUint64(w.x) {
				return MakeBool(v.x == uint64(w.x))
			}
			return MakeBool(float64(v.x) == w.x)
		}

	case syntax.Neq:
		switch w := w.(type) {
		case intValue: // unsigned != signed
			return MakeBool(w.x < 0 || v.x != uint64(w.x))
		case uintValue:
			return MakeBool(v.x == w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanUint64(w.x) {
				return MakeBool(v.x != uint64(w.x))
			}
			return MakeBool(float64(v.x) != w.x)
		}

	case syntax.Lss:
		switch w := w.(type) {
		case intValue: // unsigned < signed
			return MakeBool(w.x >= 0 && v.x < uint64(w.x))
		case uintValue:
			return MakeBool(v.x < w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanUint64(w.x) {
				return MakeBool(v.x < uint64(w.x))
			}
			return MakeBool(float64(v.x) < w.x)
		}

	case syntax.Leq:
		switch w := w.(type) {
		case intValue: // unsigned <= signed
			return MakeBool(w.x >= 0 && v.x <= uint64(w.x))
		case uintValue:
			return MakeBool(v.x <= w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanUint64(w.x) {
				return MakeBool(v.x <= uint64(w.x))
			}
			return MakeBool(float64(v.x) <= w.x)
		}

	case syntax.Gtr:
		switch w := w.(type) {
		case intValue: // unsigned > signed
			return MakeBool(w.x < 0 || v.x > uint64(w.x))
		case uintValue:
			return MakeBool(v.x > w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanUint64(w.x) {
				return MakeBool(v.x > uint64(w.x))
			}
			return MakeBool(float64(v.x) > w.x)
		}

	case syntax.Geq:
		switch w := w.(type) {
		case intValue: // unsigned >= signed
			return MakeBool(w.x < 0 || v.x >= uint64(w.x))
		case uintValue:
			return MakeBool(v.x >= w.x)
		case floatValue:
			if math.IsInf(w.x, 0) || math.IsNaN(w.x) {
				return MakeBool(false)
			}
			if floatCanUint64(w.x) {
				return MakeBool(v.x >= uint64(w.x))
			}
			return MakeBool(float64(v.x) >= w.x)
		}

	case syntax.Add:
		switch w := w.(type) {
		case intValue:
			return MakeUint(v.x + uint64(w.x))
		case uintValue:
			return MakeUint(v.x + w.x)
		case floatValue:
			return MakeFloat(float64(v.x) + w.x)
		}

	case syntax.Sub:
		switch w := w.(type) {
		case intValue:
			return MakeUint(v.x - uint64(w.x))
		case uintValue:
			return MakeUint(v.x - w.x)
		case floatValue:
			return MakeFloat(float64(v.x) - w.x)
		}

	case syntax.Or:
		switch w := w.(type) {
		case intValue:
			return MakeUint(v.x | uint64(w.x))
		case uintValue:
			return MakeUint(v.x | w.x)
		}

	case syntax.Xor:
		switch w := w.(type) {
		case intValue:
			return MakeUint(v.x ^ uint64(w.x))
		case uintValue:
			return MakeUint(v.x ^ w.x)
		}

	case syntax.Mul:
		switch w := w.(type) {
		case intValue:
			return MakeUint(v.x * uint64(w.x))
		case uintValue:
			return MakeUint(v.x * w.x)
		case floatValue:
			return MakeFloat(float64(v.x) * w.x)
		}

	case syntax.Div:
		switch w := w.(type) {
		case intValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeUint(v.x / uint64(w.x))
		case uintValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeUint(v.x / w.x)
		case floatValue:
			if w.x == 0.0 {
				return Undefined
			}
			return MakeFloat(float64(v.x) / w.x)
		}

	case syntax.Rem:
		switch w := w.(type) {
		case intValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeUint(v.x % uint64(w.x))
		case uintValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeUint(v.x % w.x)
		}

	case syntax.And:
		switch w := w.(type) {
		case intValue:
			return MakeUint(v.x & uint64(w.x))
		case uintValue:
			return MakeUint(v.x & w.x)
		}

	case syntax.Shl:
		switch w := w.(type) {
		case intValue:
			if w.x < 0 {
				return Undefined
			}
			return MakeUint(v.x << w.x)
		case uintValue:
			return MakeUint(v.x << w.x)
		}

	case syntax.Shr:
		switch w := w.(type) {
		case intValue:
			if w.x < 0 {
				return Undefined
			}
			return MakeUint(v.x >> w.x)
		case uintValue:
			return MakeUint(v.x >> w.x)
		}
	}

	return Undefined
}

func (v uintValue) Convert(to Kind) Value {
	if to == v.Kind() {
		return v
	}

	if to.IsSigned() {
		if n := kindbits(to); n > v.bits {
			return intValue{int64(zext(v.x, v.bits)), n}
		} else {
			return intValue{sext(int64(v.x), n), n}
		}
	}

	if to.IsUnsigned() {
		if n := kindbits(to); n > v.bits {
			return uintValue{zext(v.x, v.bits), n}
		} else {
			return uintValue{zext(v.x, n), n}
		}
	}

	if to.IsFloat() {
		if n := kindbits(to); n == 32 {
			return floatValue{float64(float32(v.x)), n}
		} else {
			return floatValue{float64(v.x), n}
		}
	}

	return Undefined
}

// floatValue is a floating-point value
type floatValue struct {
	x    float64
	bits int // 32 or 64
}

// MakeFloat returns a floating-point value with the provided float.
//
// It defaults toa 32-bit float, but uses a 64-bit float if x is not
// representable in just 32 bits.
func MakeFloat(x float64) Value {
	if float64(float32(x)) == x {
		return floatValue{x, 32}
	}
	return floatValue{x, 64}
}

func (v floatValue) Kind() Kind {
	switch v.bits {
	case 32:
		return TFLOAT32
	case 64:
		return TFLOAT64
	}
	panic("unreachable")
}

func (v floatValue) String() string {
	return strconv.FormatFloat(v.x, 'f', -1, v.bits)
}

func (v floatValue) Unary(op syntax.Operator) Value {
	switch op {
	case syntax.Inc:
		v.x += 1
	case syntax.Dec:
		v.x -= 1
	case syntax.Add:
		// no-op
	case syntax.Sub:
		v.x = -v.x
	}

	return MakeFloat(v.x)
}

func (v floatValue) Binary(op syntax.Operator, w Value) Value {
	switch op {
	case syntax.Eql:
		switch w := w.(type) {
		case intValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanInt64(v.x) {
				return MakeBool(int64(v.x) == w.x)
			}
			return MakeBool(v.x == float64(w.x))
		case uintValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanUint64(v.x) {
				return MakeBool(uint64(v.x) == w.x)
			}
			return MakeBool(v.x == float64(w.x))
		case floatValue:
			return MakeBool(v.x == w.x)
		}

	case syntax.Neq:
		switch w := w.(type) {
		case intValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanInt64(v.x) {
				return MakeBool(int64(v.x) != w.x)
			}
			return MakeBool(v.x != float64(w.x))
		case uintValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanUint64(v.x) {
				return MakeBool(uint64(v.x) != w.x)
			}
			return MakeBool(v.x != float64(w.x))
		case floatValue:
			return MakeBool(v.x != w.x)
		}

	case syntax.Lss:
		switch w := w.(type) {
		case intValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanInt64(v.x) {
				return MakeBool(int64(v.x) < w.x)
			}
			return MakeBool(v.x < float64(w.x))
		case uintValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanUint64(v.x) {
				return MakeBool(uint64(v.x) < w.x)
			}
			return MakeBool(v.x < float64(w.x))
		case floatValue:
			return MakeBool(v.x < w.x)
		}

	case syntax.Leq:
		switch w := w.(type) {
		case intValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanInt64(v.x) {
				return MakeBool(int64(v.x) <= w.x)
			}
			return MakeBool(v.x <= float64(w.x))
		case uintValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanUint64(v.x) {
				return MakeBool(uint64(v.x) <= w.x)
			}
			return MakeBool(v.x <= float64(w.x))
		case floatValue:
			return MakeBool(v.x <= w.x)
		}

	case syntax.Gtr:
		switch w := w.(type) {
		case intValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanInt64(v.x) {
				return MakeBool(int64(v.x) > w.x)
			}
			return MakeBool(v.x > float64(w.x))
		case uintValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanUint64(v.x) {
				return MakeBool(uint64(v.x) > w.x)
			}
			return MakeBool(v.x > float64(w.x))
		case floatValue:
			return MakeBool(v.x > w.x)
		}

	case syntax.Geq:
		switch w := w.(type) {
		case intValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanInt64(v.x) {
				return MakeBool(int64(v.x) >= w.x)
			}
			return MakeBool(v.x >= float64(w.x))
		case uintValue:
			if math.IsInf(v.x, 0) || math.IsNaN(v.x) {
				return MakeBool(false)
			}
			if floatCanUint64(v.x) {
				return MakeBool(uint64(v.x) >= w.x)
			}
			return MakeBool(v.x >= float64(w.x))
		case floatValue:
			return MakeBool(v.x >= w.x)
		}

	case syntax.Add:
		switch w := w.(type) {
		case intValue:
			return MakeFloat(v.x + float64(w.x))
		case uintValue:
			return MakeFloat(v.x + float64(w.x))
		case floatValue:
			return MakeFloat(v.x + w.x)
		}

	case syntax.Sub:
		switch w := w.(type) {
		case intValue:
			return MakeFloat(v.x - float64(w.x))
		case uintValue:
			return MakeFloat(v.x - float64(w.x))
		case floatValue:
			return MakeFloat(v.x - w.x)
		}

	case syntax.Mul:
		switch w := w.(type) {
		case intValue:
			return MakeFloat(v.x * float64(w.x))
		case uintValue:
			return MakeFloat(v.x * float64(w.x))
		case floatValue:
			return MakeFloat(v.x * w.x)
		}

	case syntax.Div:
		switch w := w.(type) {
		case intValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeFloat(v.x / float64(w.x))
		case uintValue:
			if w.x == 0 {
				return Undefined
			}
			return MakeFloat(v.x / float64(w.x))
		case floatValue:
			if w.x == 0.0 {
				return Undefined
			}
			return MakeFloat(v.x / w.x)
		}
	}

	return Undefined
}

func (v floatValue) Convert(to Kind) Value {
	if to == v.Kind() {
		return v
	}

	if to.IsSigned() {
		n := kindbits(to)
		return intValue{sext(int64(v.x), n), n}
	}

	if to.IsUnsigned() {
		n := kindbits(to)
		return uintValue{zext(uint64(v.x), n), n}
	}

	if to.IsFloat() {
		if n := kindbits(to); n == 32 {
			return floatValue{float64(float32(v.x)), n}
		} else {
			return v
		}
	}

	return Undefined
}

// ----------------------------------------------------------------------------
// Utilities

func sext(x int64, n int) int64 {
	x &^= int64(-1) << n
	bit := x & (int64(1) << (n - 1))
	mask := (bit << (64 - n)) >> (64 - n)
	return x | mask
}

func zext(x uint64, n int) uint64 {
	mask := ^uint64(0) >> (64 - n)
	return x & mask
}

func kindbits(k Kind) int {
	switch k {
	case TINT8, TUINT8:
		return 8
	case TINT16, TUINT16:
		return 16
	case TINT32, TUINT32, TFLOAT32:
		return 32
	case TINT64, TUINT64, TFLOAT64:
		return 64
	}
	panic("unreachable")
}

func floatCanInt64(f float64) bool {
	return f == math.Trunc(f) &&
		f >= float64(math.MinInt64) &&
		f <= float64(math.MaxInt64)
}

func floatCanUint64(f float64) bool {
	return f == math.Trunc(f) &&
		f >= 0 &&
		f <= float64(math.MaxUint64)
}
