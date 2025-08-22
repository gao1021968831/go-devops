package services

import (
	"fmt"
	"time"

	"go-devops/internal/executor"
	"go-devops/internal/logger"
	"go-devops/internal/models"

	"gorm.io/gorm"
)

// ExecutionService 执行服务
type ExecutionService struct {
	db       *gorm.DB
	executor *executor.ScriptExecutor
}

// NewExecutionService 创建执行服务实例
func NewExecutionService(db *gorm.DB) *ExecutionService {
	return &ExecutionService{
		db:       db,
		executor: executor.NewScriptExecutor(),
	}
}

// ExecuteScriptOnHost 在指定主机上执行脚本的统一方法
func (s *ExecutionService) ExecuteScriptOnHost(execution *models.JobExecution, script *models.Script, host *models.Host) {
	logger.Logger.WithFields(map[string]interface{}{
		"execution_id": execution.ID,
		"host_id":      host.ID,
		"host_name":    host.Name,
		"script_type":  script.Type,
		"script_name":  script.Name,
	}).Info("开始执行脚本")

	// 更新执行状态为运行中
	execution.Status = "running"
	execution.StartTime = time.Now()
	s.db.Save(execution)

	// 检查SSH连接
	if host.AuthType == "" {
		s.handleExecutionError(execution, "主机认证信息不完整")
		return
	}

	// 执行脚本
	output, errorOutput, err := s.executor.ExecuteScript(host, script)

	// 执行时长会在前端计算显示
	// 处理执行结果
	if err != nil {
		s.handleExecutionError(execution, fmt.Sprintf("执行失败: %v", err))
		if errorOutput != "" {
			execution.Error = errorOutput
		}
	} else {
		execution.Status = "completed"
		execution.Output = output
		if errorOutput != "" {
			execution.Error = errorOutput
		}
	}

	// 保存执行结果
	s.db.Save(execution)

	// 记录执行完成日志
	status := execution.Status
	logger.Logger.WithFields(map[string]interface{}{
		"execution_id": execution.ID,
		"host_id":      execution.HostID,
		"status":       status,
		"success":      status == "completed",
	}).Info("脚本执行完成")
}

// handleExecutionError 处理执行错误
func (s *ExecutionService) handleExecutionError(execution *models.JobExecution, errorMsg string) {
	execution.Status = "failed"
	execution.Error = errorMsg
	endTime := time.Now()
	execution.EndTime = &endTime
	
	// 执行时长会在前端计算显示

	logger.Logger.WithFields(map[string]interface{}{
		"execution_id": execution.ID,
		"error":        errorMsg,
	}).Error("脚本执行失败")
}

// CreateJobExecution 创建作业执行记录
func (s *ExecutionService) CreateJobExecution(jobID uint, hostID uint, scriptContent, scriptType string, isQuickExec bool, executedBy uint, jobName, scriptName string) (*models.JobExecution, error) {
	execution := &models.JobExecution{
		JobID:         jobID,
		HostID:        hostID,
		Status:        "pending",
		ScriptContent: scriptContent,
		ScriptType:    scriptType,
		IsQuickExec:   isQuickExec,
		ExecutedBy:    executedBy,
		JobName:       jobName,
		ScriptName:    scriptName,
	}

	if err := s.db.Create(execution).Error; err != nil {
		return nil, fmt.Errorf("创建执行记录失败: %v", err)
	}

	return execution, nil
}

// UpdateJobStatus 更新作业状态
func (s *ExecutionService) UpdateJobStatus(jobID uint) error {
	var executions []models.JobExecution
	if err := s.db.Where("job_id = ?", jobID).Find(&executions).Error; err != nil {
		return err
	}

	if len(executions) == 0 {
		return nil
	}

	// 统计执行状态
	var completedCount, failedCount, runningCount int
	for _, exec := range executions {
		switch exec.Status {
		case "completed":
			completedCount++
		case "failed":
			failedCount++
		case "running":
			runningCount++
		}
	}

	// 确定作业整体状态
	var jobStatus string
	if runningCount > 0 {
		jobStatus = "running"
	} else if failedCount > 0 {
		if completedCount > 0 {
			jobStatus = "partial_failed"
		} else {
			jobStatus = "failed"
		}
	} else {
		jobStatus = "completed"
	}

	// 更新作业状态
	return s.db.Model(&models.Job{}).Where("id = ?", jobID).Update("status", jobStatus).Error
}
