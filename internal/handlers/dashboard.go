package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-devops/internal/logger"
	"go-devops/internal/models"
	"go-devops/internal/services"
)

type DashboardHandler struct {
	db              *gorm.DB
	activityService *services.ActivityService
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{
		db:              db,
		activityService: services.NewActivityService(db),
	}
}

// 活动记录结构
type Activity struct {
	ID        uint      `json:"id"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // success, warning, error, info
	Timestamp time.Time `json:"timestamp"`
	UserID    uint      `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Resource  string    `json:"resource,omitempty"`
}

// 分页响应结构
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Size       int         `json:"size"`
	TotalPages int64       `json:"total_pages"`
}

// 分页活动响应结构
type ActivitiesResponse struct {
	Items []Activity `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

// 作业趋势数据结构
type JobTrendData struct {
	Date    string `json:"date"`
	Success int    `json:"success"`
	Failed  int    `json:"failed"`
}

// 获取最近活动
func (h *DashboardHandler) GetRecentActivities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// 使用新的活动记录系统
	userActivities, err := h.activityService.GetRecentActivities(limit)
	if err != nil {
		logger.Errorf("获取最近活动失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取活动记录失败",
		})
		return
	}

	// 转换为前端需要的格式
	var activities []Activity
	for _, ua := range userActivities {
		var activityType string
		switch ua.Status {
		case "success":
			activityType = "success"
		case "failed":
			activityType = "error"
		default:
			activityType = "info"
		}

		activity := Activity{
			ID:        ua.ID,
			Message:   ua.Description,
			Type:      activityType,
			Timestamp: ua.CreatedAt,
			UserID:    ua.UserID,
			Username:  ua.User.Username,
		}

		activities = append(activities, activity)
	}

	c.JSON(http.StatusOK, activities)
}

// 获取作业趋势数据
func (h *DashboardHandler) GetJobTrend(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 7
	}

	var trendData []JobTrendData

	// 计算日期范围
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		
		// 查询当天的成功和失败作业数量
		var successCount int64
		var failedCount int64

		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		// 统计成功的作业执行
		h.db.Model(&models.JobExecution{}).
			Where("status = ? AND created_at >= ? AND created_at < ?", "completed", startOfDay, endOfDay).
			Count(&successCount)

		// 统计失败的作业执行
		h.db.Model(&models.JobExecution{}).
			Where("status = ? AND created_at >= ? AND created_at < ?", "failed", startOfDay, endOfDay).
			Count(&failedCount)

		trendData = append(trendData, JobTrendData{
			Date:    date.Format("01-02"),
			Success: int(successCount),
			Failed:  int(failedCount),
		})
	}

	c.JSON(http.StatusOK, trendData)
}

// 获取仪表盘统计数据
func (h *DashboardHandler) GetDashboardStats(c *gin.Context) {
	stats := gin.H{}

	// 主机统计
	var totalHosts int64
	var onlineHosts int64
	h.db.Model(&models.Host{}).Count(&totalHosts)
	h.db.Model(&models.Host{}).Where("status = ?", "online").Count(&onlineHosts)
	
	stats["total_hosts"] = totalHosts
	stats["online_hosts"] = onlineHosts

	// 作业统计
	var totalJobs int64
	var runningJobs int64
	h.db.Model(&models.Job{}).Count(&totalJobs)
	h.db.Model(&models.JobExecution{}).Where("status = ?", "running").Count(&runningJobs)
	
	stats["total_jobs"] = totalJobs
	stats["running_jobs"] = runningJobs

	// 脚本统计
	var totalScripts int64
	h.db.Model(&models.Script{}).Count(&totalScripts)
	stats["total_scripts"] = totalScripts


	c.JSON(http.StatusOK, stats)
}

// 获取主机状态分布
func (h *DashboardHandler) GetHostStatusDistribution(c *gin.Context) {
	var results []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	err := h.db.Model(&models.Host{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error

	if err != nil {
		logger.Errorf("获取主机状态分布失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取主机状态分布失败",
		})
		return
	}

	c.JSON(http.StatusOK, results)
}

// 获取全部活动（支持分页和筛选）
func (h *DashboardHandler) GetAllActivities(c *gin.Context) {
	// 解析查询参数
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")
	activityType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	keyword := c.Query("keyword")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 || size > 100 {
		size = 20
	}

	// 转换活动类型筛选
	var status string
	switch activityType {
	case "success":
		status = "success"
	case "error":
		status = "failed"
	case "info":
		// 不设置状态筛选，显示所有
	}

	// 使用新的活动记录系统
	userActivities, total, err := h.activityService.GetActivities(
		page, size, nil, "", "", status, startDate, endDate, keyword,
	)
	if err != nil {
		logger.Errorf("获取全部活动失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取活动记录失败",
		})
		return
	}

	// 转换为前端需要的格式
	var activities []Activity
	for _, ua := range userActivities {
		var activityType string
		switch ua.Status {
		case "success":
			activityType = "success"
		case "failed":
			activityType = "error"
		default:
			activityType = "info"
		}

		resource := ua.Resource
		if ua.ResourceID != nil {
			resource += " (ID: " + strconv.Itoa(int(*ua.ResourceID)) + ")"
		}

		activity := Activity{
			ID:        ua.ID,
			Message:   ua.Description,
			Type:      activityType,
			Timestamp: ua.CreatedAt,
			UserID:    ua.UserID,
			Username:  ua.User.Username,
			Resource:  resource,
		}

		activities = append(activities, activity)
	}

	response := PaginatedResponse{
		Data:       activities,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: (total + int64(size) - 1) / int64(size),
	}

	c.JSON(http.StatusOK, response)
}
