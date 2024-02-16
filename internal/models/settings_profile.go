package models

type ProfileSettings struct {
	Name  string `form:"name" json:"name" validate:"required"`
	About string `form:"about" json:"about"`
	Slug  string `form:"slug" json:"slug"`
}
