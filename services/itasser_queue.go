package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"fmt"
)

// Add To ITasser Queue
func AddToITasserQueue(sequence string) {
	var itasserQueue models.ITasserQueue
	if err := database.Database.Where("sequence = ?", sequence).Find(&itasserQueue).Error; err != nil {
		return
	}
	if itasserQueue.ID == 0 {
		itasserQueue.Sequence = sequence
		if err := database.Database.Create(&itasserQueue).Error; err != nil {
			fmt.Printf("create alphafold queue failed: %v", err)
		}
	}
}
