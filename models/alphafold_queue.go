package models

import (
	"gorm.io/gorm"
)

type AlphaFoldQueue struct {
	gorm.Model
	Sequence string `gorm:"not null;type:text" form:"sequence"`
	IsSubseq int64  `gorm:"not null;default:0" form:"is_subseq"`
	ParentId *int64 `gorm:"default:null" form:"parent_id"`
	Status   string `gorm:"not null;default:'pending'" form:"status"` // pending, processing, completed, failed
}
