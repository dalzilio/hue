// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

import (
	"errors"
	"sort"
)

// iterator is a structure which allows us to iterate through all the possible
// valid associations between variables (in a transition Env) and values (in the
// marking of places). We need to keep track of one position for each "split"
// pattern in the input arc of a transition.
type Iterator struct {
	*Net
	Env
	Operation
	venv     VEnv
	idx      int // index of the arcs we are currently trying to match
	arcs     []*arcIterator
	finished bool // report that we exhausted the search
}

type arcIterator struct {
	exps     []Expression
	pl       int           // index of the related place in the net
	pos      []int         // current position in the multiset of values
	mults    []int         // current multiplicities
	match    IteratorMatch // to store the last potential match. A len of 0 means no match.
	idx      int           // index of the exps we are currently trying to match
	finished bool          // report that we exhausted the search
}

// ----------------------------------------------------------------------

// NewIterator initialize a slice of Iterators for a transition. Pats and pl are
// two slices of equal length; pats is the list of "split" arc patterns and pl
// gives the index to the related input places.
func NewIterator(net *Net, env Env, cond Operation, pats [][]Expression, pl []int) (Iterator, error) {
	iter := Iterator{
		Net:       net,
		Env:       env,
		Operation: cond,
	}
	iter.arcs = make([]*arcIterator, len(pats))
	// We need to test that all the variables in env are in the input arcs;
	// otherwise we may miss constraints during the unification.
	insEnv := make(Env, 0)
	for k := range pats {
		iter.arcs[k] = &arcIterator{
			exps:     pats[k],
			pl:       pl[k],
			pos:      make([]int, len(pats[k])),
			mults:    make([]int, len(pats[k])),
			match:    make(IteratorMatch, 0),
			idx:      0,
			finished: false,
		}
		for _, p := range pats[k] {
			insEnv = p.AddEnv(insEnv)
		}
	}
	if !stringListIncludes(insEnv, cond.AddEnv(nil)) {
		return iter, errors.New("not enough variables in input arcs patterns")
	}
	return iter, nil
}

// returns true if all the strings in sl2 are also in sl1. WARNING: We sort the
// input strings.
func stringListIncludes(sl1, sl2 []string) bool {
	sort.Strings(sl1)
	for _, s := range sl2 {
		if x := sort.SearchStrings(sl1, s); (x >= len(sl1)) || (sl1[x] != s) {
			return false
		}
	}
	return true
}

// ----------------------------------------------------------------------

// Reset is used when we need to iterate over a new marking.
func (iter *Iterator) Reset() {
	iter.venv = make(VEnv)
	iter.idx = 0
	iter.finished = false
	for k := range iter.arcs {
		iter.ResetOneArc(k)
	}
}

// ResetOneArc is used when we need to reset only one arcIterator.
func (iter *Iterator) ResetOneArc(i int) {
	p := iter.arcs[i]
	for k := range p.pos {
		p.pos[k] = 0
		p.mults[k] = 0
	}
	p.match = p.match[:0]
	p.idx = 0
	p.finished = false
}

// ----------------------------------------------------------------------

// Step increments the search. We return false if we finished.
func (iter *Iterator) Step(m []Hue) bool {
	if iter.finished {
		return false
	}
	for {
		if !iter.StepOneArc(iter.idx, m) {
			// the current arcIterator has finished
			if iter.idx == 0 {
				iter.finished = true
				return false
			}
			iter.ResetOneArc(iter.idx)
			iter.idx--
			continue
		}
		return true
	}
}

// StepOneArc increments the position for the arcIterator at position i. it
// returns false if we finished.
func (iter *Iterator) StepOneArc(i int, m []Hue) bool {
	a := iter.arcs[i]
	h := m[a.pl]
	// there is nothing to match if the place marking is empty or if we finished
	if len(h) == 0 || a.finished {
		return false
	}
	for {
		if a.pos[a.idx] == len(h)-1 {
			if a.idx == 0 {
				a.finished = true
				return false
			}
			a.pos[a.idx] = 0
			a.mults[a.idx] = 0
			a.idx--
			continue
		}
		a.pos[a.idx]++
		return true
	}
}

// ----------------------------------------------------------------------

// Check computes the next possible assignment (VEnv) that matches all the arc
// patterns and also the conditon (Operation). It returns false if it is not
// possible.
func (iter *Iterator) Check(m []Hue) bool {
	for {
		for i := range iter.arcs {
			if !iter.CheckOneArc(i, m) {
				iter.finished = true
				return false
			}
		}
		// If we have a match for all the input places, we have a possible
		// match. We need to check that it is a solution for the condition in
		// the transition
		if !iter.Operation.OK(iter.Net, iter.venv) {
			// if it is not, we need to start iterating to the next potential candidate
			if !iter.Step(m) {
				return false
			}
			continue
		}
		return true
	}
}

// CheckOneArc computes the next possible assignment for the set of arc patterns
// associated with the same input place. We return false if there are no match.
// Otherwise  the potential match for the current place is stored in a.match.
func (iter *Iterator) CheckOneArc(i int, m []Hue) bool {
	a := iter.arcs[i]
	h := m[a.pl]
	// there is nothing to match if the place marking is empty or if we finished
	if len(h) == 0 || a.finished {
		return false
	}
	// Otherwise check if we can unify the arc conditions with the values at the
	// given position. We also need to collect the multiplicities to make sure
	// that we have enough tokens of the right value in h. The index of the
	// current subpattern we are testing is the value of a.idx. We keep a copy
	// of the current VEnv in case we need to backtrack changes made during a
	// failed unification.
	venv2 := iter.venv.duplicate()
	// We zero a.mults
	for i := range a.mults {
		a.mults[i] = 0
	}
	for {
		a.mults[a.idx] = a.exps[a.idx].Unify(iter.Net, h[a.pos[a.idx]].Value, iter.venv)
		if a.mults[a.idx] == 0 {
			// Unification failed. We need to change the position.
			if !iter.StepOneArc(i, m) {
				// we finished exploring for this arc
				return false
			}
			// we restore the VEnv and reset mult
			iter.venv.copy(venv2)
			a.mults[a.idx] = 0
			continue
		}
		if a.idx == len(a.exps)-1 {
			// We matched the last arc pattern, so we have a possible match ! We
			// need to check the multiplicities to make sure that we have enough
			// of them in the marking

			if !a.testCapacity(h) {
				// This is not a real match. We need to step to the next
				// possible match.
				if !iter.StepOneArc(i, m) {
					iter.venv.copy(venv2)
					a.mults[a.idx] = 0
					return false
				}
				continue
			}

			// We have found a match.
			return true
		}
		a.idx++
	}
}

// ----------------------------------------------------------------------

// IteratorMatch is a slice of (mult x position) that can be used to reconstruct
// the tokens that have been match by an iterator on a place marking
type IteratorMatch []struct {
	pos  int
	mult int
}

// insert a match point keeping as an invariant that values are sorted in order
// of position.
func insertMatchPoint(m IteratorMatch, pos int, mult int) IteratorMatch {
	for k := range m {
		if m[k].pos == pos {
			m[k].mult += mult
			return m
		}
		if m[k].pos > pos {
			// insert a new element at position k
			return append(m[:k], append(IteratorMatch{{pos: pos, mult: mult}}, m[k:]...)...)
		}
	}
	return append(m, struct {
		pos  int
		mult int
	}{pos: pos, mult: mult})
}

// testCapacity takes the sum of multiplicities of tokens at the same positions
// (given by a.pos) and check if it is less than the multiplicy found in h. If
// so we have a valid IteratorMatch.
func (a *arcIterator) testCapacity(h Hue) bool {
	a.match = a.match[:0]
	for i := range a.pos {
		a.match = insertMatchPoint(a.match, a.pos[i], a.mults[i])
	}
	for k := range a.match {
		if a.match[k].mult > h[a.match[k].pos].Mult {
			a.match = a.match[:0]
			return false
		}
	}
	return true
}

// ----------------------------------------------------------------------

// Witness returns a marking (a []Hue) for the current match. We return nil if
// iter did not match.
func (iter *Iterator) Witness(m []Hue) []Hue {
	res := make([]Hue, len(m))
	for i, a := range iter.arcs {
		res[a.pl] = iter.witnessOneArc(i, m)
	}
	return res
}

func (iter *Iterator) witnessOneArc(i int, m []Hue) Hue {
	a := iter.arcs[i]
	h := m[a.pl]
	if len(a.match) == 0 {
		// not a match
		return nil
	}
	res := make(Hue, len(a.match))
	for k := range res {
		res[k] = Atom{
			Value: h[a.match[k].pos].Value,
			Mult:  a.match[k].mult,
		}
	}
	return res
}
