// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import (
	"github.com/dalzilio/hue/pkg/formula"
)

// EvaluateQueries reports whether a query is true, false, or still undefined
func (s *Stepper) EvaluateQueries(q formula.Query) formula.Bool {
	switch f := q.Formula.(type) {
	case formula.BooleanConstant:
		return formula.NewBool(bool(f))
	default:
		v, ok := s.HasReached(f).Value()
		if ok && q.IsEF && v {
			return formula.TRUE
		}
		if ok && !q.IsEF && !v {
			return formula.FALSE
		}
		return formula.UNDEF
	}
}

// EvaluateAndTestSimplify checks whether the formula in a query evaluates to
// the same result than its simplification on marking m.
func (s *Stepper) EvaluateAndTestSimplify(q formula.Query) bool {
	return formula.BoolCompatible(s.HasReached(q.Original), s.HasReached(q.Formula))
}

// HasReached reports if formula f is true for the current marking m
func (s *Stepper) HasReached(f formula.Formula) formula.Bool {
	switch f := f.(type) {
	case formula.BooleanConstant:
		return formula.NewBool(bool(f))
	case formula.Negation:
		return formula.BoolNot(s.HasReached(f.Formula))
	case formula.Disjunction:
		return formula.FoldOr(func(x formula.Formula) formula.Bool { return s.HasReached(x) }, f)
	case formula.Conjunction:
		return formula.FoldAnd(func(x formula.Formula) formula.Bool { return s.HasReached(x) }, f)
	case formula.IntegerLe:
		return formula.NewBool(s.ComputeIntegerConstant(f.Left) <= s.ComputeIntegerConstant(f.Right))
	case formula.IsFireable:
		return formula.FoldOr(func(x string) formula.Bool { return s.isfireable(x) }, f)
	default:
		panic("wrong formula type in Reached")
	}
}

// ComputeIntegerConstant returns the value of an integer formula
func (s *Stepper) ComputeIntegerConstant(f formula.Formula) int {
	switch f := f.(type) {
	case formula.IntegerConstant:
		return int(f)
	case formula.TokensCount:
		res := 0
		m := s.State
		for _, pname := range f {
			res += m.PT[pname]
		}
		return res
	default:
		panic("wrong formula type in Compute")
	}
}

func (s *Stepper) isfireable(name string) formula.Bool {
	if _, ok := s.forbidden[name]; ok {
		return formula.UNDEF
	}
	b, ok := s.Enabled[name]
	if !ok {
		return formula.UNDEF
	}
	return formula.NewBool(b)
}
