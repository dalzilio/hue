// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/dalzilio/hue/pkg/formula"
	"github.com/dalzilio/hue/pkg/internal/util"
	"github.com/dalzilio/hue/pkg/pnml"
)

// Stepper is the type of marked hlnet, that is a pair consisting of a
// high-level net and its current marking
type Stepper struct {
	steppermut sync.Mutex // to protect for concurrent access when we need to copy m0 or update forbidEnabled
	*Net
	m0            *State              // initial state
	lpn           int                 // length of the longest place name, for pretty printing markings
	forbidFiring  map[int]struct{}    // index of transitions we should never fire
	forbidEnabled map[int]struct{}    // index of transitions for which we cannot decide enabledness
	forbidden     map[string]struct{} // name of transitions in forbidEnabled

}

// Worker includes a Stepper, a list of iterators used for checking that
// transitions are enabled (we use different iterators on different workers to
// avoid concurrent writes) and a current state.
type Worker struct {
	*Stepper
	iter []*Iterator // we keep iterators for each transitions in the net
	*State
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
		m0:            &m0,
		lpn:           lpn,
		forbidFiring:  make(map[int]struct{}),
		forbidEnabled: make(map[int]struct{}),
		forbidden:     make(map[string]struct{}),
	}

	return &s
}

// NewWorker returns a fresh Worker initialized with m0
func NewWorker(s *Stepper) *Worker {
	w := Worker{Stepper: s}
	w.iter = make([]*Iterator, len(s.Trans))
	for k := range s.Trans {
		w.iter[k] = s.newIterator(k)
	}
	s.steppermut.Lock()
	w.State = w.m0.Clone()
	// if we never computed the enabledness relation before
	if w.m0.Enabled == nil {
		w.Enabled = make(map[string]formula.Bool)
		w.computeEnabled()
	}
	// we copy the result back to m0 for later use
	w.m0 = w.State.Clone()
	s.steppermut.Unlock()
	return &w
}

func (w *Worker) Initialize() {
	if w.Enabled != nil {
		return
	}
	w.steppermut.Lock()
	if w.m0.Enabled != nil {
		w.State = w.m0.Clone()
	} else {
		w.Enabled = make(map[string]formula.Bool)
		w.computeEnabled()
	}
	w.steppermut.Unlock()
}

func (w *Worker) String() string {
	res := w.printMarkingAligned(w.State, w.Stepper.lpn, 65, 25)
	if len(w.iter) != 0 {
		res += "Fireable: " + w.PrintEnabled()
	}
	return res
}

func (w *Worker) PrintCOL(m pnml.Marking) string {
	res := ""
	for k, v := range w.Places {
		res += fmt.Sprintf("%s : %s\n", v.Name, w.PrintHue(m[k]))
	}
	return res
}

func (w *Worker) PrintCOLShort(m pnml.Marking) string {
	res := []string{}
	for k, v := range w.Places {
		if len(m[k]) == 0 {
			continue
		}
		res = append(res, fmt.Sprintf("%s: %s", v.Name, w.PrintHue(m[k])))
	}
	return util.ZipString(res, "[", "]", ", ")
}

// Enabled returns the set of transitions which are enabled at the current
// marking. As a side effect, it also add a list of witnesses in w for each of
// the enabled transitions.
func (w *Worker) computeEnabled() {
	// We could be more clever and only update the transition whose input places
	// have been modified.
	for tname := range w.TPosition {
		w.Enabled[tname] = formula.FALSE
	}
	for tname := range w.forbidden {
		w.Enabled[tname] = formula.UNDEF
	}
	for k, t := range w.Trans {
		if _, ok := w.forbidEnabled[k]; ok {
			continue
		}
		ok, err := w.check(k)
		if err != nil {
			// BEWARE!!: possible concurrent write to Stepper
			w.steppermut.Lock()
			w.forbidEnabled[k] = struct{}{}
			w.forbidden[t.Name] = struct{}{}
			w.steppermut.Unlock()
			continue
		}

		if ok {
			w.Enabled[t.Name] = formula.TRUE
		}
	}
}

// FireAtRandom chooses one of the enabled transitions and fires it. If we have
// no transitions to fire, we start again from the initial marking.
func (w *Worker) FireAtRandom(verbose bool) error {
	if w.Enabled == nil {
		w.Enabled = make(map[string]formula.Bool)
		w.computeEnabled()
	}

	choose := []string{}
	for k, v := range w.Enabled {
		if v.IsTrue() {
			// we also need to check that we can fire the transition
			if _, ok := w.forbidFiring[w.TPosition[k]]; !ok {
				choose = append(choose, k)
			}
		}
	}

	if len(choose) == 0 {
		w.Restart(false)
		if verbose {
			fmt.Println("----------------------------------")
			if len(w.Enabled) == 0 {
				fmt.Printf("[deadlock ; restarting]\n")
			} else {
				fmt.Printf("[only \"forbidden\" transitions ; restarting]\n")
			}
			fmt.Println("----------------------------------")
		}
		return fmt.Errorf("nothing to fire; restarting")
	}

	tname := choose[rand.Intn(len(choose))]

	w.Fire(w.TPosition[tname], verbose)
	return nil
}

// Restart starts exploration again from the initial marking
func (w *Worker) Restart(verbose bool) {
	w.restart()
	w.computeEnabled()
	if verbose {
		fmt.Println("----------------------------------")
		fmt.Println("[ repeat-limit ; restarting ]")
		fmt.Println("----------------------------------")
	}
}

// Fire updates the stepper with the result of firing transition t. If there are
// no witnesses, we try to compute one. We do nothing when t is not enabled.
func (w *Worker) Fire(k int, verbose bool) {
	if _, ok := w.forbidFiring[k]; ok {
		return
	}
	if w.After[k] == nil {
		w.computeEnabled()
	}
	if !w.Enabled[w.Trans[k].Name].IsTrue() {
		return
	}
	if verbose {
		fmt.Println("----------------------------------")
		fmt.Printf("%s %s\n", w.Trans[k].Name, w.iter[k].PrintVEnv(w.Net))
		fmt.Println("----------------------------------")
	}
	w.update(w.After[k])
}

// restart updates the State of w with the initial marking; m0 in the stepper.
func (w *Worker) restart() {
	w.steppermut.Lock()
	w.State = w.m0.Clone()
	w.steppermut.Unlock()
}

// update changes the State of w with the colored marking m.
func (w *Worker) update(m pnml.Marking) {
	w.COL = m.Clone()
	for k, v := range w.Places {
		w.PT[v.Name] = m[k].Sum()
	}
	w.computeEnabled()
}
