// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import (
	"fmt"
	"sort"

	"github.com/dalzilio/hue/pkg/internal/util"
	"github.com/dalzilio/hue/pkg/pnml"
)

// iterator is a structure used to iterate through all the possible valid
// associations between variables (in a transition Env) and values (in the
// marking of places).
type Iterator struct {
	*Net
	tid            int            // index of the transition in the net
	idx            int            // index of the arc we are currently iterating
	pnml.Operation                // condition associated with the transition
	checkpoints    []pnml.Env     // keep track of which variable to update when backtracking
	venv           pnml.VEnv      // current environment
	arcs           []*arcIterator // sub-iterator for each input place
}

type arcIterator struct {
	pl             int               // index of the related place in the net
	arcidx         int               // index of the subarc we are currently iterating
	finished       bool              // to indicate we finished the exploration
	pre            []pnml.Expression // patterm related to place pl
	arccheckpoints []pnml.Env        // keep track of which variable to update when backtracking
	pos            []int             // current position in the multiset of values
	mults          []int             // current multiplicities
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
	pls := []int{}
	iter.checkpoints = []pnml.Env{}
	insEnv := pnml.Env{}

	for _, a := range s.Trans[k].Ins {
		pats = append(pats, a.Pattern)
		pls = append(pls, a.Place)
		extra := insEnv.Extra(a.Env)
		iter.checkpoints = append(iter.checkpoints, extra)
		insEnv = append(insEnv, extra...)
	}

	// We test that all the variables in the condition are in the input arcs;
	// otherwise we may miss constraints during unification and cannot decide if
	// a transition is enabled. This is the case with model PhilosopherDyn for
	// example. We also have the case where there are variables in the out arcs
	// not appearing in the inputs. They should be associated with an <all> on
	// all the values with the type of the place. The only knwown case is
	// DatabaseWithMutex.
	if !util.StringListIncludes(insEnv, s.Trans[k].Cond.AddEnv(nil)) {
		s.forbidFiring[k] = struct{}{}
		s.forbidEnabled[k] = struct{}{}
		s.forbidden[s.Trans[k].Name] = struct{}{}
		return iter
	}

	iter.venv = pnml.NewVEnv(insEnv)

	// If we have extra variables, we must prevent the stepper from firing this
	// transition. But we may may still want to test if it is enabled.
	if len(s.Trans[k].Env) > len(insEnv) {
		s.forbidFiring[k] = struct{}{}
	}

	iter.arcs = make([]*arcIterator, len(pls))
	insEnv = pnml.Env{}
	for i := range pls {
		iter.arcs[i] = &arcIterator{
			pl:             pls[i],
			pre:            pats[i],
			arccheckpoints: []pnml.Env{},
			pos:            make([]int, len(pats[i])),
			mults:          make([]int, len(pats[i])),
		}
		for _, p := range pats[i] {
			iter.arcs[i].arccheckpoints = append(iter.arcs[i].arccheckpoints, insEnv.Extra(p.AddEnv(nil)))
		}
	}

	return iter
}

// Environment returns the VEnv association related to a match.
func (iter *Iterator) Environment() pnml.VEnv {
	return iter.venv.Clone()
}

// ----------------------------------------------------------------------

// reset is used when we need to iterate over a new marking.
func (s *Stepper) reset(k int) {
	s.After[k] = nil
	s.iter[k].idx = 0
	s.iter[k].venv.ResetAll()
	for _, a := range s.iter[k].arcs {
		a.arcidx = 0
		for j := range a.pos {
			a.pos[j] = 0
			a.mults[j] = 0
		}
	}
}

// ----------------------------------------------------------------------

// step tries to increment the search and updates idx with the arc we need to
// start looking next. We return false if we have no more choices.
func (s *Stepper) step(k int) bool {
	a := s.iter[k].arcs[s.iter[k].idx]
	for {
		if a.finished {
			if s.iter[k].idx == 0 {
				// there are no matches
				return false
			}
			a.finished = false
			s.iter[k].idx--
			a = s.iter[k].arcs[s.iter[k].idx]
			continue
		}
		if s.stepCurrentPlace(k) {
			return true
		}
	}
}

// stepCurrentPlace increments the position for the arcIterator at position
// [idx][arcidx] in the iterator for transition k. We update arcidx and return
// false if we are no more choices.
func (s *Stepper) stepCurrentPlace(k int) bool {
	a := s.iter[k].arcs[s.iter[k].idx]
	if a.arcidx == len(a.pos) {
		// it means we found a possible marking at this place before
		a.arcidx--
	}
	for {
		if a.pos[a.arcidx] < len(s.COL[a.pl])-1 {
			s.iter[k].venv.Reset(a.arccheckpoints[a.arcidx])
			a.pos[a.arcidx]++
			a.mults[a.arcidx] = 0
			return true
		}
		a.pos[a.arcidx] = 0
		a.mults[a.arcidx] = 0
		s.iter[k].venv.Reset(a.arccheckpoints[a.arcidx])
		if a.arcidx == 0 {
			a.finished = true
			return false
		}
		a.arcidx--
	}
}

// ----------------------------------------------------------------------

// check computes the next possible assignment (VEnv) that matches all the arc
// patterns and also the condition (Operation). It returns false if we do not
// find a witness. It can also return an error if we find a problem while
// unifying an arc pattern (for instance if the arc contains an <All>
// expression, like with model DatabaseWithMutex).
func (s *Stepper) check(k int) (bool, error) {
	it := s.iter[k]
	s.reset(k)

	// There is nothing to match if one of the input place marking is empty
	for _, a := range it.arcs {
		if len(s.COL[a.pl]) == 0 {
			return false, nil
		}
	}

	for {
		// if it.idx < len(it.arcs) {
		// 	fmt.Printf("-[%s(place %s) with %s\n", s.Trans[k].Name, s.Places[it.arcs[it.idx].pl].Name, it.PrintVEnv(s.Net))
		// } else {
		// 	fmt.Printf("--[testing %s with %s\n", s.Trans[k].Name, it.PrintVEnv(s.Net))
		// }
		if it.idx == len(it.arcs) {
			// We have a potential match. We need to check that it is a solution
			// for the condition in the transition
			if it.Operation.OK(s.Net.Net, it.venv) {
				// we should not compute the result of forbidden transitions
				if _, ok := s.forbidFiring[k]; !ok {
					s.After[k] = s.getWitness(k)
				}
				return true, nil
			}

			// It is not a match, we need to start iterating to the next
			// potential candidate
			s.iter[k].idx--
			if !s.step(k) {
				return false, nil
			}
		}

		b, err := s.checkCurrentPlace(k)

		if err != nil {
			return false, err
		}

		if !b {
			if !s.step(k) {
				return false, nil
			}
			continue
		}

		it.idx++
	}
}

// checkCurrentPlace computes the next possible assignment for the patterns
// associated with arc[idx] of transition[k]. We return false if there are no
// match.
func (s *Stepper) checkCurrentPlace(k int) (bool, error) {
	i := s.iter[k].idx
	a := s.iter[k].arcs[i]
	s.iter[k].venv.Reset(s.iter[k].checkpoints[i])
	a.arcidx = 0
	var err error
	for {
		if a.arcidx == len(a.pre) {
			// We matched the last arc pattern, so we have a possible match ! We
			// need to check the multiplicities to make sure that we have enough
			// of them in the marking
			if a.testCapacity(s.COL[a.pl]) {
				// We have found a match.
				return true, nil
			}
			// This is not a real match. We need to step to the next possible
			// match.
			if !s.stepCurrentPlace(k) {
				return false, nil
			}
		}

		a.mults[a.arcidx], err = a.pre[a.arcidx].Unify(s.Net.Net, s.COL[a.pl][a.pos[a.arcidx]].Value, s.iter[k].venv)

		if err != nil {
			return false, err
		}

		if a.mults[a.arcidx] == 0 {
			if !s.stepCurrentPlace(k) {
				return false, nil
			}
			continue
		}

		a.arcidx++
	}
}

func (iter *Iterator) PrintVEnv(net *Net) string {
	res := []string{}
	for vname, val := range iter.Environment() {
		res = append(res, fmt.Sprintf("%s : %s", vname, net.PrintValue(val)))
	}
	sort.Strings(res)
	return util.ZipString(res, "(", ")", ", ")
}
