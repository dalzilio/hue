// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package hlnet

import (
	"fmt"
	"strings"

	"github.com/dalzilio/hue/pkg/internal/util"
	"github.com/dalzilio/hue/pkg/pnml"
)

// Marking is the type of High-Level Nets (hlnet) markings. It is a slice
// containing the PMarking of each places in the same order than in Places. We
// also use hashmaps to have the "compound" marking for each place (useful for
// computing tokens-count) and the list of enabled transitions (for
// is-fireable), which helps simplify the evaluation of formulas.
type Marking struct {
	COL       []pnml.Hue      // Colored marking
	PT        map[string]int  // Compound P/T marking
	Enabled   map[string]bool // list of enabled transition (after ComputeEnabled)
	Witnesses []*pnml.Witness // list of witnesses that can be used to fire transitions
}

// ----------------------------------------------------------------------

// func (net *Net) PrintMarking(m Marking) string {
// 	return net.PrintCOL(m.COL)
// }

func (net *Net) PrintEnabled(m Marking) string {
	s := make([]string, len(net.Trans))
	for k, v := range m.Enabled {
		if v {
			s[net.TPosition[k]] = fmt.Sprintf("%s(+)", k)
		} else {
			s[net.TPosition[k]] = fmt.Sprintf("%s(-)", k)
		}
	}
	return util.ZipString(s, "", "", " ")
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
