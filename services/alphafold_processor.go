package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Struct
type AlphaProcessor struct {
	workerChan chan struct{} // A semaphore channel used to control concurrency
}

// The & and * are pointer related symbols
func NewAlphaProcessor(maxWorkers int) *AlphaProcessor {
	processor := &AlphaProcessor{
		workerChan: make(chan struct{}, maxWorkers), // Control the maximum number of concurrent requests
	}
	return processor
}

// Start processing - 这个方法现在由队列调度器调用
func (p *AlphaProcessor) Process() {
	// 这个方法现在由队列调度器管理，不再自动处理队列
	logger.Info("AlphaFold处理器已就绪，等待队列调度器分配任务")
}

func (p *AlphaProcessor) buildModel(id uint, sequence string) {
	// 记录开始时间
	startTime := time.Now()
	logger.Info("AlphaFold任务 ID %d 开始处理，序列长度: %d", id, len(sequence))

	// 确保输入目录存在并清空内容
	inputDir := "alphafold_input"
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		logger.Error("创建输入目录失败: %v", err)
		return
	}

	// 清空输入目录中的所有文件
	if err := p.cleanDirectory(inputDir); err != nil {
		logger.Error("清空输入目录失败: %v", err)
		return
	}

	// Create a FASTA file
	if err := p.createFastaFile(sequence); err != nil {
		logger.Error("创建FASTA文件失败: %v", err)
		return
	}

	// Run the AlphaFold command
	logger.Info("开始执行AlphaFold命令...")
	cmd := exec.Command("bash", "-c", "source /root/miniconda3/etc/profile.d/conda.sh && conda activate alphafold && bash ../alphafold/run_alphafold.sh -d ../alphadata -o ./alphafold_output -f ./alphafold_input/query.fasta -t 2021-11-01 -g False -c reduced_dbs")

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("执行AlphaFold失败: %v, 输出: %s", err, output)
		return
	}

	// 计算处理时间
	duration := time.Since(startTime)
	logger.Info("AlphaFold任务 ID %d 执行完成，耗时: %.2f秒", id, duration.Seconds())

	// Processing result
	if err := p.processResult(id, sequence, duration); err != nil {
		logger.Error("处理结果失败: %v", err)
		return
	}
}
func (p *AlphaProcessor) processResult(id uint, seq string, duration time.Duration) error {
	// Update queue status to completed
	if err := p.updateQueueStatus(id, "completed"); err != nil {
		return fmt.Errorf("update queue status failure: %v", err)
	}

	// find id
	var proteinInformation models.ProteinInformation
	if err := database.Database.Where("sequence = ?", seq).Find(&proteinInformation).Error; err != nil {
		return fmt.Errorf("find protein information failed: %v", err)
	}

	// 保存处理时间到数据库
	durationSeconds := duration.Seconds()
	if err := database.Database.Model(&models.ProteinInformation{}).Where("id = ?", proteinInformation.ID).Update("duration", durationSeconds).Error; err != nil {
		logger.Error("保存处理时间失败: %v", err)
	} else {
		logger.Info("已保存AlphaFold任务处理时间: %.2f秒", durationSeconds)
	}

	// Move the generated file to the static folder
	if err := p.moveModelFile(proteinInformation.ID); err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}

	// Calculate parameters
	CalculateProteinInfomationWithPath(proteinInformation)

	// 保存RCSB PDB结构数量到数据库
	SaveStructureNum(proteinInformation.ID)

	// 异步任务完成后，更新相关主任务的ModelId
	UpdateTaskModelIdAfterAsyncCompletion(proteinInformation.ID)

	return nil
}

func (p *AlphaProcessor) createFastaFile(sequence string) error {
	filePath := "./alphafold_input/query.fasta"
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file failed: %v", err)
	}
	defer file.Close()

	// 写入FASTA格式内容（包含头部）
	_, err = file.WriteString(">query\n")
	if err != nil {
		return fmt.Errorf("write fasta header failed: %v", err)
	}

	_, err = file.WriteString(sequence + "\n")
	if err != nil {
		return fmt.Errorf("write sequence failed: %v", err)
	}

	return nil
}

func (p *AlphaProcessor) moveModelFile(id uint) error {
	// Define source and destination paths
	sourcePath := filepath.Join("alphafold_output", "query", "unrelaxed_model_1.pdb")
	destDir := filepath.Join("static", "models")
	destPath := filepath.Join(destDir, fmt.Sprintf("%d.pdb", id))

	// 确保目标目录存在
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Move the file
	err := os.Rename(sourcePath, destPath)
	if err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

func (p *AlphaProcessor) findQueueItem() (*models.AlphaFoldQueue, error) {
	var item models.AlphaFoldQueue
	if err := database.Database.Where("status = ?", "pending").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (p *AlphaProcessor) updateQueueStatus(id uint, status string) error {
	// 更新队列记录的状态
	return database.Database.Model(&models.AlphaFoldQueue{}).Where("id = ?", id).Update("status", status).Error
}

func IsFasta(seq string) bool {
	// Define valid amino acid characters
	validAminoAcids := "ACDEFGHIKLMNPQRSTVWY"

	// Convert sequence to uppercase for case-insensitive comparison
	seq = strings.ToUpper(seq)

	// Check each character in the sequence
	for _, char := range seq {
		// If character is not found in validAminoAcids, return false
		if !strings.ContainsRune(validAminoAcids, char) {
			return false
		}
	}

	// All characters are valid
	return true
}

// cleanDirectory 清空指定目录中的所有文件和子目录
func (p *AlphaProcessor) cleanDirectory(dirPath string) error {
	dir, err := os.Open(dirPath)
	if err != nil {
		return fmt.Errorf("打开目录失败: %v", err)
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return fmt.Errorf("读取目录内容失败: %v", err)
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dirPath, name))
		if err != nil {
			return fmt.Errorf("删除文件/目录失败 %s: %v", name, err)
		}
	}

	return nil
}
