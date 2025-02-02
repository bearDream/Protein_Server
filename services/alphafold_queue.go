package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"fmt"
)

// Add To Alpha Fold Queue
func AddToAlphaFoldQueue(sequence string) {
	var alphafoldQueue models.AlphaFoldQueue
	if err := database.Database.Where("sequence = ?", sequence).Find(&alphafoldQueue).Error; err != nil {
		return
	}
	if alphafoldQueue.ID == 0 {
		alphafoldQueue.Sequence = sequence
		if err := database.Database.Create(&alphafoldQueue).Error; err != nil {
			fmt.Printf("create alphafold queue failed: %v", err)
		}
	}
}
