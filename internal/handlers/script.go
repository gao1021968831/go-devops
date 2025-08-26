package handlers

import (
	"encoding/json"
	"fmt"
	"go-devops/internal/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ScriptHandler struct {
	db *gorm.DB
}

func NewScriptHandler(db *gorm.DB) *ScriptHandler {
	return &ScriptHandler{db: db}
}

// GetScripts 获取脚本列表
func (h *ScriptHandler) GetScripts(c *gin.Context) {
	var scripts []models.Script
	var total int64

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	scriptType := c.Query("type")

	offset := (page - 1) * pageSize

	// 构建查询
	query := h.db.Model(&models.Script{})
	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if scriptType != "" {
		query = query.Where("type = ?", scriptType)
	}

	// 获取总数
	query.Count(&total)

	// 获取数据
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&scripts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scripts": scripts,
		"total":   total,
		"page":    page,
		"size":    pageSize,
	})
}

// GetScript 获取脚本详情
func (h *ScriptHandler) GetScript(c *gin.Context) {
	id := c.Param("id")
	var script models.Script

	if err := h.db.First(&script, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "脚本不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本失败"})
		}
		return
	}

	c.JSON(http.StatusOK, script)
}

// CreateScript 创建脚本
func (h *ScriptHandler) CreateScript(c *gin.Context) {
	var req models.ScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	script := models.Script{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Type:        req.Type,
		CreatedBy:   userID.(uint),
	}

	if err := h.db.Create(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建脚本失败"})
		return
	}

	// 记录活动日志
	activity := models.UserActivity{
		UserID:      userID.(uint),
		Action:      "create",
		Resource:    "script",
		ResourceID:  &script.ID,
		Description: "创建脚本: " + script.Name,
	}
	h.db.Create(&activity)

	c.JSON(http.StatusCreated, script)
}

// UpdateScript 更新脚本
func (h *ScriptHandler) UpdateScript(c *gin.Context) {
	id := c.Param("id")
	var script models.Script

	if err := h.db.First(&script, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "脚本不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本失败"})
		}
		return
	}

	var req models.ScriptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 更新脚本信息
	script.Name = req.Name
	script.Description = req.Description
	script.Content = req.Content
	script.Type = req.Type

	if err := h.db.Save(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新脚本失败"})
		return
	}

	// 记录活动日志
	activity := models.UserActivity{
		UserID:      userID.(uint),
		Action:      "update",
		Resource:    "script",
		ResourceID:  &script.ID,
		Description: "更新脚本: " + script.Name,
	}
	h.db.Create(&activity)

	c.JSON(http.StatusOK, script)
}

// DeleteScript 删除脚本
func (h *ScriptHandler) DeleteScript(c *gin.Context) {
	id := c.Param("id")
	var script models.Script

	if err := h.db.First(&script, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "脚本不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取脚本失败"})
		}
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查是否有关联的作业
	var jobCount int64
	h.db.Model(&models.Job{}).Where("script_id = ?", id).Count(&jobCount)
	if jobCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该脚本被作业引用，无法删除"})
		return
	}

	if err := h.db.Delete(&script).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除脚本失败"})
		return
	}

	// 记录活动日志
	activity := models.UserActivity{
		UserID:      userID.(uint),
		Action:      "delete",
		Resource:    "script",
		ResourceID:  &script.ID,
		Description: "删除脚本: " + script.Name,
	}
	h.db.Create(&activity)

	c.JSON(http.StatusOK, gin.H{"message": "脚本删除成功"})
}

// BatchDeleteScripts 批量删除脚本
func (h *ScriptHandler) BatchDeleteScripts(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 检查是否有关联的作业
	var jobCount int64
	h.db.Model(&models.Job{}).Where("script_id IN ?", req.IDs).Count(&jobCount)
	if jobCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "部分脚本被作业引用，无法删除"})
		return
	}

	// 获取要删除的脚本名称
	var scripts []models.Script
	h.db.Where("id IN ?", req.IDs).Find(&scripts)

	// 批量删除
	if err := h.db.Where("id IN ?", req.IDs).Delete(&models.Script{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除脚本失败"})
		return
	}

	// 记录活动日志
	for _, script := range scripts {
		activity := models.UserActivity{
			UserID:      userID.(uint),
			Action:      "delete",
			Resource:    "script",
			ResourceID:  &script.ID,
			Description: "批量删除脚本: " + script.Name,
		}
		h.db.Create(&activity)
	}

	c.JSON(http.StatusOK, gin.H{"message": "批量删除成功"})
}

// ExportScripts 导出脚本
func (h *ScriptHandler) ExportScripts(c *gin.Context) {
	// 检查是否有指定的脚本ID
	idsParam := c.Query("ids")
	var scripts []models.Script
	
	if idsParam != "" {
		// 导出指定的脚本
		var ids []uint
		if err := json.Unmarshal([]byte(idsParam), &ids); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的脚本ID参数"})
			return
		}
		if err := h.db.Where("id IN ?", ids).Find(&scripts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导出脚本失败"})
			return
		}
	} else {
		// 导出所有脚本
		if err := h.db.Find(&scripts).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "导出脚本失败"})
			return
		}
	}

	// 生成CSV内容
	var csvContent strings.Builder
	
	// 添加UTF-8 BOM以解决中文乱码问题
	csvContent.WriteString("\xEF\xBB\xBF")
	
	// CSV头部
	csvContent.WriteString("ID,名称,类型,描述,内容,创建者ID,创建时间,更新时间\n")
	
	// CSV数据行
	for _, script := range scripts {
		// 转义CSV字段中的特殊字符
		name := strings.ReplaceAll(script.Name, "\"", "\"\"")
		scriptType := strings.ReplaceAll(script.Type, "\"", "\"\"")
		description := strings.ReplaceAll(script.Description, "\"", "\"\"")
		content := strings.ReplaceAll(script.Content, "\"", "\"\"")
		content = strings.ReplaceAll(content, "\n", "\\n")
		content = strings.ReplaceAll(content, "\r", "\\r")
		
		createdAt := script.CreatedAt.Format("2006-01-02 15:04:05")
		updatedAt := script.UpdatedAt.Format("2006-01-02 15:04:05")
		
		csvContent.WriteString(fmt.Sprintf("%d,\"%s\",\"%s\",\"%s\",\"%s\",%d,\"%s\",\"%s\"\n",
			script.ID, name, scriptType, description, content, script.CreatedBy, createdAt, updatedAt))
	}

	// 设置正确的CSV响应头
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=scripts.csv")
	
	// 返回CSV内容
	c.String(http.StatusOK, csvContent.String())
}

// BatchImportScripts 批量导入脚本
func (h *ScriptHandler) BatchImportScripts(c *gin.Context) {
	// 获取当前用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要导入的CSV文件"})
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "打开文件失败"})
		return
	}
	defer src.Close()

	// 读取文件内容
	content := make([]byte, file.Size)
	_, err = src.Read(content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}

	// 解析CSV内容，确保UTF-8编码正确处理
	csvContent := string(content)
	
	// 移除BOM标记（如果存在）
	csvContent = strings.TrimPrefix(csvContent, "\xEF\xBB\xBF")

	// 统一换行符处理
	csvContent = strings.ReplaceAll(csvContent, "\r\n", "\n")
	csvContent = strings.ReplaceAll(csvContent, "\r", "\n")
	
	lines := strings.Split(csvContent, "\n")
	if len(lines) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV文件格式错误，至少需要包含头部和一行数据"})
		return
	}

	var scripts []models.Script
	var successCount, errorCount int
	var errors []string
	
	// 跳过头部行，从第二行开始解析
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue // 跳过空行
		}

		// 解析CSV行
		fields := parseCSVLine(line)
		if len(fields) < 5 { // 至少需要：ID,名称,类型,描述,内容
			errorCount++
			errors = append(errors, fmt.Sprintf("第%d行格式错误：字段数量不足", i+1))
			continue
		}

		// 清理和验证字段
		name := strings.TrimSpace(fields[1])
		scriptType := strings.TrimSpace(fields[2])
		description := strings.TrimSpace(fields[3])
		content := strings.TrimSpace(fields[4])
		
		// 还原转义的换行符
		content = strings.ReplaceAll(content, "\\n", "\n")
		content = strings.ReplaceAll(content, "\\r", "\r")
		content = strings.ReplaceAll(content, "\\t", "\t")
		
		// 验证必填字段
		if name == "" {
			errorCount++
			errors = append(errors, fmt.Sprintf("第%d行错误：脚本名称不能为空", i+1))
			continue
		}
		
		if scriptType == "" {
			scriptType = "shell" // 默认类型
		}

		script := models.Script{
			Name:        name,
			Type:        scriptType,
			Description: description,
			Content:     content,
			CreatedBy:   userID.(uint),
		}

		scripts = append(scripts, script)
		successCount++
	}

	totalCount := successCount + errorCount
	
	if len(scripts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "没有找到有效的脚本数据",
			"total_count":   totalCount,
			"success_count": 0,
			"error_count":   errorCount,
			"errors":        errors,
		})
		return
	}

	// 批量创建脚本
	if err := h.db.Create(&scripts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         "批量导入脚本失败: " + err.Error(),
			"total_count":   totalCount,
			"success_count": 0,
			"error_count":   totalCount,
		})
		return
	}

	// 记录活动日志
	activity := models.UserActivity{
		UserID:      userID.(uint),
		Action:      "import",
		Resource:    "script",
		Description: fmt.Sprintf("批量导入脚本，成功%d个，失败%d个", successCount, errorCount),
	}
	h.db.Create(&activity)

	// 构建响应消息
	message := fmt.Sprintf("批量导入完成，成功%d个", successCount)
	if errorCount > 0 {
		message += fmt.Sprintf("，失败%d个", errorCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       message,
		"count":         successCount,
		"total_count":   totalCount,
		"success_count": successCount,
		"error_count":   errorCount,
		"errors":        errors,
	})
}

// parseCSVLine 解析CSV行，处理引号包围的字段
func parseCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	
	for i, char := range line {
		switch char {
		case '"':
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				// 双引号转义
				current.WriteRune('"')
				i++ // 跳过下一个引号
			} else {
				inQuotes = !inQuotes
			}
		case ',':
			if !inQuotes {
				fields = append(fields, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}
	
	// 添加最后一个字段
	fields = append(fields, current.String())
	
	return fields
}

// DownloadImportTemplate 下载导入模板
func (h *ScriptHandler) DownloadImportTemplate(c *gin.Context) {
	// 生成CSV模板内容
	var csvContent strings.Builder
	
	// 添加UTF-8 BOM以解决中文乱码问题
	csvContent.WriteString("\xEF\xBB\xBF")
	
	// CSV头部（与导出格式保持一致）
	csvContent.WriteString("ID,名称,类型,描述,内容,创建者ID,创建时间,更新时间\n")
	
	// 示例数据行
	csvContent.WriteString("1,\"示例Shell脚本\",\"shell\",\"这是一个Shell脚本示例\",\"#!/bin/bash\\necho 'Hello World'\\ndate\",1,\"2024-01-01 12:00:00\",\"2024-01-01 12:00:00\"\n")
	csvContent.WriteString("2,\"示例Python脚本\",\"python3\",\"这是一个Python脚本示例\",\"#!/usr/bin/env python3\\nprint('Hello Python')\\nimport datetime\\nprint(datetime.datetime.now())\",1,\"2024-01-01 12:00:00\",\"2024-01-01 12:00:00\"\n")
	csvContent.WriteString("3,\"系统信息检查\",\"shell\",\"检查系统基本信息\",\"#!/bin/bash\\necho '=== 系统信息 ==='\\nuname -a\\necho '=== 内存信息 ==='\\nfree -h\\necho '=== 磁盘信息 ==='\\ndf -h\",1,\"2024-01-01 12:00:00\",\"2024-01-01 12:00:00\"\n")

	// 设置正确的CSV响应头
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=script_import_template.csv")
	
	// 返回CSV内容
	c.String(http.StatusOK, csvContent.String())
}
