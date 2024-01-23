package models

//go:generate stringer -type=NetCheckinKind -trimprefix=NetCheckinKind
type NetCheckinKind int

const (
	NetCheckinKindUnknown NetCheckinKind = iota
	NetCheckinKindRoutine
	NetCheckinKindPriority
	NetCheckinKindTraffic
)
