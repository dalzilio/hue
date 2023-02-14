// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"
	"strings"

	"github.com/dalzilio/hue/pkg/pnml"
)

// Marking is the type of High-Level Nets (hlnet) markings. It is a slice
// containing the PMarking of each places in the same order than in Places. We
// also use hashmaps to have the "compound" marking for each place
// (tokens-count) and the list of enabled transitions (for is-fireable), which
// helps simplify the evaluation of formulas.
type Marking struct {
	COL     []pnml.Hue
	PT      map[string]int
	Enabled map[string]bool
}

// ----------------------------------------------------------------------

func (net *Net) PrintMarking(m Marking) string {
	s := ""
	for k, v := range net.Places {
		s += fmt.Sprintf("%s : %s\n", v.Name, net.PrintHue(m.COL[k]))
	}
	return s
}

func (net *Net) printMarkingAligned(m Marking, left int, trunc int) string {
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
