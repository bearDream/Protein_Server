package services

import (
	"Protein_Server/logger"
	"fmt"
	"os"
	"os/exec"
)

func Ramachandran(protein_id string) {
	// 确保输出目录存在
	outputDir := "static/imgs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("创建Ramachandran输出目录失败: %v", err)
		return
	}

	// 构建输入和输出路径
	inputPath := fmt.Sprintf("./static/models/%s.pdb", protein_id)
	outputPath := fmt.Sprintf("./static/imgs/%s.png", protein_id)

	// 构建命令字符串，使用正确的字符串格式化
	cmdStr := fmt.Sprintf("source /root/miniconda3/etc/profile.d/conda.sh && conda activate alphafold && python py-scripts/pdb2img.py %s %s", inputPath, outputPath)

	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("运行Ramachandran失败: %v, 输出: %s", err, output)
	} else {
		logger.Info("成功生成Ramachandran图: %s.png", protein_id)
	}
}
