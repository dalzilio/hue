// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package formula

// Bool is the type of ternary Boolean values (True, False, Undef).
type Bool int

const (
	UNDEF Bool = iota
	FALSE
	TRUE
)

func (b Bool) String() string {
	switch b {
	case FALSE:
		return "FALSE"
	case TRUE:
		return "TRUE"
	case UNDEF:
		return "UNDEF"
	default:
		panic("wrong Bool value. It should not happen")
	}
}

func NewBool(b bool) Bool {
	if b {
		return TRUE
	}
	return FALSE
}

// Value returns the Booolean value and sets ok to true if b is not UNDEF.
func (b Bool) Value() (bool, bool) {
	switch b {
	case FALSE:
		return false, true
	case TRUE:
		return true, true
	case UNDEF:
		return true, false
	default:
		panic("wrong Bool value. It should not happen")
	}
}

// Negation
func BoolNot(b Bool) Bool {
	switch b {
	case FALSE:
		return TRUE
	case TRUE:
		return FALSE
	case UNDEF:
		return UNDEF
	default:
		panic("wrong Bool value. It should not happen")
	}
}

func BoolOr(b1, b2 Bool) Bool {
	v1, ok1 := b1.Value()
	if ok1 {
		if v1 {
			// (TRUE || b) is always TRUE
			return TRUE
		}
		// (FALSE || b) is b
		return b2
	}
	if v2, ok2 := b2.Value(); v2 && ok2 {
		// (UNDEF || TRUE) is TRUE
		return TRUE
	}
	return UNDEF
}

func BoolAnd(b1, b2 Bool) Bool {
	v1, ok1 := b1.Value()
	if ok1 {
		if !v1 {
			// (FALSE && b) is always FALSE
			return FALSE
		}
		// (TRUE && b) is b
		return b2
	}
	if v2, ok2 := b2.Value(); !v2 && ok2 {
		// (UNDEF && FALSE) is FALSE
		return FALSE
	}
	return UNDEF
}

func BoolIff(b1, b2 Bool) Bool {
	return BoolOr(BoolAnd(b1, b2), BoolAnd(BoolNot(b1), BoolNot(b2)))
}

// BoolCompatible checks that b1 and b2 are equal when they are both defined
// (not UNDEF)
func BoolCompatible(b1, b2 Bool) bool {
	v1, ok1 := b1.Value()
	if !ok1 {
		// UNDEF compat b
		return true
	}
	v2, ok2 := b2.Value()
	if !ok2 {
		return true
	}
	if v1 {
		return v2
	}
	return !v2
}

// FoldOr returns the disjunction of all the values computed from bs
func FoldOr[T any](f func(T) Bool, bs []T) Bool {
	res := FALSE
	for _, b := range bs {
		res = BoolOr(res, f(b))
		if v, ok := res.Value(); v && ok {
			return TRUE
		}
	}
	return res
}

// FoldAnd returns the conjunction of all the values computed from bs
func FoldAnd[T any](f func(T) Bool, bs []T) Bool {
	res := TRUE
	for _, b := range bs {
		res = BoolAnd(res, f(b))
		if v, ok := res.Value(); !v && ok {
			return FALSE
		}
	}
	return res
}

// If calls the function f with the value if the value is present.
func (b Bool) If(f func(bool)) {
	if v, ok := b.Value(); ok {
		f(v)
	}
}
