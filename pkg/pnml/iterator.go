// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

import (
	"errors"

	"github.com/dalzilio/hue/pkg/internal/util"
)

// iterator is a structure which allows us to iterate through all the possible
// valid associations between variables (in a transition Env) and values (in the
// marking of places). We need to keep track of one position for each "split"
// pattern in the input arc of a transition.
type Iterator struct {
	*Net
	Operation
	idx         int            // index of the arc we are currently trying to match
	checkpoints []VEnv         // copy of Venv used when backtracking
	arcs        []*arcIterator // sub-iterator for each input place
	finished    bool           // report that we exhausted the search
}

type arcIterator struct {
	pre         []Expression
	pl          int           // index of the related place in the net
	pos         []int         // current position in the multiset of values
	mults       []int         // current multiplicities
	match       IteratorMatch // to store the last potential match. A len of 0 means no match.
	idx         int           // index of the exps we are currently trying to match
	checkpoints []VEnv        // copy of Venv when backtarcking over exprs
	finished    bool          // report that we exhausted the search
}

// ----------------------------------------------------------------------

// NewIterator initialize a slice of Iterators for a transition. Pats and pl are
// two slices of equal length; pats is the list of "split" arc patterns and pl
// gives the index to the related input places.
func NewIterator(net *Net, tname string, cond Operation, pats [][]Expression, pl []int) (Iterator, error) {
	iter := Iterator{
		Net:         net,
		Operation:   cond,
		checkpoints: make([]VEnv, len(pats)+1),
		arcs:        make([]*arcIterator, len(pats)),
	}
	// We compute the set of variables that are matched by the patterns. We use
	// the fact that the order in which we unify VEnv is deterministic, which
	// means that we know wich variables may already have been set. The first
	// checkpoint is always empty and is used only to reset the VEnv.
	insEnv := make(Env, 0)
	iter.checkpoints[0] = make(VEnv)
	for k := range pats {
		iter.arcs[k] = &arcIterator{
			pre:         pats[k],
			pl:          pl[k],
			pos:         make([]int, len(pats[k])),
			mults:       make([]int, len(pats[k])),
			match:       make(IteratorMatch, 0),
			checkpoints: make([]VEnv, len(pats[k])),
		}
		for i, p := range pats[k] {
			insEnv = p.AddEnv(insEnv)
			iter.arcs[k].checkpoints[i] = newVenv(insEnv)
		}
		iter.checkpoints[k+1] = newVenv(insEnv)
	}
	// We test that all the variables in the condition are in the input arcs;
	// otherwise we may miss constraints during  unification and cannot decide
	// if a transition is enabled. This is the case with model PhilosopherDyn
	// for example.
	//
	// We also have the case where there variables in the out arcs not appearing
	// in the inputs. They should be associated with an <all> on all the values
	// with the type of the place. The only knwown case is DatabaseWithMutex.
	//
	// TODO: use a ternary system for enabled transition, for this particular
	// case.
	if !util.StringListIncludes(insEnv, cond.AddEnv(nil)) {
		return iter, errors.New("not enough variables in input arcs patterns")
	}
	return iter, nil
}

// Venv returns the tight VEnv association when we have found a match. It can be
// found in the last checkpoint of the iterator.
func (iter *Iterator) Venv() VEnv {
	return iter.checkpoints[len(iter.arcs)]
}

// ----------------------------------------------------------------------

// Reset is used when we need to iterate over a new marking.
func (iter *Iterator) Reset() {
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
// patterns and also the condition (Operation). It returns false if it is not
// possible. The first return value is a witness: a sub marking of m that
// corresponds to the preconiditions of the transition.
func (iter *Iterator) Check(m []Hue) (*Witness, bool) {
	iter.Reset()
	for {
		if iter.finished {
			return nil, false
		}

		if !iter.CheckOneArc(iter.idx, m) {
			if !iter.Step(m) {
				// there is no more possibilities
				iter.finished = true
				return nil, false
			}
			continue
		}

		if iter.idx == len(iter.arcs)-1 {
			// If we have a match for all the input places, we have a possible
			// match. We need to check that it is a solution for the condition
			// in the transition
			if !iter.Operation.OK(iter.Net, iter.Venv()) {
				// it is not a match, we need to start iterating to the next
				// potential candidate
				if !iter.Step(m) {
					return nil, false
				}
				continue
			}
			return iter.GetWitness(m), true
		}
		iter.idx++
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

	// We zero a.mults
	for i := range a.mults {
		a.mults[i] = 0
	}
	for {
		// we restore the VEnv and reset mult
		if a.idx == 0 {
			a.checkpoints[a.idx].copy(iter.checkpoints[i])
		} else {
			a.checkpoints[a.idx].copy(a.checkpoints[a.idx-1])
		}
		a.mults[a.idx] = a.pre[a.idx].Unify(iter.Net, h[a.pos[a.idx]].Value, a.checkpoints[a.idx])
		if a.mults[a.idx] == 0 {
			// Unification failed. We need to change the position.
			if !iter.StepOneArc(i, m) {
				// we finished exploring for this arc
				return false
			}
			continue
		}
		if a.idx == len(a.pre)-1 {
			// We matched the last arc pattern, so we have a possible match ! We
			// need to check the multiplicities to make sure that we have enough
			// of them in the marking
			if !a.testCapacity(h) {
				// This is not a real match. We need to step to the next
				// possible match.
				if !iter.StepOneArc(i, m) {
					return false
				}
				continue
			}
			// We have found a match.
			iter.checkpoints[i+1].copy(a.checkpoints[a.idx])
			a.idx = 0
			return true
		}
		a.idx++
	}
}
