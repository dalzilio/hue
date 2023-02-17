// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

import "fmt"

// ----------------------------------------------------------------------

// Env is an environment, that is a sorted list of variable (names) occuring in
// an Expressions.
type Env []string

func (p Env) String() string {
	s := "["
	for k, vname := range p {
		if k != 0 {
			s += ", "
		}
		s += vname
	}
	return s + "]"
}

// VEnv is the type of association lists between environment variables and
// values.
type VEnv map[string]*Value

func (venv VEnv) copy() VEnv {
	res := make(VEnv)
	for k, v := range venv {
		res[k] = v
	}
	return res
}

func (venv VEnv) restore(venv2 VEnv) {
	for k := range venv {
		delete(venv, k)
	}
	for k, v := range venv2 {
		venv[k] = v
	}
}

func (venv VEnv) PrintVEnv(net *Net) string {
	res := ""
	for vname, val := range venv {
		res += fmt.Sprintf("%s : %s\n", vname, net.PrintValue(val))
	}
	return res
}
