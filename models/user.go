package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"not null;unique" form:"email" binding:"required"`
	Password string `gorm:"not null" form:"password" binding:"required"`
	NewCount int64  `gorm:"not null"`
}
