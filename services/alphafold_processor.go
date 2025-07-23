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

	// Empty the input folder
	if err := os.RemoveAll("alphafold_input/*"); err != nil {
		logger.Error("清空输入文件夹失败: %v", err)
		return
	}

	// Create a FASTA file
	if err := p.createFastaFile(sequence); err != nil {
		logger.Error("创建FASTA文件失败: %v", err)
		return
	}

	// Run the AlphaFold command
	cmd := exec.Command("bash", "-c", "source /root/miniconda3/etc/profile.d/conda.sh && conda activate alphafold && bash ../alphafold/run_alphafold.sh -d ../alphadata -o ./alphafold_output -f ./alphafold_input/query.fasta -t 2021-11-01 -g False -c reduced_dbs")

	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("执行AlphaFold失败: %v, 输出: %s", err, output)
		return
	}

	// Processing result
	if err := p.processResult(id, sequence); err != nil {
		logger.Error("处理结果失败: %v", err)
		return
	}
}
func (p *AlphaProcessor) processResult(id uint, seq string) error {
	// Delete queue record
	if err := p.deleteFromQueue(id); err != nil {
		return fmt.Errorf("deleted the queue record failure: %v", err)
	}

	// find id
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

func (p *AlphaProcessor) createFastaFile(sequence string) error {
	file, err := os.Create("./alphafold_input/query.fasta")
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

func (p *AlphaProcessor) moveModelFile(id uint) error {
	// Define source and destination paths
	sourcePath := filepath.Join("alphafold_output", "query", "unrelaxed_model_1.pdb")
	destPath := filepath.Join("static", "models", fmt.Sprintf("%d.pdb", id))

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

func (p *AlphaProcessor) deleteFromQueue(id uint) error {
	return database.Database.Model(&models.AlphaFoldQueue{}).Where("id = ?", id).Update("status", "completed").Error
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
