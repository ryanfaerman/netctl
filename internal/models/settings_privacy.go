package models

type PrivacySettings struct {
	Location    string `json:"location" form:"location" conform:"lower,trim" validate:"required,oneof=public protected private"`
	Visiblility string `json:"visibility" form:"visibility" validate:"required,oneof=public protected private"`
}
