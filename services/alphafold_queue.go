package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
)

// Add To Alpha Fold Queue
func AddToAlphaFoldQueue(sequence string) {
	AddToAlphaFoldQueueWithParent(sequence, nil)
}

// Add To Alpha Fold Queue with parent ID
func AddToAlphaFoldQueueWithParent(sequence string, parentId *int64) {
	var alphafoldQueue models.AlphaFoldQueue
	if err := database.Database.Where("sequence = ?", sequence).Find(&alphafoldQueue).Error; err != nil {
		return
	}
	if alphafoldQueue.ID == 0 {
		alphafoldQueue.Sequence = sequence
		alphafoldQueue.ParentId = parentId
		if err := database.Database.Create(&alphafoldQueue).Error; err != nil {
			logger.Error("创建AlphaFold队列失败: %v", err)
		}
	}
}
