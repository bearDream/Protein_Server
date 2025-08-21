package controllers

import (
	"Protein_Server/database"
	"Protein_Server/logger"
	"Protein_Server/models"
	"Protein_Server/services"
	"Protein_Server/utils"
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// ViewNote 查看用户对某个序列的注释
// POST /viewNote {"sequenceId": 123}
type ViewNoteRequest struct {
	SequenceId uint `json:"sequenceId" binding:"required"`
}

func ViewNote(c *gin.Context) {
	var req ViewNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}

	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	result := services.ViewNote(userByToken.ID, req.SequenceId)
	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}

	utils.Success(c, gin.H{"data": result.Data}, "ok")
}

// UpdateNote 更新或创建用户对某个序列的注释
// POST /updateNote {"no": "注释内容", "sequenceId": 123}
type UpdateNoteRequest struct {
	No         string `json:"no" binding:"required"`
	SequenceId uint   `json:"sequenceId" binding:"required"`
}

func UpdateNote(c *gin.Context) {
	var req UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}

	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	result := services.UpdateNote(req.No, userByToken.ID, req.SequenceId)
	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}

	utils.Success(c, gin.H{"msg": result.Message}, "ok")
}

// FoldRequest fold请求结构体
type FoldRequest struct {
	Codes []string `json:"codes" binding:"required"`
	Title string   `json:"title" binding:"required"`
	Type  string   `json:"type" binding:"required"`
}

// GetAllModelNotMe 获取与当前序列不同的已建模列表（排除自身）
type GetAllModelNotMeRequest struct {
	Seq string `json:"seq" binding:"required"`
}

func Fold(c *gin.Context) {
	var foldRequest FoldRequest
	// 绑定请求参数
	if err := c.Bind(&foldRequest); err != nil {
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

	// 调用 fold 服务
	result := services.Fold(foldRequest.Codes, foldRequest.Title, foldRequest.Type, int64(user.ID))

	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}

	utils.Success(c, gin.H{"id": result.ID}, "ok")
}

func GetAllModelNotMe(c *gin.Context) {
	var req GetAllModelNotMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}

	result := services.GetAllModelNotMe(req.Seq)
	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}

	utils.Success(c, result.Data, "ok")
}

// UploadPDB 上传PDB文件
// POST /uploadPDB (multipart/form-data)
func UploadPDB(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "No file uploaded or file upload error")
		return
	}

	// 检查文件扩展名
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".pdb") {
		utils.Error(c, 400, "Only PDB files are allowed")
		return
	}

	// 生成唯一的文件名（使用时间戳）
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)

	// 确保上传目录存在
	uploadDir := "uploads/pdb"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		utils.Error(c, 500, "Failed to create upload directory")
		return
	}

	// 构建文件保存路径
	filePath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.Error(c, 500, "Failed to save file")
		return
	}

	// 返回文件路径
	utils.Success(c, filePath, "File uploaded successfully")
}

// PDB2X3D 将PDB文件转换为X3D格式
// POST /pdb2x3d (multipart/form-data)
func PDB2X3D(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "No file uploaded or file upload error")
		return
	}

	// 检查文件是否为PDB格式
	if !strings.Contains(file.Filename, ".pdb") {
		utils.Error(c, 400, "file type error")
		return
	}

	// 生成唯一的文件名（使用时间戳）
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)

	// 确保临时目录存在
	tempDir := "temp/pdb2x3d"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		utils.Error(c, 500, "Failed to create temp directory")
		return
	}

	// 构建文件保存路径
	filePath := filepath.Join(tempDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.Error(c, 500, "Failed to save file")
		return
	}

	// 获取不带扩展名的文件路径
	purePath := strings.Split(filePath, ".pdb")[0]

	// 执行chimera转换命令
	cmd := fmt.Sprintf(`chimera --script "./py-script/pdb_export.py %s"`, purePath)
	if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
		logger.Error("Chimera转换失败: %v", err)
		utils.Error(c, 500, "convert error")
		return
	}

	// 构建X3D文件路径
	x3dPath := purePath + ".x3d"

	// 检查X3D文件是否生成成功
	if _, err := os.Stat(x3dPath); os.IsNotExist(err) {
		utils.Error(c, 500, "X3D file not generated")
		return
	}

	// 设置响应头
	c.Header("Content-Type", "model/x3d+xml")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.x3d", strings.Split(file.Filename, ".pdb")[0]))

	// 返回文件
	c.File(x3dPath)

	// 清理临时文件（可选，也可以定期清理）
	go func() {
		time.Sleep(5 * time.Minute) // 5分钟后删除临时文件
		os.Remove(filePath)
		os.Remove(x3dPath)
	}()
}

// PDB2OBJ 将PDB文件转换为OBJ格式并打包成ZIP
// POST /pdb2obj (multipart/form-data)
func PDB2OBJ(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "No file uploaded or file upload error")
		return
	}

	// 检查文件是否为PDB格式
	if !strings.Contains(file.Filename, ".pdb") {
		utils.Error(c, 400, "file type error")
		return
	}

	// 生成唯一的文件名（使用时间戳）
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)

	// 确保临时目录存在
	tempDir := "temp/pdb2obj"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		utils.Error(c, 500, "Failed to create temp directory")
		return
	}

	// 构建文件保存路径
	filePath := filepath.Join(tempDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.Error(c, 500, "Failed to save file")
		return
	}

	// 获取不带扩展名的文件路径
	purePath := strings.Split(filePath, ".pdb")[0]

	// 步骤1: 执行chimera转换命令 (pdb -> x3d)
	chimeraCmd := fmt.Sprintf(`chimera --script "./py-script/pdb_export.py %s"`, purePath)
	if err := exec.Command("sh", "-c", chimeraCmd).Run(); err != nil {
		logger.Error("Chimera转换失败: %v", err)
		utils.Error(c, 500, "convert error")
		return
	}

	// 步骤2: 执行blender转换命令 (x3d -> obj)
	blenderCmd := fmt.Sprintf("blender --background --python ./py-script/x3d_export_obj.py -- %s", purePath)
	logger.Info("执行Blender命令: %s", blenderCmd)
	if err := exec.Command("sh", "-c", blenderCmd).Run(); err != nil {
		logger.Error("Blender转换失败: %v", err)
		utils.Error(c, 500, "convert error")
		return
	}

	// 构建文件路径
	objPath := purePath + ".obj"
	mtlPath := purePath + ".mtl"

	// 检查生成的文件是否存在
	if _, err := os.Stat(objPath); os.IsNotExist(err) {
		utils.Error(c, 500, "OBJ file not generated")
		return
	}

	// 创建目录用于存放重命名的文件
	if err := os.MkdirAll(purePath, 0755); err != nil {
		logger.Error("创建目录失败: %v", err)
		utils.Error(c, 500, "Failed to create directory")
		return
	}

	// 重命名文件
	newObjPath := filepath.Join(purePath, "pdb.obj")
	newMtlPath := filepath.Join(purePath, "pdb.mtl")

	if err := os.Rename(objPath, newObjPath); err != nil {
		logger.Error("重命名OBJ文件失败: %v", err)
		utils.Error(c, 500, "Failed to rename OBJ file")
		return
	}

	// MTL文件可能不存在，所以错误不是致命的
	if _, err := os.Stat(mtlPath); err == nil {
		if err := os.Rename(mtlPath, newMtlPath); err != nil {
			logger.Error("重命名MTL文件失败: %v", err)
		}
	}

	// 创建ZIP文件
	zipPath := purePath + ".zip"
	if err := createZipFromDir(purePath, zipPath); err != nil {
		logger.Error("创建ZIP文件失败: %v", err)
		utils.Error(c, 500, "Failed to create ZIP file")
		return
	}

	// 检查ZIP文件是否生成成功
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		utils.Error(c, 500, "ZIP file not generated")
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", strings.Split(file.Filename, ".pdb")[0]))

	// 返回ZIP文件
	c.File(zipPath)

	// 清理临时文件（可选，也可以定期清理）
	go func() {
		time.Sleep(5 * time.Minute) // 5分钟后删除临时文件
		os.Remove(filePath)
		os.RemoveAll(purePath)
		os.Remove(zipPath)
		os.Remove(purePath + ".x3d") // 清理中间生成的x3d文件
	}()
}

// PDB2FBX 将PDB文件转换为FBX格式
// POST /pdb2fbx (multipart/form-data)
func PDB2FBX(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "No file uploaded or file upload error")
		return
	}

	// 检查文件是否为PDB格式
	if !strings.Contains(file.Filename, ".pdb") {
		utils.Error(c, 400, "file type error")
		return
	}

	// 生成唯一的文件名（使用时间戳）
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)

	// 确保临时目录存在
	tempDir := "temp/pdb2fbx"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		utils.Error(c, 500, "Failed to create temp directory")
		return
	}

	// 构建文件保存路径
	filePath := filepath.Join(tempDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		utils.Error(c, 500, "Failed to save file")
		return
	}

	// 获取不带扩展名的文件路径
	purePath := strings.Split(filePath, ".pdb")[0]

	// 步骤1: 执行chimera转换命令 (pdb -> x3d)
	chimeraCmd := fmt.Sprintf(`chimera --script "./py-script/pdb_export.py %s"`, purePath)
	if err := exec.Command("sh", "-c", chimeraCmd).Run(); err != nil {
		logger.Error("Chimera转换失败: %v", err)
		utils.Error(c, 500, "convert error")
		return
	}

	// 步骤2: 执行blender转换命令 (x3d -> fbx)
	blenderCmd := fmt.Sprintf("blender --background --python ./py-script/x3d_export_fbx.py -- %s", purePath)
	if err := exec.Command("sh", "-c", blenderCmd).Run(); err != nil {
		logger.Error("Blender转换失败: %v", err)
		utils.Error(c, 500, "convert error")
		return
	}

	// 构建FBX文件路径
	fbxPath := purePath + ".fbx"

	// 检查FBX文件是否生成成功
	if _, err := os.Stat(fbxPath); os.IsNotExist(err) {
		utils.Error(c, 500, "FBX file not generated")
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.fbx", strings.Split(file.Filename, ".pdb")[0]))

	// 返回文件
	c.File(fbxPath)

	// 清理临时文件（可选，也可以定期清理）
	go func() {
		time.Sleep(5 * time.Minute) // 5分钟后删除临时文件
		os.Remove(filePath)
		os.Remove(fbxPath)
		os.Remove(purePath + ".x3d") // 清理中间生成的x3d文件
	}()
}

func GetBlastList(c *gin.Context) {
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	// 解析分页参数
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	title := c.Query("title")
	category := c.Query("category")
	createStart := c.Query("createStart")
	createEnd := c.Query("createEnd")

	// 查询
	result, err := services.GetBlastList(int64(userByToken.ID), current, pageSize, title, category, createStart, createEnd)
	if err != nil {
		utils.Success(c, 500, err.Error())
		return
	}
	utils.Success(c, result, "ok")
}

func GetBlastResult(c *gin.Context) {
	var req struct {
		ID string `json:"id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}

	// 调用服务层获取结果
	result, err := services.GetBlastResult(req.ID)
	if err != nil {
		utils.Error(c, 500, err.Error())
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

// ShareResponse 分享记录响应结构体
type ShareResponse struct {
	ID     uint  `json:"id"`
	TaskId uint  `json:"taskId"`
	ToId   uint  `json:"toId"`
	Status int64 `json:"status"`
	FromId uint  `json:"fromId"`
	SeqId  uint  `json:"seqId"`
}

func ShowShare(c *gin.Context) {
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	var shares []models.Share
	// 查询与当前用户相关的分享，只查询状态为0（未处理）的记录
	if err := database.Database.Where("to_id = ? AND status = ?", userByToken.ID, 0).Find(&shares).Error; err != nil {
		utils.Error(c, 500, "Network error.")
		return
	}

	// 转换为前端期望的格式
	var shareResponses []ShareResponse
	for _, share := range shares {
		shareResponses = append(shareResponses, ShareResponse{
			ID:     share.ID,
			TaskId: share.TaskId,
			ToId:   share.ToId,
			Status: share.Status,
			FromId: share.FromId,
			SeqId:  share.SeqId,
		})
	}

	if len(shareResponses) == 0 {
		utils.Success(c, shareResponses, "No share found.")
		return
	}
	utils.Success(c, shareResponses, "ok")
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
		TaskId: taskid,
		ToId:   userid,
		Status: 0,
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
		UserId:                  int64(share.ToId),
		SubSequence:             task.SubSequence,
	}
	if err := database.Database.Create(&newtask).Error; err != nil {
		utils.Error(c, 400, "Failed to copy task")
		return
	}
	utils.Success(c, nil, "Agreed successfully")
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

// createZipFromDir 创建ZIP文件从目录
func createZipFromDir(sourceDir, zipPath string) error {
	// 创建ZIP文件
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// 创建ZIP写入器
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 遍历目录
	return filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if info.IsDir() {
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}

		// 在ZIP中创建文件
		zipFileWriter, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// 打开源文件
		srcFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// 复制文件内容到ZIP
		_, err = io.Copy(zipFileWriter, srcFile)
		return err
	})
}

// ShareBlastRequest shareBlast请求结构体
type ShareBlastRequest struct {
	SeqId  uint `json:"seqId" binding:"required"`  // 序列ID
	UserId uint `json:"userId" binding:"required"` // 目标用户ID
}

// ShareBlast 分享Blast结果给其他用户
func ShareBlast(c *gin.Context) {
	var req ShareBlastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}

	// 从token中获取当前用户ID
	account, _ := c.Get("account")
	userByToken := account.(*services.AccountClaims).User

	// 创建分享记录
	share := &models.Share{
		FromId: userByToken.ID,
		ToId:   req.UserId,
		TaskId: req.SeqId,
		SeqId:  req.SeqId,
		Status: 0, // 0: 未处理
	}

	if err := database.Database.Create(&share).Error; err != nil {
		utils.Error(c, 500, "database error!")
		return
	}

	utils.Success(c, nil, "Shared successfully")
}

// AgreeShareBlastRequest agreeShareBlast请求结构体
type AgreeShareBlastRequest struct {
	Id uint `json:"id" binding:"required"` // 分享记录ID
}

// AgreeShareBlast 同意分享Blast结果
func AgreeShareBlast(c *gin.Context) {
	var req AgreeShareBlastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}

	// 查找分享记录
	var share models.Share
	if err := database.Database.Where("id = ?", req.Id).First(&share).Error; err != nil {
		utils.Error(c, 400, "Share record not found")
		return
	}

	// 查找原始任务（用于复制任务信息）
	var originalTask models.Task
	if err := database.Database.Where("id = ?", share.TaskId).First(&originalTask).Error; err != nil {
		utils.Error(c, 400, "Original task not found")
		return
	}

	// 为接收用户创建新的任务记录
	newTask := models.Task{
		Title:                   originalTask.Title,
		Sequence:                originalTask.Sequence,
		Type:                    originalTask.Type,
		StructurePredictionTool: originalTask.StructurePredictionTool,
		UserId:                  int64(share.ToId),
		SubSequence:             originalTask.SubSequence,
		ModelId:                 originalTask.ModelId,
	}

	if err := database.Database.Create(&newTask).Error; err != nil {
		utils.Error(c, 500, "Network error!")
		return
	}

	// 更新分享状态为同意
	share.Status = 1
	if err := database.Database.Save(&share).Error; err != nil {
		utils.Error(c, 500, "Network error!")
		return
	}

	utils.Success(c, nil, "Agreed successfully")
}

// RefuseShareBlastRequest refuseShareBlast请求结构体
type RefuseShareBlastRequest struct {
	Id uint `json:"id" binding:"required"` // 分享记录ID
}

// RefuseShareBlast 拒绝分享Blast结果
func RefuseShareBlast(c *gin.Context) {
	var req RefuseShareBlastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "Parameter error")
		return
	}

	// 查找分享记录
	var share models.Share
	if err := database.Database.Where("id = ?", req.Id).First(&share).Error; err != nil {
		utils.Error(c, 400, "Share record not found")
		return
	}

	// 更新分享状态为拒绝
	share.Status = 2
	if err := database.Database.Save(&share).Error; err != nil {
		utils.Error(c, 500, "Network error!")
		return
	}

	utils.Success(c, true, "Refused successfully")
}

func SearchPdbByParam(c *gin.Context) {
	// 获取查询参数
	currentStr := c.DefaultQuery("current", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	sequence := c.Query("sequence")
	pdbId := c.Query("pdbId")
	rcScore := c.DefaultQuery("rcScore", ",")
	hydrophobicity := c.DefaultQuery("hydrophobicity", ",")
	instability := c.DefaultQuery("instability", ",")
	isoelectricPoint := c.DefaultQuery("isoelectricPoint", ",")
	size := c.DefaultQuery("size", ",")
	solventAccesibility := c.DefaultQuery("solventAccesibility", ",")

	// 转换分页参数
	current, err := strconv.Atoi(currentStr)
	if err != nil {
		utils.Error(c, 400, "Invalid current parameter")
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		utils.Error(c, 400, "Invalid pageSize parameter")
		return
	}

	result := services.SearchPdbByParam(
		sequence,
		pdbId,
		rcScore,
		hydrophobicity,
		instability,
		isoelectricPoint,
		size,
		solventAccesibility,
		current,
		pageSize,
		"score,DESC",
	)

	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}

	utils.Success(c, result.Data, "ok")
}

func GetPDBInformationById(c *gin.Context) {
	// 获取查询参数
	pdbId := c.Query("pdbId")
	if pdbId == "" {
		utils.Error(c, 400, "pdbId parameter is required")
		return
	}

	result := services.GetPDBInformationById(pdbId)
	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}

	utils.Success(c, result.Data, "ok")
}

func GetSeqTimeTable(c *gin.Context) {
	result := services.GetSeqTimeTable()
	if result.Error != "" {
		c.JSON(400, gin.H{"error": result.Error})
		return
	}

	// 直接返回数组，不包装在其他字段中
	c.JSON(200, result.Data)
}

func GetPDBParameterList(c *gin.Context) {
	// 获取查询参数
	currentStr := c.DefaultQuery("current", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	sort := c.Query("sort")
	pdbId := c.Query("pdbId")
	fasta := c.Query("fasta")
	rcScore := c.DefaultQuery("rcScore", ",")
	hydrophobicity := c.DefaultQuery("hydrophobicity", ",")
	instability := c.DefaultQuery("instability", ",")
	isoelectricPoint := c.DefaultQuery("isoelectricPoint", ",")
	size := c.DefaultQuery("size", ",")
	solventAccesibility := c.DefaultQuery("solventAccesibility", ",")

	// 转换分页参数
	current, err := strconv.Atoi(currentStr)
	if err != nil {
		utils.Error(c, 400, "Invalid current parameter")
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		utils.Error(c, 400, "Invalid pageSize parameter")
		return
	}

	result := services.GetPDBParameterList(
		fasta,
		pdbId,
		rcScore,
		hydrophobicity,
		instability,
		isoelectricPoint,
		size,
		solventAccesibility,
		current,
		pageSize,
		sort,
	)

	if result.Error != "" {
		utils.Error(c, 400, result.Error)
		return
	}

	utils.Success(c, result.Data, "ok")
}
