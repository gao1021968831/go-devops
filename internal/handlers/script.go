package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-devops/internal/models"
	"gorm.io/gorm"
)

type ScriptHandler struct {
	db *gorm.DB
}

func NewScriptHandler(db *gorm.DB) *ScriptHandler {
	return &ScriptHandler{db: db}
}

// 获取脚本列表
func (h *ScriptHandler) GetScripts(c *gin.Context) {
	var scripts []models.Script
	if err := h.db.Preload("User").Find(&scripts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本列表失败"})
		return
	}

	c.JSON(http.StatusOK, scripts)
}

// 创建脚本
func (h *ScriptHandler) CreateScript(c *gin.Context) {
	var script models.Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 设置创建者
	script.CreatedBy = c.GetUint("user_id")

	if err := h.db.Create(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建脚本失败"})
		return
	}

	// 加载用户信息
	h.db.Preload("User").First(&script, script.ID)

	c.JSON(http.StatusCreated, script)
}

// 获取单个脚本
func (h *ScriptHandler) GetScript(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的脚本ID"})
		return
	}

	var script models.Script
	if err := h.db.Preload("User").First(&script, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "脚本不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本信息失败"})
		}
		return
	}

	c.JSON(http.StatusOK, script)
}

// 更新脚本
func (h *ScriptHandler) UpdateScript(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的脚本ID"})
		return
	}

	var script models.Script
	if err := h.db.First(&script, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "脚本不存在"})
		return
	}

	// 检查权限（只有创建者或管理员可以修改）
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if script.CreatedBy != userID && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限修改此脚本"})
		return
	}

	var updateData models.Script
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if err := h.db.Model(&script).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新脚本失败"})
		return
	}

	// 重新加载脚本信息
	h.db.Preload("User").First(&script, script.ID)

	c.JSON(http.StatusOK, script)
}

// 删除脚本
func (h *ScriptHandler) DeleteScript(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的脚本ID"})
		return
	}

	var script models.Script
	if err := h.db.First(&script, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "脚本不存在"})
		return
	}

	// 检查权限
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if script.CreatedBy != userID && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限删除此脚本"})
		return
	}

	if err := h.db.Delete(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除脚本失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "脚本删除成功"})
}
