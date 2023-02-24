// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package formula

import (
	"fmt"

	"github.com/dalzilio/hue/pkg/internal/util"
)

// Query is the type of reachability queries we need to check. When IsEF is true
// the query is (EF phi), meaning we report  true if we can reach a state where
// phi is true. Conversely, when IsEF is false, the query is (AG phi), meaning
// we report false if (not phi) is reachable.
type Query struct {
	ID       string
	IsEF     bool    // true if we have a formula of the form (EF phi)
	Original Formula // formula before simplification. For debugging purpose
	Formula
}

// Formula is the interface that wraps the type of XML Formula.
//
// Includes reports if one of the constant in the formula is part of the set of
// names
type Formula interface {
	String() string
}

// ----------------------------------------------------------------------

// BooleanConstant defines a constant True or False.
type BooleanConstant bool

func (f BooleanConstant) String() string {
	if f {
		return "true"
	}
	return "false"
}

// ----------------------------------------------------------------------

// Negation is the type of all expressions.
type Negation struct {
	Formula
}

func (f Negation) String() string {
	return "(not " + f.Formula.String() + ")"
}

// ----------------------------------------------------------------------

// Disjunction is the disjunction of two or more formulas.
type Disjunction []Formula

func (f Disjunction) String() string {
	return util.ZipPrint(f, "(or ", ")", " ")
}

// ----------------------------------------------------------------------

// Conjunction is the conjunction of two or more formulas.
type Conjunction []Formula

func (f Conjunction) String() string {
	return util.ZipPrint(f, "(and ", ")", " ")
}

// ----------------------------------------------------------------------

// IntegerLe returns true if the value of the Left integer expression is less
// than or equal to the value of the Right one.
type IntegerLe struct {
	Left  Formula
	Right Formula
}

func (f IntegerLe) String() string {
	return fmt.Sprintf("(<= %s %s)", f.Left.String(), f.Right.String())
}

// ----------------------------------------------------------------------

// IsFireable evaluates to true if one of the declared transition is fireable.
type IsFireable []string

func (f IsFireable) String() string {
	if len(f) == 1 {
		return f[0]
	}
	return util.ZipString(f, "(or ", ")", " ")
}

// ----------------------------------------------------------------------

// TokensCount returns the exact number of tokens in a set of places. It is an
// atomic integer expression.
type TokensCount []string

func (f TokensCount) String() string {
	if len(f) == 1 {
		return f[0]
	}
	return util.ZipString(f, "(+ ", ")", " ")
}

// ----------------------------------------------------------------------

// IntegerConstant defines an integer constant. It is an atomic integer
// expression.
type IntegerConstant int

func (f IntegerConstant) String() string {
	return fmt.Sprintf("%d", f)
}

// ----------------------------------------------------------------------
//
// The following elements are not used in practice
//
// ----------------------------------------------------------------------

// // IntegerSum defines the sum between two or more integer expressions. We
// // assume, but do not check beforehand, that all the Formula in IntegerSum are
// // integer expressions.
// type IntegerSum []Formula

// func (f IntegerSum) String() string {
// 	s := f[0].String()
// 	for k := 1; k < len(f); k++ {
// 		s += "," + f[k].String()
// 	}
// 	return "sum(" + s + ")"
// }

// // ----------------------------------------------------------------------

// // IntegerDifference defines the difference between two or more integer
// // expressions. We assume, but do not check beforehand, that all the Formula in
// // IntegerDifference are integer expressions.
// type IntegerDifference []Formula

// func (f IntegerDifference) String() string {
// 	s := f[0].String()
// 	for k := 1; k < len(f); k++ {
// 		s += "," + f[k].String()
// 	}
// 	return "diff(" + s + ")"
// }
