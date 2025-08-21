package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
)

type ViewNoteResult struct {
	Data  string `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// ViewNote 查看用户对某个序列的注释
func ViewNote(userId uint, sequenceId uint) ViewNoteResult {
	var notes []models.Note
	
	// 查找指定用户对指定任务的注释记录（根据实际模型结构调整）
	if err := database.Database.Where("user_id = ? AND task_id = ?", userId, sequenceId).Find(&notes).Error; err != nil {
		return ViewNoteResult{Error: "Network error."}
	}
	
	// 如果没有记录，返回空字符串
	if len(notes) == 0 {
		return ViewNoteResult{Data: ""}
	}
	
	// 返回该结构的注释内容
	return ViewNoteResult{Data: notes[0].Note}
}

type ModelInfo struct {
	ID  uint   `json:"id"`
	Seq string `json:"seq"`
}

type GetAllModelNotMeResult struct {
	Data  []ModelInfo `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// GetAllModelNotMe 获取与当前序列不同的已建模列表（排除自身）
func GetAllModelNotMe(seq string) GetAllModelNotMeResult {
	var proteinInfos []models.ProteinInformation
	
	// 查找与当前序列不同的蛋白质信息记录
	if err := database.Database.Where("sequence <> ?", seq).Find(&proteinInfos).Error; err != nil {
		return GetAllModelNotMeResult{Error: "Network error!"}
	}
	
	// 转换为返回格式
	result := make([]ModelInfo, 0, len(proteinInfos))
	for _, info := range proteinInfos {
		result = append(result, ModelInfo{
			ID:  info.ID,
			Seq: info.Sequence,
		})
	}
	
	return GetAllModelNotMeResult{Data: result}
}

type UpdateNoteResult struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// UpdateNote 更新或创建用户对某个序列的注释
func UpdateNote(noteContent string, userId uint, sequenceId uint) UpdateNoteResult {
	// 查询是否已有该用户对该任务的记录
	var noteCount int64
	if err := database.Database.Model(&models.Note{}).Where("user_id = ? AND task_id = ?", userId, sequenceId).Count(&noteCount).Error; err != nil {
		return UpdateNoteResult{Error: "Network error."}
	}
	
	if noteCount > 0 {
		// 如果有记录，执行更新操作
		if err := database.Database.Model(&models.Note{}).
			Where("user_id = ? AND task_id = ?", userId, sequenceId).
			Update("note", noteContent).Error; err != nil {
			return UpdateNoteResult{Error: "Network error."}
		}
	} else {
		// 如果没有记录，创建新的注释记录
		newNote := models.Note{
			Note:   noteContent,
			UserId: int64(userId),
			TaskId: int64(sequenceId),
		}
		if err := database.Database.Create(&newNote).Error; err != nil {
			return UpdateNoteResult{Error: "Network error."}
		}
	}
	
	return UpdateNoteResult{Message: "Update successfully!"}
} 