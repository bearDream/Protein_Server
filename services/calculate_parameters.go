package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"bytes"
	"fmt"
	"github.com/tealeg/xlsx/v3"
	"math"
	"os"
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

func CalcSa(protein_id string) float64 {
	bashCmd := fmt.Sprintf("source /root/miniconda3/etc/profile.d/conda.sh && conda activate alphafold && python py-scripts/rsa_calculation.py %s", protein_id)
	logger.Info("[CalcSa] 计算可及表面积，输入: %s，命令: %s", protein_id, bashCmd)
	cmd := exec.Command("bash", "-c", bashCmd)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		logger.Error("[CalcSa] 无法计算RSA: %v，输出: %s", err, stdout.String())
		return 0
	}
	score, err := strconv.ParseFloat(strings.Split(stdout.String(), "\n")[0], 64)
	if err != nil {
		logger.Error("[CalcSa] 无法解析RSA: %v，原始输出: %s", err, stdout.String())
		return 0
	}
	return score
}

// CalcRcWithPath 支持传入完整pdb路径，自动提取 protein_id 并调用原有 CalcRc, 删除后缀名
func CalcRcWithPath(pdbPath string) float64 {
	base := pdbPath
	if strings.HasSuffix(base, ".pdb") {
		base = base[:len(base)-4]
	}
	logger.Info("[CalcRcWithPath] 计算RCScore，输入pdb路径: %s，protein_id: %s", pdbPath, base)
	score := CalcRc(base)
	logger.Info("[CalcRcWithPath] RCScore计算结果: %f (protein_id: %s)", score, base)
	return score
}

func CalcRc(protein_id string) float64 {
	cmdPath := "py-scripts/calc_rc"
	logger.Info("[CalcRc] 调用命令: %s %s", cmdPath, protein_id)
	cmd := exec.Command(cmdPath, protein_id)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		logger.Error("[CalcRc] 计算RCScore失败: %v，输出: %s", err, stdout.String())
		return 0
	}
	score, err := strconv.ParseFloat(stdout.String(), 64)
	if err != nil {
		logger.Error("[CalcRc] 解析RCScore失败: %v，原始输出: %s", err, stdout.String())
		return 0
	}
	return score
}

func CalcIi(fasta string) float64 {
	if len(fasta) == 0 {
		return 0
	}
	totalScore := 0.0
	for i := 0; i < len(fasta)-1; i++ {
		dipeptide := fasta[i : i+2]
		value, exists := DIWV[dipeptide]
		if !exists {
			value = 0
		}
		totalScore += value
	}
	return math.Round((10.0/float64(len(fasta))*totalScore)*100) / 100
}

func CalcMw(fasta string) float64 {
	totalWeight := 0.0
	for _, aa := range fasta {
		weight, exists := AMINO_AVERAGE[string(aa)]
		if !exists {
			weight = 0
		}
		totalWeight += weight
	}
	totalWeight += AMINO_AVERAGE["H2O"]
	result := totalWeight / 1000.0
	return math.Round(result*100) / 100 // 保留2位小数
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
	result := sum / float64(count)
	return math.Round(result*10000) / 10000 // 保留4位小数
}

func CalcIp(fasta string) float64 {
	if len(fasta) == 0 {
		return 0
	}

	type PKAScale struct {
		Cterm float64
		PKAsp float64
		PKGlu float64
		PKCys float64
		PKTyr float64
		PKHis float64
		Nterm float64
		PKLys float64
		PKArg float64
	}

	var ipcProteinScale = PKAScale{
		Cterm: 2.869,
		PKAsp: 3.872,
		PKGlu: 4.412,
		PKCys: 7.555,
		PKTyr: 10.85,
		PKHis: 5.637,
		Nterm: 9.094,
		PKLys: 9.052,
		PKArg: 11.84,
	}

	countAminoAcid := func(seq string, aa rune) int {
		count := 0
		for _, r := range seq {
			if r == aa {
				count++
			}
		}
		return count
	}

	scale := ipcProteinScale
	pH := 6.51 // starting pH
	pHprev := 0.0
	pHnext := 14.0
	precision := 0.01 // epsilon for precision

	for {
		// Calculate negative charges
		QN1 := -1.0 / (1.0 + math.Pow(10, (scale.Cterm-pH)))                                 // C-terminus
		QN2 := -float64(countAminoAcid(fasta, 'D')) / (1.0 + math.Pow(10, (scale.PKAsp-pH))) // Aspartic acid
		QN3 := -float64(countAminoAcid(fasta, 'E')) / (1.0 + math.Pow(10, (scale.PKGlu-pH))) // Glutamic acid
		QN4 := -float64(countAminoAcid(fasta, 'C')) / (1.0 + math.Pow(10, (scale.PKCys-pH))) // Cysteine
		QN5 := -float64(countAminoAcid(fasta, 'Y')) / (1.0 + math.Pow(10, (scale.PKTyr-pH))) // Tyrosine

		// Calculate positive charges
		QP1 := float64(countAminoAcid(fasta, 'H')) / (1.0 + math.Pow(10, (pH-scale.PKHis))) // Histidine
		QP2 := 1.0 / (1.0 + math.Pow(10, (pH-scale.Nterm)))                                 // N-terminus
		QP3 := float64(countAminoAcid(fasta, 'K')) / (1.0 + math.Pow(10, (pH-scale.PKLys))) // Lysine
		QP4 := float64(countAminoAcid(fasta, 'R')) / (1.0 + math.Pow(10, (pH-scale.PKArg))) // Arginine

		// Net charge
		netCharge := QN1 + QN2 + QN3 + QN4 + QN5 + QP1 + QP2 + QP3 + QP4

		// Bisection method to find pH where net charge = 0
		if netCharge < 0.0 {
			// Net charge is negative, need to decrease pH
			temp := pH
			pH = pH - ((pH - pHprev) / 2.0)
			pHnext = temp
		} else {
			// Net charge is positive, need to increase pH
			temp := pH
			pH = pH + ((pHnext - pH) / 2.0)
			pHprev = temp
		}

		// Check if we've reached desired precision
		if (pH-pHprev < precision) && (pHnext-pH < precision) {
			return math.Round(pH*100) / 100 // Round to 2 decimal places
		}
	}
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
		logger.Error("无法更新参数: %v", err)
	}
}

// ReadAndCalcParameterExcel 读取parameter_calc.xlsx，批量计算参数，测试参数计算正确性
func ReadAndCalcParameterExcel() error {
	excelPath := "parameter_calc.xlsx"
	pdbDir := "pdb_test_files"

	file, err := xlsx.OpenFile(excelPath)
	if err != nil {
		return err
	}

	if len(file.Sheets) == 0 {
		return fmt.Errorf("Excel文件没有sheet")
	}
	sheet := file.Sheets[0]

	type RowResult struct {
		Protein        string
		Fasta          string
		Size           float64
		Instability    float64
		Hydrophobicity float64
		Isoelectric    float64
		Accessibility  float64
		RCScore        float64
	}

	var results []RowResult

	rowIndex := 0
	sheet.ForEachRow(func(row *xlsx.Row) error {
		if rowIndex == 0 {
			rowIndex++
			return nil // 跳过表头
		}
		// 获取单元格内容，row.GetCell(idx)
		protein := ""
		fasta := ""
		if cell := row.GetCell(0); cell != nil {
			protein = cell.String()
		}
		if cell := row.GetCell(1); cell != nil {
			fasta = cell.String()
		}
		if protein == "" || fasta == "" {
			rowIndex++
			return nil
		}

		// 计算参数
		mw := CalcMw(fasta)
		instability := CalcIi(fasta)
		hydro := CalcH(fasta)
		iso := CalcIp(fasta)

		// 计算Accessibility和RC.Score
		pdbPath := fmt.Sprintf("%s/%s.pdb", pdbDir, protein)
		if _, err := os.Stat(pdbPath); err != nil {
			fmt.Printf("%s 缺少PDB文件，跳过\n", protein)
			rowIndex++
			return nil
		}
		access := CalcSa(pdbPath)
		rc := CalcRcWithPath(pdbPath)

		result := RowResult{
			Protein:        protein,
			Fasta:          fasta,
			Size:           mw,
			Instability:    instability,
			Hydrophobicity: hydro,
			Isoelectric:    iso,
			Accessibility:  access,
			RCScore:        rc,
		}
		results = append(results, result)
		fmt.Printf("%s\t%s\t%.2f\t%.2f\t%.4f\t%.2f\t%.2f\t%.2f\n", protein, fasta, mw, instability, hydro, iso, access, rc)
		rowIndex++
		return nil
	})
	return nil
}
