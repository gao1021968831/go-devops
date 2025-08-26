package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"go-devops/internal/logger"
	"go-devops/internal/models"
	"go-devops/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// JobExecutionHandler 作业执行处理器
type JobExecutionHandler struct {
	db               *gorm.DB
	executionService *services.ExecutionService
	activityService  *services.ActivityService
}

// NewJobExecutionHandler 创建作业执行处理器
func NewJobExecutionHandler(db *gorm.DB) *JobExecutionHandler {
	return &JobExecutionHandler{
		db:               db,
		executionService: services.NewExecutionService(db),
		activityService:  services.NewActivityService(db),
	}
}

// ExecuteJob 执行作业
func (h *JobExecutionHandler) ExecuteJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	// 获取作业信息
	var job models.Job
	if err := h.db.Preload("Script").First(&job, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业信息失败"})
		}
		return
	}

	// 解析主机ID列表
	var hostIDs []uint
	if err := json.Unmarshal([]byte(job.HostIDs), &hostIDs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "主机配置错误"})
		return
	}

	if len(hostIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未配置执行主机"})
		return
	}

	// 获取主机信息
	var hosts []models.Host
	if err := h.db.Where("id IN ?", hostIDs).Find(&hosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取主机信息失败"})
		return
	}

	if len(hosts) != len(hostIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "部分主机不存在"})
		return
	}

	// 更新作业状态为运行中
	job.Status = "running"
	h.db.Save(&job)

	executedBy := c.GetUint("user_id")
	var executions []models.JobExecution

	// 为每个主机创建执行记录并启动执行
	for _, host := range hosts {
		jobID := uint(id)
		execution, err := h.executionService.CreateJobExecution(
			&jobID,
			host.ID,
			job.Script.Content,
			job.Script.Type,
			false, // 不是快速执行
			executedBy,
			job.Name,        // 作业名称
			job.Script.Name, // 脚本名称
		)
		if err != nil {
			logger.Logger.WithFields(map[string]interface{}{
				"job_id":  job.ID,
				"host_id": host.ID,
				"error":   err.Error(),
			}).Error("创建执行记录失败")
			continue
		}

		executions = append(executions, *execution)

		// 获取输入文件（如果有）
		var inputFiles []models.File
		if job.InputFileIDs != "" {
			var fileIDs []uint
			if err := json.Unmarshal([]byte(job.InputFileIDs), &fileIDs); err == nil {
				h.db.Where("id IN ?", fileIDs).Find(&inputFiles)
			}
		}

		// 启动异步执行（支持文件功能）
		go h.executionService.ExecuteScriptOnHostWithOptions(
			execution,
			&job.Script,
			&host,
			inputFiles,
			job.SaveOutput,
			job.SaveError,
			job.OutputCategory,
		)
	}

	// 启动作业状态监控
	go h.monitorJobCompletion(job.ID)

	// 记录作业执行活动
	userID := c.GetUint("user_id")
	h.activityService.LogSuccess(c, userID, "execute", "job", &job.ID,
		fmt.Sprintf("执行作业 '%s' 在 %d 台主机上", job.Name, len(hosts)))

	c.JSON(http.StatusOK, gin.H{
		"message":    "作业执行已启动",
		"job_id":     job.ID,
		"executions": executions,
	})
}

// SaveExecutionResult 保存执行结果为文件
func (h *JobExecutionHandler) SaveExecutionResult(c *gin.Context) {
	var request models.SaveExecutionResultRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 获取执行记录
	var execution models.JobExecution
	if err := h.db.First(&execution, request.ExecutionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "执行记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取执行记录失败"})
		}
		return
	}

	// 检查权限
	userID := c.GetUint("user_id")
	if execution.ExecutedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限操作此执行记录"})
		return
	}

	// 保存执行结果为文件
	category := request.OutputCategory
	if category == "" {
		category = "script_output"
	}

	// 直接通过执行服务访问脚本执行器
	err := h.executionService.SaveExecutionResultAsFile(&execution, execution.Output, execution.Error, category, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("保存文件失败: %v", err)})
		return
	}

	// 记录活动
	h.activityService.LogSuccess(c, userID, "save_result", "job_execution", &execution.ID,
		fmt.Sprintf("保存执行结果为文件 - 执行ID: %d", execution.ID))

	c.JSON(http.StatusOK, gin.H{
		"message":   "执行结果已保存为文件",
		"execution": execution,
	})
}

// QuickExecuteScript 快速执行脚本
func (h *JobExecutionHandler) QuickExecuteScript(c *gin.Context) {
	var request struct {
		Name          string `json:"name" binding:"required"`
		ScriptContent string `json:"script_content" binding:"required"`
		ScriptType    string `json:"script_type" binding:"required"`
		HostIDs       []uint `json:"host_ids" binding:"required,min=1"`
		InputFileIDs  []uint `json:"input_file_ids"`
		Description   string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取主机信息
	var hosts []models.Host
	if err := h.db.Where("id IN ?", request.HostIDs).Find(&hosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取主机信息失败"})
		return
	}

	if len(hosts) != len(request.HostIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "部分主机不存在"})
		return
	}

	// 获取输入文件信息
	var inputFiles []models.File
	if len(request.InputFileIDs) > 0 {
		logger.Logger.WithFields(map[string]interface{}{
			"input_file_ids": request.InputFileIDs,
			"count":          len(request.InputFileIDs),
		}).Info("快速执行：开始获取输入文件")
		
		if err := h.db.Where("id IN ?", request.InputFileIDs).Find(&inputFiles).Error; err != nil {
			logger.Logger.WithFields(map[string]interface{}{
				"input_file_ids": request.InputFileIDs,
				"error":          err.Error(),
			}).Error("快速执行：获取输入文件失败")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取输入文件失败"})
			return
		}
		
		logger.Logger.WithFields(map[string]interface{}{
			"requested_count": len(request.InputFileIDs),
			"found_count":     len(inputFiles),
			"files":           inputFiles,
		}).Info("快速执行：输入文件查询结果")
		
		if len(inputFiles) != len(request.InputFileIDs) {
			logger.Logger.WithFields(map[string]interface{}{
				"requested_ids": request.InputFileIDs,
				"found_files":   inputFiles,
			}).Error("快速执行：部分输入文件不存在")
			c.JSON(http.StatusBadRequest, gin.H{"error": "部分输入文件不存在"})
			return
		}
		
		// 详细记录每个文件信息
		for _, file := range inputFiles {
			logger.Logger.WithFields(map[string]interface{}{
				"file_id":   file.ID,
				"file_name": file.Name,
				"file_path": file.Path,
				"file_size": file.Size,
			}).Info("快速执行：输入文件详情")
		}
	} else {
		logger.Logger.Info("快速执行：未选择输入文件")
	}

	executedBy := c.GetUint("user_id")
	var executions []models.JobExecution

	// 创建临时脚本对象
	script := &models.Script{
		Name:    request.Name,
		Type:    request.ScriptType,
		Content: request.ScriptContent,
	}

	// 为每个主机创建执行记录并启动执行
	for _, host := range hosts {
		execution, err := h.executionService.CreateJobExecution(
			nil, // 快速执行没有关联的作业ID
			host.ID,
			request.ScriptContent,
			request.ScriptType,
			true, // 标记为快速执行
			executedBy,
			request.Name, // 快速执行的任务名称
			"快速执行脚本",     // 脚本名称
		)
		if err != nil {
			logger.Logger.WithFields(map[string]interface{}{
				"host_id": host.ID,
				"error":   err.Error(),
			}).Error("创建快速执行记录失败")
			continue
		}

		executions = append(executions, *execution)

		// 启动异步执行，传递输入文件
		go h.executionService.ExecuteScriptOnHostWithOptions(execution, script, &host, inputFiles, false, false, "")
	}

	// 记录快速执行日志
	logger.Logger.WithFields(map[string]interface{}{
		"action":          "quick_execute_script",
		"resource":        "scripts",
		"script_type":     request.ScriptType,
		"host_count":      len(hosts),
		"execution_count": len(executions),
		"executed_by":     executedBy,
	}).Info("快速脚本执行已启动")

	// 记录快速执行活动
	h.activityService.LogSuccess(c, executedBy, "execute", "script", nil,
		fmt.Sprintf("快速执行脚本 '%s' 在 %d 台主机上", request.Name, len(hosts)))

	c.JSON(http.StatusOK, gin.H{
		"message":    "快速执行已启动",
		"executions": executions,
	})
}

// GetJobExecutions 获取作业执行记录
func (h *JobExecutionHandler) GetJobExecutions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	var executions []models.JobExecution
	if err := h.db.Preload("Host").Where("job_id = ?", id).
		Order("created_at DESC").Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取执行记录失败"})
		return
	}

	c.JSON(http.StatusOK, executions)
}

// GetAllExecutions 获取所有执行记录
func (h *JobExecutionHandler) GetAllExecutions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	var executions []models.JobExecution
	var total int64

	query := h.db.Model(&models.JobExecution{})

	// 搜索过滤
	if search := c.Query("search"); search != "" {
		query = query.Joins("LEFT JOIN hosts ON job_executions.host_id = hosts.id").
			Where("hosts.name LIKE ? OR job_executions.script_content LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 状态过滤
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 脚本类型过滤
	if scriptType := c.Query("script_type"); scriptType != "" {
		query = query.Where("script_type = ?", scriptType)
	}

	// 获取总数
	query.Count(&total)

	// 获取数据
	if err := query.Preload("Host").Preload("Job").Preload("ExecutedUser").
		Offset(offset).Limit(size).
		Order("created_at DESC").
		Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取执行记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  executions,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetExecutionDetail 获取执行详情
func (h *JobExecutionHandler) GetExecutionDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的执行ID"})
		return
	}

	var execution models.JobExecution
	if err := h.db.Preload("Host").Preload("Job").Preload("Job.Script").Preload("ExecutedUser").
		First(&execution, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "执行记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取执行详情失败"})
		}
		return
	}

	c.JSON(http.StatusOK, execution)
}

// DeleteJobExecution 删除作业执行记录（仅限admin）
func (h *JobExecutionHandler) DeleteJobExecution(c *gin.Context) {
	// 检查admin权限
	userRole := c.GetString("role")
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有管理员可以删除作业执行记录"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的执行记录ID"})
		return
	}

	// 获取执行记录信息
	var execution models.JobExecution
	if err := h.db.Preload("Host").Preload("Job").First(&execution, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "执行记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取执行记录失败"})
		}
		return
	}

	// 删除关联的输出文件（如果存在）
	if execution.OutputFileID != nil {
		var outputFile models.File
		if err := h.db.First(&outputFile, *execution.OutputFileID).Error; err == nil {
			// 删除物理文件
			if err := os.Remove(outputFile.Path); err != nil && !os.IsNotExist(err) {
				logger.Warnf("删除输出文件失败: %v", err)
			}
			// 删除文件记录
			h.db.Delete(&outputFile)
		}
	}

	// 删除执行记录
	if err := h.db.Delete(&execution).Error; err != nil {
		logger.Errorf("删除执行记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除执行记录失败"})
		return
	}

	// 记录删除活动
	userID := c.GetUint("user_id")
	description := fmt.Sprintf("删除作业执行记录 - ID: %d", execution.ID)
	if execution.Job != nil {
		description += fmt.Sprintf(", 作业: %s", execution.Job.Name)
	}
	if execution.Host.ID != 0 {
		description += fmt.Sprintf(", 主机: %s", execution.Host.Name)
	}
	
	h.activityService.LogSuccess(c, userID, "delete", "job_execution", &execution.ID, description)

	logger.Infof("管理员删除执行记录成功 - ID: %d, 操作者: %d", execution.ID, userID)

	c.JSON(http.StatusOK, gin.H{
		"message": "执行记录删除成功",
	})
}

// BatchDeleteJobExecutions 批量删除作业执行记录（仅限admin）
func (h *JobExecutionHandler) BatchDeleteJobExecutions(c *gin.Context) {
	// 检查admin权限
	userRole := c.GetString("role")
	if userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "只有管理员可以删除作业执行记录"})
		return
	}

	var req struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取要删除的执行记录
	var executions []models.JobExecution
	if err := h.db.Preload("Host").Preload("Job").Where("id IN ?", req.IDs).Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取执行记录失败"})
		return
	}

	if len(executions) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到要删除的执行记录"})
		return
	}

	// 删除关联的输出文件
	var outputFileIDs []uint
	for _, execution := range executions {
		if execution.OutputFileID != nil {
			outputFileIDs = append(outputFileIDs, *execution.OutputFileID)
		}
	}

	if len(outputFileIDs) > 0 {
		var outputFiles []models.File
		if err := h.db.Where("id IN ?", outputFileIDs).Find(&outputFiles).Error; err == nil {
			// 删除物理文件
			for _, file := range outputFiles {
				if err := os.Remove(file.Path); err != nil && !os.IsNotExist(err) {
					logger.Warnf("删除输出文件失败: %v", err)
				}
			}
			// 删除文件记录
			h.db.Where("id IN ?", outputFileIDs).Delete(&models.File{})
		}
	}

	// 批量删除执行记录
	if err := h.db.Where("id IN ?", req.IDs).Delete(&models.JobExecution{}).Error; err != nil {
		logger.Errorf("批量删除执行记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除执行记录失败"})
		return
	}

	// 记录批量删除活动
	userID := c.GetUint("user_id")
	for _, execution := range executions {
		description := fmt.Sprintf("批量删除作业执行记录 - ID: %d", execution.ID)
		if execution.Job != nil {
			description += fmt.Sprintf(", 作业: %s", execution.Job.Name)
		}
		if execution.Host.ID != 0 {
			description += fmt.Sprintf(", 主机: %s", execution.Host.Name)
		}
		
		h.activityService.LogSuccess(c, userID, "delete", "job_execution", &execution.ID, description)
	}

	logger.Infof("管理员批量删除执行记录成功 - 数量: %d, 操作者: %d", len(executions), userID)

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("成功删除 %d 条执行记录", len(executions)),
	})
}

// monitorJobCompletion 监控作业完成状态
func (h *JobExecutionHandler) monitorJobCompletion(jobID uint) {
	// 这里可以实现作业完成状态的监控逻辑
	// 例如定期检查所有执行是否完成，然后更新作业状态
	go func() {
		// 简单的状态更新，实际可以用更复杂的监控机制
		if err := h.executionService.UpdateJobStatus(jobID); err != nil {
			logger.Logger.WithFields(map[string]interface{}{
				"job_id": jobID,
				"error":  err.Error(),
			}).Error("更新作业状态失败")
		}
	}()
}
