package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
)

// Protein Information
func ProteinInformation(sequence string, blastinformation string, structurePredictionTool int64) {
	ProteinInformationWithParent(sequence, blastinformation, structurePredictionTool, 0)
}

// Protein Information with parent ID
func ProteinInformationWithParent(sequence string, blastinformation string, structurePredictionTool int64, parentId uint) {
	var proteinInformation models.ProteinInformation
	if err := database.Database.Where("sequence = ?", sequence).Find(&proteinInformation).Error; err != nil {
		return
	}
	if proteinInformation.ID == 0 {
		proteinInformation.Sequence = sequence
		proteinInformation.BlastInformation = blastinformation
		if parentId > 0 {
			proteinInformation.ParentId = parentId
		}
		// Create data
		if err := database.Database.Create(&proteinInformation).Error; err != nil {
			logger.Error("创建蛋白质信息失败: %v", err)
			return
		}
		
		// 将parentId转换为*int64类型以匹配队列函数参数
		var parentIdPtr *int64
		if parentId > 0 {
			parentIdInt64 := int64(parentId)
			parentIdPtr = &parentIdInt64
		}
		
		if structurePredictionTool == 1 {
			// AlphaFold2 (修复：使用正确的函数和ParentId)
			AddToAlphaFoldQueueWithParent(sequence, parentIdPtr)
		}
		if structurePredictionTool == 2 {
			// I-Tasser (修复：使用正确的函数和ParentId)
			AddToITasserQueueWithParent(sequence, parentIdPtr)
		}
		if structurePredictionTool == 3 {
			// ESMFold
			ESMFold(sequence)
		}

	}
}
