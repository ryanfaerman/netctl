package models

type Settings struct {
	PrivacySettings    `json:"privacy"`
	AppearanceSettings `json:"appearance"`
}

var t = true

var DefaultSettings = Settings{
	PrivacySettings: PrivacySettings{
		Location:    "public",
		Visiblility: "public",
	},
	AppearanceSettings: AppearanceSettings{
		ActivityGraphs: "on",
	},
}
