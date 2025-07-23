package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type ItasserProcessor struct {
	workerChan chan struct{} // A semaphore channel used to control concurrency
}

func NewItasserProcessor(maxWorkers int) *ItasserProcessor {
	processor := &ItasserProcessor{
		workerChan: make(chan struct{}, maxWorkers), // Control the maximum number of concurrent requests
	}
	return processor
}

func (p *ItasserProcessor) Process() {
	// 这个方法现在由队列调度器管理，不再自动处理队列
	logger.Info("I-TASSER处理器已就绪，等待队列调度器分配任务")
}

func (p *ItasserProcessor) buildModel(id uint, sequence string) {

	// Empty the input folder
	if err := os.RemoveAll("itasser_example/*"); err != nil {
		logger.Error("清空输入文件夹失败: %v", err)
		return
	}

	// Create a FASTA file
	if err := p.createFastaFile(sequence); err != nil {
		logger.Error("创建FASTA文件失败: %v", err)
		return
	}

	// Run the I-Tasser command
	cmd := exec.Command("../I-TASSER5.1/I-TASSERmod/runI-TASSER.pl",
		"-libdir", "../I-TASSER5.1/lib",
		"-seqname", "itasser_example",
		"-datadir", "./itasser_example",
		"-light", "true",
		"-nmodel", "1",
		"-hours", "2")

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("执行I-Tasser失败: %v, 输出: %s", err, output)
		return
	}

	if err := p.processResult(id, sequence); err != nil {
		logger.Error("处理结果失败: %v", err)
		return
	}
}
func (p *ItasserProcessor) processResult(id uint, seq string) error {
	// Delete queue record
	if err := p.deleteFromQueue(id); err != nil {
		return fmt.Errorf("deleted the queue record failure: %v", err)
	}

	// get id
	var proteinInformation models.ProteinInformation
	if err := database.Database.Where("sequence = ?", seq).Find(&proteinInformation).Error; err != nil {
		return fmt.Errorf("find protein information failed: %v", err)
	}

	// Move the generated file to the static folder
	if err := p.moveModelFile(proteinInformation.ID); err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}

	// Calculate parameters
	CalculateProteinInfomatio(proteinInformation)

	// Generate Ramachandran plot
	Ramachandran(fmt.Sprintf("%d", proteinInformation.ID))

	return nil
}

func (p *ItasserProcessor) createFastaFile(sequence string) error {
	file, err := os.Create("./itasser_example/seq.fasta")
	if err != nil {
		return fmt.Errorf("create file failed: %v", err)
	}
	_, err = file.WriteString(sequence)
	if err != nil {
		return fmt.Errorf("write file failed: %v", err)
	}
	file.Close()
	return nil
}

func (p *ItasserProcessor) moveModelFile(id uint) error {
	// Define source and destination paths
	sourcePath := filepath.Join("itasser_example", "model1.pdb")
	destPath := filepath.Join("static", "models", fmt.Sprintf("%d.pdb", id))

	// Move the file
	err := os.Rename(sourcePath, destPath)
	if err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

func (p *ItasserProcessor) findQueueItem() (*models.ITasserQueue, error) {
	var item models.ITasserQueue
	if err := database.Database.Where("status = ?", "pending").First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (p *ItasserProcessor) deleteFromQueue(id uint) error {
	return database.Database.Model(&models.ITasserQueue{}).Where("id = ?", id).Update("status", "completed").Error
}
