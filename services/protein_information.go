package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"fmt"
)

// Protein Information
func ProteinInformation(sequence string, blastinformation string, structurePredictionTool int64) {
	var proteinInformation models.ProteinInformation
	if err := database.Database.Where("sequence = ?", sequence).Find(&proteinInformation).Error; err != nil {
		return
	}
	if proteinInformation.ID == 0 {
		proteinInformation.Sequence = sequence
		proteinInformation.BlastInformation = blastinformation
		// Create data
		if err := database.Database.Create(&proteinInformation).Error; err != nil {
			fmt.Printf("Create protein information failed: %v", err)
			return
		}
		if structurePredictionTool == 3 {
			// ESMFOld
			ESMFold(sequence)
		}
		if structurePredictionTool == 2 {
			// AlphaFold 2
			AddToAlphaFoldQueue(sequence)
		}
		if structurePredictionTool == 1 {
			// I-Tasser
			AddToITasserQueue(sequence)
		}

	}
}
