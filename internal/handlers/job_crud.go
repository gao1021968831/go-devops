package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-devops/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// JobCRUDHandler 作业CRUD操作处理器
type JobCRUDHandler struct {
	db *gorm.DB
}

// NewJobCRUDHandler 创建作业CRUD处理器
func NewJobCRUDHandler(db *gorm.DB) *JobCRUDHandler {
	return &JobCRUDHandler{db: db}
}

// GetJobs 获取作业列表
func (h *JobCRUDHandler) GetJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	var jobs []models.Job
	var total int64

	query := h.db.Model(&models.Job{})
	
	// 搜索过滤
	if search := c.Query("search"); search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// 状态过滤
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	query.Count(&total)

	// 获取数据
	if err := query.Preload("Script").Preload("User").
		Offset(offset).Limit(size).
		Order("created_at DESC").
		Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  jobs,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// CreateJob 创建作业
func (h *JobCRUDHandler) CreateJob(c *gin.Context) {
	var job models.Job
	if err := c.ShouldBindJSON(&job); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 设置创建者
	job.CreatedBy = c.GetUint("user_id")

	if err := h.db.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建作业失败"})
		return
	}

	// 加载关联信息
	h.db.Preload("Script").Preload("User").First(&job, job.ID)

	c.JSON(http.StatusCreated, job)
}

// GetJob 获取单个作业
func (h *JobCRUDHandler) GetJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	var job models.Job
	if err := h.db.Preload("Script").Preload("User").First(&job, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业信息失败"})
		}
		return
	}

	c.JSON(http.StatusOK, job)
}

// UpdateJob 更新作业
func (h *JobCRUDHandler) UpdateJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	var job models.Job
	if err := h.db.First(&job, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业信息失败"})
		}
		return
	}

	var updateData models.Job
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 更新字段
	job.Name = updateData.Name
	job.ScriptID = updateData.ScriptID
	job.HostIDs = updateData.HostIDs

	if err := h.db.Save(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新作业失败"})
		return
	}

	// 加载关联信息
	h.db.Preload("Script").Preload("User").First(&job, job.ID)

	c.JSON(http.StatusOK, job)
}

// DeleteJob 删除作业
func (h *JobCRUDHandler) DeleteJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的作业ID"})
		return
	}

	// 检查作业是否存在
	var job models.Job
	if err := h.db.First(&job, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "作业不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取作业信息失败"})
		}
		return
	}

	// 检查是否有正在运行的执行
	var runningCount int64
	h.db.Model(&models.JobExecution{}).Where("job_id = ? AND status = ?", id, "running").Count(&runningCount)
	if runningCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "作业正在执行中，无法删除"})
		return
	}

	// 删除相关的执行记录
	h.db.Where("job_id = ?", id).Delete(&models.JobExecution{})

	// 删除作业
	if err := h.db.Delete(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除作业失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "作业删除成功"})
}

// BatchDeleteJobs 批量删除作业
func (h *JobCRUDHandler) BatchDeleteJobs(c *gin.Context) {
	var request struct {
		IDs []uint `json:"ids" binding:"required,min=1"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	userID := c.GetUint("user_id")
	userRole := c.GetString("role")

	// 查询要删除的作业
	var jobs []models.Job
	if err := h.db.Where("id IN ?", request.IDs).Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询作业失败"})
		return
	}

	// 检查权限和运行状态
	var deletableIDs []uint
	var skippedCount int
	var runningCount int
	
	for _, job := range jobs {
		// 检查权限
		if job.CreatedBy != userID && userRole != "admin" {
			skippedCount++
			continue
		}
		
		// 检查是否有正在运行的执行
		var execCount int64
		h.db.Model(&models.JobExecution{}).Where("job_id = ? AND status = ?", job.ID, "running").Count(&execCount)
		if execCount > 0 {
			runningCount++
			continue
		}
		
		deletableIDs = append(deletableIDs, job.ID)
	}

	if len(deletableIDs) == 0 {
		message := "没有可删除的作业"
		if skippedCount > 0 {
			message += fmt.Sprintf("（%d个无权限）", skippedCount)
		}
		if runningCount > 0 {
			message += fmt.Sprintf("（%d个正在运行）", runningCount)
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}

	// 删除相关的执行记录
	h.db.Where("job_id IN ?", deletableIDs).Delete(&models.JobExecution{})

	// 执行批量删除
	if err := h.db.Where("id IN ?", deletableIDs).Delete(&models.Job{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除作业失败"})
		return
	}

	message := fmt.Sprintf("成功删除 %d 个作业", len(deletableIDs))
	if skippedCount > 0 {
		message += fmt.Sprintf("，跳过 %d 个无权限作业", skippedCount)
	}
	if runningCount > 0 {
		message += fmt.Sprintf("，跳过 %d 个正在运行的作业", runningCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"deleted_count": len(deletableIDs),
		"skipped_count": skippedCount,
		"running_count": runningCount,
	})
}

// ExportJobs 导出作业为CSV
func (h *JobCRUDHandler) ExportJobs(c *gin.Context) {
	idsParam := c.Query("ids")
	var jobs []models.Job
	
	if idsParam != "" {
		// 导出指定作业
		var ids []uint
		if err := json.Unmarshal([]byte(idsParam), &ids); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "作业ID参数格式错误"})
			return
		}
		
		if err := h.db.Preload("Script").Preload("User").Where("id IN ?", ids).Find(&jobs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询作业失败"})
			return
		}
	} else {
		// 导出所有作业
		if err := h.db.Preload("Script").Preload("User").Find(&jobs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询作业失败"})
			return
		}
	}

	// 设置响应头
	filename := fmt.Sprintf("jobs_export_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// 创建CSV写入器
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入BOM以支持中文
	c.Writer.Write([]byte{0xEF, 0xBB, 0xBF})

	// 写入CSV头部
	headers := []string{"ID", "作业名称", "描述", "脚本ID", "脚本名称", "脚本类型", "主机ID列表", "参数", "超时时间", "状态", "创建者", "创建时间", "更新时间"}
	if err := writer.Write(headers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入CSV头部失败"})
		return
	}

	// 写入数据行
	for _, job := range jobs {
		// 处理参数中的换行符和引号
		parameters := strings.ReplaceAll(job.Parameters, "\n", "\\n")
		parameters = strings.ReplaceAll(parameters, "\r", "\\r")
		
		record := []string{
			fmt.Sprintf("%d", job.ID),
			job.Name,
			job.Description,
			fmt.Sprintf("%d", job.ScriptID),
			job.Script.Name,
			job.Script.Type,
			job.HostIDs,
			parameters,
			fmt.Sprintf("%d", job.Timeout),
			job.Status,
			job.User.Username,
			job.CreatedAt.Format("2006-01-02 15:04:05"),
			job.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		
		if err := writer.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "写入CSV数据失败"})
			return
		}
	}
}
