package models

import "strings"

//go:generate stringer -type=NetCheckinKind -trimprefix=NetCheckinKind
type NetCheckinKind int

const (
	NetCheckinKindUnknown NetCheckinKind = iota
	NetCheckinKindRoutine
	NetCheckinKindPriority
	NetCheckinKindWelfare
	NetCheckinKindEmergency
	NetCheckinKindTraffic
	NetCheckinKindQuestion
	NetCheckinKindAnnouncement
	NetCheckinKindComment
	NetCheckinKindWeather
)

func ParseNetCheckinKind(s string) NetCheckinKind {
	switch strings.ToLower(s) {
	case "routine":
		return NetCheckinKindRoutine
	case "priority":
		return NetCheckinKindPriority
	case "welfare":
		return NetCheckinKindWelfare
	case "emergency":
		return NetCheckinKindEmergency
	case "traffic":
		return NetCheckinKindTraffic
	case "question":
		return NetCheckinKindQuestion
	case "announcement":
		return NetCheckinKindAnnouncement
	case "comment":
		return NetCheckinKindComment
	case "weather":
		return NetCheckinKindWeather
	default:
		return NetCheckinKindUnknown
	}
}
