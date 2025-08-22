package handlers

import (
	"net/http"
	"strconv"

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
