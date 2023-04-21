// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package formula

import (
	"github.com/dalzilio/rudd"
)

// IsTrivial returns true if the query is a Boolean constant.
func (q Query) IsTrivial() bool {
	switch q.Formula.(type) {
	case BooleanConstant:
		return true
	default:
		return false
	}
}

// Simplify performs basic simplifications on formulas.
func Simplify(f Formula) Formula {
	switch f := f.(type) {
	case Negation:
		sf := Simplify(f.Formula)
		switch sf := sf.(type) {
		case Negation:
			//  Simplify(! ! f) = Simplify(f)
			return sf.Formula
		case BooleanConstant:
			// Simplify(! b) = !b
			return BooleanConstant(!sf)
		default:
			//  Simplify(! f) = ! Simplify(f)
			return Negation{sf}
		}
	case Disjunction:
		sfa := []Formula{}
		for _, v := range f {
			sv := Simplify(v)
			switch sv := sv.(type) {
			case BooleanConstant:
				//  Simplify(True | f) = True
				if sv {
					return BooleanConstant(true)
				}
				//  Simplify(False | f) = Simplify(f)
				continue
			default:
				//  Simplify(f1 | f2) = Simplify(f1) | Simplify(f2)
				sfa = append(sfa, sv)
			}
		}
		if len(sfa) == 0 {
			return BooleanConstant(false)
		}
		if len(sfa) == 1 {
			return sfa[0]
		}
		return Disjunction(sfa)
	case Conjunction:
		sfa := []Formula{}
		for _, v := range f {
			sv := Simplify(v)
			switch sv := sv.(type) {
			case BooleanConstant:
				//  Simplify(False & f) = True
				if !sv {
					return BooleanConstant(false)
				}
				//  Simplify(True & f) = Simplify(f)
				continue
			default:
				//  Simplify(f1 & f2) = Simplify(f1) & Simplify(f2)
				sfa = append(sfa, sv)
			}
		}
		if len(sfa) == 0 {
			return BooleanConstant(true)
		}
		if len(sfa) == 1 {
			return sfa[0]
		}
		return Conjunction(sfa)
	case IntegerLe:
		sfl := Simplify(f.Left)
		sfr := Simplify(f.Right)
		switch sfl := sfl.(type) {
		case IntegerConstant:
			//  Simplify(0 <= f) = True
			if sfl == 0 {
				return BooleanConstant(true)
			}
		case TokensCount:
			switch sfr := sfr.(type) {
			case TokensCount:
				//  Simplify(a <= a) = True
				if stringSlicesEqual(sfl, sfr) {
					return BooleanConstant(true)
				}
			default:
				break
			}
		}
		return IntegerLe{Left: sfl, Right: sfr}
	default:
		return f
	}
}

func collectAtoms(tr map[string]int, idx int, f Formula) int {
	// we update tr with new identifiers in f using idx as index of new vars
	switch f := f.(type) {
	case BooleanConstant:
		return idx
	case Negation:
		return collectAtoms(tr, idx, f.Formula)
	case Disjunction:
		for _, v := range f {
			idx = collectAtoms(tr, idx, v)
		}
		return idx
	case Conjunction:
		for _, v := range f {
			idx = collectAtoms(tr, idx, v)
		}
		return idx
	case IsFireable:
		for _, tname := range f {
			if _, ok := tr[tname]; !ok {
				tr[tname] = idx
				idx++
			}
		}
		return idx
	case IntegerLe:
		panic("IntegerLe in collectAtoms")
	}
	panic("unknown formula in collectAtoms")
}

// BddFireabilitySimplify generates a BDD from a fireability formula, f.
func BddFireabilitySimplify(f Formula) (Formula, error) {
	// We build a map between IsFireable constants and integer
	tr := make(map[string]int)
	collectAtoms(tr, 0, f)
	// We build a BDD and use it to simplify the formula
	bdd, err := rudd.New(len(tr), rudd.Nodesize(10000), rudd.Cachesize(5000))
	if err != nil {
		return nil, err
	}
	n := bddEncode(bdd, f, tr)
	if *n == 0 {
		return BooleanConstant(false), nil
	}
	if *n == 1 {
		return BooleanConstant(true), nil
	}
	return Simplify(f), nil
}

// func bddDecode(bdd *rudd.BDD, n rudd.Node, tr map[string]int) Formula {
// 	// we build a list between transition (position) and the related IsFireable formula
// 	trf := make([]Formula, len(tr))
// 	for k, v := range tr {
// 		trf[v] = IsFireable([]string{k})
// 	}
// 	// we build a map of all the ids accessible from node n. We associate to
// 	// each id its level, and the low and high successors. The id 0 and 1 are
// 	// reserved for Boolean constants.
// 	tab := make(map[int][3]int)
// 	bdd.Allnodes(func(id, level, low, high int) error {
// 		if id > 1 {
// 			tab[id] = [3]int{level, low, high}
// 		}
// 		return nil
// 	}, n)

// 	return tab2formula(tab, trf, int(*n))
// }

// func tab2formula(tab map[int][3]int, trf []Formula, n int) Formula {
// 	if n == 0 {
// 		return BooleanConstant(false)
// 	}
// 	if n == 1 {
// 		return BooleanConstant(true)
// 	}

// 	res, ok := tab[n]
// 	if !ok {
// 		panic("error while generating formula from BDD")
// 	}

// 	varf := trf[res[0]]

// 	return ITE{
// 		Condition: varf,
// 		Then:      tab2formula(tab, trf, res[2]),
// 		Else:      tab2formula(tab, trf, res[1]),
// 	}

// 	// if res[1] == 0 {
// 	// 	if res[2] == 1 {
// 	// 		// low: FALSE high: TRUE
// 	// 		return varf
// 	// 	}
// 	// 	return Conjunction{varf, tab2formula(tab, trf, res[2])}
// 	// }

// 	// if res[1] == 1 {
// 	// 	if res[2] == 0 {
// 	// 		// low: TRUE high: FALSE
// 	// 		return Negation{varf}
// 	// 	}
// 	// 	return Disjunction{Negation{varf}, Conjunction{varf, tab2formula(tab, trf, res[2])}}
// 	// }

// 	// if res[2] == 0 {
// 	// 	return Conjunction{Negation{varf}, tab2formula(tab, trf, res[1])}
// 	// }

// 	// if res[2] == 1 {
// 	// 	return Disjunction{varf, Conjunction{Negation{varf}, tab2formula(tab, trf, res[1])}}
// 	// }

// 	// return Disjunction{
// 	// 	Conjunction{varf, tab2formula(tab, trf, res[2])},
// 	// 	Conjunction{Negation{varf}, tab2formula(tab, trf, res[1])},
// 	// }
// }

func bddEncode(bdd *rudd.BDD, f Formula, tr map[string]int) rudd.Node {
	switch f := f.(type) {
	case BooleanConstant:
		return bdd.From(bool(f))
	case Negation:
		return bdd.Not(bddEncode(bdd, f.Formula, tr))
	case Disjunction:
		na := bdd.False()
		for _, v := range f {
			na = bdd.Or(na, bddEncode(bdd, v, tr))
		}
		return na
	case Conjunction:
		na := bdd.True()
		for _, v := range f {
			na = bdd.And(na, bddEncode(bdd, v, tr))
		}
		return na
	case IsFireable:
		na := bdd.False()
		for _, tname := range f {
			na = bdd.Or(na, bdd.Ithvar(tr[tname]))
		}
		return na
	}
	panic("unexpected formula element")
}

func stringSlicesEqual(a1, a2 []string) bool {
	if len(a1) != len(a2) {
		return false
	}
	if a1 == nil {
		return a2 == nil
	}
	if a2 == nil {
		return false
	}
	for k := range a1 {
		if a1[k] != a2[k] {
			return false
		}
	}
	return true
}

// Size returns the number of operators and atomic propositions in f
func Size(f Formula) int {
	acc := new(int)
	accSize(f, acc)
	return *acc
}

func accSize(f Formula, acc *int) {
	switch f := f.(type) {
	case Negation:
		accSize(f.Formula, acc)
	case Disjunction:
		for _, v := range f {
			accSize(v, acc)
		}
	case Conjunction:
		for _, v := range f {
			accSize(v, acc)
		}
	}
	*acc = *acc + 1
}
