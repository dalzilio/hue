// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import "github.com/dalzilio/hue/pkg/pnml"

// IteratorMatch is a slice of (mult x position) that can be used to reconstruct
// the tokens that have been match by an iterator on a place marking
type IteratorMatch []struct {
	pos  int
	mult int
}

// insert a match point keeping as an invariant that values are sorted in order
// of position.
func insertMatchPoint(m IteratorMatch, pos int, mult int) IteratorMatch {
	for k := range m {
		if m[k].pos == pos {
			m[k].mult += mult
			return m
		}
		if m[k].pos > pos {
			// insert a new element at position k
			return append(m[:k], append(IteratorMatch{{pos: pos, mult: mult}}, m[k:]...)...)
		}
	}
	return append(m, struct {
		pos  int
		mult int
	}{pos: pos, mult: mult})
}

// testCapacity takes the sum of multiplicities of tokens at the same positions
// (given by a.pos) and check if it is less than the multiplicy found in h. If
// so we have a valid IteratorMatch.
func (a *arcIterator) testCapacity(h pnml.Hue) bool {
	a.match = a.match[:0]
	for i := range a.pos {
		a.match = insertMatchPoint(a.match, a.pos[i], a.mults[i])
	}
	for k := range a.match {
		if a.match[k].mult > h[a.match[k].pos].Mult {
			a.match = a.match[:0]
			return false
		}
	}
	return true
}

// ----------------------------------------------------------------------

// Witness are values that can be used to prove that a transition is enabled and
// to compute a marking after firing a transition
type Witness struct {
	*Iterator
	Tr     int             // index of arc in the iterator
	Places []int           // list of input places
	Pre    []IteratorMatch // tokens for the pre-condition
	Assoc  pnml.VEnv       // values for the variables
}

// ApplyPreconditions returns a marking where we removed the necessary tokens.
// We assume m is equal or larger than the marking used to instantiate the
// witness. We do not compact the marking to eliminate  atoms with multiplicity
// 0 since it will be done after we add the post.
func (w *Witness) ApplyPreconditions(m pnml.Marking) pnml.Marking {
	m1 := m.Clone()
	for k, pl := range w.Places {
		// we remove the Pre.
		if pre := w.Pre[k]; len(pre) != 0 {
			for _, p := range pre {
				m1[pl][p.pos].Mult -= p.mult
			}
		}
	}
	return m1
}

// GetWitness returns a witness for the current match. The function should be
// used right after a call to Check. We return nil if iter did not match.
func (iter *Iterator) GetWitness(m pnml.Marking) *Witness {
	res := Witness{
		Iterator: iter,
		Places:   make([]int, len(iter.arcs)),
		Pre:      make([]IteratorMatch, len(iter.arcs)),
		Assoc:    iter.Venv(),
	}
	for i, a := range iter.arcs {
		res.Places[i] = a.pl
		res.Pre[i] = make(IteratorMatch, len(a.match))
		copy(res.Pre[i], a.match)
	}
	return &res
}
