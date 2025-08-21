package services

import (
	"Protein_Server/logger"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// BLAST Processing
func BlastProcessing(sequence string) ([]string, []map[string]string) {
	// Create a fasta.txt file
	file, err := os.Create("fasta.txt")
	if err != nil {
		logger.Error("创建文件失败: %v", err)
		return nil, nil
	}
	// Write sequence to fasta.txt file
	_, err = file.WriteString(sequence)
	if err != nil {
		logger.Error("写入文件失败: %v", err)
		return nil, nil
	}
	file.Close()

	// rpsblast - used to calculate CD-Search (Conserved Domain Search)
	// -db - blast/bin/db/Cdd database, use Cdd(Conserved Domain database)
	// -outfmt - 11 is asn format
	// cmd := "../RpsbProc-x64-linux/rpsblast -query fasta.txt -db ../RpsbProc-x64-linux/db/Cdd -evalue 0.01 -outfmt 11 -out fasta.asn"
	// err = exec.Command(cmd).Run()
	rpsblastPath := "../RpsbProc-x64-linux/rpsblast"
	err = exec.Command(rpsblastPath,
		"-query", "fasta.txt",
		"-db", "../RpsbProc-x64-linux/db/Cdd",
		"-evalue", "0.01",
		"-outfmt", "11",
		"-out", "fasta.asn").Run()
	if err != nil {
		logger.Error("运行rpsblast失败: %v", err)
		return nil, nil
	}

	// rpsbproc is used to process rpsblast results
	// -i the input
	// -o the output
	// -e the evalue
	// -m the result mode, std is the standard result, rep is the concise result, full is the full result
	// -t doms only needs domains
	// cmd = "../RpsbProc-x64-linux/rpsbproc -i fasta.asn -o fasta.out -e 0.01 -m std -t doms"
	// err = exec.Command(cmd).Run()
	rpsbprocPath := "../RpsbProc-x64-linux/rpsbproc"
	err = exec.Command(rpsbprocPath,
		"-i", "fasta.asn",
		"-o", "fasta.out",
		"-e", "0.01",
		"-m", "std",
		"-t", "doms").Run()
	if err != nil {
		logger.Error("运行rpsbproc失败: %v", err)
		return nil, nil
	}

	results, err := parseFastaResult()
	if err != nil {
		logger.Error("解析 fasta 结果失败: %v", err)
		return nil, nil
	}

	if len(results) == 0 {
		logger.Error("没有找到有效的子序列")
		return nil, nil
	}

	var subSequences []string

	// 在 parseFastaResult 函数中，先打印一下结果内容
	logger.Info("解析到的结果: %+v", results)

	// 修改处理子序列的代码
	for _, item := range results {
		// 打印每个 item 的内容，看看实际的数据结构
		logger.Info("处理的 item: %+v", item)

		// 检查 From 和 To 字段是否存在且非空
		// 尝试多种可能的字段名
		var fromStr, toStr string
		var fromExists, toExists bool
		
		// 尝试常见的字段名变体
		if val, exists := item["from"]; exists {
			fromStr, fromExists = val, exists
		} else if val, exists := item["From"]; exists {
			fromStr, fromExists = val, exists
		} else if val, exists := item["FROM"]; exists {
			fromStr, fromExists = val, exists
		}
		
		if val, exists := item["to"]; exists {
			toStr, toExists = val, exists
		} else if val, exists := item["To"]; exists {
			toStr, toExists = val, exists
		} else if val, exists := item["TO"]; exists {
			toStr, toExists = val, exists
		}

		if !fromExists || !toExists {
			logger.Error("缺少必要的 From 或 To 字段: %+v", item)
			logger.Error("可用的字段名: %+v", getMapKeys(item))
			continue
		}

		if fromStr == "" || toStr == "" {
			logger.Error("From 或 To 字段为空: From=%s, To=%s", fromStr, toStr)
			continue
		}

		from, err := strconv.Atoi(fromStr)
		if err != nil {
			logger.Error("解析 From 字段失败: %v, 原始值: %s", err, fromStr)
			continue
		}

		to, err := strconv.Atoi(toStr)
		if err != nil {
			logger.Error("解析 To 字段失败: %v, 原始值: %s", err, toStr)
			continue
		}

		// 检查序列范围是否有效
		if from <= 0 || to > len(sequence) || from > to {
			logger.Error("无效的序列范围: From=%d, To=%d, 序列长度=%d", from, to, len(sequence))
			continue
		}

		subSequence := sequence[from-1 : to]
		subSequences = append(subSequences, subSequence)
	}

	return subSequences, results
}

func parseFastaResult() ([]map[string]string, error) {
	file, err := os.Open("fasta.out")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var result []map[string]string
	var keys []string
	readKey := false
	readData := false

	logger.Info("=== 开始解析 fasta.out 文件 ===")

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行
		if line == "" {
			continue
		}

		// 当遇到 #DOMAINS 时，下一行将是字段名
		if line == "#DOMAINS" {
			readKey = true
			continue
		}

		// 读取字段名行（紧跟在 #DOMAINS 后的行）
		if readKey {
			// 去掉开头的 # 并按 tab 分割，然后去掉每个字段的 < >
			keyLine := strings.TrimPrefix(line, "#")
			rawKeys := strings.Split(keyLine, "\t")
			keys = make([]string, len(rawKeys))
			for i, key := range rawKeys {
				// 去掉字段名两边的 < > 和空格
				trimmed := strings.TrimSpace(key)
				if strings.HasPrefix(trimmed, "<") && strings.HasSuffix(trimmed, ">") {
					keys[i] = trimmed[1 : len(trimmed)-1]
				} else {
					keys[i] = trimmed
				}
			}
			readKey = false
			continue
		}

		// 跳过其他注释行
		if strings.HasPrefix(line, "#") {
			continue
		}

		// 当遇到 DOMAINS 时开始读取数据
		if line == "DOMAINS" {
			readData = true
			continue
		}

		// 当遇到 ENDDOMAINS 时停止读取数据
		if line == "ENDDOMAINS" {
			readData = false
			continue
		}

		// 读取数据行
		if readData {
			values := strings.Split(line, "\t")
			if len(values) >= len(keys) && len(keys) > 0 {
				element := make(map[string]string)
				// 按字段名映射值
				for i, key := range keys {
					if i < len(values) {
						element[key] = strings.TrimSpace(values[i])
						logger.Info("映射: %s = %s", key, strings.TrimSpace(values[i]))
					}
				}
				result = append(result, element)
			} else {
				logger.Error("数据行字段数量不足或字段名未解析: 需要%d个字段，实际%d个，字段名数量%d", len(keys), len(values), len(keys))
			}
		}

		if line == "ENDDATA" {
			if len(result) == 0 {
				return nil, fmt.Errorf("code: 205, message: Wrong Sequence")
			}
			if len(result) > 5 {
				return result[:5], nil
			}
			return result, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	logger.Info("=== 解析 fasta.out 文件结束 ===")
	return nil, nil
}

// getMapKeys 获取 map 的所有键名，用于调试
func getMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
