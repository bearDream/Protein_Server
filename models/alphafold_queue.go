package models

import (
	"gorm.io/gorm"
)

type AlphaFoldQueue struct {
	gorm.Model
	Sequence string `gorm:"not null;unique" form:"sequence"`
}
