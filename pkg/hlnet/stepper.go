// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/dalzilio/hue/pkg/pnml"
)

// Stepper is the type of marked hlnet, that is a pair consisting of a
// high-level net and its current marking
type Stepper struct {
	*Net
	State
	m0            pnml.Marking        //copy of the initial marking
	iter          []Iterator          // we keep iterators for each transitions in the net
	lpn           int                 // length of the longest place name, for pretty printing markings
	forbidFiring  map[int]struct{}    // index of transitions we should never fire
	forbidEnabled map[int]struct{}    // index of transitions for which we cannot decide enabledness
	forbidden     map[string]struct{} // name of transitions in forbidEnabled

}

// NewStepper returns a fresh Stepper starting with the initial marking of n
func NewStepper(n *Net) *Stepper {
	m0 := State{
		COL:       make(pnml.Marking, len(n.Places)),
		PT:        make(map[string]int),
		Enabled:   make(map[string]bool),
		Witnesses: make([]*Witness, len(n.Trans)),
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

	s.iter = make([]Iterator, len(n.Trans))
	// we should compute enabled after testing reachability queries on the initial marking.
	for k := range s.Trans {
		s.iter[k] = s.newIterator(k)
	}
	s.ComputeEnabled()

	return &s
}

func (s *Stepper) String() string {
	res := s.printMarkingAligned(s.State, s.lpn, 90)
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
	res := ""
	for k, v := range s.Places {
		if len(m[k]) == 0 {
			continue
		}
		res += fmt.Sprintf("%s : %s\n", v.Name, s.PrintHue(m[k]))
	}
	return res
}

// ExistMatch reports if there is a set of values that match the condition of
// the transition with index t.
func (s *Stepper) ExistMatch(t int) (bool, error) {
	s.iter[t].Reset()

	w, ok, err := s.iter[t].Check(s.COL)

	if err != nil {
		return false, err
	}

	if ok {
		s.State.Witnesses[t] = w
		return true, nil
	}
	return false, nil
}

func (s *Stepper) PrintWitnesses() {
	for k := range s.Trans {
		s.PrintWitness(k)
	}
}

// PrintWitness displays information about a witness that the transition with
// index t is enabled.
func (s *Stepper) PrintWitness(t int) {
	if _, ok := s.forbidEnabled[t]; ok {
		return
	}
	if len(s.Witnesses) == 0 {
		return
	}
	if !s.Enabled[s.Trans[t].Name] {
		return
	}
	fmt.Println("----------------------------------")
	fmt.Printf("%s enabled\n", s.Trans[t].Name)
	fmt.Println("witness:")
	fmt.Print(s.PrintCOLShort(s.showWitness(t)))
	fmt.Println(s.iter[t].PrintVEnv(s.Net))
	fmt.Println("----------------------------------")
}

// showWitness returns a colored marking corresponding to a witness for
// transition with index t.
func (s *Stepper) showWitness(t int) pnml.Marking {
	res := make(pnml.Marking, len(s.COL))
	w := s.Witnesses[t]
	for i := range w.Places {
		h := s.COL[i]
		hh := make(pnml.Hue, len(w.Pre[i]))
		for k, match := range w.Pre[i] {
			hh[k] = pnml.Atom{
				Value: h[match.pos].Value,
				Mult:  match.mult,
			}
		}
		res[i] = hh
	}
	return res
}

// Enabled returns the set of transition (a slice of ordered indexes in n.Trans)
// which are enabled at the current marking. As a side effect, it also add a
// list of witnesses in s for each of the enabled transitions.
func (s *Stepper) ComputeEnabled() {
	// We could be more clever and only update the transition whose input places
	// have been modified.
	for tname := range s.TPosition {
		s.Enabled[tname] = false
	}
	for k, t := range s.Trans {
		if _, ok := s.forbidEnabled[k]; ok {
			continue
		}
		ok, err := s.ExistMatch(k)

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
func (s *Stepper) FireAtRandom() error {
	choose := []string{}
	for k, v := range s.Enabled {
		if v {
			choose = append(choose, k)
		}
	}

	if len(choose) == 0 {
		// return fmt.Errorf("nothing to fire")
		s.update(s.m0)
		s.ComputeEnabled()
		return fmt.Errorf("nothing to fire; restarting")
	}

	tname := choose[rand.Intn(len(choose))]

	s.Fire(s.TPosition[tname])
	return nil
}

// Fire updates the stepper with the result of firing transition t. If there are
// no witnesses, we try to compute one. We do nothing when t is not enabled.
func (s *Stepper) Fire(t int) {
	if _, ok := s.forbidFiring[t]; ok {
		return
	}
	if s.Witnesses[t] == nil {
		s.ComputeEnabled()
	}
	if !s.Enabled[s.Trans[t].Name] {
		return
	}
	s.update(s.fire(t))
	s.ComputeEnabled()
	// fmt.Println("----------------------------------")
	// fmt.Printf("[firing: %s]\n", s.Trans[t].Name)
	// fmt.Println(s.PrintCOL(s.COL))
	// fmt.Println(s.PrintEnabled())
	// fmt.Println("----------------------------------")
}

// update changes the marking to m in the stepper
func (s *Stepper) update(m pnml.Marking) {
	s.COL = m
	for k, v := range s.Places {
		s.PT[v.Name] = m[k].Sum()
	}
}

func (s *Stepper) fire(t int) pnml.Marking {
	w := s.Witnesses[t]
	m1 := w.ApplyPreconditions(s.COL)
	tr := s.Trans[t]
	// We add the Post.
	for _, a := range tr.Outs {
		for _, e := range a.Pattern {
			m1[a.Place] = append(m1[a.Place], e.Eval(w.Net.Net, w.Assoc)...)
		}
	}
	// We need to shorten the marking but also to remove duplicates. We
	// start to sort the slice to simplify the logic.
	for k, h := range m1 {
		sort.Slice(h, func(i, j int) bool { return pnml.AtomIsLess(h[i], h[j]) })
		removed := 0
		j := 0
		for i := 0; i < len(h); i++ {
			if h[i].Mult == 0 {
				removed++
				continue
			}
			if (j != 0) && (h[i].Value == h[j-1].Value) {
				removed++
				h[j-1].Mult += h[j].Mult
				continue
			}
			h[j] = h[i]
			j++
		}
		m1[k] = h[:len(m1[k])-removed]
	}
	return m1
}
