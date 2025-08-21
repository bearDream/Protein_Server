package models

import (
	"gorm.io/gorm"
)

type Share struct {
	gorm.Model
	TaskId uint `gorm:"default:0" form:"taskid" binding:"required"`
	ToId   uint `gorm:"not null" form:"toId" binding:"required"`
	// 0 Undisposed; 1 agree; 2 reject
	Status int64 `gorm:"not null" form:"status"`
	FromId uint  `gorm:"not null" form:"fromId"`
	SeqId  uint  `gorm:"default:0" form:"seqId"`
}
