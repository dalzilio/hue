// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"github.com/dalzilio/hue/pkg/pnml"
)

// Stepper is the type of marked hlnet, that is a pair consisting of a
// high-level net and its current marking
type Stepper struct {
	*Net
	Marking
	iter []pnml.Iterator // we keep iterators for each transitions in the net
	lpn  int             // length of the longest place name, for pretty printing markings
}

// NewStepper returns a fresh Stepper starting with the initial marking of n
func NewStepper(n *Net) *Stepper {
	m0 := Marking{
		COL:     make([]pnml.Hue, len(n.Places)),
		PT:      make(map[string]int),
		Enabled: make(map[string]bool),
	}
	lpn := 0
	for k, v := range n.Places {
		m0.COL[k] = v.Init
		m0.PT[v.Name] = v.Init.Sum()
		if len(v.Name) > lpn {
			lpn = len(v.Name)
		}
	}
	s := &Stepper{
		Net:     n,
		Marking: m0,
		iter:    make([]pnml.Iterator, len(n.Trans)),
		lpn:     lpn,
	}
	// we should compute enabled after testing reachability queries on the initial marking.
	s.SetInitialEnabled()
	s.ComputeEnabled()
	return s
}

func (s *Stepper) String() string {
	return s.printMarkingAligned(s.Marking, s.lpn, 90) + "\nEnabled: " + s.PrintEnabled(s.Marking) + "\n"
}

// SetInitialEnabled computes the enable relation for the initial marking as
// well as initialize some important data structure. We keep this separate from
// the initializer because we can sometimes answer reachability verdict before
// this step, which may be expensive on very large hlnets.
func (s *Stepper) SetInitialEnabled() {
	for k, t := range s.Trans {
		// We collect the patterns and the marking of the places for the input arcs.
		inse := [][]pnml.Expression{}
		insm := []int{}
		for _, a := range t.Ins {
			inse = append(inse, a.Pattern)
			insm = append(insm, a.Place)
		}
		s.iter[k] = pnml.NewIterator(s.Net.Net, t.Env, t.Cond, inse, insm)
	}
}

// Enabled returns the set of transition (a slice of ordered indexes in n.Trans)
// which are enabled at the current marking
func (s *Stepper) ComputeEnabled() {
	// We could be more clever and only update the transition whose input places
	// have been modified.
	for tname := range s.Enabled {
		s.Enabled[tname] = false
	}
	for k, t := range s.Trans {
		if s.iter[k].ExistMatch(s.COL) {
			s.Enabled[t.Name] = true
		}
	}
}
