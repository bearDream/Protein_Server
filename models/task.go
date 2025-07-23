package models

import (
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title    string `gorm:"not null;type:longtext" form:"title" binding:"required"`
	Sequence string `gorm:"not null;type:longtext" form:"sequence" binding:"required"`
	// 1 Sequence Search; 2 Structure Prediction; 3 Parameters Calculation; 4 Superimpose
	Type int64 `gorm:"not null" form:"type"`
	// 1 I-Tasser; 2 AlphaFold2; 3 ESMFold
	StructurePredictionTool *int64 `gorm:"default:null" form:"structurepredictiontool"`
	UserId                  int64  `gorm:"not null" form:"userid" binding:"required"`
	// Sequences from Sequence Search
	SubSequence string `gorm:"type:longtext" form:"subsequence"`
	ModelId     string `gorm:"not null;type:longtext" form:"model_id"`
}
