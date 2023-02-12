// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"
	"sort"

	"github.com/dalzilio/hue/pkg/pnml"
)

// Net is the concrete type of symmetric nets.
type Net struct {
	Name      string
	PPosition map[string]int // gives the index of a place in Places
	TPosition map[string]int // gives the index of a transition in Trans
	Places    []*Place
	Trans     []*Transition
}

// Place is the concrete type of symmetric nets places.
type Place struct {
	Name   string
	Init   pnml.Expression
	Type   string
	Stable bool
}

// Transition is the concrete type of symmetric nets transitions.
type Transition struct {
	Name string
	Cond pnml.Operation
	Env  pnml.Env
	Ins  []*Arcs
	Outs []*Arcs
}

// Arcs is the concrete type of symmetric nets arcs.
type Arcs struct {
	Pattern pnml.Expression
	Place   int
}

// ----------------------------------------------------------------------

func pInit(p pnml.Expression) string {
	if p == nil {
		return "-"
	}
	return p.String()
}

// ----------------------------------------------------------------------

// Build returns an hlnet from a PNML net. This structure is easier to deal
// with.
func Build(n *pnml.Net) (*Net, error) {

	var net = Net{Name: n.Name}

	// We build the list of places in lexicographic order by first sorting the
	// Places in the PNML page.
	pnames := make([]string, len(n.Page.Places))
	for k, p := range n.Page.Places {
		pnames[k] = p.ID
	}
	sort.Strings(pnames)
	net.PPosition = make(map[string]int, len(pnames))
	for k, name := range pnames {
		net.PPosition[name] = k
	}
	net.Places = make([]*Place, len(pnames))
	for _, p := range n.Page.Places {
		net.Places[net.PPosition[p.ID]] = &Place{Name: p.ID, Init: p.InitialMarking, Type: p.Type.ID}
	}

	// Next we do the same with transitions.
	tnames := make([]string, len(n.Page.Trans))
	for k, p := range n.Page.Trans {
		tnames[k] = p.ID
	}
	sort.Strings(tnames)
	net.TPosition = make(map[string]int, len(tnames))
	for k, name := range tnames {
		net.TPosition[name] = k
	}
	net.Trans = make([]*Transition, len(tnames))
	for _, t := range n.Page.Trans {
		env := make(pnml.Env)
		var cond pnml.Operation
		if t.Condition == nil {
			cond = pnml.Operation{Op: pnml.NIL}
		} else {
			cond = t.Condition.(pnml.Operation)
		}
		cond.AddEnv(env)
		for varname := range env {
			if _, ok := n.Env[varname]; !ok {
				return nil, fmt.Errorf("variable \"%s\" used in condition of transition %s was never declared", varname, t.ID)
			}
		}
		net.Trans[net.TPosition[t.ID]] = &Transition{Name: t.ID, Cond: cond, Env: env}
	}

	for _, a := range n.Page.Arcs {
		e := Arcs{Pattern: a.Pattern}
		if e.Pattern == nil {
			e.Pattern = pnml.Operation{Op: pnml.NIL}
		}
		env := make(pnml.Env)
		e.Pattern.AddEnv(env)
		for varname := range env {
			if _, ok := n.Env[varname]; !ok {
				return nil, fmt.Errorf("variable \"%s\" used in pattern of arc from %s to %s was never declared", varname, a.Source, a.Target)
			}
		}
		if p, ok := net.PPosition[a.Source]; ok {
			// arc source is a place, target is a transition. The edge is an
			// input arc (Ins). We add the variables in the pattern to env.
			t := net.Trans[net.TPosition[a.Target]]
			e.Pattern.AddEnv(t.Env)
			e.Place = p
			t.Ins = append(t.Ins, &e)
		}
		if p, ok := net.PPosition[a.Target]; ok {
			// arc source is a transition, target is a place. The edge is an
			// output arc.
			t := net.Trans[net.TPosition[a.Source]]
			e.Pattern.AddEnv(t.Env)
			e.Place = p
			t.Outs = append(t.Outs, &e)
		}
	}

	// if there are less than 200 places (we chose an arbitrary value) we try to
	// find "stable places", that is a place p such that, for every transition
	// t, there is an edge t--exp-->p if and only if there is an edge
	// p--exp-->t. When a place is stable we know that the tokens in place p
	// (the possible pairs p x val) are exactly the same than in the initial
	// marking.
	if len(net.Places) < 200 {
		for _, p := range net.Places {
			p.Stable = net.IsPlaceStable(p)
		}
	}

	return &net, nil
}
