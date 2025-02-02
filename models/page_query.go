package models

type PageQuery struct {
	Page int `form:"page" binding:"required"`
	Size int `form:"size" binding:"required"`
}
