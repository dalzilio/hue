package set

// Set is a container type for lists of integers without duplication.
// Conventions: Items are listed in increasing order of keys.
type Set []int

// Clone returns a deep copy of a set
func Clone(s Set) Set {
	r := make(Set, len(s))
	copy(r, s)
	return r
}

// Shift returns a copy of set s in which every values has been shifted by d.
// Shift does not allocate when s is nil.
func Shift(s Set, d int) Set {
	if len(s) == 0 {
		return nil
	}
	r := make(Set, len(s))
	for k, v := range s {
		r[k] = v + d
	}
	return r
}

// Add returns a Set obtained by adding the value v to s. We return a new slice
// only if the result is different from s.
func Add(s Set, v int) Set {
	if len(s) == 0 {
		return Set{v}
	}
	for i := range s {
		if s[i] == v {
			return s
		}
		if s[i] > v {
			res := make(Set, len(s)+1)
			copy(res[:i], s[:i])
			copy(res[i+1:], s[i:])
			res[i] = v
			return res
		}
	}
	res := make(Set, len(s))
	copy(res, s)
	res = append(res, v)
	return res
}

// Delete returns a Set containing the values of s, except v. It also returns
// the index in the set where the value is found (or -1 if v is not in s). We
// return a new slice only if the result is different from s.
func Delete(s Set, v int) (int, Set) {
	if s == nil {
		return -1, nil
	}
	for i := range s {
		if s[i] == v {
			res := make(Set, len(s)-1)
			copy(res, s[:i])
			copy(res[i:], s[i+1:])
			return i, res
		}
	}
	return -1, s
}

// Setminus returns the sets of elements in s1 but not in s2. It always returns
// a new slice.
func Setminus(s1, s2 Set) Set {
	i1, i2 := 0, 0
	s := Set{}
	for {
		switch {
		case i1 == len(s1):
			return s
		case i2 == len(s2):
			return append(s, s1[i1:]...)
		case s1[i1] == s2[i2]:
			i1++
			i2++
		case s1[i1] < s2[i2]:
			s = append(s, s1[i1])
			i1++
		case s1[i1] > s2[i2]:
			i2++
		}
	}
}

// Intersection returns the sets of elements that occur both in s1 and s2. It
// always returns a new slice.
func Intersection(s1, s2 Set) Set {
	i1, i2 := 0, 0
	s := Set{}
	for {
		switch {
		case i1 == len(s1):
			return s
		case i2 == len(s2):
			return s
		case s1[i1] == s2[i2]:
			s = append(s, s1[i1])
			i1++
			i2++
		case s1[i1] < s2[i2]:
			i1++
		case s1[i1] > s2[i2]:
			i2++
		}
	}
}

// Included returns true if all the elements in slice s1 are also in s2. This is
// equivalent to len(s1) == len(Intersection(s1,s2)) but does not need to
// allocate memory.
func Included(s1, s2 Set) bool {
	i1, i2 := 0, 0
	for {
		switch {
		case i1 == len(s1):
			return true
		case i2 == len(s2):
			return false
		case s1[i1] == s2[i2]:
			i1++
			i2++
		case s1[i1] < s2[i2]:
			return false
		case s1[i1] > s2[i2]:
			i2++
		}
	}
}

// Disjoint returns true is all the elements in slice s1 are also in s2. This is
// equivalent to len(Intersection(s1,s2)) == 0 but does not need to allocate
// memory.
func Disjoint(s1, s2 Set) bool {
	i1, i2 := 0, 0
	for {
		switch {
		case i1 == len(s1):
			return true
		case i2 == len(s2):
			return true
		case s1[i1] == s2[i2]:
			return false
		case s1[i1] < s2[i2]:
			i1++
		case s1[i1] > s2[i2]:
			i2++
		}
	}
}

// Relation takes two sets s1, s2 such that s1 ⊆ s2, and returns the index in s2
// of elements in s1. The result is a slice with same size than s1. We return
// nil if s1 ⊈ s2.
func Relation(s1, s2 Set) Set {
	i1, i2 := 0, 0
	s := Set{}
	for {
		switch {
		case i1 == len(s1):
			if len(s) != len(s1) {
				return nil
			}
			return s
		case i2 == len(s2):
			return nil
		case s1[i1] == s2[i2]:
			s = append(s, i2)
			i1++
			i2++
		case s1[i1] < s2[i2]:
			i1++
		case s1[i1] > s2[i2]:
			i2++
		}
	}
}

// Member returns the index in s at which element v occurs, or -1 if v does not
// appear in s.
func Member(s Set, v int) int {
	for k, i := range s {
		if i == v {
			return k
		}
	}
	return -1
}

// Intersect returns the index in s1 at which an element of s2 occurs, or -1 if
// the two sets do not intersect.
func Intersect(s1, s2 Set) int {
	k1, k2 := 0, 0
	for {
		if len(s1) == k1 || len(s2) == k2 {
			return -1
		}
		if s1[k1] == s2[k2] {
			return k1
		}
		if s2[k2] > s1[k1] {
			k1++
			continue
		}
		if s1[k1] > s2[k2] {
			k2++
		}
	}
}

// Equal is a (deep) equality between sets
func Equal(s1, s2 Set) bool {
	if len(s1) != len(s2) {
		return false
	}
	for k, v := range s1 {
		if v != s2[k] {
			return false
		}
	}
	return true
}

// Compare compare two set using lexicographic ordering, a positive return value
// means that s1 is bigger than s2.
func Compare(s1, s2 Set) int {
	i1, i2 := 0, 0
	for {
		if i1 == len(s1) {
			return i2 - len(s2)
		}
		if i2 == len(s2) {
			return len(s1) - i1
		}
		if s1[i1] != s2[i2] {
			return s1[i1] - s2[i2]
		}
		i1++
		i2++
	}
}

// Union does set union between s1 and s2.
func Union(s1, s2 Set) Set {
	res := make(Set, len(s1))
	copy(res, s1)
	for _, v := range s2 {
		res = Add(res, v)
	}
	return res
}

// Range returns a set containing all the integer between begin (included) and
// end (excluded).
func Range(begin, end int) Set {
	if end <= begin {
		return nil
	}
	res := make(Set, end-begin)
	for i := begin; i < end; i++ {
		res[i-begin] = i
	}
	return res
}
