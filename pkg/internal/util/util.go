// Copyright 2023. Silvano DAL ZILIO (LAAS-CNRS). All rights reserved. Use of
// this source code is governed by the GNU Affero license that can be found in
// the LICENSE file.

package util

import (
	"fmt"
	"sort"
)

// ----------------------------------------------------------------------
// Strings stuff
// ----------------------------------------------------------------------

// returns true if all the strings in sl2 are also in sl1. WARNING: We sort the
// input strings.
func StringListIncludes(sl1, sl2 []string) bool {
	sort.Strings(sl1)
	for _, s := range sl2 {
		if x := sort.SearchStrings(sl1, s); (x >= len(sl1)) || (sl1[x] != s) {
			return false
		}
	}
	return true
}

// returns true if the string liss  sl containes s.
func StringListContains(sl []string, s string) bool {
	for _, x := range sl {
		if x == s {
			return true
		}
	}
	return false
}

func ZipString(a []string, start, end, sep string) string {
	res := start
	for k, aa := range a {
		if k != 0 {
			res += sep
		}
		res += aa
	}
	return res + end
}

func ZipPrint[T fmt.Stringer](a []T, start, end, sep string) string {
	res := start
	for k, aa := range a {
		if k != 0 {
			res += sep
		}
		res += aa.String()
	}
	return res + end
}

// ----------------------------------------------------------------------
// Bool stuff
// ----------------------------------------------------------------------

func IfAndOnlyIf(b1, b2 bool) bool {
	if b1 {
		return b2
	}
	return !b2
}

func FoldOr[T any](f func(T) bool, bs []T) bool {
	for _, b := range bs {
		if f(b) {
			return true
		}
	}
	return false
}
