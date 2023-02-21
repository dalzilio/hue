// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import (
	"sort"

	"github.com/dalzilio/hue/pkg/pnml"
)

// testCapacity takes the sum of multiplicities of tokens at the same positions
// (given by a.pos) and check if it is less than the multiplicy found in h. If
// so we have a valid IteratorMatch.
func (a *arcIterator) testCapacity(h pnml.Hue) bool {
	z := make(map[int]int)
	// we start by merging duplicates (same pos) in the matched values
	for i := range a.pos {
		z[a.pos[i]] = z[a.pos[i]] + a.mults[i]
	}
	for k, v := range z {
		if h[k].Mult < v {
			return false
		}
	}
	return true
}

// getWitness returns a witness for the current match on transition k.
func (s *Stepper) getWitness(k int) pnml.Marking {
	m1 := s.COL.Clone()
	it := s.iter[k]
	tr := s.Trans[k]

	// We start by removing the Pre.
	for _, a := range it.arcs {
		for j, p := range a.pos {
			m1[a.pl][p].Mult -= a.mults[j]
		}
	}

	// We add the Post.
	for _, a := range tr.Outs {
		for _, e := range a.Pattern {
			m1[a.Place] = append(m1[a.Place], e.Eval(s.Net.Net, it.venv)...)
		}
	}

	// We need to shorten the marking but also to remove duplicates. We
	// start to sort the slice to simplify the logic.
	for pl, h := range m1 {
		sort.Slice(h, func(a, b int) bool { return pnml.AtomIsLess(h[a], h[b]) })
		removed := 0
		j := 0
		for i := 0; i < len(h); i++ {
			if h[i].Mult == 0 {
				removed++
				continue
			}
			if (j != 0) && (h[i].Value == h[j-1].Value) {
				removed++
				h[j-1].Mult += h[i].Mult
				continue
			}
			h[j] = h[i]
			j++
		}
		m1[pl] = h[:len(m1[pl])-removed]
	}
	return m1
}
