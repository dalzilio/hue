// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import (
	"fmt"

	"github.com/dalzilio/hue/pkg/internal/util"
	"github.com/dalzilio/hue/pkg/pnml"
)

// iterator is a structure which allows us to iterate through all the possible
// valid associations between variables (in a transition Env) and values (in the
// marking of places). We need to keep track of one position for each "split"
// pattern in the input arc of a transition.
type Iterator struct {
	*Net
	tid            int            // index of the transition in the net
	pnml.Operation                // condition associated with the transition
	idx            int            // index of the arc we are currently trying to match
	checkpoints    []pnml.VEnv    // copy of Venv used when backtracking
	arcs           []*arcIterator // sub-iterator for each input place
	finished       bool           // report that we exhausted the search
}

type arcIterator struct {
	pre         []pnml.Expression
	pl          int           // index of the related place in the net
	pos         []int         // current position in the multiset of values
	mults       []int         // current multiplicities
	match       IteratorMatch // to store the last potential match. A len of 0 means no match.
	idx         int           // index of the exps we are currently trying to match
	checkpoints []pnml.VEnv   // copy of Venv when backtarcking over exprs
	finished    bool          // report that we exhausted the search
}

// ----------------------------------------------------------------------

// newIterator initializes a slice of Iterators for transition with index k in
// the Net.
func (s *Stepper) newIterator(k int) Iterator {
	t := s.Trans[k]
	iter := Iterator{
		Net:       s.Net,
		tid:       k,
		Operation: t.Cond,
	}

	// We compute the environment of patterns in the input arcs to detect the
	// case where we have extra variables (see DatabaseWithMutex)
	pats := [][]pnml.Expression{}
	pl := []int{}
	insEnv := pnml.Env{}
	for _, a := range t.Ins {
		pats = append(pats, a.Pattern)
		pl = append(pl, a.Place)
		for _, e := range a.Pattern {
			insEnv = e.AddEnv(insEnv)
		}
	}

	// We test that all the variables in the condition are in the input arcs;
	// otherwise we may miss constraints during  unification and cannot decide
	// if a transition is enabled. This is the case with model PhilosopherDyn
	// for example.
	//
	// We also have the case where there variables in the out arcs not appearing
	// in the inputs. They should be associated with an <all> on all the values
	// with the type of the place. The only knwown case is DatabaseWithMutex.
	if !util.StringListIncludes(insEnv, t.Cond.AddEnv(nil)) {
		s.forbidFiring[k] = struct{}{}
		s.forbidEnabled[k] = struct{}{}
		s.forbidden[t.Name] = struct{}{}
		return iter
	}

	// if we have extra variables, we must prevent the stepper from firing this
	// transition.
	if len(t.Env) > len(insEnv) {
		s.forbidFiring[k] = struct{}{}
	}

	iter.checkpoints = make([]pnml.VEnv, len(pl)+1)
	iter.arcs = make([]*arcIterator, len(pl))

	// We compute again the set of variables that are matched by the patterns.
	// We use the fact that the order in which we unify VEnv is deterministic,
	// which means that we know wich variables may already have been set. The
	// first checkpoint is always empty and is used only to reset the VEnv.
	insEnv = make(pnml.Env, 0)
	iter.checkpoints[0] = make(pnml.VEnv)
	for i := range pats {
		iter.arcs[i] = &arcIterator{
			pre:         pats[i],
			pl:          pl[i],
			pos:         make([]int, len(pats[i])),
			mults:       make([]int, len(pats[i])),
			match:       make(IteratorMatch, 0),
			checkpoints: make([]pnml.VEnv, len(pats[i])),
		}
		for j, p := range pats[i] {
			insEnv = p.AddEnv(insEnv)
			iter.arcs[i].checkpoints[j] = newVenv(insEnv)
		}
		iter.checkpoints[i+1] = newVenv(insEnv)
	}

	return iter
}

// Venv returns the tight VEnv association when we have found a match. It can be
// found in the last checkpoint of the iterator.
func (iter *Iterator) Venv() pnml.VEnv {
	return iter.checkpoints[len(iter.arcs)]
}

func newVenv(env pnml.Env) pnml.VEnv {
	res := make(pnml.VEnv)
	for _, s := range env {
		res[s] = nil
	}
	return res
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
func (iter *Iterator) Step(m pnml.Marking) bool {
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
func (iter *Iterator) StepOneArc(i int, m pnml.Marking) bool {
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
func (iter *Iterator) Check(m pnml.Marking) (*Witness, bool) {
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
			if !iter.Operation.OK(iter.Net.Net, iter.Venv()) {
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
func (iter *Iterator) CheckOneArc(i int, m pnml.Marking) bool {
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
			a.checkpoints[a.idx].Copy(iter.checkpoints[i])
		} else {
			a.checkpoints[a.idx].Copy(a.checkpoints[a.idx-1])
		}
		a.mults[a.idx] = a.pre[a.idx].Unify(iter.Net.Net, h[a.pos[a.idx]].Value, a.checkpoints[a.idx])
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
			iter.checkpoints[i+1].Copy(a.checkpoints[a.idx])
			a.idx = 0
			return true
		}
		a.idx++
	}
}

func (iter *Iterator) PrintVEnv(net *Net) string {
	res := ""
	for vname, val := range iter.Venv() {
		res += fmt.Sprintf("%s : %s\n", vname, net.PrintValue(val))
	}
	return res
}
