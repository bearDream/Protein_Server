package services

import (
	"Protein_Server/database"
	"Protein_Server/models"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
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

	// 1. 处理第一个PDB文件
	fasta1, err := pdb2fasta(path1)
	if err != nil {
		return SuperimposeResult{Error: "pdb2fasta error."}
	}

	// 2. 查找或创建第一个蛋白质信息记录
	var mainInfo models.ProteinInformation
	isNewMainInfo := false
	if err := database.Database.Where("sequence = ?", fasta1).First(&mainInfo).Error; err != nil {
		// 如果没找到，创建新记录
		mainInfo = models.ProteinInformation{
			Sequence: fasta1,
		}
		if err := database.Database.Create(&mainInfo).Error; err != nil {
			return SuperimposeResult{Error: "DB error."}
		}
		isNewMainInfo = true
	}

	// 3. 移动第一个PDB文件
	modelPath1 := fmt.Sprintf("static/models/%d.pdb", mainInfo.ID)
	if err := exec.Command("mv", path1, modelPath1).Run(); err != nil {
		return SuperimposeResult{Error: "Move pdb1 error."}
	}

	// 只有新记录才需要计算参数
	if isNewMainInfo {
		CalculateProteinInfomationWithPath(mainInfo)
	}

	// 4. 处理第二个PDB文件
	fasta2, err := pdb2fasta(path2)
	if err != nil {
		return SuperimposeResult{Error: "pdb2fasta2 error."}
	}

	// 5. 查找或创建第二个蛋白质信息记录（作为子序列）
	var subInfo models.ProteinInformation
	isNewSubInfo := false
	if err := database.Database.Where("sequence = ?", fasta2).First(&subInfo).Error; err != nil {
		// 如果没找到，创建新记录，设置ParentId为主序列的ID
		subInfo = models.ProteinInformation{
			Sequence: fasta2,
			ParentId: mainInfo.ID,
		}
		if err := database.Database.Create(&subInfo).Error; err != nil {
			return SuperimposeResult{Error: "DB error."}
		}
		isNewSubInfo = true
	}

	// 6. 移动第二个PDB文件
	modelPath2 := fmt.Sprintf("static/models/%d.pdb", subInfo.ID)
	if err := exec.Command("mv", path2, modelPath2).Run(); err != nil {
		return SuperimposeResult{Error: "Move pdb2 error."}
	}

	// 只有新记录才需要计算参数
	if isNewSubInfo {
		CalculateProteinInfomationWithPath(subInfo)
	}

	// 7. 创建主任务（只创建一条任务记录）
	mainTask := models.Task{
		Title:       title,
		Sequence:    fasta1,
		SubSequence: fasta2,
		Type:        4, // superimpose
		UserId:      userId,
		ModelId:     fmt.Sprintf("%d,%d", mainInfo.ID, subInfo.ID), // 包含两个蛋白质信息的ID
	}
	if err := database.Database.Create(&mainTask).Error; err != nil {
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
	// 查找或创建 protein_information
	var mainInfo models.ProteinInformation
	isNewInfo := false
	if err := database.Database.Where("sequence = ?", fasta).First(&mainInfo).Error; err != nil {
		// 如果没找到，创建新记录
		mainInfo = models.ProteinInformation{
			Sequence: fasta,
		}
		if err := database.Database.Create(&mainInfo).Error; err != nil {
			return SuperimposeResult{Error: "DB error."}
		}
		isNewInfo = true
	}
	// 移动pdb文件
	modelPath := fmt.Sprintf("static/models/%d.pdb", mainInfo.ID)
	if err := exec.Command("mv", path, modelPath).Run(); err != nil {
		return SuperimposeResult{Error: "Move pdb error."}
	}
	// 只有新记录才需要计算参数
	if isNewInfo {
		CalculateProteinInfomationWithPath(mainInfo)
	}

	// 创建主任务
	mainTask := models.Task{
		Title:    title,
		Sequence: fasta,
		Type:     3, // analysis
		UserId:   userId,
		ModelId:  fmt.Sprintf("%d", mainInfo.ID),
	}
	if err := database.Database.Create(&mainTask).Error; err != nil {
		return SuperimposeResult{Error: "DB error."}
	}
	return SuperimposeResult{ID: mainTask.ID}
}

// atomMap 氨基酸三字母代码到单字母代码的映射
var atomMap = map[string]string{
	"ALA": "A", "ARG": "R", "ASN": "N", "ASP": "D", "CYS": "C",
	"GLU": "E", "GLN": "Q", "GLY": "G", "HIS": "H", "ILE": "I",
	"LEU": "L", "LYS": "K", "MET": "M", "PHE": "F", "PRO": "P",
	"SER": "S", "THR": "T", "TRP": "W", "TYR": "Y", "VAL": "V",
	"SEC": "U", "PYL": "O", // 非标准氨基酸
}

// pdb2fasta 将pdb文件转换为fasta序列
func pdb2fasta(pdbPath string) (string, error) {
	file, err := os.Open(pdbPath)
	if err != nil {
		return "", fmt.Errorf("无法打开PDB文件: %v", err)
	}
	defer file.Close()

	var code strings.Builder
	lastIndex := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// 检查是否是ATOM行
		if strings.HasPrefix(line, "ATOM") && len(line) >= 26 {
			// 提取残基序号 (位置22-26)
			indexStr := strings.TrimSpace(line[22:26])
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				continue // 跳过无法解析序号的行
			}

			// 只处理新的残基序号
			if index > lastIndex && len(line) >= 20 {
				// 提取氨基酸三字母代码 (位置17-20)
				key := strings.TrimSpace(line[17:20])

				// 查找对应的单字母代码
				if singleLetter, exists := atomMap[key]; exists {
					code.WriteString(singleLetter)
					lastIndex = index
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("读取PDB文件时出错: %v", err)
	}

	return code.String(), nil
}
