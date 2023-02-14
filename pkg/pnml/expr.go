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
// initial marking), together with their multiplicities. The method returns two
// slices that have equal length.
type Expression interface {
	String() string
	AddEnv(Env) Env
	Eval(*Net) ([]*Value, []int)
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

func (p All) Eval(net *Net) ([]*Value, []int) {
	f := net.World[string(p)]
	m := make([]int, len(f))
	for i := range f {
		m[i] = 1
	}
	return f, m
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

func (p Add) Eval(net *Net) ([]*Value, []int) {
	var res []*Value
	var mult []int
	for i := range p {
		f, m := p[i].Eval(net)
		res = append(res, f...)
		mult = append(mult, m...)
	}
	return res, mult
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

func (p Subtract) Eval(net *Net) ([]*Value, []int) {
	fa, ma := p[0].Eval(net)
	if fa == nil || len(p) == 1 {
		return fa, ma
	}
	for i := 1; i < len(p); i++ {
		fb, mb := p[i].Eval(net)
		if fb == nil {
			continue
		}
		fa, ma = subtract(fa, ma, fb, mb)
	}
	return fa, ma
}

// subtract computes multiset difference, taking into account multiplicities
func subtract(fa []*Value, ma []int, fb []*Value, mb []int) ([]*Value, []int) {
	var f []*Value
	var m []int
OUTER:
	for i, a := range fa {
		for j, b := range fb {
			if a == b {
				if ma[i]-mb[j] <= 0 {
					continue OUTER
				}
				ma[i] = ma[i] - mb[j]
			}
		}
		f = append(f, a)
		m = append(m, ma[i])
	}
	return f, m
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

func (p Tuple) Eval(net *Net) ([]*Value, []int) {
	res := []*Value{nil}
	for i := len(p) - 1; i >= 0; i-- {
		f, _ := p[i].Eval(net)
		foo := []*Value{}
		for _, v1 := range f {
			for _, v2 := range res {
				foo = append(foo, net.Unique[Value{Head: v1.Head, Tail: v2}])
			}
		}
		res = foo
	}
	mres := []int{}
	for i := 0; i < len(res); i++ {
		mres = append(mres, 1)
	}
	return res, mres
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

func (p Operation) Eval(net *Net) ([]*Value, []int) {
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
		v1, _ := p.Elem[0].Eval(net)
		v2, _ := p.Elem[1].Eval(net)
		if len(v1) == 0 || len(v2) == 0 {
			return false
		}
		if len(v1) > 1 || len(v2) > 1 {
			panic("problem in conditional, too many results")
		}
		return net.compareExp(v1[0], v2[0], p.Op)
	}
}

// ----------------------------------------------------------------------

// Constant is the type of constant expressions.
type Constant string

func (p Constant) String() string {
	return string(p)
}

func (p Constant) AddEnv(env Env) Env { return env }

func (p Constant) Eval(net *Net) ([]*Value, []int) {
	pval, found := net.order[string(p)]
	if !found {
		f, wfound := net.World[string(p)]
		if !wfound {
			log.Fatalf("identifier %s is not a constant or a known type", string(p))
		}
		m := make([]int, len(f))
		for i := range f {
			m[i] = 1
		}
		return f, m
	}
	return []*Value{pval}, []int{1}
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

func (p FIRConstant) Eval(net *Net) ([]*Value, []int) {
	return []*Value{net.order[p.stringify()]}, []int{1}
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

func (p Var) Eval(net *Net) ([]*Value, []int) {
	panic("Eval not authorized on Var")
}

// ----------------------------------------------------------------------

// Dot is the type of dot constants.
type Dot struct{}

func (p Dot) String() string {
	return "o"
}

func (p Dot) AddEnv(env Env) Env { return env }

func (p Dot) Eval(net *Net) ([]*Value, []int) {
	return []*Value{net.vdot}, []int{1}
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

func (p Successor) Eval(net *Net) ([]*Value, []int) {
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

func (p Numberof) Eval(net *Net) ([]*Value, []int) {
	f, m := p.Expression.Eval(net)
	for i := range f {
		m[i] = p.Mult
	}
	return f, m
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
