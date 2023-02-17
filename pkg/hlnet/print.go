// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

import (
	"fmt"

	"github.com/dalzilio/hue/pkg/pnml"
)

func zipString(a []string, start, end, sep string) string {
	res := start
	for k, aa := range a {
		if k != 0 {
			res += sep
		}
		res += aa
	}
	return res + end
}

func zipPrint[T fmt.Stringer](a []T, start, end, sep string) string {
	res := start
	for k, aa := range a {
		if k != 0 {
			res += sep
		}
		res += aa.String()
	}
	return res + end
}

func (net Net) String() string {
	// s := fmt.Sprintf("# net %s\n", net.Name)
	s := fmt.Sprintf("%s\n\n", net.Net)
	for _, v := range net.Places {
		s += fmt.Sprintf("# pl %s : %s\n", v.Name, v.Type)
	}
	for _, v := range net.Trans {
		s += fmt.Sprintf("# tr %s %s %s\n", v.Name, v.Cond, v.Env)
		for _, e := range v.Ins {
			s += fmt.Sprintf("#\t%s -->( %s )\n", net.Places[e.Place].Name, zipPrint(e.Pattern, "[", "]", ","))
		}
		for _, e := range v.Outs {
			s += fmt.Sprintf("#\t%s <--( %s )\n", net.Places[e.Place].Name, zipPrint(e.Pattern, "[", "]", ","))
		}
	}
	return s
}

// ----------------------------------------------------------------------

// Tina outputs an hlnet in a format that can be displayed by Tina' s nd tool.
func (net Net) Tina() string {
	s := fmt.Sprintf("net {%s}\n", net.Name)
	for _, v := range net.Places {
		if is := net.PrintHue(v.Init); is != "-" {
			s += fmt.Sprintf("pl {%s} : {%s} (1)\n", v.Name, is)
		} else {
			s += fmt.Sprintf("pl {%s}\n", v.Name)
		}
	}
	notes := 1
	for _, v := range net.Trans {
		if v.Cond.Op == pnml.NIL {
			s += fmt.Sprintf("tr {%s}", v.Name)
		} else {
			s += fmt.Sprintf("tr {%s} : {%s} ", v.Name, v.Cond)
		}
		for _, e := range v.Ins {
			s += fmt.Sprintf(" {%s}", net.Places[e.Place].Name)
		}
		s += " -> "
		for _, e := range v.Outs {
			s += fmt.Sprintf(" {%s}", net.Places[e.Place].Name)
		}
		s += "\n"
		// we output the Pattern of the edges as a comment because it is not
		// possible to draw them with Tina's nd tool.
		s += fmt.Sprintf("nt n%d 1 {tr %s :", notes, v.Name)
		notes++
		for _, e := range v.Ins {
			if e.Pattern != nil {
				s += fmt.Sprintf("\\\\n  %s ---( %s )--> ", net.Places[e.Place].Name, e.Pattern)
			}
		}
		for _, e := range v.Outs {
			if e.Pattern != nil {
				s += fmt.Sprintf("\\\\n  %s <--( %s )--- ", net.Places[e.Place].Name, e.Pattern)
			}
		}
		s += "}\n"
	}
	return s
}
