package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"fmt"
	"log"
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
	// Try to get the working slot
	p.workerChan <- struct{}{}

	// Finds the sequence waiting to be processed
	result, err := p.findQueueItem()
	if err != nil {
		<-p.workerChan // Release slot
		return
	}

	// Start the goroutine processing task
	go func(id uint, sequence string) {
		defer func() {
			<-p.workerChan
		}()

		if !IsFasta(sequence) {
			// Delete sequences that are not in FASTA format
			if err := p.deleteFromQueue(id); err != nil {
				log.Printf("Description Failed to delete the queue record: %v", err)
			}
			return
		}

		p.buildModel(id, sequence)
	}(result.ID, result.Sequence)
}

func (p *ItasserProcessor) buildModel(id uint, sequence string) {

	// Empty the input folder
	if err := os.RemoveAll("itasser_example/*"); err != nil {
		log.Printf("Failed to empty the input folder: %v", err)
		return
	}

	// Create a FASTA file
	if err := p.createFastaFile(sequence); err != nil {
		log.Printf("Failed to create the FASTA file: %v", err)
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
		log.Printf("Executing I-Tasser failed: %v, Output: %s", err, output)
		return
	}

	if err := p.processResult(id, sequence); err != nil {
		log.Printf("Processing result failure: %v", err)
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
	if err := database.Database.First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (p *ItasserProcessor) deleteFromQueue(id uint) error {
	return database.Database.Delete(&models.ITasserQueue{}, id).Error
}
