package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// BlastRequest 请求结构体
type BlastRequest struct {
	Code  string `json:"code" binding:"required"`
	Title string `json:"title" binding:"required"`
	Type  int64  `json:"type" binding:"required"`
}

// BlastResponse 响应结构体
type BlastResponse struct {
	ID    uint   `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// Blast 处理 blast 请求的主要函数
func Blast(code, title string, typeValue int64, userId int64) BlastResponse {
	// 检查项目是否已存在
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

	// 如果主序列既不在 protein_information 中也不在队列中，则添加到队列
	if mainProteinInfo.ID == 0 && mainQueueCount == 0 {
		addSequenceToQueue(code, typeValue)
	}

	// 创建主任务
	mainTask := models.Task{
		Title:    title,
		Sequence: code,
		Type:     1, // Sequence Search
		UserId:   userId,
		ModelId:  strconv.FormatUint(uint64(mainProteinInfo.ID), 10),
	}

	if err := database.Database.Create(&mainTask).Error; err != nil {
		return BlastResponse{Error: "创建主任务失败"}
	}

	// 处理所有子序列，只创建蛋白质信息和队列项
	for i := range subSequences {
		// 获取子序列的描述信息
		description := getDescription(blastInformations[i])
		fasta := code[description.From-1 : description.To]

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

		// 如果子序列既不在 protein_information 中也不在队列中，则添加到队列
		if subProteinInfo.ID == 0 && subQueueCount == 0 {
			// 为子序列设置父ID（主序列的蛋白质信息ID）
			var parentId *int64
			if mainProteinInfo.ID > 0 {
				parentIdInt := int64(mainProteinInfo.ID)
				parentId = &parentIdInt
			}
			addSequenceToQueueWithParent(fasta, typeValue, parentId)
		}

		// 将描述信息转换为 JSON 字符串
		informationJSON := ""
		if infoJSON, err := json.Marshal(blastInformations[i]); err == nil {
			informationJSON = string(infoJSON)
		}

		// 创建或更新蛋白质信息，设置父序列ID为主序列的蛋白质信息ID
		ProteinInformationWithParent(fasta, informationJSON, typeValue, mainProteinInfo.ID)
	}

	return BlastResponse{ID: mainTask.ID}
}

// addSequenceToQueue 根据类型将序列添加到相应的队列
func addSequenceToQueue(sequence string, typeValue int64) {
	addSequenceToQueueWithParent(sequence, typeValue, nil)
}

// addSequenceToQueueWithParent 根据类型将序列添加到相应的队列，支持父ID
func addSequenceToQueueWithParent(sequence string, typeValue int64, parentId *int64) {
	switch typeValue {
	case 1: // I-Tasser
		AddToITasserQueueWithParent(sequence, parentId)
	case 2: // AlphaFold2
		AddToAlphaFoldQueueWithParent(sequence, parentId)
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
	From int `json:"from"`
	To   int `json:"to"`
}

// getDescription 从 BLAST 信息中提取描述
func getDescription(blastInfo map[string]string) BlastDescription {
	from, _ := strconv.Atoi(blastInfo["From"])
	to, _ := strconv.Atoi(blastInfo["To"])
	return BlastDescription{
		From: from,
		To:   to,
	}
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

// GetBlastList 查询主任务的分页列表
func GetBlastList(userId int64, current, pageSize int, title, createStart, createEnd string) (BlastListResult, error) {
	var tasks []models.Task
	var total int64

	db := database.Database.Model(&models.Task{}).
		Where("user_id = ? AND type = ?", userId, 1) // type=1为主任务

	if title != "" {
		db = db.Where("title LIKE ?", "%"+title+"%")
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
		// 统计模型总数和有模型的数量
		var modelTotal int64
		var hasModelCount int64
		database.Database.Model(&models.Task{}).
			Where("id = ? OR parent_id = ?", task.ID, task.ID).
			Count(&modelTotal)
		database.Database.Model(&models.Task{}).
			Where("(id = ? OR parent_id = ?) AND model_id <> '' AND model_id <> '0'", task.ID, task.ID).
			Count(&hasModelCount)

		// 组装
		list = append(list, BlastListItem{
			Category:      1, // 你可以根据业务调整
			CreatedAt:     task.CreatedAt.UnixMilli(),
			Fasta:         task.Sequence,
			HasModelCount: hasModelCount,
			ID:            task.ID,
			Information:   nil,
			ModelId:       task.ModelId,
			ModelTotal:    modelTotal,
			ParentId:      nil,     // 主任务无parent
			SubSequence:   nil,     // 可选: &task.SubSequence
			TaskType:      "blast", // 或根据type映射
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
