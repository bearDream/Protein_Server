package models

import (
	"gorm.io/gorm"
)

type Share struct {
	gorm.Model
	TaskId   uint `gorm:"not null" form:"taskid" binding:"required"`
	ToUserId uint `gorm:"not null" form:"touserid" binding:"required"`
	// 0 Undisposed; 1 agree; 2 reject
	Status int64 `form:"status"`
}
