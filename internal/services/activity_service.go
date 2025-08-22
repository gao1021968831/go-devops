package services

import (
	"go-devops/internal/logger"
	"go-devops/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ActivityService struct {
	db *gorm.DB
}

func NewActivityService(db *gorm.DB) *ActivityService {
	return &ActivityService{db: db}
}

// 记录用户活动
func (s *ActivityService) LogActivity(c *gin.Context, userID uint, action, resource string, resourceID *uint, description string, status string, details string) {
	activity := models.UserActivity{
		UserID:      userID,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		IPAddress:   c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
		Status:      status,
		Details:     details,
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(&activity).Error; err != nil {
		logger.Errorf("记录用户活动失败: %v", err)
	}
}

// 记录成功的活动
func (s *ActivityService) LogSuccess(c *gin.Context, userID uint, action, resource string, resourceID *uint, description string) {
	s.LogActivity(c, userID, action, resource, resourceID, description, "success", "")
}

// 记录失败的活动
func (s *ActivityService) LogFailure(c *gin.Context, userID uint, action, resource string, resourceID *uint, description string, errorMsg string) {
	s.LogActivity(c, userID, action, resource, resourceID, description, "failed", errorMsg)
}

// 获取用户活动列表（支持分页和筛选）
func (s *ActivityService) GetActivities(page, size int, userID *uint, action, resource, status string, startDate, endDate, keyword string) ([]models.UserActivity, int64, error) {
	var activities []models.UserActivity
	var total int64

	query := s.db.Model(&models.UserActivity{}).Preload("User")

	// 筛选条件
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}
	if keyword != "" {
		query = query.Where("description LIKE ? OR details LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	if err := query.Order("created_at desc").Offset(offset).Limit(size).Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// 获取最近活动
func (s *ActivityService) GetRecentActivities(limit int) ([]models.UserActivity, error) {
	var activities []models.UserActivity
	
	err := s.db.Preload("User").
		Order("created_at desc").
		Limit(limit).
		Find(&activities).Error
	
	return activities, err
}

// 清理旧的活动记录（保留指定天数）
func (s *ActivityService) CleanOldActivities(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	
	result := s.db.Where("created_at < ?", cutoffDate).Delete(&models.UserActivity{})
	if result.Error != nil {
		return result.Error
	}
	
	logger.Infof("清理了 %d 条旧的活动记录", result.RowsAffected)
	return nil
}
