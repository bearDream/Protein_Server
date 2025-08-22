package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// SearchPdbByParamResult 搜索结果结构体
type SearchPdbByParamResult struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// SearchPdbByParam 根据参数搜索PDB
func SearchPdbByParam(
	sequence string,
	pdbId string,
	rcScore string,
	hydrophobicity string,
	instability string,
	isoelectricPoint string,
	size string,
	solventAccesibility string,
	current int,
	pageSize int,
	sort string,
) SearchPdbByParamResult {

	// 构建查询条件
	query := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ?", true)

	// 序列搜索（模糊匹配）
	if sequence != "" {
		query = query.Where("fasta LIKE ?", "%"+sequence+"%")
	}

	// PDB ID 精确匹配
	if pdbId != "" {
		query = query.Where("pdb_id = ?", pdbId)
	}

	// RC Score 范围搜索
	if rcScore != "," {
		parts := strings.Split(rcScore, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) >= ? AND CAST(rc_score AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL'")
			}
		}
	}

	// Hydrophobicity 范围搜索
	if hydrophobicity != "," {
		parts := strings.Split(hydrophobicity, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) >= ? AND CAST(hydrophobicity AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL'")
			}
		}
	}

	// Instability 范围搜索
	if instability != "," {
		parts := strings.Split(instability, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) >= ? AND CAST(instability AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL'")
			}
		}
	}

	// Isoelectric Point 范围搜索
	if isoelectricPoint != "," {
		parts := strings.Split(isoelectricPoint, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) >= ? AND CAST(isoelectric_point AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL'")
			}
		}
	}

	// Size 范围搜索
	if size != "," {
		parts := strings.Split(size, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("size > 0 AND size >= ? AND size <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("size > 0 AND size <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("size > 0 AND size >= ?", startVal-1e-5)
			} else {
				query = query.Where("size > 0")
			}
		}
	}

	// Solvent Accessibility 范围搜索
	if solventAccesibility != "," {
		parts := strings.Split(solventAccesibility, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) >= ? AND CAST(solvent_accesibility AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL'")
			}
		}
	}

	// 获取所有符合条件的记录用于统计
	var totalList []models.PDBParameter
	if err := query.Find(&totalList).Error; err != nil {
		return SearchPdbByParamResult{Error: "Network error!"}
	}

	// 计算统计信息
	total := int64(len(totalList))
	var missRcScore, missSolventAccesibility, missProt, missIso int64
	var totalTime float64

	for _, item := range totalList {
		// 统计缺失数据 - 考虑所有可能的缺失值情况
		if item.RcScore == "" || item.RcScore == "0" || item.RcScore == "null" || item.RcScore == "NULL" || item.RcScore == "None" || item.RcScore == "none" {
			missRcScore++
		}
		if item.SolventAccesibility == "" || item.SolventAccesibility == "0" || item.SolventAccesibility == "null" || item.SolventAccesibility == "NULL" || item.SolventAccesibility == "None" || item.SolventAccesibility == "none" {
			missSolventAccesibility++
		}
		if (item.Hydrophobicity == "" || item.Hydrophobicity == "0" || item.Hydrophobicity == "null" || item.Hydrophobicity == "NULL" || item.Hydrophobicity == "None" || item.Hydrophobicity == "none") &&
			(item.Instability == "" || item.Instability == "0" || item.Instability == "null" || item.Instability == "NULL" || item.Instability == "None" || item.Instability == "none") &&
			item.Size == 0 {
			missProt++
		}
		if item.IsoelectricPoint == "" || item.IsoelectricPoint == "0" || item.IsoelectricPoint == "null" || item.IsoelectricPoint == "NULL" || item.IsoelectricPoint == "None" || item.IsoelectricPoint == "none" {
			missIso++
		}

		// 累计时间
		totalTime += item.Duration
	}

	// 计算平均时间
	var meanTime float64
	if total > 0 {
		meanTime = totalTime / float64(total)
	}

	// 获取非蛋白质数量
	var notProtein int64
	if err := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ?", false).Count(&notProtein).Error; err != nil {
		return SearchPdbByParamResult{Error: "Network error!"}
	}

		// 读取PDB文件目录并计算大小
	var all int64
	var sizeGB float64
	
	// 尝试读取目录
	if files, err := os.ReadDir("../PROFASA-PDB-GO/data"); err == nil {
		var totalSize int64
		for _, file := range files {
			if !file.IsDir() {
				// 统计所有文件数量
				all++
				// 计算文件大小
				if info, err := file.Info(); err == nil {
					totalSize += info.Size()
				}
			}
		}
		// 转换为GB
		sizeGB = float64(totalSize) / (1024 * 1024 * 1024)
	} else {
		// 如果读取目录失败，使用默认值
		sizeGB = 186.67
	}

	// 分页查询
	offset := (current - 1) * pageSize
	if sort != "" {
		sortParts := strings.Split(sort, ",")
		if len(sortParts) == 2 {
			field := sortParts[0]
			order := sortParts[1]
			query = query.Order(field + " " + order)
		}
	}

	// 执行分页查询
	var results []models.PDBParameter
	if err := query.Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return SearchPdbByParamResult{Error: "Network error!"}
	}

	// 构建返回结果
	data := map[string]interface{}{
		"all":                      all,
		"size":                     sizeGB,
		"notProtein":               notProtein,
		"total":                    total,
		"miss_rcScore":             missRcScore,
		"miss_solventAccesibility": missSolventAccesibility,
		"miss_prot":                missProt,
		"miss_iso":                 missIso,
		"total_time":               totalTime,
		"mean_time":                meanTime,
		"list":                     results,
	}

	return SearchPdbByParamResult{Data: data}
}

// GetPDBInformationByIdResult 根据ID获取PDB信息的结果结构体
type GetPDBInformationByIdResult struct {
	Data  *models.PDBParameter `json:"data,omitempty"`
	Error string               `json:"error,omitempty"`
}

// GetPDBInformationById 根据PDB ID获取PDB信息
func GetPDBInformationById(pdbId string) GetPDBInformationByIdResult {
	var result models.PDBParameter

	if err := database.Database.Where("pdb_id = ?", pdbId).First(&result).Error; err != nil {
		if err.Error() == "record not found" {
			return GetPDBInformationByIdResult{Data: nil}
		}
		return GetPDBInformationByIdResult{Error: "Network error!"}
	}

	return GetPDBInformationByIdResult{Data: &result}
}

// SeqTimeTableItem 序列时间表项结构体
type SeqTimeTableItem struct {
	Seq      string `json:"seq"`
	Length   int    `json:"length"`
	Duration string `json:"duration"`
	Type     string `json:"type"`
}

// GetSeqTimeTableResult 获取序列时间表的结果结构体
type GetSeqTimeTableResult struct {
	Data  []SeqTimeTableItem `json:"data,omitempty"`
	Error string             `json:"error,omitempty"`
}

// formatDuring 格式化持续时间
func formatDuring(durationMs int64) string {
	hours := durationMs / (1000 * 60 * 60)
	minutes := (durationMs % (1000 * 60 * 60)) / (1000 * 60)
	seconds := (durationMs % (1000 * 60)) / 1000

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// GetSeqTimeTable 获取序列时间表
func GetSeqTimeTable() GetSeqTimeTableResult {
	var proteinInformations []models.ProteinInformation

	// 查询所有蛋白质信息记录
	if err := database.Database.Find(&proteinInformations).Error; err != nil {
		return GetSeqTimeTableResult{Error: "Network error!"}
	}

	var result []SeqTimeTableItem

	for _, item := range proteinInformations {
		// 计算处理时间（毫秒）
		durationMs := item.UpdatedAt.UnixMilli() - item.CreatedAt.UnixMilli()

		// 过滤掉处理时间超过100小时的记录
		if durationMs/(1000*60*60) >= 100 {
			continue
		}

		// 格式化持续时间
		duration := formatDuring(durationMs)

		// 根据处理时间判断类型
		var seqType string
		if durationMs/(1000*60*60) > 2 {
			seqType = "itasser"
		} else {
			seqType = "alpha"
		}

		// 创建时间表项
		timeTableItem := SeqTimeTableItem{
			Seq:      item.Sequence,
			Length:   len(item.Sequence),
			Duration: duration,
			Type:     seqType,
		}

		result = append(result, timeTableItem)
	}

	return GetSeqTimeTableResult{Data: result}
}

// GetPDBParameterListResult 获取PDB参数列表的结果结构体
type GetPDBParameterListResult struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// GetPDBParameterList 获取PDB参数列表（包含统计信息）
func GetPDBParameterList(
	fasta string,
	pdbId string,
	rcScore string,
	hydrophobicity string,
	instability string,
	isoelectricPoint string,
	size string,
	solventAccesibility string,
	current int,
	pageSize int,
	sort string,
) GetPDBParameterListResult {

	// 构建查询条件 - 复用SearchPdbByParam的逻辑
	query := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ?", true)

	// PDB ID 精确匹配
	if pdbId != "" {
		query = query.Where("pdb_id = ?", pdbId)
	}

	// 序列搜索（子字符串匹配）
	if fasta != "" {
		query = query.Where("fasta LIKE ?", "%"+fasta+"%")
	}

	// RC Score 范围搜索
	if rcScore != "," {
		parts := strings.Split(rcScore, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) >= ? AND CAST(rc_score AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL'")
			}
		}
	}

	// Hydrophobicity 范围搜索
	if hydrophobicity != "," {
		parts := strings.Split(hydrophobicity, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) >= ? AND CAST(hydrophobicity AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL'")
			}
		}
	}

	// Instability 范围搜索
	if instability != "," {
		parts := strings.Split(instability, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) >= ? AND CAST(instability AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL'")
			}
		}
	}

	// Isoelectric Point 范围搜索
	if isoelectricPoint != "," {
		parts := strings.Split(isoelectricPoint, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) >= ? AND CAST(isoelectric_point AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL'")
			}
		}
	}

	// Size 范围搜索
	if size != "," {
		parts := strings.Split(size, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("size > 0 AND size >= ? AND size <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("size > 0 AND size <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("size > 0 AND size >= ?", startVal-1e-5)
			} else {
				query = query.Where("size > 0")
			}
		}
	}

	// Solvent Accessibility 范围搜索
	if solventAccesibility != "," {
		parts := strings.Split(solventAccesibility, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) >= ? AND CAST(solvent_accesibility AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				query = query.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				query = query.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				query = query.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL'")
			}
		}
	}

	// 创建基础查询条件的副本，用于统计查询
	// 确保只查询6个参数都不为空的数据
	baseQuery := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ? AND (rc_score IS NOT NULL AND rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND rc_score != 'None' AND rc_score != 'none') AND (solvent_accesibility IS NOT NULL AND solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND solvent_accesibility != 'None' AND solvent_accesibility != 'none') AND (hydrophobicity IS NOT NULL AND hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND hydrophobicity != 'None' AND hydrophobicity != 'none') AND (isoelectric_point IS NOT NULL AND isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND isoelectric_point != 'None' AND isoelectric_point != 'none') AND (instability IS NOT NULL AND instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND instability != 'None' AND instability != 'none') AND (size IS NOT NULL AND size != 0)", true)

	// 应用所有筛选条件到基础查询
	if pdbId != "" {
		baseQuery = baseQuery.Where("pdb_id = ?", pdbId)
	}
	if fasta != "" {
		baseQuery = baseQuery.Where("fasta LIKE ?", "%"+fasta+"%")
	}
	if rcScore != "," {
		parts := strings.Split(rcScore, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) >= ? AND CAST(rc_score AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				baseQuery = baseQuery.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND CAST(rc_score AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				baseQuery = baseQuery.Where("rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL'")
			}
		}
	}
	// 继续应用其他筛选条件...
	if hydrophobicity != "," {
		parts := strings.Split(hydrophobicity, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) >= ? AND CAST(hydrophobicity AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				baseQuery = baseQuery.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND CAST(hydrophobicity AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				baseQuery = baseQuery.Where("hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL'")
			}
		}
	}
	if instability != "," {
		parts := strings.Split(instability, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) >= ? AND CAST(instability AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				baseQuery = baseQuery.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND CAST(instability AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				baseQuery = baseQuery.Where("instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL'")
			}
		}
	}
	if isoelectricPoint != "," {
		parts := strings.Split(isoelectricPoint, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) >= ? AND CAST(isoelectric_point AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				baseQuery = baseQuery.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND CAST(isoelectric_point AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				baseQuery = baseQuery.Where("isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL'")
			}
		}
	}
	if size != "," {
		parts := strings.Split(size, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("size > 0 AND size >= ? AND size <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("size > 0 AND size <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				baseQuery = baseQuery.Where("size > 0 AND size >= ?", startVal-1e-5)
			} else {
				baseQuery = baseQuery.Where("size > 0")
			}
		}
	}
	if solventAccesibility != "," {
		parts := strings.Split(solventAccesibility, ",")
		if len(parts) == 2 {
			start, end := parts[0], parts[1]
			if start != "" && end != "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) >= ? AND CAST(solvent_accesibility AS DECIMAL(10,5)) <= ?", startVal-1e-5, endVal+1e-5)
			} else if start == "" && end != "" {
				endVal, _ := strconv.ParseFloat(end, 64)
				baseQuery = baseQuery.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) <= ?", endVal+1e-5)
			} else if start != "" && end == "" {
				startVal, _ := strconv.ParseFloat(start, 64)
				baseQuery = baseQuery.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND CAST(solvent_accesibility AS DECIMAL(10,5)) >= ?", startVal-1e-5)
			} else {
				baseQuery = baseQuery.Where("solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL'")
			}
		}
	}

	// 统计6个参数都不为空的数据量（rc_score, solvent_accesibility, hydrophobicity, isoelectric_point, instability, size）
	var total int64
	if err := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ? AND (rc_score IS NOT NULL AND rc_score != '' AND rc_score != '0' AND rc_score != 'null' AND rc_score != 'NULL' AND rc_score != 'None' AND rc_score != 'none') AND (solvent_accesibility IS NOT NULL AND solvent_accesibility != '' AND solvent_accesibility != '0' AND solvent_accesibility != 'null' AND solvent_accesibility != 'NULL' AND solvent_accesibility != 'None' AND solvent_accesibility != 'none') AND (hydrophobicity IS NOT NULL AND hydrophobicity != '' AND hydrophobicity != '0' AND hydrophobicity != 'null' AND hydrophobicity != 'NULL' AND hydrophobicity != 'None' AND hydrophobicity != 'none') AND (isoelectric_point IS NOT NULL AND isoelectric_point != '' AND isoelectric_point != '0' AND isoelectric_point != 'null' AND isoelectric_point != 'NULL' AND isoelectric_point != 'None' AND isoelectric_point != 'none') AND (instability IS NOT NULL AND instability != '' AND instability != '0' AND instability != 'null' AND instability != 'NULL' AND instability != 'None' AND instability != 'none') AND (size IS NOT NULL AND size != 0)", true).Count(&total).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}

	// 使用全表统计（不应用筛选条件）计算缺失数据统计
	var missRcScore, missSolventAccesibility, missProt, missIso int64
	var totalTime float64
	var avgTime float64

		// 统计缺失数据（基于全表）- 考虑所有可能的缺失值情况
	
	// 统计缺失 rc_score 的记录数（NULL、空字符串、'0'、'null'、'NULL'、'None'等）
	if err := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ? AND (rc_score IS NULL OR rc_score = '' OR rc_score = '0' OR rc_score = 'null' OR rc_score = 'NULL' OR rc_score = 'None' OR rc_score = 'none')", true).Count(&missRcScore).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}
	
	// 统计缺失 solvent_accesibility 的记录数（NULL、空字符串、'0'、'null'、'NULL'、'None'等）
	if err := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ? AND (solvent_accesibility IS NULL OR solvent_accesibility = '' OR solvent_accesibility = '0' OR solvent_accesibility = 'null' OR solvent_accesibility = 'NULL' OR solvent_accesibility = 'None' OR solvent_accesibility = 'none')", true).Count(&missSolventAccesibility).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}
	
	// 统计缺失蛋白质参数的记录数（疏水性和不稳定性为缺失值，且大小为0或NULL）
	if err := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ? AND (hydrophobicity IS NULL OR hydrophobicity = '' OR hydrophobicity = '0' OR hydrophobicity = 'null' OR hydrophobicity = 'NULL' OR hydrophobicity = 'None' OR hydrophobicity = 'none') AND (instability IS NULL OR instability = '' OR instability = '0' OR instability = 'null' OR instability = 'NULL' OR instability = 'None' OR instability = 'none') AND (size IS NULL OR size = 0)", true).Count(&missProt).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}
	
	// 统计缺失 isoelectric_point 的记录数（NULL、空字符串、'0'、'null'、'NULL'、'None'等）
	if err := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ? AND (isoelectric_point IS NULL OR isoelectric_point = '' OR isoelectric_point = '0' OR isoelectric_point = 'null' OR isoelectric_point = 'NULL' OR isoelectric_point = 'None' OR isoelectric_point = 'none')", true).Count(&missIso).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}

	// 计算总时间和平均时间（基于全表）
	var timeStats struct {
		TotalTime float64 `gorm:"column:total_time"`
		AvgTime   float64 `gorm:"column:avg_time"`
	}
	if err := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ?", true).Select("SUM(duration) as total_time, AVG(duration) as avg_time").Scan(&timeStats).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}

	totalTime = timeStats.TotalTime
	avgTime = timeStats.AvgTime

	// 获取非蛋白质数量
	var notProtein int64
	if err := database.Database.Model(&models.PDBParameter{}).Where("is_protein = ?", false).Count(&notProtein).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}

		// 获取所有记录数（包括蛋白质和非蛋白质）
	var all int64
	var sizeGB float64
	
	// 查询所有记录数
	if err := database.Database.Model(&models.PDBParameter{}).Count(&all).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}
	
	// 计算 ../PROFASA-PDB-GO/data 文件夹下所有文件的总大小
	if files, err := os.ReadDir("../PROFASA-PDB-GO/data"); err == nil {
		var totalSize int64
		for _, file := range files {
			if !file.IsDir() {
				if info, err := file.Info(); err == nil {
					totalSize += info.Size()
				}
			}
		}
		// 转换为GB
		sizeGB = float64(totalSize) / (1024 * 1024 * 1024)
	} else {
		// 如果读取目录失败，使用默认值
		sizeGB = 196.5
	}

	// 分页查询
	offset := (current - 1) * pageSize
	if sort != "" {
		sortParts := strings.Split(sort, ",")
		if len(sortParts) == 2 {
			field := sortParts[0]
			order := sortParts[1]
			baseQuery = baseQuery.Order(field + " " + order)
		}
	}

	// 执行分页查询
	var results []models.PDBParameter
	if err := baseQuery.Offset(offset).Limit(pageSize).Find(&results).Error; err != nil {
		return GetPDBParameterListResult{Error: "Network error!"}
	}

	// 构建返回结果
	data := map[string]interface{}{
		"all":                      all,
		"size":                     sizeGB,
		"notProtein":               notProtein,
		"total":                    total,
		"miss_rcScore":             missRcScore,
		"miss_solventAccesibility": missSolventAccesibility,
		"miss_prot":                missProt,
		"miss_iso":                 missIso,
		"total_time":               totalTime,
		"mean_time":                avgTime,
		"list":                     results,
	}

	return GetPDBParameterListResult{Data: data}
}

// CalcAllPDBParamsResult 批量计算PDB参数的结果结构
type CalcAllPDBParamsResult struct {
	Data  map[string]interface{} `json:"data"`
	Error string                 `json:"error"`
}

// CalcAllPDBParams 批量计算所有PDB参数的6个参数
func CalcAllPDBParams(batchSize int) CalcAllPDBParamsResult {
	logger.Info("[CalcAllPDBParams] 开始批量计算PDB参数，批处理大小: %d", batchSize)
	
	var totalProcessed int64 = 0
	var totalCalculated int64 = 0
	var totalSkipped int64 = 0
	var totalErrors int64 = 0
	
	// 获取所有需要计算的记录
	var allRecords []models.PDBParameter
	if err := database.Database.Where("is_protein = ?", true).Find(&allRecords).Error; err != nil {
		return CalcAllPDBParamsResult{Error: "Database query error: " + err.Error()}
	}
	
	totalRecords := len(allRecords)
	logger.Info("[CalcAllPDBParams] 找到 %d 条蛋白质记录需要处理", totalRecords)
	
	// 分批处理
	for i := 0; i < totalRecords; i += batchSize {
		end := i + batchSize
		if end > totalRecords {
			end = totalRecords
		}
		
		batch := allRecords[i:end]
		logger.Info("[CalcAllPDBParams] 处理批次 %d-%d (共 %d 条记录)", i+1, end, len(batch))
		
		// 处理当前批次
		for _, record := range batch {
			totalProcessed++
			
			// 检查PDB文件是否存在
			pdbPath := fmt.Sprintf("../PROFASA-PDB-GO/data/%s.pdb", record.PdbId)
			if _, err := os.Stat(pdbPath); os.IsNotExist(err) {
				logger.Warn("[CalcAllPDBParams] PDB文件不存在，跳过: %s", pdbPath)
				totalSkipped++
				continue
			}
			
			// 检查序列是否为空
			if record.Fasta == "" {
				logger.Warn("[CalcAllPDBParams] 序列为空，跳过: %s", record.PdbId)
				totalSkipped++
				continue
			}
			
			// 计算6个参数
			rc, sa, ii, mw, h, ip := CalcAllWithPath(record.Fasta, record.PdbId)
			
			// 更新记录
			record.RcScore = fmt.Sprintf("%f", rc)
			record.SolventAccesibility = fmt.Sprintf("%f", sa)
			record.Instability = fmt.Sprintf("%f", ii)
			record.Size = mw
			record.Hydrophobicity = fmt.Sprintf("%f", h)
			record.IsoelectricPoint = fmt.Sprintf("%f", ip)
			
			// 批量更新数据库
			if err := database.Database.Model(&models.PDBParameter{}).Where("id = ?", record.ID).Updates(map[string]interface{}{
				"rc_score":             record.RcScore,
				"solvent_accesibility": record.SolventAccesibility,
				"instability":          record.Instability,
				"size":                 record.Size,
				"hydrophobicity":       record.Hydrophobicity,
				"isoelectric_point":    record.IsoelectricPoint,
			}).Error; err != nil {
				logger.Error("[CalcAllPDBParams] 更新记录失败: %s, 错误: %v", record.PdbId, err)
				totalErrors++
			} else {
				totalCalculated++
				logger.Info("[CalcAllPDBParams] 成功计算并更新: %s (rc: %.2f, sa: %.2f, ii: %.2f, mw: %.2f, h: %.4f, ip: %.2f)", 
					record.PdbId, rc, sa, ii, mw, h, ip)
			}
		}
		
		// 每批次处理完后输出进度
		progress := float64(totalProcessed) / float64(totalRecords) * 100
		logger.Info("[CalcAllPDBParams] 进度: %.2f%% (%d/%d)", progress, totalProcessed, totalRecords)
	}
	
	logger.Info("[CalcAllPDBParams] 批量计算完成 - 总计: %d, 成功: %d, 跳过: %d, 错误: %d", 
		totalProcessed, totalCalculated, totalSkipped, totalErrors)
	
	result := map[string]interface{}{
		"total_processed":  totalProcessed,
		"total_calculated": totalCalculated,
		"total_skipped":    totalSkipped,
		"total_errors":     totalErrors,
		"batch_size":       batchSize,
	}
	
	return CalcAllPDBParamsResult{Data: result}
}
