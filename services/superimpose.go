package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"fmt"
	"os/exec"
	"strings"
)

type SuperimposeResult struct {
	ID    uint   `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// path: [pdb1, pdb2]
func Superimpose(paths []string, title string, userId int64) SuperimposeResult {
	if len(paths) != 2 {
		return SuperimposeResult{Error: "Params error."}
	}
	path1, path2 := paths[0], paths[1]

	// 1. pdb2fasta
	fasta1, err := pdb2fasta(path1)
	if err != nil {
		return SuperimposeResult{Error: "pdb2fasta error."}
	}
	// 2. 写入 protein_information
	mainInfo := models.ProteinInformation{
		Sequence: fasta1,
	}
	if err := database.Database.Create(&mainInfo).Error; err != nil {
		return SuperimposeResult{Error: "DB error."}
	}
	// 3. 移动pdb文件
	modelPath1 := fmt.Sprintf("static/models/%d.pdb", mainInfo.ID)
	if err := exec.Command("mv", path1, modelPath1).Run(); err != nil {
		return SuperimposeResult{Error: "Move pdb1 error."}
	}
	// 4. 生成图片
	imgPath1 := fmt.Sprintf("static/imgs/%d.png", mainInfo.ID)
	if err := exec.Command("bash", "-c", "source /root/miniconda3/etc/profile.d/conda.sh && conda activate alphafold && python3 py-scripts/pdb2img.py", modelPath1, imgPath1).Run(); err != nil {
		return SuperimposeResult{Error: "pdb2img1 error."}
	}
	// 5. 创建主任务
	mainTask := models.Task{
		Title:    title,
		Sequence: fasta1,
		Type:     4, // superimpose
		UserId:   userId,
		ModelId:  fmt.Sprintf("%d", mainInfo.ID),
		// 其他字段可补充
	}
	if err := database.Database.Create(&mainTask).Error; err != nil {
		return SuperimposeResult{Error: "DB error."}
	}

	// 6. 处理第二个pdb
	fasta2, err := pdb2fasta(path2)
	if err != nil {
		return SuperimposeResult{Error: "pdb2fasta2 error."}
	}
	subInfo := models.ProteinInformation{
		Sequence: fasta2,
		ParentId: mainInfo.ID,
	}
	if err := database.Database.Create(&subInfo).Error; err != nil {
		return SuperimposeResult{Error: "DB error."}
	}
	modelPath2 := fmt.Sprintf("static/models/%d.pdb", subInfo.ID)
	if err := exec.Command("mv", path2, modelPath2).Run(); err != nil {
		return SuperimposeResult{Error: "Move pdb2 error."}
	}
	imgPath2 := fmt.Sprintf("static/imgs/%d.png", subInfo.ID)
	if err := exec.Command("bash", "-c", "source /root/miniconda3/etc/profile.d/conda.sh && conda activate alphafold && python3 py-scripts/pdb2img.py", modelPath2, imgPath2).Run(); err != nil {
		return SuperimposeResult{Error: "pdb2img2 error."}
	}
	// 7. 创建子任务
	subTask := models.Task{
		Title:    title,
		Sequence: fasta2,
		Type:     4, // superimpose
		UserId:   userId,
		ModelId:  fmt.Sprintf("%d", subInfo.ID),
	}
	if err := database.Database.Create(&subTask).Error; err != nil {
		return SuperimposeResult{Error: "DB error."}
	}

	return SuperimposeResult{ID: mainTask.ID}
}

// 单个pdb上传,analysis
func Single(path string, title string, userId int64) SuperimposeResult {
	fasta, err := pdb2fasta(path)
	if err != nil {
		return SuperimposeResult{Error: "pdb2fasta error."}
	}
	// 写入 protein_information
	mainInfo := models.ProteinInformation{
		Sequence: fasta,
	}
	if err := database.Database.Create(&mainInfo).Error; err != nil {
		return SuperimposeResult{Error: "DB error."}
	}
	// 移动pdb文件
	modelPath := fmt.Sprintf("static/models/%d.pdb", mainInfo.ID)
	if err := exec.Command("mv", path, modelPath).Run(); err != nil {
		return SuperimposeResult{Error: "Move pdb error."}
	}
	// 生成图片
	imgPath := fmt.Sprintf("static/imgs/%d.png", mainInfo.ID)
	if err := exec.Command("bash", "-c", "source /root/miniconda3/etc/profile.d/conda.sh && conda activate alphafold && python3 py-scripts/pdb2img.py", modelPath, imgPath).Run(); err != nil {
		return SuperimposeResult{Error: "pdb2img error."}
	}
	// 创建主任务
	mainTask := models.Task{
		Title:    title,
		Sequence: fasta,
		Type:     3, // single
		UserId:   userId,
		ModelId:  fmt.Sprintf("%d", mainInfo.ID),
		// 其他字段可补充
	}
	if err := database.Database.Create(&mainTask).Error; err != nil {
		return SuperimposeResult{Error: "DB error."}
	}
	return SuperimposeResult{ID: mainTask.ID}
}

// 调用python脚本将pdb转为fasta
func pdb2fasta(pdbPath string) (string, error) {
	out, err := exec.Command("python3", "-c", "source /root/miniconda3/etc/profile.d/conda.sh && conda activate alphafold && py-scripts/pdb2fasta.py", pdbPath).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
