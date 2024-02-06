package models

import "strings"

//go:generate stringer -type=Permission -trimprefix=Permission
type Permission uint

const (
	PermissionNone Permission = 0
	PermissionEdit Permission = 1 << iota
	PermissionRunNet

	PermissionOwner = PermissionEdit | PermissionRunNet
)

func (p Permission) Has(flag Permission) bool          { return p&flag != 0 }
func (p Permission) Grant(flag Permission) Permission  { return p | flag }
func (p Permission) Revoke(flag Permission) Permission { return p &^ flag }

func ParsePermission(s string) Permission {
	switch strings.ToLower(s) {
	case "edit":
		return PermissionEdit
	case "run-net", "runnet":
		return PermissionRunNet
	case "owner":
		return PermissionOwner

	}
	return PermissionNone
}
