// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"

	"github.com/dalzilio/hue/pkg/pnml"
)

// Stepper is the type of marked hlnet, that is a pair consisting of a
// high-level net and its current marking
type Stepper struct {
	*Net
	Marking
	iter        []pnml.Iterator // we keep iterators for each transitions in the net
	lpn         int             // length of the longest place name, for pretty printing markings
	showwitness bool            // print witness when we find a fireable transition
}

// NewStepper returns a fresh Stepper starting with the initial marking of n
func NewStepper(n *Net, showwitness bool, needfireable bool) (*Stepper, error) {
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
	s := Stepper{
		Net:         n,
		Marking:     m0,
		iter:        nil,
		lpn:         lpn,
		showwitness: showwitness,
	}

	if needfireable || showwitness {
		s.iter = make([]pnml.Iterator, len(n.Trans))
		// we should compute enabled after testing reachability queries on the initial marking.
		for k, t := range s.Trans {
			// We collect the patterns and the marking of the places for the input arcs.
			inse := [][]pnml.Expression{}
			insm := []int{}
			for _, a := range t.Ins {
				inse = append(inse, a.Pattern)
				insm = append(insm, a.Place)
			}
			var err error
			s.iter[k], err = pnml.NewIterator(s.Net.Net, t.Env, t.Cond, inse, insm)
			if err != nil {
				return &s, err
			}
		}
		s.ComputeEnabled()
	}

	return &s, nil
}

func (s *Stepper) String() string {
	res := s.printMarkingAligned(s.Marking, s.lpn, 90)
	if len(s.iter) != 0 {
		res += "\nFireable: " + s.PrintEnabled(s.Marking)
	}
	return res + "\n"
}

// ExistMatch reports if there is a set of values that match the condition of
// the transition with index t.
func (s *Stepper) ExistMatch(t int) bool {
	s.iter[t].Reset()
	if s.iter[t].Check(s.COL) {
		if s.showwitness {
			fmt.Println("----------------------------------")
			fmt.Printf("%s enabled\n", s.Trans[t].Name)
			fmt.Printf("witness:\n%s\n", s.PrintCOL(s.iter[t].Witness(s.COL)))
			fmt.Println("----------------------------------")
		}
		return true
	}
	return false
}

// Enabled returns the set of transition (a slice of ordered indexes in n.Trans)
// which are enabled at the current marking
func (s *Stepper) ComputeEnabled() {
	// We could be more clever and only update the transition whose input places
	// have been modified.
	for tname := range s.TPosition {
		s.Enabled[tname] = false
	}
	for k, t := range s.Trans {
		if s.ExistMatch(k) {
			s.Enabled[t.Name] = true
		}
	}
}
