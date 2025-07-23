package models

import (
	"gorm.io/gorm"
)

type PDBParameter struct {
	gorm.Model
	PdbId              string  `gorm:"not null;type:longtext" form:"pdb_id"`
	Fasta              string  `gorm:"type:text" form:"fasta"`
	RcScore            string  `gorm:"default:'';type:varchar(191)" form:"rc_score"`
	Hydrophobicity     string  `gorm:"default:'';type:varchar(191)" form:"hydrophobicity"`
	Instability        string  `gorm:"default:'';type:varchar(191)" form:"instability"`
	IsoelectricPoint   string  `gorm:"default:'';type:varchar(191)" form:"isoelectric_point"`
	Size               float64 `gorm:"default:0" form:"size"`
	SolventAccesibility string  `gorm:"default:'';type:varchar(191)" form:"solvent_accesibility"`
	Duration           float64 `gorm:"default:0" form:"duration"`
	IsProtein          bool    `gorm:"default:true" form:"is_protein"`
	Score              *float32 `gorm:"default:null" form:"score"`
} 