package models

import (
	"gorm.io/gorm"
)

type ProteinInformation struct {
	gorm.Model
	Sequence            string `gorm:"not null;unique" form:"sequence"`
	BlastInformation    string `form:"blastinformation"`
	RcScore             string `form:"rcscore"`
	Hydrophobicity      string `form:"hydrophobicity"`
	Instability         string `form:"instability"`
	IsoelectricPoint    string `form:"isoelectricpoint"`
	MolecularWeight     string `form:"molecularweight"`
	SolventAccesibility string `form:"solventaccesibility"`
	PdbId               string `form:"pdbid"`
}
