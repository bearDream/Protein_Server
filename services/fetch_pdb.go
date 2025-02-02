package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
)

var fetch_processing = false

func FetchPDB() {
	// If fetch processing is true, it's processing now. Not continue.
	if fetch_processing {
		return
	}
	// If fetch processing is false, start process.
	fetch_processing = true

	// -------  get ids   -------
	resp, err := resty.New().R().Get("https://data.rcsb.org/rest/v1/holdings/current/entry_ids")
	if err != nil {
		log.Println("Unable to make request: ", err)
		fetch_processing = false
		return
	}

	//  -------  parse json  -------
	var entry_ids []string
	if err = json.Unmarshal(resp.Body(), &entry_ids); err != nil {
		log.Println("Unable to parse json: ", err)
		fetch_processing = false
		return
	}

	// -------  build worker pool  -------
	jobs := make(chan string)
	for i := 1; i <= 20; i++ {
		go fetch_worker(jobs)
	}

	// -------  start download job  -------
	for _, entry_id := range entry_ids {
		var count int64
		if err := database.Database.Model(&models.ProteinInformation{}).Where("pdb_id = ?", entry_id).Count(&count).Error; err != nil {
			log.Println("Unable to find protein_information: ", err)
		}
		if count != 0 {
			continue
		}
		jobs <- entry_id
	}
	close(jobs)
}

func fetch_worker(jobs <-chan string) {
	defer func() { fetch_processing = false }()
	// download files
	for j := range jobs {
		// get pdb
		pdb_response, err := resty.New().R().Get("https://files.rcsb.org/view/" + j + ".pdb")
		if err != nil {
			log.Println("Unable to download pdb file: ", err)
		} else {
			// get fasta (sequence)
			fasta_response, err := resty.New().R().Get("https://www.rcsb.org/fasta/entry/" + j + "/display")
			if err != nil {
				log.Println("Unable to download fasta file: ", err)
			} else {
				// get first chain
				lines := strings.Split(fasta_response.String(), "\n")
				if len(lines) >= 2 {
					sequence := strings.TrimSpace(lines[1])
					// save data to  protein_information, then you can get an protein id
					proteinInformation := models.ProteinInformation{
						Sequence: sequence,
						PdbId:    j,
					}
					if err := database.Database.Create(&proteinInformation).Error; err != nil {
						log.Println("Unable to create protein information: ", err)
					} else {
						// get protein img form pdb website directly
						img_response, err := resty.New().R().Get("https://cdn.rcsb.org/images/structures/" + j + "_chain-A.jpeg")
						if err != nil {
							log.Println("Unable to download img file: ", err)
						}
						// img's name is protein id name
						// imgs be save to "static/imgs" folder
						img_filename := filepath.Join("static/imgs", fmt.Sprintf("%d.jpg", proteinInformation.ID))
						if err := os.WriteFile(img_filename, img_response.Body(), 0644); err != nil {
							log.Println("Unable to save img file", err)
						}
						// pdbs be save to "static/imgs" folder
						pdb_filename := filepath.Join("static/models", fmt.Sprintf("%d.pdb", proteinInformation.ID))
						if err := os.WriteFile(pdb_filename, pdb_response.Body(), 0644); err != nil {
							log.Println("Unable to save pdb file", err)
						}
						fmt.Println("Downloaded", j)
						CalculateProteinInfomatio(proteinInformation)
						fmt.Println("Calculated", j)
					}
				}
			}
		}
	}
}
