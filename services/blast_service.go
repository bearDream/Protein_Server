package services

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// BlastRequest 请求结构体
type BlastRequest struct {
	Code  string `json:"code" binding:"required"`
	Title string `json:"title" binding:"required"`
	Type  string `json:"type" binding:"required"` // "alpha", "itasser", "esm"
}

// BlastResponse 响应结构体
type BlastResponse struct {
	ID    uint   `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// FoldResponse fold响应结构体
type FoldResponse struct {
	ID    uint   `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func BlastTypeStringToInt(typeStr string) int64 {
	switch typeStr {
	case "alpha":
		return 1
	case "itasser":
		return 2
	case "esm":
		return 3
	default:
		return 0 // 非法类型
	}
}

// Blast 处理 blast 请求的主要函数
func Blast(code, title string, typeStr string, userId int64) BlastResponse {
	typeValue := BlastTypeStringToInt(typeStr)
	if typeValue == 0 {
		return BlastResponse{Error: "Invalid type."}
	}

	// 检查是否已存在相同序列的任务（不考虑标题和用户，只检查序列）
	var existingTask models.Task
	if err := database.Database.Where("sequence = ?", code).First(&existingTask).Error; err == nil {
		// 如果找到相同序列的任务，直接返回该任务的ID
		return BlastResponse{ID: existingTask.ID}
	}

	// 检查项目是否已存在（标题和用户都相同）
	var existingTasks []models.Task
	if err := database.Database.Where("sequence = ? AND title = ? AND user_id = ?", code, title, userId).Find(&existingTasks).Error; err != nil {
		return BlastResponse{Error: "数据库查询错误"}
	}
	if len(existingTasks) != 0 {
		return BlastResponse{Error: "项目已存在"}
	}

	// 调用 BlastProcessing 处理序列
	subSequences, blastInformations := BlastProcessing(code)
	if subSequences == nil {
		return BlastResponse{Error: "BLAST 处理失败"}
	}

	// 查找主序列在 protein_information 表中是否存在
	var mainProteinInfo models.ProteinInformation
	if err := database.Database.Where("sequence = ?", code).Find(&mainProteinInfo).Error; err != nil {
		return BlastResponse{Error: "查询蛋白质信息失败"}
	}

	// 查找主序列在所有队列中是否存在
	var mainQueueCount int64
	if err := database.Database.Model(&models.AlphaFoldQueue{}).Where("sequence = ?", code).Count(&mainQueueCount).Error; err != nil {
		return BlastResponse{Error: "查询 AlphaFold 队列失败"}
	}
	if mainQueueCount == 0 {
		if err := database.Database.Model(&models.ITasserQueue{}).Where("sequence = ?", code).Count(&mainQueueCount).Error; err != nil {
			return BlastResponse{Error: "查询 I-Tasser 队列失败"}
		}
	}
	if mainQueueCount == 0 {
		if err := database.Database.Model(&models.ESMQueue{}).Where("sequence = ?", code).Count(&mainQueueCount).Error; err != nil {
			return BlastResponse{Error: "查询 ESM 队列失败"}
		}
	}

	// 如果主序列既不在 protein_information 中也不在队列中，则添加到队列或同步处理
	if mainProteinInfo.ID == 0 && mainQueueCount == 0 {
		// 只有新序列才需要创建记录和添加到队列
		if typeValue == 3 { // ESM - 同步处理
			// 对ESM，先创建蛋白质信息记录
			ProteinInformation(code, "", typeValue)
			// 重新查询获取创建后的ID
			if err := database.Database.Where("sequence = ?", code).Find(&mainProteinInfo).Error; err != nil {
				return BlastResponse{Error: "查询主序列蛋白质信息失败"}
			}
		} else {
			// 对AlphaFold和I-Tasser，先创建蛋白质信息记录，再添加到队列
			ProteinInformation(code, "", typeValue)
			// 重新查询获取创建后的ID
			if err := database.Database.Where("sequence = ?", code).Find(&mainProteinInfo).Error; err != nil {
				return BlastResponse{Error: "查询主序列蛋白质信息失败"}
			}
		}
	} else if mainProteinInfo.ID == 0 {
		// 如果蛋白质信息不存在但在队列中，仍需创建蛋白质信息记录
		ProteinInformation(code, "", typeValue)
		// 重新查询获取创建后的ID
		if err := database.Database.Where("sequence = ?", code).Find(&mainProteinInfo).Error; err != nil {
			return BlastResponse{Error: "查询主序列蛋白质信息失败"}
		}
	}

	// 确保 mainProteinInfo.ID 不为0后创建主任务
	if mainProteinInfo.ID == 0 {
		return BlastResponse{Error: "无法获取主序列蛋白质信息ID"}
	}

	// 创建主任务（先不设置 ModelId，后续收集完所有 ID 后再更新）
	mainTask := models.Task{
		Title:                   title,
		Sequence:                code,
		Type:                    1,          // Sequence Search
		StructurePredictionTool: &typeValue, // 设置结构预测工具类型
		UserId:                  userId,
		ModelId:                 "", // 暂时为空，后续更新
	}

	if err := database.Database.Create(&mainTask).Error; err != nil {
		return BlastResponse{Error: "创建主任务失败"}
	}

	// 构建子序列字符串（用于保存到 task.SubSequence 字段）
	var processedSubSequences []string

	// 收集所有相关序列的 protein_information ID（包括主序列）
	var allProteinIds []string
	proteinIdSet := make(map[string]bool) // 使用map来去重

	// 添加主序列ID
	mainIdStr := strconv.FormatUint(uint64(mainProteinInfo.ID), 10)
	allProteinIds = append(allProteinIds, mainIdStr)
	proteinIdSet[mainIdStr] = true

	// 处理所有子序列，只创建蛋白质信息和队列项
	for i := range subSequences {
		// 获取子序列的描述信息
		description := getDescription(blastInformations[i])
		fasta := code[description.From-1 : description.To]

		// 将子序列添加到列表中
		processedSubSequences = append(processedSubSequences, fasta)

		// 查找子序列在 protein_information 表中是否存在
		var subProteinInfo models.ProteinInformation
		if err := database.Database.Where("sequence = ?", fasta).Find(&subProteinInfo).Error; err != nil {
			continue
		}

		// 查找子序列在所有队列中是否存在
		var subQueueCount int64
		if err := database.Database.Model(&models.AlphaFoldQueue{}).Where("sequence = ?", fasta).Count(&subQueueCount).Error; err != nil {
			continue
		}
		if subQueueCount == 0 {
			if err := database.Database.Model(&models.ITasserQueue{}).Where("sequence = ?", fasta).Count(&subQueueCount).Error; err != nil {
				continue
			}
		}
		if subQueueCount == 0 {
			if err := database.Database.Model(&models.ESMQueue{}).Where("sequence = ?", fasta).Count(&subQueueCount).Error; err != nil {
				continue
			}
		}

		// 将描述信息转换为 JSON 字符串
		informationJSON := ""
		if infoJSON, err := json.Marshal(blastInformations[i]); err == nil {
			informationJSON = string(infoJSON)
		}

		// 如果子序列既不在 protein_information 中也不在队列中，则处理
		if subProteinInfo.ID == 0 && subQueueCount == 0 {
			// 只有新序列才需要创建记录和添加到队列
			if typeValue == 3 { // ESM - 同步处理
				// 对ESM，直接创建蛋白质信息记录并同步处理
				ProteinInformationWithParent(fasta, informationJSON, typeValue, mainProteinInfo.ID)
			} else {
				// 对AlphaFold和I-Tasser，先创建蛋白质信息记录
				ProteinInformationWithParent(fasta, informationJSON, typeValue, mainProteinInfo.ID)
			}
			// 重新查询获取创建后的记录ID
			if err := database.Database.Where("sequence = ?", fasta).Find(&subProteinInfo).Error; err == nil && subProteinInfo.ID > 0 {
				idStr := strconv.FormatUint(uint64(subProteinInfo.ID), 10)
				// 使用map去重，只有不存在的ID才添加
				if !proteinIdSet[idStr] {
					allProteinIds = append(allProteinIds, idStr)
					proteinIdSet[idStr] = true
				}
			}
		} else if subProteinInfo.ID == 0 {
			// 如果蛋白质信息不存在但在队列中，仍需创建蛋白质信息记录
			ProteinInformationWithParent(fasta, informationJSON, typeValue, mainProteinInfo.ID)
			// 重新查询获取创建后的记录ID
			if err := database.Database.Where("sequence = ?", fasta).Find(&subProteinInfo).Error; err == nil && subProteinInfo.ID > 0 {
				idStr := strconv.FormatUint(uint64(subProteinInfo.ID), 10)
				// 使用map去重，只有不存在的ID才添加
				if !proteinIdSet[idStr] {
					allProteinIds = append(allProteinIds, idStr)
					proteinIdSet[idStr] = true
				}
			}
		} else {
			// 如果子序列已存在，直接添加其ID，不重复处理
			idStr := strconv.FormatUint(uint64(subProteinInfo.ID), 10)
			// 使用map去重，只有不存在的ID才添加
			if !proteinIdSet[idStr] {
				allProteinIds = append(allProteinIds, idStr)
				proteinIdSet[idStr] = true
			}
		}
	}

	// 更新主任务的子序列字段
	if len(processedSubSequences) > 0 {
		subSequenceStr := strings.Join(processedSubSequences, "|")
		if err := database.Database.Model(&mainTask).Update("sub_sequence", subSequenceStr).Error; err != nil {
			logger.Error("更新任务子序列字段失败: %v", err)
		}
	}

	// 更新主任务的 ModelId 字段
	// 只有 ESM 任务（typeValue == 3）才立即设置完整的 ModelId
	if typeValue == 3 && len(allProteinIds) > 0 {
		modelIdStr := strings.Join(allProteinIds, ",")
		if err := database.Database.Model(&mainTask).Update("model_id", modelIdStr).Error; err != nil {
			logger.Error("更新任务ModelId字段失败: %v", err)
		}
		logger.Info("ESM Task %d ModelId 设置为: %s", mainTask.ID, modelIdStr)
	} else if len(allProteinIds) > 0 {
		// 对于 Alpha 和 I-Tasser 任务，只设置主序列的 ModelId
		// 子序列的 ModelId 将在队列处理完成后更新
		modelIdStr := allProteinIds[0] // 只使用主序列的 ID
		if err := database.Database.Model(&mainTask).Update("model_id", modelIdStr).Error; err != nil {
			logger.Error("更新任务ModelId字段失败: %v", err)
		}
		logger.Info("Alpha/I-Tasser Task %d ModelId 初始设置为: %s（仅主序列）", mainTask.ID, modelIdStr)
	}

	return BlastResponse{ID: mainTask.ID}
}

// addSequenceToQueue 根据类型将序列添加到相应的队列（无父ID版本）
func addSequenceToQueue(sequence string, typeValue int64) {
	addSequenceToQueueWithParent(sequence, typeValue, nil)
}

// addSequenceToQueueWithParent 根据类型将序列添加到相应的队列，支持父ID
func addSequenceToQueueWithParent(sequence string, typeValue int64, parentId *int64) {
	switch typeValue {
	case 1:
		AddToAlphaFoldQueueWithParent(sequence, parentId)
	case 2:
		AddToITasserQueueWithParent(sequence, parentId)
	case 3: // ESMFold
		// 直接同步生成模型和图片
		SyncESMFoldAndImage(sequence, parentId)
	}
}

// SyncESMFoldAndImage 同步生成ESMFold模型和Ramachandran图片
func SyncESMFoldAndImage(sequence string, parentId *int64) {
	// 1. 写入protein_information表（如果不存在）
	var proteinInfo models.ProteinInformation
	database.Database.Where("sequence = ?", sequence).Find(&proteinInfo)
	if proteinInfo.ID == 0 {
		proteinInfo = models.ProteinInformation{
			Sequence: sequence,
		}
		if parentId != nil {
			proteinInfo.ParentId = uint(*parentId)
		}
		database.Database.Create(&proteinInfo)
	}

	// 2. 调用ESMFold生成结构和模型文件
	ESMFold(sequence)

	// 3. 生成Ramachandran图片
	Ramachandran(fmt.Sprintf("%d", proteinInfo.ID))
}

// BlastDescription BLAST 描述信息结构体
type BlastDescription struct {
	From      int    `json:"from"`
	To        int    `json:"to"`
	Accession string `json:"accession"`
	Title     string `json:"title"`
	Comment   string `json:"comment"`
}

// getDescription 从 BLAST 信息中提取描述，并读取 ACD 文件获取详细信息
func getDescription(blastInfo map[string]string) BlastDescription {
	// 尝试多种可能的字段名变体
	var fromStr, toStr, accession string

	// 查找 from 字段
	if val, exists := blastInfo["from"]; exists {
		fromStr = val
	} else if val, exists := blastInfo["From"]; exists {
		fromStr = val
	} else if val, exists := blastInfo["FROM"]; exists {
		fromStr = val
	}

	// 查找 to 字段
	if val, exists := blastInfo["to"]; exists {
		toStr = val
	} else if val, exists := blastInfo["To"]; exists {
		toStr = val
	} else if val, exists := blastInfo["TO"]; exists {
		toStr = val
	}

	// 查找 accession 字段
	if val, exists := blastInfo["accession"]; exists {
		accession = val
	} else if val, exists := blastInfo["Accession"]; exists {
		accession = val
	} else if val, exists := blastInfo["ACCESSION"]; exists {
		accession = val
	}

	from, _ := strconv.Atoi(fromStr)
	to, _ := strconv.Atoi(toStr)

	// 初始化结果
	result := BlastDescription{
		From:      from,
		To:        to,
		Accession: accession,
		Title:     "",
		Comment:   "",
	}

	// 如果没有 accession，直接返回基本信息
	if accession == "" {
		return result
	}

	// 尝试读取 ACD 文件获取详细信息
	acdPath := fmt.Sprintf("../RpsbProc-x64-linux/acd/%s.acd", accession)
	title, comment := parseAcdFile(acdPath)

	result.Title = title
	result.Comment = comment

	return result
}

// parseAcdFile 解析 ACD 文件，提取 title 和 comment 信息
func parseAcdFile(filePath string) (string, string) {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("无法打开ACD文件: %s, 错误: %v", filePath, err)
		return "", ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var title, comment strings.Builder
	var readTitle, readComment bool
	var firstTitle bool = true

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// 检查是否开始读取 title
		if strings.HasPrefix(trimmedLine, "title \"") {
			readTitle = true
		}

		// 读取 title 内容
		if readTitle && firstTitle {
			title.WriteString(trimmedLine)
		}

		// 检查是否开始读取 comment
		if strings.HasPrefix(trimmedLine, "comment \"") {
			readComment = true
		}

		// 读取 comment 内容
		if readComment {
			comment.WriteString(trimmedLine)
		}

		// 检查是否结束当前字段的读取
		if strings.HasSuffix(trimmedLine, "\",") || strings.HasSuffix(trimmedLine, "\"") {
			readComment = false
			if readTitle {
				firstTitle = false
				readTitle = false
			}
		}

		// 当遇到 seqannot 时，处理结果并返回
		if strings.HasPrefix(trimmedLine, "seqannot") {
			// 提取双引号内的内容
			titleStr := title.String()
			commentStr := comment.String()

			// 从 title 中提取双引号内的内容
			if titleStr != "" {
				if parts := strings.Split(titleStr, "\""); len(parts) > 1 {
					titleStr = parts[1]
				}
			}

			// 从 comment 中提取双引号内的内容
			if commentStr != "" {
				if parts := strings.Split(commentStr, "\""); len(parts) > 1 {
					commentStr = parts[1]
				}
			}

			return titleStr, commentStr
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("读取ACD文件时出错: %v", err)
	}

	// 如果没有遇到 seqannot，也要处理已读取的内容
	titleStr := title.String()
	commentStr := comment.String()

	// 从 title 中提取双引号内的内容
	if titleStr != "" {
		if parts := strings.Split(titleStr, "\""); len(parts) > 1 {
			titleStr = parts[1]
		}
	}

	// 从 comment 中提取双引号内的内容
	if commentStr != "" {
		if parts := strings.Split(commentStr, "\""); len(parts) > 1 {
			commentStr = parts[1]
		}
	}

	return titleStr, commentStr
}

type BlastListItem struct {
	Category      int64       `json:"category"`
	CreatedAt     int64       `json:"createdAt"`
	Fasta         string      `json:"fasta"`
	HasModelCount int64       `json:"hasModelCount"`
	ID            uint        `json:"id"`
	Information   interface{} `json:"information"`
	ModelId       string      `json:"modelId"`
	ModelTotal    int64       `json:"modelTotal"`
	ParentId      *uint       `json:"parentId"`
	SubSequence   *string     `json:"subSequence"`
	TaskType      string      `json:"taskType"`
	Title         string      `json:"title"`
	ToolType      string      `json:"toolType"`
	Type          int64       `json:"type"`
	UpdatedAt     int64       `json:"updatedAt"`
	UserId        int64       `json:"userId"`
}

type BlastListResult struct {
	List  []BlastListItem `json:"list"`
	Total int64           `json:"total"`
}

// GetBlastList 查询任务的分页列表
func GetBlastList(userId int64, current, pageSize int, title, category, createStart, createEnd string) (BlastListResult, error) {
	var tasks []models.Task
	var total int64

	db := database.Database.Model(&models.Task{}).
		Where("user_id = ?", userId) // 查询用户的所有任务

	if title != "" {
		db = db.Where("title LIKE ?", "%"+title+"%")
	}

	if category != "" {
		categoryInt, _ := strconv.ParseInt(category, 10, 64)
		db = db.Where("structure_prediction_tool = ?", categoryInt)
	}
	if createStart != "" && createEnd != "" {
		start, _ := strconv.ParseInt(createStart, 10, 64)
		end, _ := strconv.ParseInt(createEnd, 10, 64)
		db = db.Where("created_at BETWEEN ? AND ?", time.UnixMilli(start), time.UnixMilli(end))
	} else if createStart != "" {
		start, _ := strconv.ParseInt(createStart, 10, 64)
		db = db.Where("created_at > ?", time.UnixMilli(start))
	} else if createEnd != "" {
		end, _ := strconv.ParseInt(createEnd, 10, 64)
		db = db.Where("created_at < ?", time.UnixMilli(end))
	}

	// 统计总数
	db.Count(&total)

	// 分页
	db = db.Order("created_at DESC").Offset((current - 1) * pageSize).Limit(pageSize)
	if err := db.Find(&tasks).Error; err != nil {
		return BlastListResult{}, err
	}

	// 组装返回
	list := make([]BlastListItem, 0, len(tasks))
	for _, task := range tasks {
		// 统计模型总数和有模型的数量（基于 ModelId 字段）
		var modelTotal int64 = 0
		var hasModelCount int64 = 0

		if task.ModelId != "" && task.ModelId != "0" {
			// ModelId 是逗号分隔的 ID 字符串，计算数量
			modelIds := strings.Split(task.ModelId, ",")
			modelTotal = int64(len(modelIds))
			hasModelCount = modelTotal // 有 ModelId 就认为有模型
		}

		// 确定任务类型字符串
		var taskType string
		switch task.Type {
		case 1:
			taskType = "blast"
		case 2:
			taskType = "fold"
		case 3:
			taskType = "analysis"
		case 4:
			taskType = "superimpose"
		default:
			taskType = "unknown"
		}

		// 处理 SubSequence 字段
		var subSequence *string = nil
		if task.SubSequence != "" {
			subSequence = &task.SubSequence
		}

		// 组装
		list = append(list, BlastListItem{
			Category:      task.Type,
			CreatedAt:     task.CreatedAt.UnixMilli(),
			Fasta:         task.Sequence,
			HasModelCount: hasModelCount,
			ID:            task.ID,
			Information:   nil,
			ModelId:       task.ModelId,
			ModelTotal:    modelTotal,
			ParentId:      nil, // 主任务无parent
			SubSequence:   subSequence,
			TaskType:      taskType,
			Title:         task.Title,
			ToolType:      "",
			Type:          task.Type,
			UpdatedAt:     task.UpdatedAt.UnixMilli(),
			UserId:        task.UserId,
		})
	}

	return BlastListResult{
		List:  list,
		Total: total,
	}, nil
}

// BlastResultItem 表示 blast result 的单个项目
type BlastResultItem struct {
	ID                    uint    `json:"id"`
	AccessibilityFraction float64 `json:"accessibilityFraction"`
	Category              int     `json:"category"`
	Fasta                 string  `json:"fasta"`
	Hydrophobicity        float64 `json:"hydrophobicity"`
	Information           *string `json:"information"`
	Instability           float64 `json:"instability"`
	IsoelectricPoint      float64 `json:"isoelectricPoint"`
	ModelId               int     `json:"modelId"`
	ParentId              *uint   `json:"parentId"`
	RcScore               float64 `json:"rcScore"`
	Size                  float64 `json:"size"`
	SolventAccesibility   float64 `json:"solventAccesibility"`
	Title                 string  `json:"title"`
	TotalNum              int     `json:"totalNum"`
	Type                  int     `json:"type"`
	UserId                int64   `json:"userId"`
}

// RCSBQuery RCSB PDB 查询结构体
type RCSBQuery struct {
	Query          RCSBQueryDetail    `json:"query"`
	ReturnType     string             `json:"return_type"`
	RequestOptions RCSBRequestOptions `json:"request_options"`
}

type RCSBQueryDetail struct {
	Type       string              `json:"type"`
	Service    string              `json:"service"`
	Parameters RCSBQueryParameters `json:"parameters"`
}

type RCSBQueryParameters struct {
	EvalueCutoff   float64 `json:"evalue_cutoff"`
	IdentityCutoff int     `json:"identity_cutoff"`
	Target         string  `json:"target"`
	Value          string  `json:"value"`
}

type RCSBRequestOptions struct {
	ReturnCounts bool `json:"return_counts"`
}

type RCSBResponse struct {
	TotalCount int `json:"total_count"`
}

// GetBlastResult 获取 blast 结果详情
func GetBlastResult(idStr string) ([]BlastResultItem, error) {
	// 将 string 类型的 ID 转换为 uint
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("无效的ID格式: %v", err)
	}

	// 首先查询主任务信息
	var mainTask models.Task
	if err := database.Database.Where("id = ?", uint(id)).First(&mainTask).Error; err != nil {
		return nil, fmt.Errorf("查询任务信息失败: %v", err)
	}

	var proteinInfos []models.ProteinInformation

	// 修复：正确解析Task.ModelId字段来查询相关的蛋白质信息
	if mainTask.ModelId == "" {
		return []BlastResultItem{}, nil // 如果没有ModelId，返回空结果
	}

	// 解析ModelId字段（逗号分隔的ID列表）
	modelIdStrs := strings.Split(mainTask.ModelId, ",")
	var proteinIds []uint
	for _, idStr := range modelIdStrs {
		if trimmedId := strings.TrimSpace(idStr); trimmedId != "" && trimmedId != "0" {
			if proteinId, err := strconv.ParseUint(trimmedId, 10, 32); err == nil {
				proteinIds = append(proteinIds, uint(proteinId))
			}
		}
	}

	if len(proteinIds) == 0 {
		return []BlastResultItem{}, nil // 如果没有有效的蛋白质ID，返回空结果
	}

	// 根据解析出的蛋白质ID查询相关记录
	if err := database.Database.Where("id IN ?", proteinIds).Find(&proteinInfos).Error; err != nil {
		return nil, fmt.Errorf("查询蛋白质信息失败: %v", err)
	}

	var result []BlastResultItem

	// 特殊处理fold类型任务
	if mainTask.Type == 2 { // Structure Prediction (fold任务)
		// 对于fold任务，需要返回：
		// 1. 输入序列（子序列，type=2）
		// 2. fold结果序列（主序列，type=1）

		// 首先添加输入序列（子序列）
		subSequenceIndex := 1
		for _, proteinInfo := range proteinInfos {
			if proteinInfo.ParentId != 0 {
				// 为子序列生成标题
				subTitle := fmt.Sprintf("%s (输入序列 %d)", mainTask.Title, subSequenceIndex)

				// 创建子序列结果项
				item := createBlastResultItem(&proteinInfo, mainTask, 2, &subTitle)
				result = append(result, item)
				subSequenceIndex++
			}
		}

		// 然后添加fold结果序列（主序列）
		var mainProteinInfo *models.ProteinInformation
		for i := range proteinInfos {
			if proteinInfos[i].ParentId == 0 {
				mainProteinInfo = &proteinInfos[i]
				break
			}
		}

		if mainProteinInfo != nil {
			// 添加fold结果序列
			item := createBlastResultItem(mainProteinInfo, mainTask, 1, nil)
			result = append(result, item)
		}
	} else {
		// 对于非fold任务，使用原有逻辑
		for _, proteinInfo := range proteinInfos {
			item := createBlastResultItem(&proteinInfo, mainTask, 0, nil)
			result = append(result, item)
		}
	}

	return result, nil
}

// createBlastResultItem 创建BlastResultItem的辅助函数
func createBlastResultItem(proteinInfo *models.ProteinInformation, mainTask models.Task, forceType int, customTitle *string) BlastResultItem {
	// 使用数据库中保存的结构数量，避免重复API调用
	totalNum := proteinInfo.StructureNum

	// 如果数据库中没有结构数量信息，异步获取并保存
	if totalNum == 0 {
		go SaveStructureNum(proteinInfo.ID) // 异步执行，不阻塞当前请求
	}

	// 转换字符串字段为数值类型
	hydrophobicity, _ := strconv.ParseFloat(proteinInfo.Hydrophobicity, 64)
	instability, _ := strconv.ParseFloat(proteinInfo.Instability, 64)
	isoelectricPoint, _ := strconv.ParseFloat(proteinInfo.IsoelectricPoint, 64)
	rcScore, _ := strconv.ParseFloat(proteinInfo.RcScore, 64)
	size, _ := strconv.ParseFloat(proteinInfo.Size, 64)
	solventAccesibility, _ := strconv.ParseFloat(proteinInfo.SolventAccesibility, 64)

	// 计算 AccessibilityFraction (solventAccesibility / 100)
	accessibilityFraction := solventAccesibility / 100.0

	// 确定类型
	var itemType int
	if forceType > 0 {
		itemType = forceType
	} else {
		// 如果有 ParentId 则为子序列 (type=2)，否则为主序列 (type=1)
		itemType = 1
		if proteinInfo.ParentId != 0 {
			itemType = 2
		}
	}

	// 处理 Information 字段：如果是子序列且有 BlastInformation，则使用它；否则为 null
	var information *string
	if itemType == 2 && proteinInfo.BlastInformation != "" {
		information = &proteinInfo.BlastInformation
	}

	// 处理 ParentId：如果为 0 则设为 nil
	var parentId *uint
	if proteinInfo.ParentId != 0 {
		parentId = &proteinInfo.ParentId
	}

	// 解析 ModelId 从字符串转为整数
	modelId, _ := strconv.Atoi(proteinInfo.PdbId)

	// 确定标题
	title := mainTask.Title
	if customTitle != nil {
		title = *customTitle
	}

	// 组装结果项
	return BlastResultItem{
		ID:                    proteinInfo.ID,
		AccessibilityFraction: accessibilityFraction,
		Category:              int(mainTask.Type),
		Fasta:                 proteinInfo.Sequence,
		Hydrophobicity:        hydrophobicity,
		Information:           information,
		Instability:           instability,
		IsoelectricPoint:      isoelectricPoint,
		ModelId:               modelId,
		ParentId:              parentId,
		RcScore:               rcScore,
		Size:                  size,
		SolventAccesibility:   solventAccesibility,
		Title:                 title,
		TotalNum:              totalNum,
		Type:                  itemType,
		UserId:                mainTask.UserId,
	}
}

// saveStructureNum 查询RCSB PDB数据库获取结构数量并保存到数据库
func SaveStructureNum(proteinInfoId uint) {
	// 查找蛋白质信息记录
	var proteinInfo models.ProteinInformation
	if err := database.Database.Where("id = ?", proteinInfoId).First(&proteinInfo).Error; err != nil {
		logger.Error("查找蛋白质信息失败: %v", err)
		return
	}

	// 如果已经有结构数量信息，跳过
	if proteinInfo.StructureNum > 0 {
		logger.Info("蛋白质信息 %d 已有结构数量信息，跳过查询", proteinInfoId)
		return
	}

	// 查询RCSB PDB数据库
	structureNum := getStructureNum(proteinInfo.Sequence)

	// 保存到数据库
	if err := database.Database.Model(&models.ProteinInformation{}).Where("id = ?", proteinInfoId).Update("structure_num", structureNum).Error; err != nil {
		logger.Error("保存结构数量失败: %v", err)
	} else {
		logger.Info("已保存蛋白质信息 %d 的结构数量: %d", proteinInfoId, structureNum)
	}
}

// BatchUpdateStructureNum 批量更新所有缺少结构数量信息的蛋白质记录
func BatchUpdateStructureNum() {
	var proteinInfos []models.ProteinInformation

	// 查找所有没有结构数量信息的记录
	if err := database.Database.Where("structure_num = 0 OR structure_num IS NULL").Find(&proteinInfos).Error; err != nil {
		logger.Error("查询缺少结构数量的蛋白质信息失败: %v", err)
		return
	}

	if len(proteinInfos) == 0 {
		logger.Info("所有蛋白质信息记录都已有结构数量信息")
		return
	}

	logger.Info("开始批量更新结构数量，找到 %d 条需要更新的记录", len(proteinInfos))

	// 批量处理，避免同时发起太多API请求
	const batchSize = 5 // 减小批次大小，避免API限制
	processedCount := 0

	for i := 0; i < len(proteinInfos); i += batchSize {
		end := i + batchSize
		if end > len(proteinInfos) {
			end = len(proteinInfos)
		}

		logger.Info("处理第 %d-%d 条记录", i+1, end)

		// 并发处理一批记录
		for j := i; j < end; j++ {
			go func(id uint) {
				SaveStructureNum(id)
				processedCount++
			}(proteinInfos[j].ID)
		}

		// 每批之间稍作延迟，避免API限制
		if end < len(proteinInfos) {
			logger.Info("等待 10 秒后处理下一批...")
			time.Sleep(10 * time.Second)
		}
	}

	logger.Info("批量更新结构数量任务已启动，共处理 %d 条记录", len(proteinInfos))
}

// getStructureNum 查询 RCSB PDB 数据库获取结构数量（内部函数）
func getStructureNum(sequence string) int {
	// 构建查询 JSON
	query := RCSBQuery{
		Query: RCSBQueryDetail{
			Type:    "terminal",
			Service: "sequence",
			Parameters: RCSBQueryParameters{
				EvalueCutoff:   0.1,
				IdentityCutoff: 0,
				Target:         "pdb_protein_sequence",
				Value:          sequence,
			},
		},
		ReturnType: "entry",
		RequestOptions: RCSBRequestOptions{
			ReturnCounts: true,
		},
	}

	// 将查询转换为 JSON 字符串
	queryJSON, err := json.Marshal(query)
	if err != nil {
		logger.Error("序列化 RCSB 查询失败: %v", err)
		return 0
	}

	// 构建 URL
	baseURL := "https://search.rcsb.org/rcsbsearch/v2/query"
	queryParam := url.QueryEscape(string(queryJSON))
	fullURL := fmt.Sprintf("%s?json=%s", baseURL, queryParam)

	// 发起 GET 请求
	resp, err := http.Get(fullURL)
	if err != nil {
		logger.Error("请求 RCSB PDB 失败: %v", err)
		return 0
	}
	defer resp.Body.Close()

	// 解析响应
	var rcsResponse RCSBResponse
	if err := json.NewDecoder(resp.Body).Decode(&rcsResponse); err != nil {
		logger.Error("解析 RCSB 响应失败: %v", err)
		return 0
	}

	return rcsResponse.TotalCount
}

// UpdateTaskModelIdAfterAsyncCompletion 异步任务完成后更新主任务的ModelId
func UpdateTaskModelIdAfterAsyncCompletion(proteinInfoId uint) {
	// 查找该蛋白质信息记录
	var proteinInfo models.ProteinInformation
	if err := database.Database.Where("id = ?", proteinInfoId).First(&proteinInfo).Error; err != nil {
		logger.Error("查找蛋白质信息失败: %v", err)
		return
	}

	// 确定主序列ID（如果有ParentId则使用ParentId，否则使用自身ID）
	var mainProteinId uint = proteinInfoId
	if proteinInfo.ParentId != 0 {
		mainProteinId = proteinInfo.ParentId
	}

	// 查找包含该主序列ID的任务
	var tasks []models.Task
	mainProteinIdStr := strconv.FormatUint(uint64(mainProteinId), 10)

	// 使用LIKE查询找到ModelId中包含该主序列ID的任务
	if err := database.Database.Where("model_id LIKE ? OR model_id LIKE ? OR model_id LIKE ? OR model_id = ?",
		mainProteinIdStr+",%", "%,"+mainProteinIdStr+",%", "%,"+mainProteinIdStr, mainProteinIdStr).Find(&tasks).Error; err != nil {
		logger.Error("查找相关任务失败: %v", err)
		return
	}

	for _, task := range tasks {
		// 重新收集该任务相关的所有蛋白质信息ID
		var allProteinInfos []models.ProteinInformation
		if err := database.Database.Where("id = ? OR parent_id = ?", mainProteinId, mainProteinId).Find(&allProteinInfos).Error; err != nil {
			logger.Error("查找任务相关蛋白质信息失败: %v", err)
			continue
		}

		// 构建新的ModelId字符串
		var newModelIds []string
		for _, info := range allProteinInfos {
			newModelIds = append(newModelIds, strconv.FormatUint(uint64(info.ID), 10))
		}

		if len(newModelIds) > 0 {
			newModelIdStr := strings.Join(newModelIds, ",")

			// 只有当ModelId发生变化时才更新
			if task.ModelId != newModelIdStr {
				if err := database.Database.Model(&models.Task{}).Where("id = ?", task.ID).Update("model_id", newModelIdStr).Error; err != nil {
					logger.Error("更新任务ModelId失败: %v", err)
				} else {
					logger.Info("异步任务完成，已更新Task %d 的ModelId: %s", task.ID, newModelIdStr)
				}
			}
		}
	}
}

// Fold 处理 fold 请求的主要函数
func Fold(codes []string, title string, typeStr string, userId int64) FoldResponse {
	logger.Info("Fold: 输入序列数量: %d", len(codes))

	typeValue := BlastTypeStringToInt(typeStr)
	if typeValue == 0 {
		logger.Error("Fold: 无效的类型: %s", typeStr)
		return FoldResponse{Error: "Invalid type."}
	}

	// 将多个子序列连接作为主序列
	mainSequence := strings.Join(codes, "")

	if mainSequence == "" {
		return FoldResponse{Error: "连接后的序列为空"}
	}

	// 将多个序列合并为一个主序列（用于数据库存储）
	codesString := strings.Join(codes, "|") // 子序列用"|"分割
	logger.Info("Fold: 子序列字符串: %s", codesString)

	// 查找主序列（连接后的序列）在 protein_information 表中是否存在
	var mainProteinInfo models.ProteinInformation
	if err := database.Database.Where("sequence = ?", mainSequence).Find(&mainProteinInfo).Error; err != nil {
		return FoldResponse{Error: "查询主序列蛋白质信息失败"}
	}

	// 查找主序列在所有队列中是否存在
	var mainQueueCount int64
	if err := database.Database.Model(&models.AlphaFoldQueue{}).Where("sequence = ?", mainSequence).Count(&mainQueueCount).Error; err != nil {
		return FoldResponse{Error: "查询 AlphaFold 队列失败"}
	}
	if mainQueueCount == 0 {
		if err := database.Database.Model(&models.ITasserQueue{}).Where("sequence = ?", mainSequence).Count(&mainQueueCount).Error; err != nil {
			return FoldResponse{Error: "查询 I-Tasser 队列失败"}
		}
	}
	if mainQueueCount == 0 {
		if err := database.Database.Model(&models.ESMQueue{}).Where("sequence = ?", mainSequence).Count(&mainQueueCount).Error; err != nil {
			return FoldResponse{Error: "查询 ESM 队列失败"}
		}
	}

	// 如果主序列既不在 protein_information 中也不在队列中，则添加到队列或同步处理
	if mainProteinInfo.ID == 0 && mainQueueCount == 0 {
		// 只有新序列才需要创建记录和添加到队列
		if typeValue == 3 { // ESM - 同步处理
			// 对ESM，先创建蛋白质信息记录
			ProteinInformation(mainSequence, "", typeValue)
			// 重新查询获取创建后的ID
			if err := database.Database.Where("sequence = ?", mainSequence).Find(&mainProteinInfo).Error; err != nil {
				return FoldResponse{Error: "查询主序列蛋白质信息失败"}
			}
		} else {
			// 对AlphaFold和I-Tasser，先创建蛋白质信息记录，再添加到队列
			ProteinInformation(mainSequence, "", typeValue)
			// 重新查询获取创建后的ID
			if err := database.Database.Where("sequence = ?", mainSequence).Find(&mainProteinInfo).Error; err != nil {
				return FoldResponse{Error: "查询主序列蛋白质信息失败"}
			}
		}
	} else if mainProteinInfo.ID == 0 {
		// 如果蛋白质信息不存在但在队列中，仍需创建蛋白质信息记录
		ProteinInformation(mainSequence, "", typeValue)
		// 重新查询获取创建后的ID
		if err := database.Database.Where("sequence = ?", mainSequence).Find(&mainProteinInfo).Error; err != nil {
			return FoldResponse{Error: "查询主序列蛋白质信息失败"}
		}
	}

	// 确保 mainProteinInfo.ID 不为0后创建主任务
	if mainProteinInfo.ID == 0 {
		return FoldResponse{Error: "无法获取主序列蛋白质信息ID"}
	}

	// 创建主任务（先不设置 ModelId，后续收集完所有 ID 后再更新）
	mainTask := models.Task{
		Title:                   title,
		Sequence:                mainSequence, // 使用连接后的序列
		SubSequence:             codesString,  // 子序列用"|"分割
		Type:                    2,            // Structure Prediction
		StructurePredictionTool: &typeValue,
		UserId:                  userId,
		ModelId:                 "", // 暂时为空，后续更新
	}

	if err := database.Database.Create(&mainTask).Error; err != nil {
		return FoldResponse{Error: "创建主任务失败"}
	}

	// 收集所有相关序列的 protein_information ID（包括主序列）
	var allProteinIds []string
	proteinIdSet := make(map[string]bool) // 使用map来去重

	// 添加主序列ID
	mainIdStr := strconv.FormatUint(uint64(mainProteinInfo.ID), 10)
	allProteinIds = append(allProteinIds, mainIdStr)
	proteinIdSet[mainIdStr] = true

	// 处理所有子序列，按照 Blast 函数的模式
	for _, code := range codes {
		// 查找子序列在 protein_information 表中是否存在
		var subProteinInfo models.ProteinInformation
		if err := database.Database.Where("sequence = ?", code).Find(&subProteinInfo).Error; err != nil {
			continue
		}

		// 查找子序列在所有队列中是否存在
		var subQueueCount int64
		if err := database.Database.Model(&models.AlphaFoldQueue{}).Where("sequence = ?", code).Count(&subQueueCount).Error; err != nil {
			continue
		}
		if subQueueCount == 0 {
			if err := database.Database.Model(&models.ITasserQueue{}).Where("sequence = ?", code).Count(&subQueueCount).Error; err != nil {
				continue
			}
		}
		if subQueueCount == 0 {
			if err := database.Database.Model(&models.ESMQueue{}).Where("sequence = ?", code).Count(&subQueueCount).Error; err != nil {
				continue
			}
		}

		// 计算子序列在主序列中的位置
		startPos := findSubsequencePosition(mainSequence, code)
		information := struct {
			From int `json:"from"`
			To   int `json:"to"`
		}{
			From: 1,
			To:   1,
		}
		if startPos >= 0 {
			information.From = startPos + 1
			information.To = startPos + len(code)
		} else {
			// 如果找不到精确匹配，使用默认位置
			information.To = information.From + len(code) - 1
		}

		// 将位置信息转换为JSON字符串
		informationJSON := ""
		if infoJSON, err := json.Marshal(information); err == nil {
			informationJSON = string(infoJSON)
		}

		// 如果子序列既不在 protein_information 中也不在队列中，则处理
		if subProteinInfo.ID == 0 && subQueueCount == 0 {
			// 只有新序列才需要创建记录和添加到队列
			if typeValue == 3 { // ESM - 同步处理
				// 对ESM，直接创建蛋白质信息记录并同步处理
				ProteinInformationWithParent(code, informationJSON, typeValue, mainProteinInfo.ID)
			} else {
				// 对AlphaFold和I-Tasser，先创建蛋白质信息记录
				ProteinInformationWithParent(code, informationJSON, typeValue, mainProteinInfo.ID)
			}
			// 重新查询获取创建后的记录ID
			if err := database.Database.Where("sequence = ?", code).Find(&subProteinInfo).Error; err == nil && subProteinInfo.ID > 0 {
				idStr := strconv.FormatUint(uint64(subProteinInfo.ID), 10)
				// 使用map去重，只有不存在的ID才添加
				if !proteinIdSet[idStr] {
					allProteinIds = append(allProteinIds, idStr)
					proteinIdSet[idStr] = true
				}
			}
		} else if subProteinInfo.ID == 0 {
			// 如果蛋白质信息不存在但在队列中，仍需创建蛋白质信息记录
			ProteinInformationWithParent(code, informationJSON, typeValue, mainProteinInfo.ID)
			// 重新查询获取创建后的记录ID
			if err := database.Database.Where("sequence = ?", code).Find(&subProteinInfo).Error; err == nil && subProteinInfo.ID > 0 {
				idStr := strconv.FormatUint(uint64(subProteinInfo.ID), 10)
				// 使用map去重，只有不存在的ID才添加
				if !proteinIdSet[idStr] {
					allProteinIds = append(allProteinIds, idStr)
					proteinIdSet[idStr] = true
				}
			}
		} else {
			// 如果子序列已存在，更新其ParentId和BlastInformation
			if subProteinInfo.ParentId == 0 {
				if err := database.Database.Model(&subProteinInfo).Updates(map[string]interface{}{
					"parent_id":         mainProteinInfo.ID,
					"blast_information": informationJSON,
				}).Error; err != nil {
					logger.Error("更新子序列信息失败: %v", err)
					continue
				}
			} else {
				// 如果ParentId已经存在，只更新BlastInformation
				if err := database.Database.Model(&subProteinInfo).Update("blast_information", informationJSON).Error; err != nil {
					logger.Error("更新子序列BlastInformation失败: %v", err)
					continue
				}
			}
			// 直接添加其ID，不重复处理
			idStr := strconv.FormatUint(uint64(subProteinInfo.ID), 10)
			// 使用map去重，只有不存在的ID才添加
			if !proteinIdSet[idStr] {
				allProteinIds = append(allProteinIds, idStr)
				proteinIdSet[idStr] = true
			}
		}

	}

	// 更新主任务的 ModelId 字段
	// 只有 ESM 任务（typeValue == 3）才立即设置完整的 ModelId
	if typeValue == 3 && len(allProteinIds) > 0 {
		modelIdStr := strings.Join(allProteinIds, ",")
		if err := database.Database.Model(&mainTask).Update("model_id", modelIdStr).Error; err != nil {
			logger.Error("更新任务ModelId字段失败: %v", err)
		}
		logger.Info("ESM Fold Task %d ModelId 设置为: %s", mainTask.ID, modelIdStr)
	} else if len(allProteinIds) > 0 {
		// 对于AlphaFold和I-Tasser，先设置主序列ID，后续异步更新
		modelIdStr := strings.Join(allProteinIds, ",")
		if err := database.Database.Model(&mainTask).Update("model_id", modelIdStr).Error; err != nil {
			logger.Error("更新任务ModelId字段失败: %v", err)
		}
		logger.Info("Fold Task %d ModelId 设置为: %s", mainTask.ID, modelIdStr)
	}

	return FoldResponse{ID: mainTask.ID}
}

// findSubsequencePosition 找到子序列在主序列中的位置
func findSubsequencePosition(mainSeq, subSeq string) int {
	return strings.Index(mainSeq, subSeq)
}
