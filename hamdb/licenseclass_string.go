// Code generated by "stringer -type=LicenseClass -linecomment"; DO NOT EDIT.

package hamdb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[UnknownClass-0]
	_ = x[TechnicianClass-1]
	_ = x[GeneralClass-2]
	_ = x[AdvancedClass-3]
	_ = x[ExtraClass-4]
	_ = x[ClubClass-5]
}

const _LicenseClass_name = "UnknownTechnicianGeneralAdvancedExtraClub"

var _LicenseClass_index = [...]uint8{0, 7, 17, 24, 32, 37, 41}

func (i LicenseClass) String() string {
	if i < 0 || i >= LicenseClass(len(_LicenseClass_index)-1) {
		return "LicenseClass(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _LicenseClass_name[_LicenseClass_index[i]:_LicenseClass_index[i+1]]
}
