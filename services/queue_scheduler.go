package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"sync"
	"time"
)

// 全局队列调度器实例
var globalQueueScheduler *QueueScheduler
var globalSchedulerOnce sync.Once

// QueueScheduler 统一队列调度器
type QueueScheduler struct {
	alphaProcessor *AlphaProcessor
	itasserProcessor *ItasserProcessor
	stopChan       chan struct{}
	wg             sync.WaitGroup
	isRunning      bool
	mu             sync.Mutex
}

// GetGlobalQueueScheduler 获取全局队列调度器实例
func GetGlobalQueueScheduler() *QueueScheduler {
	globalSchedulerOnce.Do(func() {
		globalQueueScheduler = NewQueueScheduler(1, 1)
	})
	return globalQueueScheduler
}

// NewQueueScheduler 创建新的队列调度器
func NewQueueScheduler(maxAlphaWorkers, maxItasserWorkers int) *QueueScheduler {
	return &QueueScheduler{
		alphaProcessor:   NewAlphaProcessor(maxAlphaWorkers),
		itasserProcessor: NewItasserProcessor(maxItasserWorkers),
		stopChan:         make(chan struct{}),
	}
}

// Start 启动队列调度器
func (qs *QueueScheduler) Start() {
	qs.mu.Lock()
	if qs.isRunning {
		qs.mu.Unlock()
		return
	}
	qs.isRunning = true
	qs.mu.Unlock()

	logger.Info("队列调度器启动...")
	
	qs.wg.Add(1)
	go qs.run()
}

// Stop 停止队列调度器
func (qs *QueueScheduler) Stop() {
	qs.mu.Lock()
	if !qs.isRunning {
		qs.mu.Unlock()
		return
	}
	qs.isRunning = false
	qs.mu.Unlock()

	logger.Info("正在停止队列调度器...")
	close(qs.stopChan)
	qs.wg.Wait()
	logger.Info("队列调度器已停止")
}

// run 主调度循环
func (qs *QueueScheduler) run() {
	defer qs.wg.Done()
	
	ticker := time.NewTicker(1 * time.Minute) // 每1分钟检查一次队列
	defer ticker.Stop()

	for {
		select {
		case <-qs.stopChan:
			return
		case <-ticker.C:
			qs.processQueues()
		}
	}
}

// processQueues 处理所有队列
func (qs *QueueScheduler) processQueues() {
	// 处理AlphaFold队列
	qs.processAlphaFoldQueue()
	
	// 处理I-TASSER队列
	qs.processItasserQueue()
	
	// 清理已完成的任务
	qs.cleanupCompletedTasks()
}

// processAlphaFoldQueue 处理AlphaFold队列
func (qs *QueueScheduler) processAlphaFoldQueue() {
	var pendingTasks []models.AlphaFoldQueue
	
	// 查找待处理的任务
	if err := database.Database.Where("status = ?", "pending").Find(&pendingTasks).Error; err != nil {
		logger.Error("查询AlphaFold待处理任务失败: %v", err)
		return
	}

	for _, task := range pendingTasks {
		// 检查是否有正在处理的任务
		var processingCount int64
		if err := database.Database.Model(&models.AlphaFoldQueue{}).Where("status = ?", "processing").Count(&processingCount).Error; err != nil {
			logger.Error("查询AlphaFold处理中任务数量失败: %v", err)
			continue
		}

		// 如果当前没有处理中的任务，开始处理新任务
		if processingCount == 0 {
			// 更新任务状态为处理中
			if err := database.Database.Model(&models.AlphaFoldQueue{}).Where("id = ?", task.ID).Update("status", "processing").Error; err != nil {
				logger.Error("更新AlphaFold任务状态失败: %v", err)
				continue
			}

			logger.Info("开始处理AlphaFold任务 ID: %d", task.ID)
			
			// 启动处理任务
			go qs.processAlphaFoldTask(task)
		} else {
			logger.Info("AlphaFold队列中有任务正在处理中，跳过新任务")
			break
		}
	}
}

// processItasserQueue 处理I-TASSER队列
func (qs *QueueScheduler) processItasserQueue() {
	var pendingTasks []models.ITasserQueue
	
	// 查找待处理的任务
	if err := database.Database.Where("status = ?", "pending").Find(&pendingTasks).Error; err != nil {
		logger.Error("查询I-TASSER待处理任务失败: %v", err)
		return
	}

	for _, task := range pendingTasks {
		// 检查是否有正在处理的任务
		var processingCount int64
		if err := database.Database.Model(&models.ITasserQueue{}).Where("status = ?", "processing").Count(&processingCount).Error; err != nil {
			logger.Error("查询I-TASSER处理中任务数量失败: %v", err)
			continue
		}

		// 如果当前没有处理中的任务，开始处理新任务
		if processingCount == 0 {
			// 更新任务状态为处理中
			if err := database.Database.Model(&models.ITasserQueue{}).Where("id = ?", task.ID).Update("status", "processing").Error; err != nil {
				logger.Error("更新I-TASSER任务状态失败: %v", err)
				continue
			}

			logger.Info("开始处理I-TASSER任务 ID: %d", task.ID)
			
			// 启动处理任务
			go qs.processItasserTask(task)
		} else {
			logger.Info("I-TASSER队列中有任务正在处理中，跳过新任务")
			break
		}
	}
}

// processAlphaFoldTask 处理单个AlphaFold任务
func (qs *QueueScheduler) processAlphaFoldTask(task models.AlphaFoldQueue) {
	defer func() {
		// 任务完成后，将状态更新为已完成
		if err := database.Database.Model(&models.AlphaFoldQueue{}).Where("id = ?", task.ID).Update("status", "completed").Error; err != nil {
			logger.Error("更新AlphaFold任务完成状态失败: %v", err)
		}
	}()

	// 验证FASTA格式
	if !IsFasta(task.Sequence) {
		logger.Error("AlphaFold任务序列格式无效，跳过处理")
		if err := database.Database.Model(&models.AlphaFoldQueue{}).Where("id = ?", task.ID).Update("status", "failed").Error; err != nil {
			logger.Error("更新AlphaFold任务失败状态失败: %v", err)
		}
		return
	}

	// 使用现有的AlphaProcessor处理任务
	qs.alphaProcessor.buildModel(task.ID, task.Sequence)
}

// processItasserTask 处理单个I-TASSER任务
func (qs *QueueScheduler) processItasserTask(task models.ITasserQueue) {
	defer func() {
		// 任务完成后，将状态更新为已完成
		if err := database.Database.Model(&models.ITasserQueue{}).Where("id = ?", task.ID).Update("status", "completed").Error; err != nil {
			logger.Error("更新I-TASSER任务完成状态失败: %v", err)
		}
	}()

	// 验证FASTA格式
	if !IsFasta(task.Sequence) {
		logger.Error("I-TASSER任务序列格式无效，跳过处理")
		if err := database.Database.Model(&models.ITasserQueue{}).Where("id = ?", task.ID).Update("status", "failed").Error; err != nil {
			logger.Error("更新I-TASSER任务失败状态失败: %v", err)
		}
		return
	}

	// 使用现有的ItasserProcessor处理任务
	qs.itasserProcessor.buildModel(task.ID, task.Sequence)
}

// cleanupCompletedTasks 清理已完成和失败的任务
func (qs *QueueScheduler) cleanupCompletedTasks() {
	// 清理AlphaFold已完成和失败任务（保留最近24小时的任务用于调试）
	yesterday := time.Now().Add(-24 * time.Hour)
	
	if err := database.Database.Where("(status = ? OR status = ?) AND updated_at < ?", "completed", "failed", yesterday).Delete(&models.AlphaFoldQueue{}).Error; err != nil {
		logger.Error("清理AlphaFold已完成和失败任务失败: %v", err)
	}

	// 清理I-TASSER已完成和失败任务
	if err := database.Database.Where("(status = ? OR status = ?) AND updated_at < ?", "completed", "failed", yesterday).Delete(&models.ITasserQueue{}).Error; err != nil {
		logger.Error("清理I-TASSER已完成和失败任务失败: %v", err)
	}
}

// GetQueueStatus 获取队列状态
func (qs *QueueScheduler) GetQueueStatus() map[string]interface{} {
	var alphaPending, alphaProcessing, alphaCompleted, alphaFailed int64
	var itasserPending, itasserProcessing, itasserCompleted, itasserFailed int64

	// 统计AlphaFold队列状态
	database.Database.Model(&models.AlphaFoldQueue{}).Where("status = ?", "pending").Count(&alphaPending)
	database.Database.Model(&models.AlphaFoldQueue{}).Where("status = ?", "processing").Count(&alphaProcessing)
	database.Database.Model(&models.AlphaFoldQueue{}).Where("status = ?", "completed").Count(&alphaCompleted)
	database.Database.Model(&models.AlphaFoldQueue{}).Where("status = ?", "failed").Count(&alphaFailed)

	// 统计I-TASSER队列状态
	database.Database.Model(&models.ITasserQueue{}).Where("status = ?", "pending").Count(&itasserPending)
	database.Database.Model(&models.ITasserQueue{}).Where("status = ?", "processing").Count(&itasserProcessing)
	database.Database.Model(&models.ITasserQueue{}).Where("status = ?", "completed").Count(&itasserCompleted)
	database.Database.Model(&models.ITasserQueue{}).Where("status = ?", "failed").Count(&itasserFailed)

	return map[string]interface{}{
		"alphafold": map[string]int64{
			"pending":    alphaPending,
			"processing": alphaProcessing,
			"completed":  alphaCompleted,
			"failed":     alphaFailed,
		},
		"itasser": map[string]int64{
			"pending":    itasserPending,
			"processing": itasserProcessing,
			"completed":  itasserCompleted,
			"failed":     itasserFailed,
		},
		"is_running": qs.isRunning,
	}
} 