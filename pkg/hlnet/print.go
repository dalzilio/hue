// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"
)

func (net Net) String() string {
	s := fmt.Sprintf("# net %s\n", net.Name)
	for _, v := range net.Places {
		s += fmt.Sprintf("# pl %s %s\n", v.Name, pInit(v.Init))
	}
	for _, v := range net.Trans {
		s += fmt.Sprintf("# tr %s %s %s\n", v.Name, v.Cond, v.Env)
		for _, e := range v.Ins {
			s += fmt.Sprintf("#\t%s -->( %s )\n", net.Places[e.Place].Name, e.Pattern)
		}
		for _, e := range v.Outs {
			s += fmt.Sprintf("#\t%s <--( %s )\n", net.Places[e.Place].Name, e.Pattern)
		}
	}
	return s
}

// ----------------------------------------------------------------------

// // Tina outputs an hlnet in a format that can be displayed by Tina' s nd tool.
// func (net Net) Tina() string {
// 	s := fmt.Sprintf("net {%s}\n", net.Name)
// 	for _, v := range net.Places {
// 		isstable := ""
// 		if v.Stable {
// 			isstable = "(stable) "
// 		}
// 		if is := pInit(v.Init); is != "-" {
// 			s += fmt.Sprintf("pl {%s} : {%s%s} (1)\n", v.Name, isstable, is)
// 		} else {
// 			s += fmt.Sprintf("pl {%s}\n", v.Name)
// 		}
// 	}
// 	notes := 1
// 	for _, v := range net.Trans {
// 		if v.Cond.Op == pnml.NIL {
// 			s += fmt.Sprintf("tr {%s}", v.Name)
// 		} else {
// 			s += fmt.Sprintf("tr {%s} : {%s} ", v.Name, v.Cond)
// 		}
// 		for _, e := range v.Ins {
// 			s += fmt.Sprintf(" {%s}", net.Places[e.Place].Name)
// 		}
// 		s += " -> "
// 		for _, e := range v.Outs {
// 			s += fmt.Sprintf(" {%s}", net.Places[e.Place].Name)
// 		}
// 		s += "\n"
// 		// we output the Pattern of the edges as a comment because it is not
// 		// possible to draw them with Tina's nd tool.
// 		note := fmt.Sprintf("nt n%d 1 {tr %s :", notes, k)
// 		notes++
// 		for k, e := range v.Ins {
// 			if e.Pattern != nil {
// 				s += fmt.Sprintf("#  %s ---( %s )--> ", net.Places[e.Place].Name, e.Pattern)
// 			}
// 			if k != 0 {
// 				s += "\n"
// 			}
// 		}
// 		for _, e := range v.Outs {
// 			if e.Pattern != nil {
// 				s += fmt.Sprintf("#  %s <--( %s )--- \n", net.Places[e.Place].Name, e.Pattern)
// 			}
// 		}
// 		s += "\n" + note + "}\n"
// 	}
// 	return s
// }
