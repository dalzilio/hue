// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package formula

import "fmt"

func (p *Query) String() string {
	s := fmt.Sprintf("Property %s :\n", p.ID)
	if p.IsEF {
		s += fmt.Sprintf("EF %v\n", p.Formula)
	} else {
		s += fmt.Sprintf("AG %v\n", p.Formula)
	}
	return s
}

func PrintQueries(queries []Query) string {
	s := ""
	for k, v := range queries {
		if k != 0 {
			s += "----------------------------------\n"
		}
		s += v.String()
	}
	return s
}
