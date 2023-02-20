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
	pl          int         // index of the related place in the net
	pos         []int       // current position in the multiset of values
	mults       []int       // current multiplicities
	idx         int         // index of the exps we are currently trying to match
	checkpoints []pnml.VEnv // copy of Venv when backtarcking over exprs
	finished    bool        // report that we exhausted the search
}

// ----------------------------------------------------------------------

// newIterator initializes a slice of Iterators for transition with index k in
// the Net.
func (s *Stepper) newIterator(k int) *Iterator {
	iter := &Iterator{
		Net:       s.Net,
		tid:       k,
		Operation: s.Trans[k].Cond,
	}

	// We compute the environment of patterns in the input arcs to detect the
	// case where we have extra variables (see DatabaseWithMutex)
	pats := [][]pnml.Expression{}
	pl := []int{}
	insEnv := pnml.Env{}
	for _, a := range s.Trans[k].Ins {
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
	// We also have the case where there are variables in the out arcs not
	// appearing in the inputs. They should be associated with an <all> on all
	// the values with the type of the place. The only knwown case is
	// DatabaseWithMutex.
	if !util.StringListIncludes(insEnv, s.Trans[k].Cond.AddEnv(nil)) {
		s.forbidFiring[k] = struct{}{}
		s.forbidEnabled[k] = struct{}{}
		s.forbidden[s.Trans[k].Name] = struct{}{}
		return iter
	}

	// if we have extra variables, we must prevent the stepper from firing this
	// transition.
	if len(s.Trans[k].Env) > len(insEnv) {
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

// Environment returns the VEnv association related to the last match. It can be
// found in the last checkpoint of the iterator.
func (iter *Iterator) Environment() pnml.VEnv {
	return iter.checkpoints[len(iter.arcs)].Clone()
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
	for k := range iter.checkpoints {
		iter.checkpoints[k].Reset()
	}
	for k := range iter.arcs {
		iter.ResetOneArc(k)
	}
}

// ResetOneArc is used when we need to reset only one arcIterator.
func (iter *Iterator) ResetOneArc(i int) {
	for k := range iter.arcs[i].pos {
		iter.arcs[i].pos[k] = 0
		iter.arcs[i].mults[k] = 0
	}
	iter.arcs[i].idx = 0
	iter.arcs[i].finished = false
	for k := range iter.arcs[i].checkpoints {
		iter.arcs[i].checkpoints[k].Reset()
	}
}

// ----------------------------------------------------------------------

// Step increments the search. We return false if we finished.
func (iter *Iterator) Step(m pnml.Marking) bool {
	if iter.finished {
		return false
	}
	for {
		if iter.StepOneArc(iter.idx, m) {
			iter.idx = 0
			return true
		}
		// the current arcIterator has finished
		if iter.idx == len(iter.arcs)-1 {
			iter.finished = true
			return false
		}
		iter.ResetOneArc(iter.idx)
		iter.idx++
		continue
	}
}

// StepOneArc increments the position for the arcIterator at position i. it
// returns false if we finished.
func (iter *Iterator) StepOneArc(i int, m pnml.Marking) bool {
	a := iter.arcs[i]
	// there is nothing to match if the place marking is empty or if we finished
	if len(m[a.pl]) == 0 || a.finished {
		return false
	}
	for {
		if a.pos[a.idx] < len(m[a.pl])-1 {
			a.pos[a.idx]++
			a.idx = 0
			return true
		}
		if a.idx == len(a.pos)-1 {
			a.finished = true
			return false
		}
		a.pos[a.idx] = 0
		a.mults[a.idx] = 0
		a.idx++
		continue
	}
}

// ----------------------------------------------------------------------

// check computes the next possible assignment (VEnv) that matches all the arc
// patterns and also the condition (Operation). It returns false if we do not
// find a witness.
func (s *Stepper) check(k int) (bool, error) {
	it := s.iter[k]
	it.Reset()
	i := 0
	for {
		if it.finished {
			return false, nil
		}

		if i == len(it.arcs) {
			// If we have a match for all the input places, we have a possible
			// match. We need to check that it is a solution for the condition
			// in the transition
			if it.Operation.OK(s.Net.Net, it.Environment()) {
				// we should not compute the result of forbidden transitions
				if _, ok := s.forbidFiring[k]; ok {
					return true, nil
				}
				s.Next[k] = s.getWitness(k)
				return true, nil
			}
			// it is not a match, we need to start iterating to the next
			// potential candidate
			if it.Step(s.COL) {
				i = 0
				continue
			}
			it.finished = true
			return false, nil
		}

		b, err := it.CheckOneArc(i, s.COL)

		if err != nil {
			return false, err
		}

		if !b {
			if it.Step(s.COL) {
				i = 0
				continue
			}
			// there is no more possibilities
			it.finished = true
			return false, nil
		}

		i++
	}
}

// CheckOneArc computes the next possible assignment for the set of arc patterns
// associated with the same input place. We return false if there are no match.
// Otherwise  the potential match for the current place is stored in a.match.
func (iter *Iterator) CheckOneArc(i int, m pnml.Marking) (bool, error) {
	a := iter.arcs[i]
	var err error

	// We zero a.mults
	for i := range a.mults {
		a.mults[i] = 0
	}

	// there is nothing to match if the place marking is empty or if we finished
	if len(m[a.pl]) == 0 || a.finished {
		return false, nil
	}

	// Otherwise check if we can unify the arc conditions with the values at the
	// given position. We also need to collect the multiplicities to make sure
	// that we have enough tokens of the right value in h. We keep a copy of the
	// current VEnv in case we need to backtrack changes made during a failed
	// unification.
	j := 0
	for {
		if j == len(a.pre) {
			// We matched the last arc pattern, so we have a possible match ! We
			// need to check the multiplicities to make sure that we have enough
			// of them in the marking
			if a.testCapacity(m[a.pl]) {
				// We have found a match.
				iter.checkpoints[i+1].Copy(a.checkpoints[len(a.pre)-1])
				return true, nil
			}
			// This is not a real match. We need to step to the next
			// possible match.
			if iter.StepOneArc(i, m) {
				j = 0
				continue
			}
			return false, nil
		}

		// we restore the VEnv and reset mult
		if j == 0 {
			a.checkpoints[j].Copy(iter.checkpoints[i])
		} else {
			a.checkpoints[j].Copy(a.checkpoints[j-1])
		}

		a.mults[j], err = a.pre[j].Unify(iter.Net.Net, m[a.pl][a.pos[j]].Value, a.checkpoints[j])

		if err != nil {
			return false, err
		}

		if a.mults[j] == 0 {
			// Unification failed. We need to change the position.
			if iter.StepOneArc(i, m) {
				j = 0
				continue
			}
			// otherwise we finished exploring this arc
			return false, nil
		}

		j++
	}
}

func (iter *Iterator) PrintVEnv(net *Net) string {
	res := []string{}
	for vname, val := range iter.Environment() {
		res = append(res, fmt.Sprintf("%s : %s", vname, net.PrintValue(val)))
	}
	return util.ZipString(res, "(", ")", ", ")
}
