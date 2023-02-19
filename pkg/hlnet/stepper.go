// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"
	"sort"

	"github.com/dalzilio/hue/pkg/pnml"
)

// Stepper is the type of marked hlnet, that is a pair consisting of a
// high-level net and its current marking
type Stepper struct {
	*Net
	Marking
	iter         []pnml.Iterator  // we keep iterators for each transitions in the net
	lpn          int              // length of the longest place name, for pretty printing markings
	showwitness  bool             // flag to print witness when we find a fireable transition
	forbidFiring map[int]struct{} // index of transitions we should never fire
}

// NewStepper returns a fresh Stepper starting with the initial marking of n
func NewStepper(n *Net, showwitness bool, needfireable bool) (*Stepper, error) {
	m0 := Marking{
		COL:       make([]pnml.Hue, len(n.Places)),
		PT:        make(map[string]int),
		Enabled:   make(map[string]bool),
		Witnesses: make([]*pnml.Witness, len(n.Trans)),
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
		Net:          n,
		Marking:      m0,
		iter:         nil,
		lpn:          lpn,
		showwitness:  showwitness,
		forbidFiring: make(map[int]struct{}),
	}

	if needfireable || showwitness {
		s.iter = make([]pnml.Iterator, len(n.Trans))
		// we should compute enabled after testing reachability queries on the initial marking.
		for k, t := range s.Trans {
			// We collect the patterns and the list of places in the input arcs.
			// We compute their environment to detect the case where we have
			// extra variables in on the output arc (see DatabaseWithMutex)
			inse := [][]pnml.Expression{}
			insm := []int{}
			insEnv := pnml.Env{}
			for _, a := range t.Ins {
				inse = append(inse, a.Pattern)
				insm = append(insm, a.Place)
				for _, e := range a.Pattern {
					insEnv = e.AddEnv(insEnv)
				}
			}
			// if we have extra variables, we must prevent the stepper from firing this transition.
			if len(t.Env) > len(insEnv) {
				s.forbidFiring[k] = struct{}{}
			}
			var err error
			s.iter[k], err = pnml.NewIterator(n.Net, t.Name, t.Cond, inse, insm)
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

func (s *Stepper) PrintCOL(m []pnml.Hue) string {
	res := ""
	for k, v := range s.Places {
		res += fmt.Sprintf("%s : %s\n", v.Name, s.PrintHue(m[k]))
	}
	return res
}

// ExistMatch reports if there is a set of values that match the condition of
// the transition with index t.
func (s *Stepper) ExistMatch(t int) bool {
	s.iter[t].Reset()
	if w, ok := s.iter[t].Check(s.COL); ok {
		if s.showwitness {
			fmt.Println("----------------------------------")
			fmt.Printf("%s enabled\n", s.Trans[t].Name)
			fmt.Println("witness:")
			fmt.Print(s.PrintCOL(w.ShowWitness(s.COL)))
			fmt.Println(s.iter[t].PrintVEnv(s.Net.Net))
			fmt.Println("----------------------------------")
		}
		s.Marking.Witnesses[t] = w
		return true
	}
	return false
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
		if s.ExistMatch(k) {
			s.Enabled[t.Name] = true
		}
	}
}

// Fire updates the stepper with the result of firing transition t. If there are
// no witnesses, we try to compute one. We do nothing when t is not enabled.
func (s *Stepper) Fire(t int) {
	if s.Witnesses[t] == nil {
		s.ComputeEnabled()
	}
	if !s.Enabled[s.iter[t].Name] {
		return
	}
	if _, ok := s.forbidFiring[t]; ok {
		return
	}
	s.fire(t)
}

func (s *Stepper) fire(t int) {
	w := s.Witnesses[t]
	m1 := w.ApplyPreconditions(s.COL)
	tr := s.Trans[t]
	// We add the Post.
	for _, a := range tr.Outs {
		for _, e := range a.Pattern {
			m1[a.Place] = append(m1[a.Place], e.Eval(w.Net, w.Assoc)...)
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
}
