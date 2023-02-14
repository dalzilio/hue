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
