package services

import (
	"Protein_Server/logger"
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

	cmd := exec.Command("bash", "-c", "source ~/anaconda3/etc/profile.d/conda.sh && conda activate alphafold && python py-scripts/ramachandran.py "+protein_id)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("运行Ramachandran失败: %v, 输出: %s", err, output)
	} else {
		logger.Info("成功生成Ramachandran图: %s.png", protein_id)
	}
}
