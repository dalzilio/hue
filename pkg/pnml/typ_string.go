package pnml

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[DOT-0]
	_ = x[CENUM-1]
	_ = x[FENUM-2]
	_ = x[FINTRANGE-3]
	_ = x[PROD-4]
	_ = x[NUMERIC-5]
}

const _TYP_name = "DOTCENUMFENUMFINTRANGEPRODNUMERIC"

var _TYP_index = [...]uint8{0, 3, 8, 13, 22, 26, 33}

func (i TYP) String() string {
	if i < 0 || i >= TYP(len(_TYP_index)-1) {
		return "TYP(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TYP_name[_TYP_index[i]:_TYP_index[i+1]]
}
