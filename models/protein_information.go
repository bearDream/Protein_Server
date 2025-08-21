package models

import (
	"gorm.io/gorm"
)

type ProteinInformation struct {
	gorm.Model
	Sequence            string  `gorm:"not null;type:longtext" form:"sequence"`
	BlastInformation    string  `gorm:"type:longtext" form:"blastinformation"`
	RcScore             string  `gorm:"type:longtext" form:"rcscore"`
	Hydrophobicity      string  `gorm:"type:longtext" form:"hydrophobicity"`
	Instability         string  `gorm:"type:longtext" form:"instability"`
	IsoelectricPoint    string  `gorm:"type:longtext" form:"isoelectricpoint"`
	MolecularWeight     string  `gorm:"type:longtext" form:"molecularweight"`
	SolventAccesibility string  `gorm:"type:longtext" form:"solventaccesibility"`
	Size                string  `gorm:"type:longtext" form:"size"`
	PdbId               string  `gorm:"type:longtext" form:"pdbid"`
	ParentId            uint    `gorm:"default:0;index:idx_protein_informations_parent_id" form:"parent_id"`
	Duration            float64 `gorm:"default:0" form:"duration"`
	StructureNum        int     `gorm:"default:0" form:"structure_num"` // RCSB PDB结构数量
}
