package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
)

// ESMFold
func ESMFold(sequence string) {
	// 记录开始时间
	startTime := time.Now()
	logger.Info("ESMFold任务开始处理，序列长度: %d", len(sequence))
	
	var proteinInformation models.ProteinInformation
	// get sequence id
	if err := database.Database.Where("sequence = ?", sequence).Find(&proteinInformation).Error; err != nil {
		logger.Error("查找蛋白质信息失败: %v", err)
		return
	}
	
	// 确保输出目录存在
	modelsDir := filepath.Join("static", "models")
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		logger.Error("创建模型目录失败: %v", err)
		return
	}
	
	// ESMFold's API requires skipping SSL authentication
	// SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	// resty.New() can get an object
	logger.Info("开始调用ESMFold API...")
	client := resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	// Use R() then can use POST GET ...
	// ESMFold API: https://api.esmatlas.com/foldSequence/v1/pdb/
	resp, err := client.R().SetBody(sequence).Post("https://api.esmatlas.com/foldSequence/v1/pdb/")
	if err != nil {
		logger.Error("请求ESMFold失败: %v", err)
		return
	}
	
	// Save pdb files in static/models fold
	// PDB file's name should be id.pdb
	filename := filepath.Join("static/models", fmt.Sprintf("%d.pdb", proteinInformation.ID))
	if err := os.WriteFile(filename, resp.Body(), 0644); err != nil {
		logger.Error("保存PDB文件失败: %v", err)
		return
	}

	// 计算处理时间并保存到数据库
	duration := time.Since(startTime)
	durationSeconds := duration.Seconds()
	if err := database.Database.Model(&models.ProteinInformation{}).Where("id = ?", proteinInformation.ID).Update("duration", durationSeconds).Error; err != nil {
		logger.Error("保存处理时间失败: %v", err)
	} else {
		logger.Info("ESMFold任务执行完成，耗时: %.2f秒", durationSeconds)
	}

	// Calculate parameters
	CalculateProteinInfomationWithPath(proteinInformation)
	
	// 保存RCSB PDB结构数量到数据库
	SaveStructureNum(proteinInformation.ID)
}
