package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
)

// Add To ESM Queue
func AddToESMQueue(sequence string) {
	AddToESMQueueWithParent(sequence, nil)
}

// Add To ESM Queue with parent ID
func AddToESMQueueWithParent(sequence string, parentId *int64) {
	var esmQueue models.ESMQueue
	if err := database.Database.Where("sequence = ?", sequence).Find(&esmQueue).Error; err != nil {
		return
	}
	if esmQueue.ID == 0 {
		esmQueue.Sequence = sequence
		esmQueue.ParentId = parentId
		if err := database.Database.Create(&esmQueue).Error; err != nil {
			logger.Error("创建ESM队列失败: %v", err)
		}
	}
} 