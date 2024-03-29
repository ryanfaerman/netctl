// Code generated by "stringer -type=NetStatus -trimprefix=NetStatus"; DO NOT EDIT.

package models

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[NetStatusUnknown-0]
	_ = x[NetStatusScheduled-1]
	_ = x[NetStatusOpened-2]
	_ = x[NetStatusClosed-3]
}

const _NetStatus_name = "UnknownScheduledOpenedClosed"

var _NetStatus_index = [...]uint8{0, 7, 16, 22, 28}

func (i NetStatus) String() string {
	if i < 0 || i >= NetStatus(len(_NetStatus_index)-1) {
		return "NetStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _NetStatus_name[_NetStatus_index[i]:_NetStatus_index[i+1]]
}
