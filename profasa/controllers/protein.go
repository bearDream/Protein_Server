package controllers

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"Protein_Server/services"
	"Protein_Server/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func SequenceSearch(c *gin.Context) {
	var task models.Task
	// Bind automatically parses the input parameters of the api to variables using the form description in the struct
	if err := c.Bind(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter error"})
		return
	}

	// call BlastProcessing and get subSequences
	task.Type = 1
	subSequences, blastinformations := services.BlastProcessing(task.Sequence)
	task.SubSequence = strings.Join(subSequences, "|")

	// find or create protein information
	// main sequence
	if task.StructurePredictionTool != nil {
		services.ProteinInformation(task.Sequence, "", *task.StructurePredictionTool)
	}
	// subSequences
	for i := range subSequences {
		blastinformationstr := ""
		blastinformation, err := json.Marshal(blastinformations[i])
		if err == nil {
			blastinformationstr = string(blastinformation)
		}
		if task.StructurePredictionTool != nil {
			services.ProteinInformation(subSequences[i], blastinformationstr, *task.StructurePredictionTool)
		}
	}

	// create a task
	if err := database.Database.Create(&task).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Network exception"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Executed successfully"})
}

func StructurePrediction(c *gin.Context) {
	var task models.Task
	// Bind automatically parses the input parameters of the api to variables using the form description in the struct
	if err := c.Bind(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter error"})
		return
	}

	// Structure Prediction
	task.Type = 2
	subSequences := strings.Split(task.SubSequence, "|")
	task.Sequence = strings.Join(subSequences, "")

	// find or create protein information
	// main sequence
	if task.StructurePredictionTool != nil {
		services.ProteinInformation(task.Sequence, "", *task.StructurePredictionTool)
	}
	// subSequences
	for i := range subSequences {
		if task.StructurePredictionTool != nil {
			services.ProteinInformation(subSequences[i], "", *task.StructurePredictionTool)
		}
	}

	// create a task
	if err := database.Database.Create(&task).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Network exception"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Executed successfully"})
}

func ParametersCalculation(c *gin.Context) {

	// c.JSON(http.StatusOK, gin.H{"msg": "Created successfully"})
}

type SuperimposeRequest struct {
	Path  []string `json:"path" binding:"required"` // 两个pdb文件的路径
	Title string   `json:"title" binding:"required"`
}

func Superimpose(c *gin.Context) {
	var req SuperimposeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Params error.")
		return
	}
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	result := services.Superimpose(req.Path, req.Title, int64(userByToken.ID))
	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}
	utils.Success(c, gin.H{"id": result.ID}, "ok")
}

// 单个pdb上传
// POST /single {"path": "xxx", "title": "xxx"}
type SingleRequest struct {
	Path  string `json:"path" binding:"required"`
	Title string `json:"title" binding:"required"`
}

func Single(c *gin.Context) {
	var req SingleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Params error.")
		return
	}
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	result := services.Single(req.Path, req.Title, int64(userByToken.ID))
	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}
	utils.Success(c, gin.H{"id": result.ID}, "ok")
}

func Blast(c *gin.Context) {
	var blastRequest services.BlastRequest
	// 绑定请求参数
	if err := c.Bind(&blastRequest); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}

	// 从 token 中获取用户 ID
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User
	var user models.User
	if err := database.Database.Where("email = ?", userByToken.Email).First(&user).Error; err != nil {
		utils.Error(c, 400, "Failed fetch user information")
		return
	}

	// 调用 blast 服务
	result := services.Blast(blastRequest.Code, blastRequest.Title, blastRequest.Type, int64(user.ID))

	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}

	utils.Success(c, gin.H{"id": result.ID}, "ok")
}

func GetBlastList(c *gin.Context) {
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	// 解析分页参数
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	title := c.Query("title")
	createStart := c.Query("createStart")
	createEnd := c.Query("createEnd")

	// 查询
	result, err := services.GetBlastList(int64(userByToken.ID), current, pageSize, title, createStart, createEnd)
	if err != nil {
		utils.Success(c, 500, err.Error())
		return
	}
	utils.Success(c, result, "ok")
}

func TaskList(c *gin.Context) {
	// get page and size
	pagequery := models.PageQuery{}
	if c.ShouldBindQuery(&pagequery) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failure to analytic parameter"})
		return
	}
	// get user id
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User
	var user models.User
	if err := database.Database.Where("email = ?", userByToken.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failure to obtain information"})
		return
	}
	// Get only your own tasks
	var task []models.Task
	var total int64
	if err := database.Database.Where("user_id=?", user.ID).Count(&total).Offset((pagequery.Page - 1) * pagequery.Size).Limit(pagequery.Size).Order("created_at desc").Find(&task).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	data := map[string]interface{}{
		"page_total": (int(total) + pagequery.Size - 1) / pagequery.Size,
		"data_list":  task,
		"data_total": total,
	}
	c.JSON(http.StatusOK, data)
}

func TaskDetails(c *gin.Context) {
	// get id
	var cid uint
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		cid = uint(id)
	}
	// get task information
	var task models.Task
	if err := database.Database.Where("id = ?", cid).Find(&task).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// get sequence infotmation
	var proteininformation models.ProteinInformation
	if err := database.Database.Where("sequence = ?", task.Sequence).Find(&proteininformation).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// get subsequence infotmation
	subsequenceList := make([]models.ProteinInformation, 0)
	if task.SubSequence != "" {
		subsequences := strings.Split(task.SubSequence, "|")
		for _, subsubsequence := range subsequences {
			var proteininformation models.ProteinInformation
			if err := database.Database.Where("sequence = ?", subsubsequence).Find(&proteininformation).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			subsequenceList = append(subsequenceList, proteininformation)
		}
	}
	data := map[string]interface{}{
		"task":          task,
		"main_sequence": proteininformation,
		"subsequences":  subsequenceList,
	}
	utils.Success(c, data, "ok")
}

func DeleteTask(c *gin.Context) {
	var task models.Task
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		task.ID = uint(id)
	}
	if err := database.Database.Delete(&task).Error; err != nil {
		logger.Error("删除任务失败: %v", err)
		utils.Error(c, 500, "Delete task failed")
		return
	}
	utils.Success(c, nil, "Deleted successfully")
}

func ShowShare(c *gin.Context) {
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	var shares []models.Share
	// 查询与当前用户相关的分享（你可以根据业务调整是 ToUserId 还是 FromId）
	if err := database.Database.Where("to_user_id = ?", userByToken.ID).Find(&shares).Error; err != nil {
		utils.Error(c, 500, "Network error.")
		return
	}
	if len(shares) == 0 {
		utils.Success(c, shares, "No share found.")
		return
	}
	utils.Success(c, shares, "ok")
}

func ShareTask(c *gin.Context) {
	// get task id
	var taskid uint
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		utils.Error(c, 400, err.Error())
		return
	} else {
		taskid = uint(id)
	}
	// get user id
	var userid uint
	if id, err := strconv.Atoi(c.Param("userid")); err != nil {
		utils.Error(c, 400, err.Error())
		return
	} else {
		userid = uint(id)
	}
	// create a share table
	share := &models.Share{
		TaskId:   taskid,
		ToUserId: userid,
		Status:   0,
	}
	if err := database.Database.Create(&share).Error; err != nil {
		utils.Error(c, 400, "Network error")
		return
	}
	utils.Success(c, nil, "Shared successfully")
}

func AgreeShare(c *gin.Context) {
	// share id
	var shareid uint
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		utils.Error(c, 400, err.Error())
		return
	} else {
		shareid = uint(id)
	}
	// find share
	var share models.Share
	if err := database.Database.Where("id = ?", shareid).Find(&share).Error; err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	share.Status = 1
	database.Database.Updates(share)
	// find task
	var task models.Task
	if err := database.Database.Where("id = ?", share.TaskId).Find(&task).Error; err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	newtask := &models.Task{
		Title:                   task.Title,
		Sequence:                task.Sequence,
		Type:                    task.Type,
		StructurePredictionTool: task.StructurePredictionTool,
		UserId:                  int64(share.ToUserId),
		SubSequence:             task.SubSequence,
	}
	if err := database.Database.Create(&newtask).Error; err != nil {
		utils.Error(c, 400, "Failed to copy task")
		return
	}
	utils.Success(c, nil, "Agreed successfully")
}

func RejectShare(c *gin.Context) {
	// share id
	var shareid uint
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		utils.Error(c, 400, err.Error())
		return
	} else {
		shareid = uint(id)
	}
	// find share
	var share models.Share
	if err := database.Database.Where("id = ?", shareid).Find(&share).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	share.Status = 2
	database.Database.Updates(share)
	utils.Success(c, nil, "Rejected successfully")
}

func ViewNotes(c *gin.Context) {
	// get task id
	var taskid uint
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		utils.Error(c, 400, err.Error())
		return
	} else {
		taskid = uint(id)
	}
	// get note
	var note models.Note
	if err := database.Database.Where("id = ?", taskid).Find(&note).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	utils.Success(c, note, "ok")
}
func UpdateNotes(c *gin.Context) {
	var note models.Note
	// Bind automatically parses the input parameters of the api to variables using the form description in the struct
	if err := c.Bind(&note); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}
	// get id
	if id, err := strconv.Atoi(c.Param("id")); err != nil {
		utils.Error(c, 400, err.Error())
		return
	} else {
		note.ID = uint(id)
	}
	if err := database.Database.Save(&note).Error; err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, nil, "Success")
}
