// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import (
	"github.com/dalzilio/hue/pkg/formula"
)

// EvaluateRCQueries reports whether a ReachabilityCardinality query is true,
// false, or still undefined. The difference with EvaluateQueries is that we can
// evaluate the formula on both the current marking, but also all the ones
// stored in After (list of marking reachable when firing an enabled
// transition).
func (s *Stepper) EvaluateCardinalityQueries(q formula.Query) formula.Bool {
	if b, ok := s.State.evaluateQ(q).Value(); ok {
		return formula.NewBool(b)
	}
	for k, m := range s.After {
		if s.Enabled[s.Trans[k].Name].IsTrue() {
			if _, ok := s.forbidFiring[k]; !ok {
				v := s.newState(m).evaluateQ(q)
				if _, ok := v.Value(); ok {
					return v
				}
			}
		}
	}
	return formula.UNDEF
}

func (s *Stepper) EvaluateFireabilityQueries(q formula.Query) formula.Bool {
	return s.evaluateQ(q)
}

// evaluateQ reports whether a query is true, false, or still undefined
func (s State) evaluateQ(q formula.Query) formula.Bool {
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
func (s *State) EvaluateAndTestSimplify(q formula.Query) bool {
	return formula.BoolCompatible(s.HasReached(q.Original), s.HasReached(q.Formula))
}

// HasReached reports if formula f is true for the current marking m
func (s *State) HasReached(f formula.Formula) formula.Bool {
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
		return formula.FoldOr(func(x string) formula.Bool { return s.Enabled[x] }, f)
	default:
		panic("wrong formula type in Reached")
	}
}

// ComputeIntegerConstant returns the value of an integer formula
func (s *State) ComputeIntegerConstant(f formula.Formula) int {
	switch f := f.(type) {
	case formula.IntegerConstant:
		return int(f)
	case formula.TokensCount:
		res := 0
		for _, pname := range f {
			res += s.PT[pname]
		}
		return res
	default:
		panic("wrong formula type in Compute")
	}
}
