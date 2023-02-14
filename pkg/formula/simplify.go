// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package formula

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
		return Disjunction(sfa)
	case Conjunction:
		sfa := []Formula{}
		for _, v := range f {
			sv := Simplify(v)
			switch sv := sv.(type) {
			case BooleanConstant:
				//  Simplify(False & f) = True
				if sv {
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
		}
		return IntegerLe{Left: sfl, Right: sfr}
	default:
		return f
	}
}