// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import (
	"fmt"
	"strings"

	"github.com/dalzilio/hue/pkg/pnml"
)

// State is the type of High-Level Nets (hlnet) markings. It is a slice
// containing the Marking of each places in the same order than in Places. We
// also use hashmaps to have the "compound" marking for each place (useful for
// computing tokens-count) and the list of enabled transitions (for
// is-fireable), which helps simplify the evaluation of formulas.
type State struct {
	COL       pnml.Marking    // Colored marking
	PT        map[string]int  // Compound P/T marking
	Enabled   map[string]bool // list of enabled transition (after ComputeEnabled)
	Witnesses []*Witness      // list of witnesses that can be used to fire transitions
}

// ----------------------------------------------------------------------

func (s *Stepper) PrintEnabled() string {
	res := ""
	for k, t := range s.Trans {
		if _, ok := s.forbidEnabled[k]; ok {
			res += fmt.Sprintf("%s(?) ", t.Name)
			continue
		}
		if v := s.Enabled[t.Name]; v {
			res += fmt.Sprintf("%s(+) ", t.Name)
		} else {
			res += fmt.Sprintf("%s(-) ", t.Name)
		}
	}
	return res
}

func (net *Net) printMarkingAligned(m State, left int, trunc int) string {
	s := ""
	for k, v := range net.Places {
		mm := []rune(net.PrintHue(m.COL[k]))
		if len(mm) > trunc {
			mm = mm[:trunc]
			mm = append(mm, []rune("...")...)
		}
		s += fmt.Sprintf("%s%s : %s\n", v.Name, strings.Repeat(" ", left-len(v.Name)), string(mm))
	}
	return s
}
