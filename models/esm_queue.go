package models

import (
	"gorm.io/gorm"
)

type ESMQueue struct {
	gorm.Model
	Sequence string `gorm:"not null;type:longtext" form:"sequence"`
	IsSubseq int64  `gorm:"not null;default:0" form:"is_subseq"`
	ParentId *int64 `gorm:"default:null" form:"parent_id"`
} 