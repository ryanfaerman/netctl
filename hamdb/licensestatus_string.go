// Code generated by "stringer -type=LicenseStatus -trimprefix=LicenseStatus"; DO NOT EDIT.

package hamdb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[LicenseStatusUnknown-0]
	_ = x[LicenseStatusActive-1]
}

const _LicenseStatus_name = "UnknownActive"

var _LicenseStatus_index = [...]uint8{0, 7, 13}

func (i LicenseStatus) String() string {
	if i < 0 || i >= LicenseStatus(len(_LicenseStatus_index)-1) {
		return "LicenseStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _LicenseStatus_name[_LicenseStatus_index[i]:_LicenseStatus_index[i+1]]
}
