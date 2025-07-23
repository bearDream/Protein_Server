package main

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"Protein_Server/services"
	"fmt"
	"time"
)

func testQueue() {
	// 初始化日志
	logger.Info("开始测试队列调度器...")

	// 启动队列调度器
	queueScheduler := services.GetGlobalQueueScheduler()
	queueScheduler.Start()
	defer queueScheduler.Stop()

	// 等待调度器启动
	time.Sleep(2 * time.Second)

	// 添加测试任务到AlphaFold队列
	testSequence := "MKTVRQERLKSIVRILERSKEPVSGAQLAEELSVSRQVIVQDIAYLRSLGYNIVATPRGYVLAGG"

	alphaTask := models.AlphaFoldQueue{
		Sequence: testSequence,
		IsSubseq: 0,
		Status:   "pending",
	}

	if err := database.Database.Create(&alphaTask).Error; err != nil {
		logger.Error("创建AlphaFold测试任务失败: %v", err)
		return
	}

	logger.Info("已添加AlphaFold测试任务，ID: %d", alphaTask.ID)

	// 添加测试任务到I-TASSER队列
	itasserTask := models.ITasserQueue{
		Sequence: testSequence,
		IsSubseq: 0,
		Status:   "pending",
	}

	if err := database.Database.Create(&itasserTask).Error; err != nil {
		logger.Error("创建I-TASSER测试任务失败: %v", err)
		return
	}

	logger.Info("已添加I-TASSER测试任务，ID: %d", itasserTask.ID)

	// 监控队列状态
	for i := 0; i < 10; i++ {
		status := queueScheduler.GetQueueStatus()
		fmt.Printf("队列状态 (第%d次检查):\n", i+1)
		fmt.Printf("  AlphaFold: pending=%d, processing=%d, completed=%d, failed=%d\n",
			status["alphafold"].(map[string]int64)["pending"],
			status["alphafold"].(map[string]int64)["processing"],
			status["alphafold"].(map[string]int64)["completed"],
			status["alphafold"].(map[string]int64)["failed"])
		fmt.Printf("  I-TASSER: pending=%d, processing=%d, completed=%d, failed=%d\n",
			status["itasser"].(map[string]int64)["pending"],
			status["itasser"].(map[string]int64)["processing"],
			status["itasser"].(map[string]int64)["completed"],
			status["itasser"].(map[string]int64)["failed"])
		fmt.Printf("  调度器运行状态: %v\n", status["is_running"])
		fmt.Println("---")

		time.Sleep(10 * time.Second)
	}

	logger.Info("队列调度器测试完成")
}
