// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package formula

import "fmt"

// Query is the type of reachability queries we need to check. When Verdict is
// true the query is (EF phi), meaning we report  true if we can reach a state
// where phi is true. Conversely, when Verdict is false, the query is (AG phi),
// meaning we report false if (not phi) is reachable.
type Query struct {
	ID      string
	Verdict bool
	Formula
}

// Formula is the interface that wraps the type of XML Formula.
//
// Match return the set of constant values that match an Expression, also called
// a pattern,  together with their multiplicities. The method returns two slices
// that have equal length.
type Formula interface {
	String() string
	// AddEnv(Env)
	// Match(*Net, Env) ([]*Value, []int)
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
	s := f[0].String()
	for k := 1; k < len(f); k++ {
		s += " " + f[k].String()
	}
	return "(or " + s + ")"
}

// ----------------------------------------------------------------------

// Conjunction is the conjunction of two or more formulas.
type Conjunction []Formula

func (f Conjunction) String() string {
	s := f[0].String()
	for k := 1; k < len(f); k++ {
		s += " " + f[k].String()
	}
	return "(and " + s + ")"
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
	s := f[0]
	for k := 1; k < len(f); k++ {
		s += "," + f[k]
	}
	// return "\"" + s + "\"?"
	return "fire_" + s
}

// ----------------------------------------------------------------------

// TokensCount returns the exact number of tokens in a set of places. It is an
// atomic integer expression.
type TokensCount []string

func (f TokensCount) String() string {
	s := f[0]
	for k := 1; k < len(f); k++ {
		s += " " + f[k]
	}
	return "(+ " + s + " 0)"
}

// ----------------------------------------------------------------------

// IntegerConstant defines an integer constant. It is an atomic integer
// expression.
type IntegerConstant int

func (f IntegerConstant) String() string {
	return fmt.Sprintf("%d", f)
}

// ----------------------------------------------------------------------

// IntegerSum defines the sum between two or more integer expressions. We
// assume, but do not check beforehand, that all the Formula in IntegerSum are
// integer expressions.
type IntegerSum []Formula

func (f IntegerSum) String() string {
	s := f[0].String()
	for k := 1; k < len(f); k++ {
		s += "," + f[k].String()
	}
	return "sum(" + s + ")"
}

// ----------------------------------------------------------------------

// IntegerDifference defines the difference between two or more integer
// expressions. We assume, but do not check beforehand, that all the Formula in
// IntegerDifference are integer expressions.
type IntegerDifference []Formula

func (f IntegerDifference) String() string {
	s := f[0].String()
	for k := 1; k < len(f); k++ {
		s += "," + f[k].String()
	}
	return "diff(" + s + ")"
}
