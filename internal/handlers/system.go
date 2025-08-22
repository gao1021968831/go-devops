package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-devops/internal/config"
	"go-devops/internal/logger"
)

type SystemHandler struct{}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// 获取系统信息
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	cfg, err := config.Load()
	if err != nil {
		logger.Errorf("获取系统配置失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统配置失败"})
		return
	}

	systemInfo := gin.H{
		"app": gin.H{
			"name":        cfg.App.Name,
			"version":     cfg.App.Version,
			"environment": cfg.App.Environment,
		},
		"features": gin.H{
			"scheduler_enabled": cfg.Scheduler.Enabled,
			"logging_to_file":   cfg.Logging.ToFile,
		},
	}

	c.JSON(http.StatusOK, systemInfo)
}

// 获取应用健康状态
func (h *SystemHandler) GetHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": "ok",
		"services": gin.H{
			"database": "connected",
			"logger":   "active",
		},
	})
}
