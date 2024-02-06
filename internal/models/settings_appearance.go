package models

type AppearanceSettings struct {
	ActivityGraphs string `json:"activity-graphs" form:"activity-graphs" validate:"required,oneof=on off"`
}
