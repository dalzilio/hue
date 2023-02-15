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
// values (with multiplicities), meaning atoms.
type VEnv map[string]Atom

// func (p VEnv) String() string {
// 	keys := []string{}
// 	for s := range p {
// 		keys = append(keys, s)
// 	}
// 	sort.Strings(keys)
// 	res := "["
// 	for k, vname := range keys {
// 		if k != 0 {
// 			res += ", "
// 		}
// 		res += fmt.Sprintf("%s : %s", vname, p[vname])
// 	}
// 	return res + "]"
// }

// // PrintEnv returns a readable description of a Value environment
// func (net *Net) PrintEnv(env Env) string {
// 	s := "["
// 	start := true
// 	var keys []string
// 	for k := range env {
// 		keys = append(keys, k)
// 	}
// 	sort.Strings(keys)
// 	for _, k := range keys {
// 		if start {
// 			s += fmt.Sprintf("%s : %s", k, net.PrintValue(env[k]))
// 			start = false
// 		} else {
// 			s += fmt.Sprintf(", %s : %s", k, net.PrintValue(env[k]))
// 		}
// 	}
// 	return s + "]"
// }
