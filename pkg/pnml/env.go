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

func (venv VEnv) Reset() {
	for k := range venv {
		venv[k] = nil
	}
}
