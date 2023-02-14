// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

// Stepper is the type of marked hlnet, that is a pair consisting of a
// high-level net and its current marking
type Stepper struct {
	*Net
	Marking
	lpn int // length of the longest place name, for pretty printing markings
}

// NewStepper returns a fresh Stepper starting with the initial marking of n
func NewStepper(n *Net) *Stepper {
	m0 := Marking{COL: make([]PMarking, len(n.Places))}
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
		lpn:     lpn,
	}
	s.ComputeEnabled()
	return s
}

func (s *Stepper) String() string {
	return s.printMarkingAligned(s.Marking, s.lpn, 90)
}

// Enabled returns the set of transition (a slice of ordered indexes in n.Trans)
// which are enabled at the current marking
func (s *Stepper) ComputeEnabled() {
	// reset the enabled list of the current marking
	for tname := range s.Enabled {
		s.Enabled[tname] = false
	}
	for _, t := range s.Trans {
		if s.CheckCondition(t) {
			s.Enabled[t.Name] = true
		}
	}
}

// CheckCondition checks the (marking) condition of the given transition.
func (s *Stepper) CheckCondition(t Transition) bool {
	for _, v := range net.cond[i] {
		if m.get(v.value) < v.mult {
			return false
		}
	}

	return true
}
