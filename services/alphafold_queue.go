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
	
	// 构建查询条件
	query := database.Database.Where("sequence = ?", sequence)
	if parentId != nil {
		query = query.Where("parent_id = ?", *parentId)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	
	if err := query.First(&alphafoldQueue).Error; err != nil {
		// 记录不存在，创建新记录
		if err.Error() == "record not found" {
			alphafoldQueue = models.AlphaFoldQueue{
				Sequence: sequence,
				ParentId: parentId,
				Status:   "pending",
			}
			if err := database.Database.Create(&alphafoldQueue).Error; err != nil {
				logger.Error("创建AlphaFold队列失败: %v", err)
			} else {
				logger.Info("已添加序列到AlphaFold队列，ID: %d", alphafoldQueue.ID)
			}
		} else {
			logger.Error("查询AlphaFold队列失败: %v", err)
		}
	} else {
		logger.Info("序列已存在于AlphaFold队列中，跳过添加，ID: %d", alphafoldQueue.ID)
	}
}
