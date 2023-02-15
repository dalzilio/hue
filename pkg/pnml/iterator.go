// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

type iterand struct {
	positions []int
}

type itchan chan iterand

// // We collect the conditions and the marking of the places for the input arcs.
// inse := []pnml.Expression{}
// insm := []pnml.Hue{}
// for _, a := range t.Ins {
// 	inse = append(inse, a.Pattern)
// 	insm = append(insm, s.COL[a.Place])
// }
// // we initalize a mapping between var names in the transition environment
// // and their values
// assoc := make(map[string]pnml.Atom)
// return pnml.ExistMatch(inse, insm, assoc, t.Cond)

// func ExistMatch(inse []Expression, insm []Hue, assoc map[string]Atom, cond Operation) bool {
// 	if len(inse) == 0
// switch exp :=
// }

// // Iterator is a structure which allows us to iterate through all the possible
// // valid environments (sequence of typed variables) corrresponding to given
// // multiset of values (a slice of Hue denoting the marking of a place). We
// // should build one iterator for each input arc in a hlnet.
// //
// // An Iterator keeps track of the current index in a Hue where it can
// // potentially find a match.
// type Iterator struct {
// 	Expression
// 	Env
// 	Hue
// 	index    int
// 	finished bool
// }

// // mkiter initialize a slice of Iterators for a transition. We only need to know
// // the number of input arcs.
// func mkiter(size int) []Iterator {
// 	iter := make([]Iterator, size)
// 	return iter
// }

// // next steps to the next possible match and returns false when we have gone
// // through all the possible iterations.
// func next(iter []Iterator) bool {
// 	for _, i := range iter {
// 		if !i.finished {
// 			return true
// 		}
// 	}
// 	return false
// }

// // step compute the next possible match; it returns false if we finish without finding one
// func step(iter []Iterator) bool {
// 	return false
// }

// // check reports if a condition is true for the current match
// func check(iter []Iterator, cond Expression) bool {
// 	return true
// }

// // func (iter *envIterator) skip(k int) {

// // func (iter *Iterator) zero(k int)
