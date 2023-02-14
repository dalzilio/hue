// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

import (
	"fmt"
)

// Value provides a more efficient representation for values
//
// {0 nil} 			is a Dot
// {i nil} 			is Constant(name) where i uniquely identifies name
// {i {j {...}}} 	is for tuples
// we encode a range value, x, using a constant named _intx
type Value struct {
	Head int
	Tail *Value
}

// Atom is a pair of a multiplicity and a colored value.
type Atom struct {
	*Value
	Mult int
}

// Hue is a slice of Atoms meant to represent the possible value in a colored
// place
type Hue []Atom

func (pm Hue) Sum() int {
	s := 0
	for _, v := range pm {
		s += v.Mult
	}
	return s
}

func (net *Net) PrintHue(pm Hue) string {
	if len(pm) == 0 {
		return "-"
	}
	s := ""
	for _, v := range pm {
		s += fmt.Sprintf("%d'%s ", v.Mult, net.PrintValue(v.Value))
	}
	return s
}

// ----------------------------------------------------------------------

// ValueIsLess reports if vi is before vj in comparaison order. We assume, and
// do not test, that vi and vj are both of the same type.
func ValueIsLess(vi, vj *Value) bool {
	if vi == nil {
		return vj == nil
	}
	if vi.Head == vj.Head {
		return ValueIsLess(vi.Tail, vj.Tail)
	}
	return vi.Head < vj.Head
}

// ----------------------------------------------------------------------

// PrintValue returns a readable description of a Value
func (net *Net) PrintValue(val *Value) string {
	if val.Tail == nil {
		return net.printHeadValue(val.Head)
	}
	c := fmt.Sprintf("(%s", net.printHeadValue(val.Head))
	return net.printTupleValue(c, val.Tail)
}

func (net *Net) printHeadValue(i int) string {
	return net.identity[i]
}

func (net *Net) printTupleValue(s string, val *Value) string {
	if val == nil {
		return s + ")"
	}
	c := net.printHeadValue(val.Head)
	return net.printTupleValue(s+", "+c, val.Tail)
}

// ----------------------------------------------------------------------

func (net *Net) enumprod(elem []string) []*Value {
	if len(elem) == 1 {
		return net.World[elem[0]]
	}
	head := net.World[elem[0]]
	tail := net.enumprod(elem[1:])

	var list []*Value
	for _, a := range head {
		for _, b := range tail {
			val := Value{Head: a.Head, Tail: b}
			pval, ok := net.Unique[val]
			if !ok {
				pval = &val
				net.Unique[val] = &val
			}
			list = append(list, pval)
		}
	}
	return list
}
