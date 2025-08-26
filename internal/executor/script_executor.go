package executor

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-devops/internal/models"
	"go-devops/internal/ssh"
	"gorm.io/gorm"
)

// ScriptExecutor 脚本执行器
type ScriptExecutor struct{
	db *gorm.DB
}

// NewScriptExecutor 创建脚本执行器实例
func NewScriptExecutor(db *gorm.DB) *ScriptExecutor {
	return &ScriptExecutor{db: db}
}

// ExecuteScript 执行脚本的统一入口
func (e *ScriptExecutor) ExecuteScript(host *models.Host, script *models.Script) (string, string, error) {
	return e.ExecuteScriptWithFiles(host, script, nil)
}

// ExecuteScriptWithFiles 执行脚本并传递输入文件
func (e *ScriptExecutor) ExecuteScriptWithFiles(host *models.Host, script *models.Script, inputFiles []models.File) (string, string, error) {
	switch script.Type {
	case "shell":
		return e.executeShellScriptWithFiles(host, script, inputFiles)
	case "python2":
		return e.executePython2ScriptWithFiles(host, script, inputFiles)
	case "python3":
		return e.executePython3ScriptWithFiles(host, script, inputFiles)
	default:
		return "", "", fmt.Errorf("不支持的脚本类型: %s", script.Type)
	}
}

// Shell脚本执行 - 优化后的实现
func (e *ScriptExecutor) executeShellScript(host *models.Host, script *models.Script) (string, string, error) {
	// 直接执行原始脚本内容，不需要额外包装
	return ssh.ExecuteScript(host, script)
}

// Shell脚本执行（带文件）
func (e *ScriptExecutor) executeShellScriptWithFiles(host *models.Host, script *models.Script, inputFiles []models.File) (string, string, error) {
	return ssh.ExecuteScriptWithFiles(host, script, inputFiles)
}

// Python2脚本执行
func (e *ScriptExecutor) executePython2Script(host *models.Host, script *models.Script) (string, string, error) {
	// 直接传递给SSH模块处理，避免双重包装
	return ssh.ExecuteScript(host, script)
}

// Python2脚本执行（带文件）
func (e *ScriptExecutor) executePython2ScriptWithFiles(host *models.Host, script *models.Script, inputFiles []models.File) (string, string, error) {
	return ssh.ExecuteScriptWithFiles(host, script, inputFiles)
}

// Python3脚本执行
func (e *ScriptExecutor) executePython3Script(host *models.Host, script *models.Script) (string, string, error) {
	// 直接传递给SSH模块处理，避免双重包装
	return ssh.ExecuteScript(host, script)
}

// Python3脚本执行（带文件）
func (e *ScriptExecutor) executePython3ScriptWithFiles(host *models.Host, script *models.Script, inputFiles []models.File) (string, string, error) {
	return ssh.ExecuteScriptWithFiles(host, script, inputFiles)
}

// SaveExecutionResultAsFile 将执行结果保存为文件
func (e *ScriptExecutor) SaveExecutionResultAsFile(execution *models.JobExecution, output, errorOutput string, category string, userID uint) error {
	uploadPath := "uploads"
	
	// 生成包含作业名称和时间的文件名前缀
	timeStr := time.Now().Format("20060102_150405")
	jobName := execution.JobName
	if jobName == "" {
		jobName = execution.ScriptName
	}
	// 清理文件名中的特殊字符
	jobName = strings.ReplaceAll(jobName, " ", "_")
	jobName = strings.ReplaceAll(jobName, "/", "_")
	jobName = strings.ReplaceAll(jobName, "\\", "_")
	jobName = strings.ReplaceAll(jobName, ":", "_")
	
	// 保存输出文件
	if output != "" && len(strings.TrimSpace(output)) > 0 {
		outputFilename := fmt.Sprintf("%s_%s_output.txt", jobName, timeStr)
		outputFile, err := e.saveContentAsFile(
			outputFilename,
			output,
			category,
			fmt.Sprintf("脚本执行输出 - %s (%s)", jobName, timeStr),
			userID,
			uploadPath,
		)
		if err == nil {
			execution.OutputFileID = &outputFile.ID
		}
	}
	
	// 保存错误日志文件
	if errorOutput != "" && len(strings.TrimSpace(errorOutput)) > 0 {
		errorFilename := fmt.Sprintf("%s_%s_error.log", jobName, timeStr)
		errorFile, err := e.saveContentAsFile(
			errorFilename,
			errorOutput,
			"error_log",
			fmt.Sprintf("脚本执行错误日志 - %s (%s)", jobName, timeStr),
			userID,
			uploadPath,
		)
		if err == nil {
			execution.ErrorFileID = &errorFile.ID
		}
	}
	
	// 更新执行记录
	return e.db.Save(execution).Error
}

// saveContentAsFile 将内容保存为文件
func (e *ScriptExecutor) saveContentAsFile(filename, content, category, description string, userID uint, uploadPath string) (*models.File, error) {
	// 确保上传目录存在
	categoryPath := filepath.Join(uploadPath, category)
	if err := os.MkdirAll(categoryPath, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %v", err)
	}
	
	// 生成唯一文件名
	timestamp := time.Now().UnixNano()
	uniqueFilename := fmt.Sprintf("%d_%s", timestamp, filename)
	filePath := filepath.Join(categoryPath, uniqueFilename)
	
	// 写入文件内容
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %v", err)
	}
	
	// 计算MD5哈希
	hash := md5.Sum([]byte(content))
	md5Hash := fmt.Sprintf("%x", hash)
	
	// 创建文件记录
	file := &models.File{
		Name:          uniqueFilename,
		OriginalName:  filename,
		Path:          filePath,
		Size:          int64(len(content)),
		MimeType:      "text/plain",
		MD5Hash:       md5Hash,
		Category:      category,
		Description:   description,
		IsPublic:      false,
		UploadedBy:    userID,
		DownloadCount: 0,
	}
	
	if err := e.db.Create(file).Error; err != nil {
		// 如果数据库操作失败，删除已创建的文件
		os.Remove(filePath)
		return nil, fmt.Errorf("创建文件记录失败: %v", err)
	}
	
	return file, nil
}

