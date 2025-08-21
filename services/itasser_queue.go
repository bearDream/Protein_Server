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
	
	// 构建查询条件
	query := database.Database.Where("sequence = ?", sequence)
	if parentId != nil {
		query = query.Where("parent_id = ?", *parentId)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	
	if err := query.First(&itasserQueue).Error; err != nil {
		// 记录不存在，创建新记录
		if err.Error() == "record not found" {
			itasserQueue = models.ITasserQueue{
				Sequence: sequence,
				ParentId: parentId,
				Status:   "pending",
			}
			if err := database.Database.Create(&itasserQueue).Error; err != nil {
				logger.Error("创建ITasser队列失败: %v", err)
			} else {
				logger.Info("已添加序列到ITasser队列，ID: %d", itasserQueue.ID)
			}
		} else {
			logger.Error("查询ITasser队列失败: %v", err)
		}
	} else {
		logger.Info("序列已存在于ITasser队列中，跳过添加，ID: %d", itasserQueue.ID)
	}
}
