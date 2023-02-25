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
func (w *Worker) EvaluateCardinalityQueries(q formula.Query) formula.Bool {
	if b, ok := evaluateQ(w.PT, w.Enabled, q).Value(); ok {
		return formula.From(b)
	}
	for k, m := range w.After {
		if w.Enabled[w.Trans[k].Name].IsTrue() {
			if _, ok := w.forbidFiring[k]; !ok {
				pt := make(map[string]int)
				for k, v := range w.Places {
					pt[v.Name] = m[k].Sum()
				}
				v := evaluateQ(pt, nil, q)
				if _, ok := v.Value(); ok {
					return v
				}
			}
		}
	}
	return formula.UNDEF
}

func (w *Worker) EvaluateFireabilityQueries(q formula.Query) formula.Bool {
	return evaluateQ(nil, w.Enabled, q)
}

// evaluateQ reports whether a query is true, false, or still undefined
func evaluateQ(tokenability map[string]int, fireability map[string]formula.Bool, q formula.Query) formula.Bool {
	switch f := q.Formula.(type) {
	case formula.BooleanConstant:
		return formula.From(bool(f))
	default:
		v, ok := hasReached(tokenability, fireability, f).Value()
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
	return formula.BoolCompatible(hasReached(s.PT, s.Enabled, q.Original),
		hasReached(s.PT, s.Enabled, q.Formula))
}

// hasReached reports if formula f is true for the current marking m
func hasReached(tokenability map[string]int, fireability map[string]formula.Bool, f formula.Formula) formula.Bool {
	switch f := f.(type) {
	case formula.BooleanConstant:
		return formula.From(bool(f))
	case formula.Negation:
		return formula.BoolNot(hasReached(tokenability, fireability, f.Formula))
	case formula.Disjunction:
		return formula.FoldOr(func(x formula.Formula) formula.Bool { return hasReached(tokenability, fireability, x) }, f)
	case formula.Conjunction:
		return formula.FoldAnd(func(x formula.Formula) formula.Bool { return hasReached(tokenability, fireability, x) }, f)
	case formula.IntegerLe:
		return formula.From(ComputeIntegerConstant(tokenability, f.Left) <= ComputeIntegerConstant(tokenability, f.Right))
	case formula.IsFireable:
		return formula.FoldOr(func(x string) formula.Bool { return fireability[x] }, f)
	// case formula.ITE:
	// 	cond := hasReached(tokenability, fireability, f.Condition)
	// 	if b, ok := cond.Value(); ok {
	// 		if b {
	// 			return hasReached(tokenability, fireability, f.Then)
	// 		}
	// 		return hasReached(tokenability, fireability, f.Else)
	// 	}
	// 	left, ok := hasReached(tokenability, fireability, f.Then).Value()
	// 	if ok {
	// 		right, ok := hasReached(tokenability, fireability, f.Else).Value()
	// 		if ok {
	// 			if left == right {
	// 				return formula.From(left)
	// 			}
	// 		}
	// 	}
	// 	return formula.UNDEF
	default:
		panic("wrong formula type in Reached")
	}
}

// ComputeIntegerConstant returns the value of an integer formula
func ComputeIntegerConstant(tokenability map[string]int, f formula.Formula) int {
	switch f := f.(type) {
	case formula.IntegerConstant:
		return int(f)
	case formula.TokensCount:
		res := 0
		for _, pname := range f {
			res += tokenability[pname]
		}
		return res
	case formula.IntegerSum:
		res := 0
		for _, ef := range f {
			res += ComputeIntegerConstant(tokenability, ef)
		}
		return res
	case formula.IntegerDifference:
		if len(f) == 0 {
			return 0
		}
		res := ComputeIntegerConstant(tokenability, f[0])
		if len(f) == 1 {
			return res
		}
		for i := 1; i < len(f); i++ {
			res = res - ComputeIntegerConstant(tokenability, f[i])
		}
		return res
	default:
		panic("wrong formula type in Compute")
	}
}
