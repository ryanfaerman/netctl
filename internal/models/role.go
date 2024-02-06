package models

type Role struct {
	Name        string     `form:"name" json:"name" validate:"required"`
	ID          int64      `form:"-" json:"id"`
	Permissions Permission `form:"permissions" json:"permissions" validate:"required"`
	Ranking     int64      `form:"ranking" json:"ranking"`
}
