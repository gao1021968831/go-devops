package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-devops/internal/logger"
	"go-devops/internal/models"
	"go-devops/internal/services"
	"go-devops/internal/ssh"
)

type FileHandler struct {
	db              *gorm.DB
	activityService *services.ActivityService
	uploadPath      string
}

func NewFileHandler(db *gorm.DB) *FileHandler {
	uploadPath := "uploads"
	// 确保上传目录存在
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		logger.Errorf("创建上传目录失败: %v", err)
	}

	return &FileHandler{
		db:              db,
		activityService: services.NewActivityService(db),
		uploadPath:      uploadPath,
	}
}

// 获取文件列表
func (h *FileHandler) GetFiles(c *gin.Context) {
	// 查询参数
	name := c.Query("name")
	category := c.Query("category")
	isPublic := c.Query("isPublic")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	// 兼容前端使用的 page_size 参数
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			size = ps
		}
	}

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	query := h.db.Preload("User").Model(&models.File{})

	// 权限过滤：普通用户只能看到公开文件和自己上传的文件
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" {
		query = query.Where("is_public = ? OR uploaded_by = ?", true, userID)
	}

	// 文件名筛选
	if name != "" {
		query = query.Where("original_name LIKE ?", "%"+name+"%")
	}

	// 分类过滤
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// 公开状态过滤
	if isPublic != "" {
		if isPublic == "true" {
			query = query.Where("is_public = ?", true)
		} else {
			query = query.Where("is_public = ?", false)
		}
	}

	var total int64
	query.Count(&total)

	var files []models.File
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&files).Error; err != nil {
		logger.Errorf("获取文件列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        files,
		"total":       total,
		"page":        page,
		"size":        size,
		"total_pages": (total + int64(size) - 1) / int64(size),
	})
}

// 上传文件
func (h *FileHandler) UploadFile(c *gin.Context) {
	// 解析表单数据
	var req models.FileUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.Warnf("文件上传请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Warnf("获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的文件"})
		return
	}
	defer file.Close()

	// 验证文件大小（限制100MB）
	maxSize := int64(100 * 1024 * 1024)
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小超过限制（100MB）"})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(header.Filename)
	fileName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(),
		strings.ReplaceAll(header.Filename[:len(header.Filename)-len(ext)], " ", "_"), ext)

	// 根据分类创建子目录
	category := req.Category
	if category == "" {
		category = "general"
	}

	categoryPath := filepath.Join(h.uploadPath, category)
	if err := os.MkdirAll(categoryPath, 0755); err != nil {
		logger.Errorf("创建分类目录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建存储目录失败"})
		return
	}

	// 先计算MD5哈希，避免重复文件的磁盘写入
	hash := md5.New()

	// 重置文件指针到开头
	if _, err := file.Seek(0, 0); err != nil {
		logger.Errorf("重置文件指针失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理文件失败"})
		return
	}

	// 计算MD5
	if _, err := io.Copy(hash, file); err != nil {
		logger.Errorf("计算文件MD5失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理文件失败"})
		return
	}

	md5Hash := fmt.Sprintf("%x", hash.Sum(nil))

	// 获取当前用户ID
	userID := c.GetUint("user_id")

	// 检查当前用户是否已上传过相同MD5的文件
	var existingFile models.File
	if err := h.db.Where("md5_hash = ? AND uploaded_by = ?", md5Hash, userID).First(&existingFile).Error; err == nil {
		// 当前用户已上传过相同文件，返回冲突错误
		c.JSON(http.StatusConflict, gin.H{
			"error":         "文件已存在",
			"message":       fmt.Sprintf("您已上传过相同的文件 '%s'，MD5: %s", existingFile.OriginalName, md5Hash),
			"existing_file": existingFile,
		})
		return
	}

	// 文件不重复，现在保存到磁盘
	filePath := filepath.Join(categoryPath, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		logger.Errorf("创建文件失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
		return
	}
	defer dst.Close()

	// 重置文件指针到开头准备写入
	if _, err := file.Seek(0, 0); err != nil {
		logger.Errorf("重置文件指针失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "处理文件失败"})
		return
	}

	// 保存文件到磁盘
	if _, err := io.Copy(dst, file); err != nil {
		logger.Errorf("写入文件失败: %v", err)
		os.Remove(filePath) // 清理失败的文件
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
		return
	}

	// 处理MIME类型，对无扩展名文件使用application/octet-stream
	mimeType := header.Header.Get("Content-Type")
	if !strings.Contains(header.Filename, ".") {
		// 无扩展名文件直接设置为二进制流
		mimeType = "application/octet-stream"
	} else if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// 保存文件信息到数据库
	fileModel := models.File{
		Name:         fileName,
		OriginalName: header.Filename,
		Path:         filePath,
		Size:         header.Size,
		MimeType:     mimeType,
		MD5Hash:      md5Hash,
		Category:     category,
		Description:  req.Description,
		IsPublic:     req.IsPublic,
		UploadedBy:   c.GetUint("user_id"),
	}

	if err := h.db.Create(&fileModel).Error; err != nil {
		logger.Errorf("保存文件信息失败: %v", err)
		os.Remove(filePath) // 清理文件
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件信息失败"})
		return
	}

	// 预加载用户信息
	h.db.Preload("User").First(&fileModel, fileModel.ID)

	logger.Infof("文件上传成功: %s, 用户: %d", header.Filename, c.GetUint("user_id"))

	// 记录活动
	h.activityService.LogSuccess(c, c.GetUint("user_id"), "upload", "file", &fileModel.ID,
		fmt.Sprintf("上传文件 '%s' (%s)", header.Filename, category))

	c.JSON(http.StatusCreated, gin.H{
		"message": "文件上传成功",
		"file":    fileModel,
	})
}

// 下载文件
func (h *FileHandler) DownloadFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var file models.File
	if err := h.db.First(&file, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件信息失败"})
		}
		return
	}

	// 权限检查：管理员或文件所有者或公开文件
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" && file.UploadedBy != userID && !file.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限下载此文件"})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(file.Path); os.IsNotExist(err) {
		logger.Errorf("文件不存在: %s", file.Path)
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 更新下载次数
	h.db.Model(&file).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1))

	// 记录活动
	h.activityService.LogSuccess(c, userID, "download", "file", &file.ID,
		fmt.Sprintf("下载文件 '%s'", file.OriginalName))

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")

	// 正确编码文件名，支持中文和特殊字符，防止浏览器自动添加扩展名
	safeFilename := strings.ReplaceAll(file.OriginalName, "\"", "")
	encodedFilename := url.QueryEscape(file.OriginalName)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		safeFilename, encodedFilename))

	// 强制处理MIME类型，防止浏览器自动添加扩展名
	contentType := file.MimeType

	// 对于无扩展名文件，强制使用 application/octet-stream
	if !strings.Contains(file.OriginalName, ".") {
		contentType = "application/octet-stream"
	} else if contentType == "" || contentType == "text/plain" {
		// 空MIME类型或text/plain也使用二进制流
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	// 添加额外的头部来强制浏览器按原文件名下载
	c.Header("X-Content-Type-Options", "nosniff")

	// 发送文件
	c.File(file.Path)
}

// 获取单个文件信息
func (h *FileHandler) GetFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var file models.File
	if err := h.db.Preload("User").First(&file, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件信息失败"})
		}
		return
	}

	// 权限检查
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" && file.UploadedBy != userID && !file.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限查看此文件"})
		return
	}

	c.JSON(http.StatusOK, file)
}

// 更新文件信息
func (h *FileHandler) UpdateFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var file models.File
	if err := h.db.First(&file, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 权限检查：只有文件所有者或管理员可以修改
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" && file.UploadedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限修改此文件"})
		return
	}

	var req models.FileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 更新字段
	updateData := make(map[string]interface{})
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Category != "" {
		updateData["category"] = req.Category
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	updateData["is_public"] = req.IsPublic

	if err := h.db.Model(&file).Updates(updateData).Error; err != nil {
		logger.Errorf("更新文件信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新文件信息失败"})
		return
	}

	// 重新查询更新后的数据
	h.db.Preload("User").First(&file, file.ID)

	// 记录活动
	h.activityService.LogSuccess(c, userID, "update", "file", &file.ID,
		fmt.Sprintf("更新文件 '%s' 信息", file.OriginalName))

	c.JSON(http.StatusOK, file)
}

// 删除文件
func (h *FileHandler) DeleteFile(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件ID"})
		return
	}

	var file models.File
	if err := h.db.First(&file, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 权限检查：只有文件所有者或管理员可以删除
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" && file.UploadedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限删除此文件"})
		return
	}

	// 检查是否有分发记录关联
	var distributionCount int64
	if err := h.db.Model(&models.FileDistribution{}).Where("file_id = ?", file.ID).Count(&distributionCount).Error; err != nil {
		logger.Errorf("检查文件分发记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查文件关联数据失败"})
		return
	}

	if distributionCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "无法删除文件",
			"message": fmt.Sprintf("文件 '%s' 存在 %d 条分发记录，无法删除", file.OriginalName, distributionCount),
		})
		return
	}

	// 检查是否有作业执行记录关联
	var jobExecutionCount int64
	if err := h.db.Model(&models.JobExecution{}).Where("output_file_id = ?", file.ID).Count(&jobExecutionCount).Error; err != nil {
		logger.Errorf("检查作业执行记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查文件关联数据失败"})
		return
	}

	if jobExecutionCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "无法删除文件",
			"message": fmt.Sprintf("文件 '%s' 被 %d 个作业执行记录引用，无法删除", file.OriginalName, jobExecutionCount),
		})
		return
	}

	// 删除物理文件
	if err := os.Remove(file.Path); err != nil && !os.IsNotExist(err) {
		logger.Warnf("删除物理文件失败: %v", err)
		// 继续删除数据库记录，即使物理文件删除失败
	}

	// 删除数据库记录
	if err := h.db.Delete(&file).Error; err != nil {
		logger.Errorf("删除文件记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件失败"})
		return
	}

	logger.Infof("删除文件成功: %s", file.OriginalName)

	// 记录活动
	h.activityService.LogSuccess(c, userID, "delete", "file", &file.ID,
		fmt.Sprintf("删除文件 '%s'", file.OriginalName))

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("文件 '%s' 删除成功", file.OriginalName),
	})
}

// 文件分发
func (h *FileHandler) DistributeFile(c *gin.Context) {
	var req models.FileDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("文件分发请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 检查文件是否存在
	var file models.File
	if err := h.db.First(&file, req.FileID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		return
	}

	// 权限检查
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" && file.UploadedBy != userID && !file.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限分发此文件"})
		return
	}

	// 验证主机ID
	var validHosts []models.Host
	if err := h.db.Where("id IN ?", req.HostIDs).Find(&validHosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询主机信息失败"})
		return
	}

	if len(validHosts) != len(req.HostIDs) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "部分主机ID无效"})
		return
	}

	// 将主机ID列表转换为JSON字符串
	hostIDsJSON, _ := json.Marshal(req.HostIDs)

	// 创建分发记录
	distribution := models.FileDistribution{
		FileID:     req.FileID,
		HostIDs:    string(hostIDsJSON),
		TargetPath: req.TargetPath,
		Status:     "pending",
		CreatedBy:  userID,
	}

	if err := h.db.Create(&distribution).Error; err != nil {
		logger.Errorf("创建文件分发记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建分发任务失败"})
		return
	}

	// 创建分发详情记录
	for _, hostID := range req.HostIDs {
		detail := models.FileDistributionDetail{
			DistributionID: distribution.ID,
			HostID:         hostID,
			Status:         "pending",
		}
		h.db.Create(&detail)
	}

	// 异步执行分发任务
	go h.executeFileDistribution(&distribution, &file, validHosts)

	// 预加载相关信息
	h.db.Preload("File").Preload("User").First(&distribution, distribution.ID)

	// 记录活动
	h.activityService.LogSuccess(c, userID, "distribute", "file", &file.ID,
		fmt.Sprintf("分发文件 '%s' 到 %d 台主机", file.OriginalName, len(req.HostIDs)))

	c.JSON(http.StatusCreated, gin.H{
		"message":      "文件分发任务创建成功",
		"distribution": distribution,
	})
}

// 执行文件分发（支持并发和重试）
func (h *FileHandler) executeFileDistribution(distribution *models.FileDistribution, file *models.File, hosts []models.Host) {
	logger.Infof("开始执行文件分发任务: %d，目标主机数: %d", distribution.ID, len(hosts))

	// 更新分发状态为运行中
	startTime := time.Now()
	h.db.Model(distribution).Updates(map[string]interface{}{
		"status":     "running",
		"start_time": &startTime,
	})

	// 并发控制：最多同时分发到3台主机
	const maxConcurrency = 3
	const maxRetries = 3

	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	successCount := 0
	totalCount := len(hosts)
	completedCount := 0

	for _, host := range hosts {
		wg.Add(1)
		go func(currentHost models.Host) {
			defer wg.Done()

			// 获取并发许可
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行单个主机的分发任务
			success := h.distributeToSingleHost(distribution, file, &currentHost, maxRetries)

			// 更新计数器（需要加锁）
			mu.Lock()
			completedCount++
			if success {
				successCount++
			}

			// 更新总体进度
			progress := (completedCount * 100) / totalCount
			h.db.Model(distribution).UpdateColumn("progress", progress)

			logger.Infof("分发进度: %d/%d (%.1f%%), 成功: %d",
				completedCount, totalCount, float64(progress), successCount)
			mu.Unlock()
		}(host)
	}

	// 等待所有分发任务完成
	wg.Wait()

	// 更新最终状态
	endTime := time.Now()
	finalStatus := "completed"
	if successCount == 0 {
		finalStatus = "failed"
	} else if successCount < totalCount {
		finalStatus = "partial"
	}

	h.db.Model(distribution).Updates(map[string]interface{}{
		"status":   finalStatus,
		"progress": 100,
		"end_time": &endTime,
	})

	logger.Infof("文件分发任务完成: %d, 成功: %d/%d, 用时: %v",
		distribution.ID, successCount, totalCount, endTime.Sub(startTime))
}

// 分发文件到单个主机（支持重试）
func (h *FileHandler) distributeToSingleHost(distribution *models.FileDistribution, file *models.File, host *models.Host, maxRetries int) bool {
	// 获取分发详情记录
	var detail models.FileDistributionDetail
	h.db.Where("distribution_id = ? AND host_id = ?", distribution.ID, host.ID).First(&detail)

	detailStartTime := time.Now()
	h.db.Model(&detail).Updates(map[string]interface{}{
		"status":     "running",
		"start_time": &detailStartTime,
	})

	// 重试逻辑
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Infof("尝试分发文件到主机 %s (第%d次，共%d次)", host.Name, attempt, maxRetries)

		err := h.transferFileToHost(file, host, distribution.TargetPath)
		if err == nil {
			// 分发成功
			detailEndTime := time.Now()
			h.db.Model(&detail).Updates(map[string]interface{}{
				"status":   "completed",
				"output":   fmt.Sprintf("文件传输成功 (尝试%d次)", attempt),
				"end_time": &detailEndTime,
			})

			logger.Infof("文件分发到主机 %s 成功 (尝试%d次)", host.Name, attempt)
			return true
		}

		lastErr = err
		logger.Warnf("文件分发到主机 %s 失败 (第%d次): %v", host.Name, attempt, err)

		// 如果不是最后一次尝试，等待一段时间再重试
		if attempt < maxRetries {
			retryDelay := time.Duration(attempt) * time.Second
			logger.Infof("等待 %v 后重试...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	// 所有重试都失败了
	detailEndTime := time.Now()
	h.db.Model(&detail).Updates(map[string]interface{}{
		"status":   "failed",
		"error":    fmt.Sprintf("重试%d次后仍然失败: %v", maxRetries, lastErr),
		"end_time": &detailEndTime,
	})

	logger.Errorf("文件分发到主机 %s 最终失败，已重试%d次: %v", host.Name, maxRetries, lastErr)
	return false
}

// 传输文件到主机
func (h *FileHandler) transferFileToHost(file *models.File, host *models.Host, targetPath string) error {
	// 检查主机SSH配置
	if host.Username == "" {
		return fmt.Errorf("主机未配置SSH用户名")
	}

	if host.AuthType == "password" && host.Password == "" {
		return fmt.Errorf("主机未配置SSH密码")
	}

	if host.AuthType == "key" && host.PrivateKey == "" {
		return fmt.Errorf("主机未配置SSH私钥")
	}

	// 创建SSH客户端
	sshClient, err := ssh.NewSSHClient(host)
	if err != nil {
		return fmt.Errorf("创建SSH连接失败: %v", err)
	}
	defer sshClient.Close()

	// 上传文件到目标主机
	return sshClient.UploadFile(file.Path, targetPath)
}

// 获取分发记录列表
func (h *FileHandler) GetDistributions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	status := c.Query("status")

	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > 100 {
		size = 20
	}

	offset := (page - 1) * size

	query := h.db.Preload("File").Preload("User").Model(&models.FileDistribution{})

	// 权限过滤：普通用户只能看到自己创建的分发记录
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" {
		query = query.Where("created_by = ?", userID)
	}

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var distributions []models.FileDistribution
	if err := query.Offset(offset).Limit(size).Order("created_at DESC").Find(&distributions).Error; err != nil {
		logger.Errorf("获取分发记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分发记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        distributions,
		"total":       total,
		"page":        page,
		"size":        size,
		"total_pages": (total + int64(size) - 1) / int64(size),
	})
}

// 获取分发详情
func (h *FileHandler) GetDistributionDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分发ID"})
		return
	}

	var distribution models.FileDistribution
	if err := h.db.Preload("File").Preload("User").First(&distribution, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "分发记录不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分发记录失败"})
		}
		return
	}

	// 权限检查
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" && distribution.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限查看此分发记录"})
		return
	}

	// 获取分发详情
	var details []models.FileDistributionDetail
	if err := h.db.Preload("Host").Where("distribution_id = ?", distribution.ID).Find(&details).Error; err != nil {
		logger.Errorf("获取分发详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分发详情失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"distribution": distribution,
		"details":      details,
	})
}

// 删除分发记录
func (h *FileHandler) DeleteDistribution(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分发记录ID"})
		return
	}

	var distribution models.FileDistribution
	if err := h.db.First(&distribution, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "分发记录不存在"})
		return
	}

	// 权限检查：只有管理员或分发创建者可以删除
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	if userRole != "admin" && distribution.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "没有权限删除此分发记录"})
		return
	}

	// 开始事务删除分发记录和详情
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除分发详情
	if err := tx.Where("distribution_id = ?", distribution.ID).Delete(&models.FileDistributionDetail{}).Error; err != nil {
		tx.Rollback()
		logger.Errorf("删除分发详情失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除分发详情失败"})
		return
	}

	// 删除分发记录
	if err := tx.Delete(&distribution).Error; err != nil {
		tx.Rollback()
		logger.Errorf("删除分发记录失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除分发记录失败"})
		return
	}

	tx.Commit()

	logger.Infof("删除分发记录成功: ID=%d", distribution.ID)

	// 记录活动
	h.activityService.LogSuccess(c, userID, "delete", "file_distribution", &distribution.ID,
		fmt.Sprintf("删除分发记录 ID:%d", distribution.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "分发记录删除成功",
	})
}
