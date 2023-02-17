// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

import "sort"

// iterator is a structure which allows us to iterate through all the possible
// valid associations between variables (in a transition Env) and values (in the
// marking of places). We need to keep track of one position for each "split"
// pattern in the input arc of a transition.
type Iterator struct {
	*Net
	Env
	Operation
	venv VEnv
	arcs []*arcIterator
}

type arcIterator struct {
	exps []Expression
	pl   int   // index of the related place in the net
	pos  []int // current position in the multiset of values
}

// NewIterator initialize a slice of Iterators for a transition. Pat and pl are
// two slices of equal length; pats is the list of "split" arc patterns and pl
// gives the index to the related input places.
func NewIterator(net *Net, env Env, cond Operation, pats [][]Expression, pl []int) Iterator {
	iter := Iterator{Env: env, Operation: cond}
	iter.arcs = make([]*arcIterator, len(pats))
	for k := range pats {
		iter.arcs[k] = &arcIterator{
			exps: pats[k],
			pl:   pl[k],
			pos:  make([]int, len(pats[k])),
		}
	}
	return iter
}

// Reset is used when we need to iterate over a new marking.
func (iter *Iterator) Reset() {
	iter.venv = make(VEnv)
	for _, p := range iter.arcs {
		for k := range p.pos {
			p.pos[k] = 0
		}
	}
}

// ExistMatch reports if there is a set of values that match the condition of
// the Iterator.
func (iter *Iterator) ExistMatch(m []Hue) bool {
	iter.Reset()
	for {
		if !iter.Step(m) {
			return false
		}
		if iter.CheckCondition() {
			return true
		}
	}
}

// Step computes the next possible assignment (VEnv) and returns false if it is
// not possible.
func (iter *Iterator) Step(m []Hue) bool {
	for _, a := range iter.arcs {
		h := m[a.pl]
		// if we reached the end of the marking with one of the arcIterator; for
		// instance because the place marking is empty
		for _, pp := range a.pos {
			if pp == len(h) {

				return false
			}
		}
		// Otherwise check if we can unify the arc condition with the values at
		// the given position. We keep a copy of the current VEnv in case we
		// need to backtrack changes made duting a failed unification. We also
		// need to collect the multiplicities to make sure that we have enough
		// tokens of the right value in h.
		mult := make([]int, len(a.exps))
		k := 0
		var e Expression
		var token Atom
		venv2 := iter.venv.duplicate()
	CHECK:
		for {
			e = a.exps[k]
			token = h[a.pos[k]]
			mult[k] = e.Unify(iter.Net, token.Value, iter.venv)
			if mult[k] == 0 {
				// Unification failed. We need to change the position. First we
				// test if we are already at the end of h, in which case we need
				// to increment previous arcs. Except if we are already at the
				// beginning, in which case we stop.
			REWIND:
				if a.pos[k] == len(h)-1 {
					if k == 0 {
						return false
					}
					a.pos[k] = 0
					mult[k] = 0
					k--
					goto REWIND
				}
				a.pos[k]++
				// we restore the VEnv and reset mult
				iter.venv.copy(venv2)
				mult[k] = 0
				goto CHECK
			}
			if k == len(a.exps)-1 {
				// we have a possible match ! We need to check the
				// multiplicities to make sure that we have enough of them in
				// the marking
				if !capacity(mult, a.pos, h) {
					// we need to REWIND, but we cannot 'goto REWIND' because it
					// would jump across a nested block. So we copy the logic
				REWIND2:
					if a.pos[k] == len(h)-1 {
						if k == 0 {
							return false
						}
						a.pos[k] = 0
						mult[k] = 0
						k--
						goto REWIND2
					}
					a.pos[k]++
					iter.venv.copy(venv2)
					mult[k] = 0
					goto CHECK
				}
				return true
			}
			k++
		}

	}
	return true
}

// capacity takes the sum of multiplicities of tokens at the same positions
// (given by pos) and check if it is less than the multiplicy found in h
func capacity(mult []int, pos []int, h Hue) bool {
	// we sort the slices so that multiplicities for the same position are
	// contiguous
	sort.Slice(pos, func(i int, j int) bool {
		if pos[i] < pos[j] {
			mult[i], mult[j] = mult[j], mult[i]
			return true
		}
		return false
	})
	// we take the sum of multiplicities at this
	current := pos[0]
	sum := 0
	for k, p := range pos {
		if p > current {
			if sum > h[current].Mult {
				return false
			}
			current = p
			sum = mult[k]
			continue
		}
		sum += mult[k]
	}
	return sum <= h[current].Mult
}

// CheckCondition reports if the transition's condition is true for the current
// environment
func (iter *Iterator) CheckCondition() bool {
	return iter.Operation.OK(iter.Net, iter.venv)
}
