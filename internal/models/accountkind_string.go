// Code generated by "stringer -type=AccountKind --linecomment"; DO NOT EDIT.

package models

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[AccountKindUser-0]
	_ = x[AccountKindClub-1]
	_ = x[AccountKindOrganization-2]
	_ = x[AccoundKindAny-3]
}

const _AccountKind_name = "usercluborganizationany"

var _AccountKind_index = [...]uint8{0, 4, 8, 20, 23}

func (i AccountKind) String() string {
	if i < 0 || i >= AccountKind(len(_AccountKind_index)-1) {
		return "AccountKind(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _AccountKind_name[_AccountKind_index[i]:_AccountKind_index[i+1]]
}
