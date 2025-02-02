package models

import (
	"gorm.io/gorm"
)

type ITasserQueue struct {
	gorm.Model
	Sequence string `gorm:"not null;unique" form:"sequence"`
}
