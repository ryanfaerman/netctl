package models

type ProfileSettings struct {
	Name  string `form:"name" validate:"required"`
	About string `form:"about" json:"about"`
}
