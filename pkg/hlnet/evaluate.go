// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import "github.com/dalzilio/hue/pkg/formula"

// Verdict is the possible result of a query evaluation
type Verdict int

const (
	UNDEF Verdict = iota
	TRUE
	FALSE
)

func (v Verdict) String() string {
	switch v {
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	default:
		return "UNDEF"
	}
}

// Evaluate reports whether a query is true, false, or still undefined
func Evaluate(q formula.Query, m Marking) Verdict {
	switch f := q.Formula.(type) {
	case formula.BooleanConstant:
		if f {
			return TRUE
		}
		return FALSE
	default:
		r := Reached(f, m)
		if q.IsEF {
			if r {
				return TRUE
			}
			return UNDEF
		}
		if r {
			return UNDEF
		}
		return FALSE
	}
}

// Reached reports if formula f is true for the current marking m
func Reached(f formula.Formula, m Marking) bool {
	switch f := f.(type) {
	case formula.BooleanConstant:
		return bool(f)
	case formula.Negation:
		return !Reached(f.Formula, m)
	case formula.Disjunction:
		for _, p := range f {
			if Reached(p, m) {
				return true
			}
		}
		return false
	case formula.Conjunction:
		for _, p := range f {
			if !Reached(p, m) {
				return false
			}
		}
		return true
	case formula.IntegerLe:
		return Compute(f.Left, m) <= Compute(f.Right, m)
	case formula.IsFireable:
		for _, tname := range f {
			if m.Enabled[tname] {
				return true
			}
		}
		return false
	default:
		panic("wrong formula type in Reached")
	}
}

// Compute returns the value of an integer formula
func Compute(f formula.Formula, m Marking) int {
	switch f := f.(type) {
	case formula.IntegerConstant:
		return int(f)
	case formula.TokensCount:
		res := 0
		for _, pname := range f {
			res += m.PT[pname]
		}
		return res
	default:
		panic("wrong formula type in Compute")
	}
}
