// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

// OP is the type of operations in a PNML expression
type OP int

const (
	NIL OP = iota
	EQ
	INEQ
	LESSTHAN
	LESSTHANEQ
	GREATTHAN
	GREATTHANEQ
	OR
	AND
)

// ----------------------------------------------------------------------

// TYP describes the possible kind of PNML types in PNML.
type TYP int

const (
	DOT TYP = iota
	CENUM
	FENUM
	FINTRANGE
	PROD
	NUMERIC
)

// ----------------------------------------------------------------------

func getpOP(s string) OP {
	switch s {
	case "equality":
		return EQ
	case "inequality":
		return INEQ
	case "lessthan":
		return LESSTHAN
	case "lessthanorequal":
		return LESSTHANEQ
	case "greaterthan":
		return GREATTHAN
	case "greaterthanorequal":
		return GREATTHANEQ
	case "or":
		return OR
	case "and":
		return AND
	}
	panic("not an operation " + s)
}
