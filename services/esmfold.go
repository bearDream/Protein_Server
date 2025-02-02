package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
)

// ESMFold
func ESMFold(sequence string) {
	var proteinInformation models.ProteinInformation
	// get sequence id
	if err := database.Database.Where("sequence = ?", sequence).Find(&proteinInformation).Error; err != nil {
		fmt.Printf("find protein information failed: %v", err)
		return
	}
	// ESMFold's API requires skipping SSL authentication
	// SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	// resty.New() can get an object
	client := resty.New().SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	// Use R() then can use POST GET ...
	// ESMFold API: https://api.esmatlas.com/foldSequence/v1/pdb/
	resp, err := client.R().SetBody(sequence).Post("https://api.esmatlas.com/foldSequence/v1/pdb/")
	if err != nil {
		fmt.Printf("request emsfold failed: %v", err)
	}
	// Save pdb files in static/models fold
	// PDB file's name should be id.pdb
	filename := filepath.Join("static/models", fmt.Sprintf("%d.pdb", proteinInformation.ID))
	if err := os.WriteFile(filename, resp.Body(), 0644); err != nil {
		fmt.Printf("save pdb failed: %v", err)
	}

	// Calculate parameters
	CalculateProteinInfomatio(proteinInformation)
}
