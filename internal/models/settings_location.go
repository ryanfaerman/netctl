package models

type LocationSettings struct {
	TimeOffset int     `form:"timeOffset"`
	Latitude   float64 `form:"latitude" validate:"latitude"`
	Longitude  float64 `form:"longitude" validate:"longitude"`
}

func (s LocationSettings) Location() (float64, float64) {
	return s.Latitude, s.Longitude
}

func (s LocationSettings) HasLocation() bool {
	return s.Latitude != 0 && s.Longitude != 0
}
