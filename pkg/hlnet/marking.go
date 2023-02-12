// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.
package hlnet

// marking is the type of High-Level Nets (hlnet) markings. It is a multiset of
// values with integer (so maybe negative) multiplicities.
//
// Conventions:
//
// - Items are of the form {key, multiplicity} ;
//
// - We use arrays of size 2 because golang lacks pairs ;
//
// - Items with weight 0 do not appear in multisets (default weight) ;
//
// - Items are ordered in increasing order of keys.
//
type marking []mset

type atom struct{ value, mult int }
type mset []atom
