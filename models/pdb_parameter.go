package models

import (
	"gorm.io/gorm"
)

type PDBParameter struct {
	gorm.Model
	PdbId               string   `gorm:"not null;type:longtext" form:"pdbId" json:"pdbId"`
	Fasta               string   `gorm:"type:text" form:"fasta" json:"fasta"`
	RcScore             string   `gorm:"default:'';type:varchar(191)" form:"rcScore" json:"rcScore"`
	Hydrophobicity      string   `gorm:"default:'';type:varchar(191)" form:"hydrophobicity" json:"hydrophobicity"`
	Instability         string   `gorm:"default:'';type:varchar(191)" form:"instability" json:"instability"`
	IsoelectricPoint    string   `gorm:"default:'';type:varchar(191)" form:"isoelectricPoint" json:"isoelectricPoint"`
	Size                float64  `gorm:"default:0" form:"size" json:"size"`
	SolventAccesibility string   `gorm:"default:'';type:varchar(191)" form:"solventAccesibility" json:"solventAccesibility"`
	Duration            float64  `gorm:"default:0" form:"duration" json:"duration"`
	IsProtein           bool     `gorm:"default:true" form:"isProtein" json:"isProtein"`
	Score               *float32 `gorm:"default:null" form:"score" json:"score"`
}
