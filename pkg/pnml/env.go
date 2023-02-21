// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package pnml

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

// Extra returns a list of varnames that are in p2 but not in p
func (p Env) Extra(p2 Env) Env {
	q := Env{}
	for _, varname := range p2 {
		for _, v := range q {
			if v == varname {
				break
			}
		}
		q = append(q, varname)
	}
	return q
}

// VEnv is the type of association lists between environment variables and
// values.
type VEnv map[string]*Value

func (venv VEnv) Copy(venv2 VEnv) {
	for k := range venv {
		venv[k] = nil
	}
	for k, v := range venv2 {
		venv[k] = v
	}
}

func (venv VEnv) Clone() VEnv {
	venv2 := make(map[string]*Value)
	for k, v := range venv {
		venv2[k] = v
	}
	return venv2
}

func (venv VEnv) ResetAll() {
	for k := range venv {
		venv[k] = nil
	}
}

func (venv VEnv) Reset(env Env) {
	for _, v := range env {
		venv[v] = nil
	}
}

func NewVEnv(env Env) VEnv {
	venv := make(VEnv)
	for _, v := range env {
		venv[v] = nil
	}
	return venv
}
