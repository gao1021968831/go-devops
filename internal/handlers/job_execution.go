package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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
				"job_id":   job.ID,
				"host_id":  host.ID,
				"error":    err.Error(),
			}).Error("创建执行记录失败")
			continue
		}

		executions = append(executions, *execution)

		// 启动异步执行
		go h.executionService.ExecuteScriptOnHost(execution, &job.Script, &host)
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

// QuickExecuteScript 快速执行脚本
func (h *JobExecutionHandler) QuickExecuteScript(c *gin.Context) {
	var request struct {
		Name          string `json:"name" binding:"required"`
		ScriptContent string `json:"script_content" binding:"required"`
		ScriptType    string `json:"script_type" binding:"required"`
		HostIDs       []uint `json:"host_ids" binding:"required,min=1"`
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
			"快速执行脚本",      // 脚本名称
		)
		if err != nil {
			logger.Logger.WithFields(map[string]interface{}{
				"host_id": host.ID,
				"error":   err.Error(),
			}).Error("创建快速执行记录失败")
			continue
		}

		executions = append(executions, *execution)

		// 启动异步执行
		go h.executionService.ExecuteScriptOnHost(execution, script, &host)
	}

	// 记录快速执行日志
	logger.Logger.WithFields(map[string]interface{}{
		"action":        "quick_execute_script",
		"resource":      "scripts",
		"script_type":   request.ScriptType,
		"host_count":    len(hosts),
		"execution_count": len(executions),
		"executed_by":   executedBy,
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
