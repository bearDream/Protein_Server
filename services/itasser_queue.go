package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
)

// Add To ITasser Queue
func AddToITasserQueue(sequence string) {
	AddToITasserQueueWithParent(sequence, nil)
}

// Add To ITasser Queue with parent ID
func AddToITasserQueueWithParent(sequence string, parentId *int64) {
	var itasserQueue models.ITasserQueue
	if err := database.Database.Where("sequence = ?", sequence).Find(&itasserQueue).Error; err != nil {
		return
	}
	if itasserQueue.ID == 0 {
		itasserQueue.Sequence = sequence
		itasserQueue.ParentId = parentId
		if err := database.Database.Create(&itasserQueue).Error; err != nil {
			logger.Error("创建ITasser队列失败: %v", err)
		}
	}
}
