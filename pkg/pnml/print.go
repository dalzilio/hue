// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package pnml

import "fmt"

func (net Net) String() string {
	s := ""
	for _, v := range net.Declaration.Sorts {
		s += "\n# type " + v.String()
	}
	for _, v := range net.Declaration.Partitions {
		s += "\n# partition " + v.ID + " : " + v.Type.ID
		for _, p := range v.Partitions {
			s += "\n#  +-- " + p.ID + " : " + p.Elem[0].ID + " ... " + p.Elem[len(p.Elem)-1].ID
		}
	}
	for _, v := range net.Declaration.Vars {
		s += "\n# var " + v.ID + " : " + v.Type.ID
	}
	return s
}

func (typ TypeDecl) String() string {
	var s = typ.ID + " : "
	switch {
	case typ.Sort == FINTRANGE:
		s += fmt.Sprintf("IntRange %d -- %d", typ.FIntRan.Start, typ.FIntRan.End)
	case (typ.Sort == CENUM) || (typ.Sort == FENUM):
		switch len(typ.Elem) {
		case 1:
			s += typ.Elem[0]
		case 2:
			s += typ.Elem[0] + " " + typ.Elem[1]
		default:
			s += typ.Elem[0] + " ... " + typ.Elem[len(typ.Elem)-1]
		}
	case typ.Sort == PROD:
		s += "("
		if typ.Product == nil {
			s += ")"
			break
		}
		s += typ.Product[0].ID
		for i := 1; i < len(typ.Product); i++ {
			s += " x " + typ.Product[i].ID
		}
		s += ")"
	default:
		s += " dot"
	}
	return s
}
