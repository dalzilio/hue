// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

import (
	"fmt"
	"log"
	"sort"
	"strconv"
)

// ----------------------------------------------------------------------

// Expression is the interface that wraps the type of PNML expression (we only
// consider symmetric nets).
//
// AddEnv is used to accumulate the free variables in the Expression to an
// existing environment. It can be used to add the free variables of an
// expression to an alreay existing environment. It creates a (sorted) slice
// containing the names of all the variables.
//
// Eval return the set of constant values that match a ground Expression (an
// expression without free variables, such as the ones used to define the
// initial marking), together with their multiplicities.
type Expression interface {
	String() string
	AddEnv(Env) Env
	Eval(*Net) []Atom
}

// ----------------------------------------------------------------------

func (net *Net) compareExp(p1, p2 *Value, op OP) bool {
	switch op {
	case EQ:
		return p1 == p2
	case GREATTHAN:
		if p1.Head >= 0 {
			return p1.Head > p2.Head
		}
		return p1.Head < p2.Head
	case GREATTHANEQ:
		if p1.Head >= 0 {
			return p1.Head >= p2.Head
		}
		return p1.Head <= p2.Head
	case INEQ:
		return p1 != p2
	case LESSTHAN:
		if p1.Head >= 0 {
			return p1.Head < p2.Head
		}
		return p1.Head > p2.Head
	case LESSTHANEQ:
		if p1.Head >= 0 {
			return p1.Head <= p2.Head
		}
		return p1.Head >= p2.Head
	}
	panic("not reachable in compareExp")
}

// ----------------------------------------------------------------------

// All is the type of all expressions.
type All string

func (p All) String() string {
	return "<ALL:" + string(p) + ">"
}

func (p All) AddEnv(env Env) Env { return env }

func (p All) Eval(net *Net) []Atom {
	f := net.World[string(p)]
	m := make([]Atom, len(f))
	for i := range f {
		m[i] = Atom{f[i], 1}
	}
	return m
}

// ----------------------------------------------------------------------

// Add is the type of add expressions.
type Add []Expression

func (p Add) String() string {
	return multstring(p, "", " + ", "")
}

func (p Add) AddEnv(env Env) Env {
	return multaddEnv(p, env)
}

func (p Add) Eval(net *Net) []Atom {
	var res []Atom
	for i := range p {
		res = append(res, p[i].Eval(net)...)
	}
	return res
}

// ----------------------------------------------------------------------

// Subtract is the type of subtract expressions. It is an array of two
// expressions denoting the left and right elements of a subtract operation.
type Subtract []Expression

func (p Subtract) String() string {
	return multstring(p, "", " - ", "")
}

func (p Subtract) AddEnv(env Env) Env {
	return multaddEnv(p, env)
}

func (p Subtract) Eval(net *Net) []Atom {
	ma := p[0].Eval(net)
	if ma == nil || len(p) == 1 {
		return ma
	}
	for i := 1; i < len(p); i++ {
		mb := p[i].Eval(net)
		if mb == nil {
			continue
		}
		ma = subtract(ma, mb)
	}
	return ma
}

// subtract computes multiset difference, taking into account multiplicities
func subtract(ma []Atom, mb []Atom) []Atom {
	var res []Atom
OUTER:
	for _, a := range ma {
		for _, b := range mb {
			if a.Value == b.Value {
				if a.Mult-b.Mult <= 0 {
					continue OUTER
				}
				a.Mult = a.Mult - b.Mult
			}
		}
		res = append(res, Atom{a.Value, a.Mult})
	}
	return res
}

// ----------------------------------------------------------------------

// Tuple is the type of tuple expressions.
type Tuple []Expression

func (p Tuple) String() string {
	return multstring(p, "(", ", ", ")")
}

func (p Tuple) AddEnv(env Env) Env {
	return multaddEnv(p, env)
}

func (p Tuple) Eval(net *Net) []Atom {
	res := []Atom{{nil, 1}}
	for i := len(p) - 1; i >= 0; i-- {
		fi := p[i].Eval(net)
		fr := []Atom{}
		for _, v1 := range fi {
			for _, v2 := range res {
				fr = append(fr, Atom{net.Unique[Value{Head: v1.Value.Head, Tail: v2.Value}], 1})
			}
		}
		res = fr
	}
	return res
}

// ----------------------------------------------------------------------

// Operation is the type of expressions that apply an operation to a slice of
// expressions.
type Operation struct {
	Op   OP
	Elem []Expression
}

func (p OP) String() string {
	switch p {
	case AND:
		return " and "
	case OR:
		return " or "
	case EQ:
		return " = "
	case INEQ:
		return " != "
	case LESSTHAN:
		return " < "
	case LESSTHANEQ:
		return " <= "
	case GREATTHAN:
		return " > "
	case GREATTHANEQ:
		return " >= "
	}
	return ""
}

func (p Operation) String() string {
	return multstring(p.Elem, "(", p.Op.String(), ")")
}

func (p Operation) AddEnv(env Env) Env {
	return multaddEnv(p.Elem, env)
}

func (p Operation) Eval(net *Net) []Atom {
	panic("Eval not authorized on Operation")
}

// OK returns whether the condition evaluates to true.
func (p Operation) OK(net *Net, env Env) bool {
	switch p.Op {
	case NIL:
		return true
	case AND:
		for _, c := range p.Elem {
			if !c.(Operation).OK(net, env) {
				return false
			}
		}
		return true
	case OR:
		for _, c := range p.Elem {
			if c.(Operation).OK(net, env) {
				return true
			}
		}
		return false
	default:
		v1 := p.Elem[0].Eval(net)
		v2 := p.Elem[1].Eval(net)
		if len(v1) == 0 || len(v2) == 0 {
			return false
		}
		if len(v1) > 1 || len(v2) > 1 {
			panic("problem in conditional, too many results")
		}
		return net.compareExp(v1[0].Value, v2[0].Value, p.Op)
	}
}

// ----------------------------------------------------------------------

// Constant is the type of constant expressions.
type Constant string

func (p Constant) String() string {
	return string(p)
}

func (p Constant) AddEnv(env Env) Env { return env }

func (p Constant) Eval(net *Net) []Atom {
	pval, found := net.order[string(p)]
	if !found {
		f, wfound := net.World[string(p)]
		if !wfound {
			log.Fatalf("identifier %s is not a constant or a known type", string(p))
		}
		m := make([]Atom, len(f))
		for i := range f {
			m[i] = Atom{f[i], 1}
		}
		return m
	}
	return []Atom{{pval, 1}}
}

// ----------------------------------------------------------------------

// FIRConstant is the type of finite int range constant expressions.
type FIRConstant struct {
	value int
	start int
	end   int
}

func (p FIRConstant) stringify() string {
	return fmt.Sprintf("_int%d'%d'%d", p.start, p.end, p.value)
}

func (p FIRConstant) String() string {
	return strconv.Itoa(p.value)
}

func (p FIRConstant) AddEnv(env Env) Env { return env }

func (p FIRConstant) Eval(net *Net) []Atom {
	return []Atom{{net.order[p.stringify()], 1}}
}

// ----------------------------------------------------------------------

// Var is the type of variables.
type Var string

func (p Var) String() string {
	return string(p)
}

func (p Var) AddEnv(env Env) Env {
	return insertEnv(env, p)
}

func (p Var) Eval(net *Net) []Atom {
	panic("Eval not authorized on Var")
}

// ----------------------------------------------------------------------

// Dot is the type of dot constants.
type Dot struct{}

func (p Dot) String() string {
	return "o"
}

func (p Dot) AddEnv(env Env) Env { return env }

func (p Dot) Eval(net *Net) []Atom {
	return []Atom{{net.vdot, 1}}
}

// ----------------------------------------------------------------------

// Successor is the type of successor and predecessor operations.
type Successor struct {
	Var
	Incr int
}

func (p Successor) String() string {
	var mod string
	if p.Incr > 0 {
		mod = "++" + strconv.Itoa(p.Incr)
	} else {
		mod = "--" + strconv.Itoa(-p.Incr)
	}
	return string(p.Var) + mod
}

func (p Successor) AddEnv(env Env) Env {
	return insertEnv(env, p.Var)
}

func (p Successor) Eval(net *Net) []Atom {
	panic("Eval not authorized on Successor")
	// c := env[string(p.Var)]
	// res := net.Next(p.Incr, c)
	// if res == nil {
	// 	return nil, nil
	// }
	// return []*Value{res}, []int{1}
}

// ----------------------------------------------------------------------

// Numberof is the type of numberof expressions in PNML. This is used to add a
// multiplicity Mult to an Expression in a multiset.
type Numberof struct {
	Expression
	Mult int
}

func (p Numberof) String() string {
	return strconv.Itoa(p.Mult) + "'" + p.Expression.String()
}

func (p Numberof) AddEnv(env Env) Env {
	return p.Expression.AddEnv(env)
}

func (p Numberof) Eval(net *Net) []Atom {
	m := p.Expression.Eval(net)
	for i := range m {
		m[i].Mult = p.Mult
	}
	return m
}

// ----------------------------------------------------------------------

func multstring(ee []Expression, start, delim, end string) string {
	s := start
	if len(ee) == 0 {
		return s + end
	}
	s = s + ee[0].String()
	if len(ee) == 1 {
		return s + end
	}
	for i := 1; i < len(ee); i++ {
		s = s + delim + ee[i].String()
	}
	return s + end
}

func multaddEnv(ee []Expression, env Env) Env {
	for _, v := range ee {
		env = v.AddEnv(env)
	}
	return env
}

func insertEnv(env Env, x Var) Env {
	i := sort.SearchStrings(env, string(x))
	if i == len(env) {
		return append(env, string(x))
	}
	if env[i] == string(x) {
		// x is already in the Env
		return env
	}
	// otherwise we shift the necessary data
	env = append(env[:i+1], env[i:]...)
	env[i] = string(x)
	return env
}

// ----------------------------------------------------------------------
