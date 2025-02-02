package services

import (
	"log"
	"os/exec"
)

func Ramachandran(protein_id string) {
	cmd := exec.Command("python", "py-script/ramachandran.py", protein_id)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to run ramachandran: %v, print: %s", err, output)
	}
}
