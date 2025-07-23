package services

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"Protein_Server/logger"
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
	cmd := "../RpsbProc-x64-linux/rpsblast -query fasta.txt -db ../RpsbProc-x64-linux/db/Cdd -evalue 0.01 -outfmt 11 -out fasta.asn"
	err = exec.Command(cmd).Run()
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
	cmd = "../RpsbProc-x64-linux/rpsbproc -i fasta.asn -o fasta.out -e 0.01 -m std -t doms"
	err = exec.Command(cmd).Run()
	if err != nil {
		logger.Error("运行rpsbproc失败: %v", err)
		return nil, nil
	}

	results, err := parseFastaResult()
	if err != nil {
		logger.Error("解析rpsbproc结果失败: %v", err)
		return nil, nil
	}

	var subSequences []string

	// split sub-sequence by from and to
	for _, item := range results {
		from, err := strconv.Atoi(item["From"])
		if err != nil {
			logger.Error("解析子序列失败: %v", err)
			continue
		}
		to, err := strconv.Atoi(item["To"])
		if err != nil {
			logger.Error("解析子序列失败: %v", err)
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

	// Scan the contents of the file into a buffer
	scanner := bufio.NewScanner(file)
	var result []map[string]string
	var keys []string
	readKey := false
	readData := false

	for scanner.Scan() {
		line := scanner.Text()
		if readKey {
			// This line is like <tag> \t <tag> \t <tag>
			keys = strings.Split(line[1:], "\t")
			for i, key := range keys {
				// remove <>
				keys[i] = key[1 : len(key)-1]
			}
		}
		if line == "#DOMAINS" {
			readKey = true
		} else {
			readKey = false
		}
		if line == "ENDDOMAINS" {
			readData = false
		}
		if readData {
			element := make(map[string]string)
			values := strings.Split(line, "\t")
			for i, value := range values {
				element[keys[i]] = value
			}
			result = append(result, element)
		}
		if line == "DOMAINS" {
			readData = true
		}
		if line == "ENDDATA" {
			if len(result) == 0 {
				return nil, fmt.Errorf("code: 205, message: Wrong Sequence")
			}
			// only top 5
			if len(result) > 5 {
				return result[:5], nil
			}
			return result, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return nil, nil
}
