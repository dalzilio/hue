// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"
	"math/rand"

	"github.com/dalzilio/hue/pkg/internal/util"
	"github.com/dalzilio/hue/pkg/pnml"
)

// Stepper is the type of marked hlnet, that is a pair consisting of a
// high-level net and its current marking
type Stepper struct {
	*Net
	State
	m0            pnml.Marking        // copy of the initial marking
	iter          []*Iterator         // we keep iterators for each transitions in the net
	lpn           int                 // length of the longest place name, for pretty printing markings
	forbidFiring  map[int]struct{}    // index of transitions we should never fire
	forbidEnabled map[int]struct{}    // index of transitions for which we cannot decide enabledness
	forbidden     map[string]struct{} // name of transitions in forbidEnabled

}

// NewStepper returns a fresh Stepper starting with the initial marking of n
func NewStepper(n *Net) *Stepper {
	m0 := State{
		COL:     make(pnml.Marking, len(n.Places)),
		PT:      make(map[string]int),
		Enabled: nil,
		After:   make([]pnml.Marking, len(n.Trans)),
	}
	lpn := 0
	for k, v := range n.Places {
		m0.COL[k] = v.Init
		m0.PT[v.Name] = m0.COL[k].Sum()
		if len(v.Name) > lpn {
			lpn = len(v.Name)
		}
	}

	s := Stepper{
		Net:           n,
		State:         m0,
		m0:            m0.COL.Clone(),
		iter:          nil,
		lpn:           lpn,
		forbidFiring:  make(map[int]struct{}),
		forbidEnabled: make(map[int]struct{}),
		forbidden:     make(map[string]struct{}),
	}

	s.iter = make([]*Iterator, len(n.Trans))

	for k := range n.Trans {
		s.iter[k] = s.newIterator(k)
	}

	return &s
}

func (s *Stepper) InitializeEnabled() {
	if s.Enabled != nil {
		return
	}
	s.Enabled = make(map[string]bool)
	s.computeEnabled()
}

func (s *Stepper) String() string {
	res := s.printMarkingAligned(s.State, s.lpn, 65, 25)
	if len(s.iter) != 0 {
		res += "Fireable: " + s.PrintEnabled()
	}
	return res
}

func (s *Stepper) PrintCOL(m pnml.Marking) string {
	res := ""
	for k, v := range s.Places {
		res += fmt.Sprintf("%s : %s\n", v.Name, s.PrintHue(m[k]))
	}
	return res
}

func (s *Stepper) PrintCOLShort(m pnml.Marking) string {
	res := []string{}
	for k, v := range s.Places {
		if len(m[k]) == 0 {
			continue
		}
		res = append(res, fmt.Sprintf("%s: %s", v.Name, s.PrintHue(m[k])))
	}
	return util.ZipString(res, "[", "]", ", ")
}

// Enabled returns the set of transitions which are enabled at the current
// marking. As a side effect, it also add a list of witnesses in s for each of
// the enabled transitions.
func (s *Stepper) computeEnabled() {
	// We could be more clever and only update the transition whose input places
	// have been modified.
	for tname := range s.TPosition {
		s.Enabled[tname] = false
	}
	for k, t := range s.Trans {
		if _, ok := s.forbidEnabled[k]; ok {
			continue
		}
		ok, err := s.check(k)
		if err != nil {
			s.forbidEnabled[k] = struct{}{}
			s.forbidden[t.Name] = struct{}{}
			continue
		}

		if ok {
			s.Enabled[t.Name] = true
		}
	}
}

// FireAtRandom chooses one of the enabled transitions and fires it. If we have
// no transitions to fire, we start again from the initial marking.
func (s *Stepper) FireAtRandom(verbose bool) error {
	if s.Enabled == nil {
		s.Enabled = make(map[string]bool)
		s.computeEnabled()
	}

	choose := []string{}
	for k, v := range s.Enabled {
		if v {
			// we also need to check that we can fire the transition
			if _, ok := s.forbidFiring[s.TPosition[k]]; !ok {
				choose = append(choose, k)
			}
		}
	}

	if len(choose) == 0 {
		s.update(s.m0)
		s.computeEnabled()
		if verbose {
			fmt.Println("----------------------------------")
			if len(s.Enabled) == 0 {
				fmt.Printf("[deadlock ; restarting]\n")
			} else {
				fmt.Printf("[only \"forbidden\" transitions ; restarting]\n")
			}
			fmt.Println("----------------------------------")
		}
		return fmt.Errorf("nothing to fire; restarting")
	}

	tname := choose[rand.Intn(len(choose))]

	s.Fire(s.TPosition[tname], verbose)
	return nil
}

// Fire updates the stepper with the result of firing transition t. If there are
// no witnesses, we try to compute one. We do nothing when t is not enabled.
func (s *Stepper) Fire(k int, verbose bool) {
	if _, ok := s.forbidFiring[k]; ok {
		return
	}
	if s.After[k] == nil {
		s.computeEnabled()
	}
	if !s.Enabled[s.Trans[k].Name] {
		return
	}
	if verbose {
		fmt.Println("----------------------------------")
		fmt.Printf("%s %s\n", s.Trans[k].Name, s.iter[k].PrintVEnv(s.Net))
		fmt.Println("----------------------------------")
	}
	s.update(s.After[k])
}

// update changes the marking to m in the stepper
func (s *Stepper) update(m pnml.Marking) {
	s.COL = m
	for k, v := range s.Places {
		s.PT[v.Name] = m[k].Sum()
	}
	s.computeEnabled()
}
