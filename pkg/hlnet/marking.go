// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dalzilio/hue/pkg/pnml"
)

// Marking is the type of High-Level Nets (hlnet) markings. It is a slice
// containing the PMarking of each places in the same order than in PLaces
type Marking []PMarking

// Atom is a pair of a multiplicity and a colored value.
type Atom struct {
	Value *pnml.Value
	Mult  int
}

// PMarking is the possible marking of a single place. It is a multiset of Atom,
// with the following conventions:
//
// - Items are of the form {key, multiplicity} and we cannot have twice the same
// key ;
//
// - Items with weight 0 do not appear in multisets (default weight) ;
//
// - Items are ordered in increasing order of keys.
type PMarking []Atom

// ----------------------------------------------------------------------

// EvaluateGround evaluates a ground expression, that is an expression with no
// free variables, into a PMarking
func (net *Net) EvaluateGround(expr pnml.Expression) PMarking {
	if expr == nil {
		return PMarking{}
	}
	vals, mults := expr.Match(net.Net, nil)
	res := make(PMarking, len(vals))
	for k, v := range vals {
		res[k] = Atom{v, mults[k]}
	}
	sort.Slice(res, func(i, j int) bool { return pnml.ValueIsLess(res[i].Value, res[j].Value) })
	return res
}

func (net *Net) PrintPMarking(pm PMarking) string {
	if len(pm) == 0 {
		return "-"
	}
	s := ""
	for _, v := range pm {
		s += fmt.Sprintf("%d'%s ", v.Mult, net.PrintValue(v.Value))
	}
	return s
}

func (net *Net) PrintMarking(m Marking) string {
	s := ""
	for k, v := range net.Places {
		s += fmt.Sprintf("%s : %s\n", v.Name, net.PrintPMarking(m[k]))
	}
	return s
}

func (net *Net) printMarkingAligned(m Marking, left int, trunc int) string {
	s := ""
	for k, v := range net.Places {
		mm := []rune(net.PrintPMarking(m[k]))
		if len(mm) > trunc {
			mm = mm[:trunc]
			mm = append(mm, []rune("...")...)
		}
		s += fmt.Sprintf("%s%s : %s\n", v.Name, strings.Repeat(" ", left-len(v.Name)), string(mm))
	}
	return s
}
