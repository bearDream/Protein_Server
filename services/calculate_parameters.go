package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// AMINO_AVERAGE：存储各氨基酸和水分子的平均分子量（单位：道尔顿）
// AMINO_AVERAGE: Stores the average molecular weight of each amino acid and water molecule (unit: Dalton)
var AMINO_AVERAGE = map[string]float64{
	"A":   71.0788,
	"R":   156.1875,
	"N":   114.1038,
	"D":   115.0886,
	"C":   103.1388,
	"E":   129.1155,
	"Q":   128.1307,
	"G":   57.0519,
	"H":   137.1411,
	"I":   113.1594,
	"L":   113.1594,
	"K":   128.1741,
	"M":   131.1926,
	"F":   147.1766,
	"P":   97.1167,
	"S":   87.0782,
	"T":   101.1051,
	"W":   186.2132,
	"Y":   163.1760,
	"V":   99.1326,
	"U":   150.0388,
	"O":   237.3018,
	"X":   0,
	"H2O": 18.01524,
}

// DIWV字典存储二肽（两个相邻氨基酸）的相互作用值，范围从-15.91到58.28
// The DIWV dictionary stores the interaction values of dipeptides (two adjacent amino acids), ranging from -15.91 to 58.28
var DIWV = map[string]float64{
	"WW": 1.0, "WC": 1.0, "WM": 24.68, "WH": 24.68, "WY": 1.0,
	"WF": 1.0, "WQ": 1.0, "WN": 13.34, "WI": 1.0, "WR": 1.0,
	"WD": 1.0, "WP": 1.0, "WT": -14.03, "WK": 1.0, "WE": 1.0,
	"WV": -7.49, "WS": 1.0, "WG": -9.37, "WA": -14.03, "WL": 13.34,
	"CW": 24.68, "CC": 1.0, "CM": 33.6, "CH": 33.6, "CY": 1.0,
	"CF": 1.0, "CQ": -6.54, "CN": 1.0, "CI": 1.0, "CR": 1.0,
	"CD": 20.26, "CP": 20.26, "CT": 33.6, "CK": 1.0, "CE": 1.0,
	"CV": -6.54, "CS": 1.0, "CG": 1.0, "CA": 1.0, "CL": 20.26,
	"MW": 1.0, "MC": 1.0, "MM": -1.88, "MH": 58.28, "MY": 24.68,
	"MF": 1.0, "MQ": -6.54, "MN": 1.0, "MI": 1.0, "MR": -6.54,
	"MD": 1.0, "MP": 44.94, "MT": -1.88, "MK": 1.0, "ME": 1.0,
	"MV": 1.0, "MS": 44.94, "MG": 1.0, "MA": 13.34, "ML": 1.0,
	"HW": -1.88, "HC": 1.0, "HM": 1.0, "HH": 1.0, "HY": 44.94,
	"HF": -9.37, "HQ": 1.0, "HN": 24.68, "HI": 44.94, "HR": 1.0,
	"HD": 1.0, "HP": -1.88, "HT": -6.54, "HK": 24.68, "HE": 1.0,
	"HV": 1.0, "HS": 1.0, "HG": -9.37, "HA": 1.0, "HL": 1.0,
	"YW": -9.37, "YC": 1.0, "YM": 44.94, "YH": 13.34, "YY": 13.34,
	"YF": 1.0, "YQ": -6.54, "YN": 1.0, "YI": 1.0, "YR": -15.91,
	"YD": 24.68, "YP": 13.34, "YT": -7.49, "YK": 1.0, "YE": -6.54,
	"YV": 1.0, "YS": 1.0, "YG": -7.49, "YA": 24.68, "YL": 1.0,
	"FW": 1.0, "FC": 1.0, "FM": 1.0, "FH": -9.37, "FY": 33.6,
	"FF": 1.0, "FQ": 1.0, "FN": 1.0, "FI": 1.0, "FR": 1.0,
	"FD": 13.34, "FP": 20.26, "FT": 1.0, "FK": -14.03, "FE": 1.0,
	"FV": 1.0, "FS": 1.0, "FG": 1.0, "FA": 1.0, "FL": 1.0,
	"QW": 1.0, "QC": -6.54, "QM": 1.0, "QH": 1.0, "QY": -6.54,
	"QF": -6.54, "QQ": 20.26, "QN": 1.0, "QI": 1.0, "QR": 1.0,
	"QD": 20.26, "QP": 20.26, "QT": 1.0, "QK": 1.0, "QE": 20.26,
	"QV": -6.54, "QS": 44.94, "QG": 1.0, "QA": 1.0, "QL": 33.6,
	"NW": -9.37, "NC": -1.88, "NM": 1.0, "NH": 1.0, "NY": 1.0,
	"NF": -14.03, "NQ": -6.54, "NN": 1.0, "NI": 44.94, "NR": 1.0,
	"ND": 1.0, "NP": -1.88, "NT": -7.49, "NK": 24.68, "NE": 1.0,
	"NV": 1.0, "NS": 1.0, "NG": -14.03, "NA": 1.0, "NL": 1.0,
	"IW": 1.0, "IC": 1.0, "IM": 1.0, "IH": 13.34, "IY": 1.0,
	"IF": 1.0, "IQ": 1.0, "IN": 1.0, "II": 1.0, "IR": 1.0,
	"ID": 1.0, "IP": -1.88, "IT": 1.0, "IK": -7.49, "IE": 44.94,
	"IV": -7.49, "IS": 1.0, "IG": 1.0, "IA": 1.0, "IL": 20.26,
	"RW": 58.28, "RC": 1.0, "RM": -6.54, "RH": 20.26, "RY": -15.91,
	"RF": 1.0, "RQ": 20.26, "RN": 13.34, "RI": 58.28, "RR": 58.28,
	"RD": 1.0, "RP": 20.26, "RT": 1.0, "RK": 1.0, "RE": 1.0,
	"RV": 1.0, "RS": 44.94, "RG": -7.49, "RA": 1.0, "RL": 20.26,
	"DW": -14.03, "DC": 1.0, "DM": 1.0, "DH": 1.0, "DY": 1.0,
	"DF": -6.54, "DQ": 1.0, "DN": 1.0, "DI": 1.0, "DR": -6.54,
	"DD": 1.0, "DP": 1.0, "DT": -14.03, "DK": -7.49, "DE": 1.0,
	"DV": 1.0, "DS": 20.26, "DG": 1.0, "DA": 1.0, "DL": 1.0,
	"PW": 1.0, "PC": -6.54, "PM": -6.54, "PH": 20.26, "PY": 1.0,
	"PF": 20.26, "PQ": 20.26, "PN": 1.0, "PI": 1.0, "PR": -6.54,
	"PD": -6.54, "PP": 20.26, "PT": 1.0, "PK": 1.0, "PE": 18.38,
	"PV": 20.26, "PS": 20.26, "PG": 1.0, "PA": 20.26, "PL": -6.54,
	"TW": -14.03, "TC": 1.0, "TM": 1.0, "TH": 1.0, "TY": 1.0,
	"TF": 13.34, "TQ": -6.54, "TN": -14.03, "TI": 1.0, "TR": 1.0,
	"TD": 1.0, "TP": 1.0, "TT": 1.0, "TK": 1.0, "TE": 20.26,
	"TV": 1.0, "TS": 1.0, "TG": -7.49, "TA": 1.0, "TL": 1.0,
	"KW": 1.0, "KC": 1.0, "KM": 33.6, "KH": 1.0, "KY": 1.0,
	"KF": -14.03, "KQ": 24.68, "KN": 1.0, "KI": -7.49, "KR": 33.6,
	"KD": 1.0, "KP": -6.54, "KT": 1.0, "KK": 1.0, "KE": 1.0,
	"KV": 1.0, "KS": 1.0, "KG": -7.49, "KA": 1.0, "KL": -7.49,
	"EW": -14.03, "EC": 44.94, "EM": 1.0, "EH": -6.54, "EY": 1.0,
	"EF": 1.0, "EQ": 20.26, "EN": 1.0, "EI": 20.26, "ER": 1.0,
	"ED": 20.26, "EP": 20.26, "ET": 1.0, "EK": 1.0, "EE": 33.6,
	"EV": 1.0, "ES": 20.26, "EG": 1.0, "EA": 1.0, "EL": 1.0,
	"VW": -7.49, "VC": 1.0, "VM": 1.0, "VH": 1.0, "VY": -6.54,
	"VF": 1.0, "VQ": 1.0, "VN": 1.0, "VI": 1.0, "VR": 1.0,
	"VD": -14.03, "VP": 20.26, "VT": -7.49, "VK": -1.88, "VE": 1.0,
	"VV": 1.0, "VS": 1.0, "VG": -7.49, "VA": 1.0, "VL": 1.0,
	"SW": 1.0, "SC": 33.6, "SM": 1.0, "SH": 1.0, "SY": 1.0,
	"SF": 1.0, "SQ": 20.26, "SN": 1.0, "SI": 1.0, "SR": 20.26,
	"SD": 1.0, "SP": 44.94, "ST": 1.0, "SK": 1.0, "SE": 20.26,
	"SV": 1.0, "SS": 20.26, "SG": 1.0, "SA": 1.0, "SL": 1.0,
	"GW": 13.34, "GC": 1.0, "GM": 1.0, "GH": 1.0, "GY": -7.49,
	"GF": 1.0, "GQ": 1.0, "GN": -14.03, "GI": -7.49, "GR": 1.0,
	"GD": 1.0, "GP": 1.0, "GT": -7.49, "GK": -7.49, "GE": -6.54,
	"GV": 1.0, "GS": 1.0, "GG": 13.34, "GA": 1.0, "GL": 1.0,
	"AW": 1.0, "AC": 44.94, "AM": 1.0, "AH": -7.49, "AY": 1.0,
	"AF": 1.0, "AQ": 1.0, "AN": 1.0, "AI": 1.0, "AR": 1.0,
	"AD": -7.49, "AP": 20.26, "AT": 1.0, "AK": 1.0, "AE": 1.0,
	"AV": 1.0, "AS": 1.0, "AG": 1.0, "AA": 1.0, "AL": 1.0,
	"LW": 24.68, "LC": 1.0, "LM": 1.0, "LH": 1.0, "LY": 1.0,
	"LF": 1.0, "LQ": 33.6, "LN": 1.0, "LI": 1.0, "LR": 20.26,
	"LD": 1.0, "LP": 20.26, "LT": 1.0, "LK": -7.49, "LE": 1.0,
	"LV": 1.0, "LS": 1.0, "LG": 1.0, "LA": 1.0, "LL": 1.0,
}

// HydropathyValues是氨基酸疏水性值对照表
// 正值表示疏水性氨基酸(如I: 4.5)
// 负值表示亲水性氨基酸(如R: -4.5)
// HydropathyValues is a comparison table of amino acid hydrophobicity values
// Positive values indicate hydrophobic amino acids (e.g. I: 4.5)
// Negative values indicate hydrophilic amino acids (e.g. R: -4.5)
var HydropathyValues = map[string]float64{
	"A": 1.8,  // Alanine
	"R": -4.5, // Arginine
	"N": -3.5, // Asparagine
	"D": -3.5, // Aspartic acid
	"C": 2.5,  // Cysteine
	"Q": -3.5, // Glutamine
	"E": -3.5, // Glutamic acid
	"G": -0.4, // Glycine
	"H": -3.2, // Histidine
	"I": 4.5,  // Isoleucine
	"L": 3.8,  // Leucine
	"K": -3.9, // Lysine
	"M": 1.9,  // Methionine
	"F": 2.8,  // Phenylalanine
	"P": -1.6, // Proline
	"S": -0.8, // Serine
	"T": -0.7, // Threonine
	"W": -0.9, // Tryptophan
	"Y": -1.3, // Tyrosine
	"V": 4.2,  // Valine
}

func CalcAll(sequence string, protein_id string) (rc, sa, ii, mw, h, ip float64) {
	// rcScore
	rc = CalcRc(protein_id)
	// Solvent Accessibility
	sa = CalcSa(protein_id)
	// Instability
	ii = CalcIi(sequence)
	// Size（Molecular weight）
	mw = CalcMw(sequence)
	// Hydrophobicity
	h = CalcH(sequence)
	// Isoelectric Point
	ip = CalcIp(sequence)
	Ramachandran(protein_id)
	return
}

func CalcRc(protein_id string) float64 {
	cmd := exec.Command("calc_rc", "static/models/"+protein_id)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return 0
	}

	score, err := strconv.ParseFloat(stdout.String(), 64)
	if err != nil {
		return 0
	}
	return score
}

func CalcSa(protein_id string) float64 {
	cmd := exec.Command("python", "py-script/rsa.py", protein_id+".pdb")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		log.Println("Unable to calculate rsa: ", err)
		return 0
	}
	score, err := strconv.ParseFloat(strings.Split(stdout.String(), "\n")[0], 64)
	if err != nil {
		log.Println("Unable to parse rsa: ", err)
		return 0
	}
	return score
}

func CalcIi(fasta string) float64 {
	sum := 0.0
	for i := 0; i < len(fasta)-1; i++ {
		dipeptide := fasta[i : i+2]
		if value, exists := DIWV[dipeptide]; exists {
			sum += value
		}
	}
	return (10.0 / float64(len(fasta))) * sum
}

func CalcMw(fasta string) float64 {
	totalWeight := 0.0
	for _, aa := range fasta {
		weight, exists := AMINO_AVERAGE[string(aa)]
		if !exists {
			continue
		}
		totalWeight += weight
	}
	totalWeight += AMINO_AVERAGE["H2O"]
	return totalWeight
}

// 计算序列的平均疏水性
// Calculate the mean hydrophobicity of the sequence
func CalcH(fasta string) float64 {
	sum := 0.0
	count := 0
	for _, aa := range fasta {
		if value, exists := HydropathyValues[string(aa)]; exists {
			sum += value
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func CalcIp(fasta string) float64 {
	cmd := exec.Command("python", "py-script/ipc.py", fasta)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return 0
	}
	score, err := strconv.ParseFloat(strings.Split(stdout.String(), "\n")[0], 64)
	if err != nil {
		return 0
	}
	return score
}

func CalculateProteinInfomatio(proteinInformation models.ProteinInformation) {
	rc, sa, ii, mw, h, ip := CalcAll(proteinInformation.Sequence, fmt.Sprintf("%d", proteinInformation.ID))
	proteinInformation.Hydrophobicity = fmt.Sprintf("%f", h)
	proteinInformation.Instability = fmt.Sprintf("%f", ii)
	proteinInformation.IsoelectricPoint = fmt.Sprintf("%f", ip)
	proteinInformation.MolecularWeight = fmt.Sprintf("%f", mw)
	proteinInformation.RcScore = fmt.Sprintf("%f", rc)
	proteinInformation.SolventAccesibility = fmt.Sprintf("%f", sa)
	if err := database.Database.Updates(&proteinInformation).Error; err != nil {
		log.Println("Unable to update parameters: ", err)
	}
}
