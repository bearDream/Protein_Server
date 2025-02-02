package models

import (
	"gorm.io/gorm"
)

type Note struct {
	gorm.Model
	TaskId int64  `gorm:"not null;unique" form:"taskid" binding:"required"`
	UserId int64  `gorm:"not null;unique" form:"userid" binding:"required"`
	Note   string `form:"note"`
}
