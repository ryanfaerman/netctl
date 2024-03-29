package models

type Settings struct {
	PrivacySettings    `json:"privacy"`
	AppearanceSettings `json:"appearance"`
	ProfileSettings    `json:"profile"`
	LocationSettings   `json:"location"`
}

var DefaultSettings = Settings{
	PrivacySettings: PrivacySettings{
		Location:    "public",
		Visiblility: "public",
	},
	AppearanceSettings: AppearanceSettings{
		ActivityGraphs: "on",
	},
}
