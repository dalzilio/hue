// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import "github.com/dalzilio/hue/pkg/set"

// Stepper is the type of marked hlnet, that is a pair consisting of a
// high-level net and its current marking
type Stepper struct {
	*Net
	Marking
	lpn int // length of the longest place name, for pretty printing markings
}

// NewStepper returns a fresh Stepper starting with the initial marking of n
func NewStepper(n *Net) *Stepper {
	m0 := make(Marking, len(n.Places))
	lpn := 0
	for k, v := range n.Places {
		m0[k] = v.Init
		if len(v.Name) > lpn {
			lpn = len(v.Name)
		}
	}
	return &Stepper{
		Net:     n,
		Marking: m0,
		lpn:     lpn,
	}
}

func (s *Stepper) String() string {
	return s.printMarkingAligned(s.Marking, s.lpn, 80)
}

// Enabled returns the set of transition (a slice of ordered indexes in n.Trans)
// which are enabled at the current marking
func (s *Stepper) Enabled() set.Set {
	enabled := set.Set{}
	// for i := range s.Trans {
	// 	// if s.condition(s.Marking, i) {
	// 	// 	enabled = append(enabled, i)
	// 	// }
	// }
	return enabled
}

// // condition checks the (marking) condition of the given transition.
// func (net *Net) condition(m Marking, i int) bool {
// 	for _, v := range net.cond[i] {
// 		if m.get(v.value) < v.mult {
// 			return false
// 		}
// 	}
// 	for _, v := range net.inhibcond[i] {
// 		if m.get(v.value) >= v.mult {
// 			return false
// 		}
// 	}
// 	return true
// }
