package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"not null;uniqueIndex:uni_users_email;type:varchar(191)" form:"email" binding:"required"`
	Password string `gorm:"not null;type:longtext" form:"password" binding:"required"`
	NewCount int64  `gorm:"not null" form:"new_count"`
}
