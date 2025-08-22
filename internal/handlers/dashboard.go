package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-devops/internal/logger"
	"go-devops/internal/models"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// 活动记录结构
type Activity struct {
	ID        uint      `json:"id"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // success, warning, error, info
	Timestamp time.Time `json:"timestamp"`
	UserID    uint      `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
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

	var activities []Activity

	// 从作业执行记录获取活动
	var executions []models.JobExecution
	err = h.db.Preload("Job").Preload("Host").
		Order("created_at desc").
		Limit(limit).
		Find(&executions).Error

	if err != nil {
		logger.Errorf("获取作业执行记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取活动记录失败",
		})
		return
	}

	// 转换为活动记录
	for _, exec := range executions {
		var activityType string
		var message string

		switch exec.Status {
		case "completed":
			activityType = "success"
			message = "作业 \"" + exec.Job.Name + "\" 在主机 " + exec.Host.Name + " 上执行成功"
		case "failed":
			activityType = "error"
			message = "作业 \"" + exec.Job.Name + "\" 在主机 " + exec.Host.Name + " 上执行失败"
		case "running":
			activityType = "info"
			message = "作业 \"" + exec.Job.Name + "\" 正在主机 " + exec.Host.Name + " 上执行"
		default:
			activityType = "info"
			message = "作业 \"" + exec.Job.Name + "\" 状态变更为 " + exec.Status
		}

		activity := Activity{
			ID:        exec.ID,
			Message:   message,
			Type:      activityType,
			Timestamp: exec.CreatedAt,
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
