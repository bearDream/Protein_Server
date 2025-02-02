package models

import (
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title    string `gorm:"not null" form:"title" binding:"required"`
	Sequence string `gorm:"not null" form:"sequence" binding:"required"`
	// 1 Sequence Search; 2 Structure Prediction; 3 Parameters Calculation; 4 Superimpose
	Type int64 `gorm:"not null" form:"type"`
	// 1 I-Tasser; 2 AlphaFold2; 3 ESMFold
	StructurePredictionTool int64 `form:"structurepredictiontool"`
	UserId                  uint  `gorm:"not null" form:"userid" binding:"required"`
	// Sequences from Sequence Search
	SubSequence string `form:"subsequence"`
}
