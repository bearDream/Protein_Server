package models

import (
	"gorm.io/gorm"
)

type Note struct {
	gorm.Model
	TaskId int64  `gorm:"not null" form:"taskid" binding:"required"`
	UserId int64  `gorm:"not null" form:"userid" binding:"required"`
	Note   string `gorm:"type:longtext" form:"note"`
}
