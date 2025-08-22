package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"go-devops/internal/logger"
	"go-devops/internal/models"
	"go-devops/internal/ssh"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gorm.io/gorm"
)

type HostHandler struct {
	db *gorm.DB
}

func NewHostHandler(db *gorm.DB) *HostHandler {
	return &HostHandler{db: db}
}

// 获取主机列表
func (h *HostHandler) GetHosts(c *gin.Context) {
	var hosts []models.Host
	if err := h.db.Find(&hosts).Error; err != nil {
		logger.Errorf("获取主机列表失败: %v", err)
		logger.LogDBOperation("select", "hosts", false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取主机列表失败"})
		return
	}

	c.JSON(http.StatusOK, hosts)
}

// 创建主机
func (h *HostHandler) CreateHost(c *gin.Context) {
	var req models.HostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("创建主机请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 创建主机对象
	host := models.Host{
		Name:        req.Name,
		IP:          req.IP,
		Port:        req.Port,
		OS:          req.OS,
		Description: req.Description,
		Tags:        req.Tags,
		AuthType:    req.AuthType,
		Username:    req.Username,
		Password:    req.Password,
		PrivateKey:  req.PrivateKey,
		Passphrase:  req.Passphrase,
	}

	// 设置默认值
	if host.Port == 0 {
		host.Port = 22
	}
	if host.AuthType == "" {
		host.AuthType = "password"
	}

	if err := h.db.Create(&host).Error; err != nil {
		logger.Errorf("创建主机失败: %v", err)
		logger.LogDBOperation("create", "hosts", false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建主机失败"})
		return
	}

	logger.Infof("创建主机成功: %s (%s)", host.Name, host.IP)
	logger.LogDBOperation("create", "hosts", true, "")

	c.JSON(http.StatusCreated, host)
}

// 获取单个主机
func (h *HostHandler) GetHost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var host models.Host
	if err := h.db.First(&host, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取主机信息失败"})
		}
		return
	}

	// 查询主机的拓扑信息
	var topology models.HostTopology
	var topologyInfo map[string]interface{}

	if err := h.db.Preload("Cluster.Environment.Business").Where("host_id = ?", host.ID).First(&topology).Error; err == nil {
		topologyInfo = map[string]interface{}{
			"business":         topology.Cluster.Environment.Business.Name,
			"environment":      topology.Cluster.Environment.Name,
			"cluster":          topology.Cluster.Name,
			"business_code":    topology.Cluster.Environment.Business.Code,
			"environment_code": topology.Cluster.Environment.Code,
			"cluster_code":     topology.Cluster.Code,
		}
	}

	// 构造响应数据
	response := map[string]interface{}{
		"id":          host.ID,
		"name":        host.Name,
		"ip":          host.IP,
		"port":        host.Port,
		"os":          host.OS,
		"status":      host.Status,
		"description": host.Description,
		"tags":        host.Tags,
		"auth_type":   host.AuthType,
		"username":    host.Username,
		"password":    host.Password,
		"private_key": host.PrivateKey,
		"passphrase":  host.Passphrase,
		"created_at":  host.CreatedAt,
		"updated_at":  host.UpdatedAt,
		"topology":    topologyInfo,
	}

	c.JSON(http.StatusOK, response)
}

// 更新主机
func (h *HostHandler) UpdateHost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var host models.Host
	if err := h.db.First(&host, uint(id)).Error; err != nil {
		logger.Warnf("更新主机时未找到主机: ID=%d", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	var req models.HostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("更新主机请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	// 更新字段
	updateData := map[string]interface{}{
		"name":        req.Name,
		"ip":          req.IP,
		"port":        req.Port,
		"os":          req.OS,
		"description": req.Description,
		"tags":        req.Tags,
		"auth_type":   req.AuthType,
		"username":    req.Username,
	}

	// 只有在提供了新密码时才更新
	if req.Password != "" {
		updateData["password"] = req.Password
	}
	if req.PrivateKey != "" {
		updateData["private_key"] = req.PrivateKey
	}
	if req.Passphrase != "" {
		updateData["passphrase"] = req.Passphrase
	}

	if err := h.db.Model(&host).Updates(updateData).Error; err != nil {
		logger.Errorf("更新主机失败: %v", err)
		logger.LogDBOperation("update", "hosts", false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新主机失败"})
		return
	}

	logger.Infof("更新主机成功: %s (%s)", host.Name, host.IP)
	logger.LogDBOperation("update", "hosts", true, "")

	// 重新查询更新后的数据
	h.db.First(&host, uint(id))
	c.JSON(http.StatusOK, host)
}

// 批量检查所有主机状态
func (h *HostHandler) CheckAllHostsStatus(c *gin.Context) {
	var hosts []models.Host
	if err := h.db.Find(&hosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取主机列表失败"})
		return
	}

	results := make([]gin.H, 0)

	for _, host := range hosts {
		var status string
		var message string

		if host.Username == "" && host.Password == "" && host.PrivateKey == "" {
			status = "unknown"
			message = "未配置SSH认证信息"
		} else {
			testResult, err := ssh.TestSSHConnection(&host)
			if err != nil {
				status = "offline"
				message = fmt.Sprintf("连接测试失败: %v", err)
			} else if testResult.Success {
				status = "online"
				message = fmt.Sprintf("连接成功，延迟: %s", testResult.Latency)
			} else {
				status = "offline"
				message = testResult.Message
			}
		}

		// 更新主机状态
		h.db.Model(&host).Update("status", status)

		results = append(results, gin.H{
			"host_id": host.ID,
			"name":    host.Name,
			"status":  status,
			"message": message,
		})
	}

	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "check_all_hosts_status", "hosts", true, fmt.Sprintf("检查了%d台主机状态", len(hosts)))

	c.JSON(http.StatusOK, gin.H{
		"message": "批量检查完成",
		"results": results,
	})
}

// 获取定时检查配置
func (h *HostHandler) GetScheduleConfig(c *gin.Context) {
	// 这里可以从数据库或配置文件读取，暂时返回默认值
	config := gin.H{
		"enabled":  true,
		"interval": 5, // 分钟
	}
	c.JSON(http.StatusOK, config)
}

// 更新定时检查配置
func (h *HostHandler) UpdateScheduleConfig(c *gin.Context) {
	var config struct {
		Enabled  bool `json:"enabled"`
		Interval int  `json:"interval"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	if config.Interval < 1 {
		config.Interval = 5 // 最小间隔5分钟
	}

	// 这里可以保存到数据库或配置文件
	logger.Infof("定时检查配置更新: 启用=%t, 间隔=%d分钟", config.Enabled, config.Interval)

	c.JSON(http.StatusOK, gin.H{
		"message": "配置更新成功",
		"config": gin.H{
			"enabled":  config.Enabled,
			"interval": config.Interval,
		},
	})
}

// 删除主机
func (h *HostHandler) DeleteHost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	if err := h.db.Delete(&models.Host{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除主机失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "主机删除成功"})
}

// 检查主机状态
func (h *HostHandler) CheckHostStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var host models.Host
	if err := h.db.First(&host, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	logger.Infof("开始检查主机状态: %s (%s)", host.Name, host.IP)

	// 实际的SSH连接测试
	var status string
	var message string

	if host.Username == "" && host.Password == "" && host.PrivateKey == "" {
		// 没有认证信息，无法进行连接测试
		status = "unknown"
		message = "主机未配置SSH认证信息，无法进行连接测试"
		logger.Warnf("主机 %s 未配置认证信息", host.Name)
	} else {
		// 使用SSH进行连接测试
		testResult, err := ssh.TestSSHConnection(&host)
		if err != nil {
			logger.Errorf("SSH连接测试出错: %v", err)
			status = "offline"
			message = fmt.Sprintf("连接测试失败: %v", err)
		} else if testResult.Success {
			status = "online"
			message = fmt.Sprintf("连接测试成功，延迟: %s", testResult.Latency)
			logger.Infof("主机 %s SSH连接测试成功", host.Name)
		} else {
			status = "offline"
			message = testResult.Message
			logger.Warnf("主机 %s SSH连接测试失败: %s", host.Name, testResult.Message)
		}
	}

	// 更新主机状态
	if err := h.db.Model(&host).Update("status", status).Error; err != nil {
		logger.Errorf("更新主机状态失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新主机状态失败"})
		return
	}

	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "check_host_status", "hosts", status == "online", message)

	c.JSON(http.StatusOK, gin.H{
		"host_id": host.ID,
		"status":  status,
		"message": message,
	})
}

// 测试SSH连接
func (h *HostHandler) TestSSHConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var host models.Host
	if err := h.db.First(&host, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	logger.Infof("测试SSH连接: %s (%s)", host.Name, host.IP)

	// 检查认证配置
	if host.Username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "主机未配置用户名",
		})
		return
	}

	if host.AuthType == "password" && host.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "密码认证需要配置密码",
		})
		return
	}

	if host.AuthType == "key" && host.PrivateKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "密钥认证需要配置私钥",
		})
		return
	}

	// 执行SSH连接测试
	testResult, err := ssh.TestSSHConnection(&host)
	if err != nil {
		logger.Errorf("SSH连接测试出错: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": fmt.Sprintf("测试出错: %v", err),
		})
		return
	}

	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "test_ssh_connection", "hosts", testResult.Success, testResult.Message)

	c.JSON(http.StatusOK, testResult)
}

// 获取主机系统信息
func (h *HostHandler) GetSystemInfo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var host models.Host
	if err := h.db.First(&host, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	logger.Infof("获取主机系统信息: %s (%s)", host.Name, host.IP)

	// 检查认证配置
	if host.Username == "" || (host.AuthType == "password" && host.Password == "") || (host.AuthType == "key" && host.PrivateKey == "") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "主机未配置完整的SSH认证信息",
		})
		return
	}

	// 获取系统信息
	systemInfo, err := ssh.GetSystemInfo(&host)
	if err != nil {
		logger.Errorf("获取系统信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取系统信息失败: %v", err),
		})
		return
	}

	logger.Infof("成功获取主机 %s 的系统信息", host.Name)
	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "get_system_info", "hosts", true, "获取系统信息成功")

	c.JSON(http.StatusOK, gin.H{
		"host_id":     host.ID,
		"host_name":   host.Name,
		"system_info": systemInfo,
	})
}

// 批量导入主机
func (h *HostHandler) BatchImportHosts(c *gin.Context) {
	var req models.BatchHostImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("批量导入主机请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	logger.Infof("开始批量导入主机，数量: %d", len(req.Hosts))

	var successHosts []models.Host
	var failedHosts []models.BatchImportError
	successCount := 0
	failedCount := 0

	for i, hostReq := range req.Hosts {
		// 创建主机对象
		host := models.Host{
			Name:        hostReq.Name,
			IP:          hostReq.IP,
			Port:        hostReq.Port,
			OS:          hostReq.OS,
			Description: hostReq.Description,
			Tags:        hostReq.Tags,
			AuthType:    hostReq.AuthType,
			Username:    hostReq.Username,
			Password:    hostReq.Password,
			PrivateKey:  hostReq.PrivateKey,
			Passphrase:  hostReq.Passphrase,
		}

		// 设置默认值
		if host.Port == 0 {
			host.Port = 22
		}
		if host.AuthType == "" {
			host.AuthType = "password"
		}

		// 验证必填字段
		if host.Name == "" || host.IP == "" {
			failedHosts = append(failedHosts, models.BatchImportError{
				Index: i,
				Host:  hostReq,
				Error: "主机名称和IP地址为必填项",
			})
			failedCount++
			continue
		}

		// 检查IP是否已存在
		var existingHost models.Host
		if err := h.db.Where("ip = ?", host.IP).First(&existingHost).Error; err == nil {
			failedHosts = append(failedHosts, models.BatchImportError{
				Index: i,
				Host:  hostReq,
				Error: fmt.Sprintf("IP地址 %s 已存在", host.IP),
			})
			failedCount++
			continue
		}

		// 创建主机
		if err := h.db.Create(&host).Error; err != nil {
			logger.Errorf("批量导入主机失败: %v, IP: %s", err, host.IP)
			failedHosts = append(failedHosts, models.BatchImportError{
				Index: i,
				Host:  hostReq,
				Error: fmt.Sprintf("数据库创建失败: %v", err),
			})
			failedCount++
		} else {
			successHosts = append(successHosts, host)
			successCount++
			logger.Infof("批量导入主机成功: %s (%s)", host.Name, host.IP)
		}
	}

	logger.Infof("批量导入完成，成功: %d, 失败: %d", successCount, failedCount)
	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "batch_import_hosts", "hosts", successCount > 0, fmt.Sprintf("导入%d台主机，成功%d台", len(req.Hosts), successCount))

	response := models.BatchHostImportResponse{
		Success:      successCount,
		Failed:       failedCount,
		Total:        len(req.Hosts),
		SuccessHosts: successHosts,
		FailedHosts:  failedHosts,
	}

	c.JSON(http.StatusOK, response)
}

// 更新主机认证信息
func (h *HostHandler) UpdateHostAuth(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var host models.Host
	if err := h.db.First(&host, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	var req models.HostPasswordUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("更新主机认证信息请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	logger.Infof("更新主机认证信息: %s (%s)", host.Name, host.IP)

	// 构建更新数据
	updateData := make(map[string]interface{})

	if req.AuthType != "" {
		updateData["auth_type"] = req.AuthType
	}
	if req.Password != "" {
		updateData["password"] = req.Password
	}
	if req.PrivateKey != "" {
		updateData["private_key"] = req.PrivateKey
	}
	if req.Passphrase != "" {
		updateData["passphrase"] = req.Passphrase
	}

	if len(updateData) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有提供要更新的认证信息"})
		return
	}

	// 更新数据库
	if err := h.db.Model(&host).Updates(updateData).Error; err != nil {
		logger.Errorf("更新主机认证信息失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新认证信息失败"})
		return
	}

	logger.Infof("主机认证信息更新成功: %s", host.Name)
	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "update_host_auth", "hosts", true, "更新认证信息成功")

	c.JSON(http.StatusOK, gin.H{
		"message": "认证信息更新成功",
		"host_id": host.ID,
	})
}

// 批量主机操作
func (h *HostHandler) BatchHostOperation(c *gin.Context) {
	var req models.BatchHostOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warnf("批量主机操作请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	logger.Infof("开始批量主机操作: %s, 主机数量: %d", req.Operation, len(req.HostIDs))

	var results []models.BatchOperationResult
	successCount := 0
	failedCount := 0

	for _, hostID := range req.HostIDs {
		var host models.Host
		if err := h.db.First(&host, hostID).Error; err != nil {
			results = append(results, models.BatchOperationResult{
				HostID:  hostID,
				Success: false,
				Message: "主机不存在",
			})
			failedCount++
			continue
		}

		switch req.Operation {
		case "test_connection":
			result := h.batchTestConnection(&host)
			results = append(results, result)
			if result.Success {
				successCount++
			} else {
				failedCount++
			}

		case "update_status":
			result := h.batchUpdateStatus(&host)
			results = append(results, result)
			if result.Success {
				successCount++
			} else {
				failedCount++
			}

		case "get_system_info":
			result := h.batchGetSystemInfo(&host)
			results = append(results, result)
			if result.Success {
				successCount++
			} else {
				failedCount++
			}

		default:
			results = append(results, models.BatchOperationResult{
				HostID:  hostID,
				Success: false,
				Message: "不支持的操作类型",
			})
			failedCount++
		}
	}

	logger.Infof("批量操作完成: %s, 成功: %d, 失败: %d", req.Operation, successCount, failedCount)
	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "batch_host_operation", "hosts", successCount > 0, fmt.Sprintf("批量%s操作，成功%d台", req.Operation, successCount))

	response := models.BatchOperationResponse{
		Success: successCount,
		Failed:  failedCount,
		Total:   len(req.HostIDs),
		Results: results,
	}

	c.JSON(http.StatusOK, response)
}

// 批量测试连接
func (h *HostHandler) batchTestConnection(host *models.Host) models.BatchOperationResult {
	if host.Username == "" || (host.AuthType == "password" && host.Password == "") || (host.AuthType == "key" && host.PrivateKey == "") {
		return models.BatchOperationResult{
			HostID:  host.ID,
			Success: false,
			Message: "主机未配置完整的SSH认证信息",
		}
	}

	testResult, err := ssh.TestSSHConnection(host)
	if err != nil {
		return models.BatchOperationResult{
			HostID:  host.ID,
			Success: false,
			Message: fmt.Sprintf("测试出错: %v", err),
		}
	}

	return models.BatchOperationResult{
		HostID:  host.ID,
		Success: testResult.Success,
		Message: testResult.Message,
		Data:    testResult,
	}
}

// 批量更新状态
func (h *HostHandler) batchUpdateStatus(host *models.Host) models.BatchOperationResult {
	var status string
	var message string

	if host.Username == "" && host.Password == "" && host.PrivateKey == "" {
		status = "unknown"
		message = "主机未配置SSH认证信息"
	} else {
		testResult, err := ssh.TestSSHConnection(host)
		if err != nil {
			status = "offline"
			message = fmt.Sprintf("连接测试失败: %v", err)
		} else if testResult.Success {
			status = "online"
			message = "连接测试成功"
		} else {
			status = "offline"
			message = testResult.Message
		}
	}

	// 更新数据库状态
	if err := h.db.Model(host).Update("status", status).Error; err != nil {
		return models.BatchOperationResult{
			HostID:  host.ID,
			Success: false,
			Message: fmt.Sprintf("更新状态失败: %v", err),
		}
	}

	return models.BatchOperationResult{
		HostID:  host.ID,
		Success: status == "online",
		Message: message,
		Data: gin.H{
			"status": status,
		},
	}
}

// 批量获取系统信息
func (h *HostHandler) batchGetSystemInfo(host *models.Host) models.BatchOperationResult {
	if host.Username == "" || (host.AuthType == "password" && host.Password == "") || (host.AuthType == "key" && host.PrivateKey == "") {
		return models.BatchOperationResult{
			HostID:  host.ID,
			Success: false,
			Message: "主机未配置完整的SSH认证信息",
		}
	}

	systemInfo, err := ssh.GetSystemInfo(host)
	if err != nil {
		return models.BatchOperationResult{
			HostID:  host.ID,
			Success: false,
			Message: fmt.Sprintf("获取系统信息失败: %v", err),
		}
	}

	return models.BatchOperationResult{
		HostID:  host.ID,
		Success: true,
		Message: "获取系统信息成功",
		Data:    systemInfo,
	}
}

// CSV批量导入主机
func (h *HostHandler) BatchImportHostsFromCSV(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Warnf("获取上传文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的CSV文件"})
		return
	}
	defer file.Close()

	// 验证文件类型
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传CSV格式的文件"})
		return
	}

	logger.Infof("开始解析CSV文件: %s, 大小: %d bytes", header.Filename, header.Size)

	// 读取文件内容并处理编码
	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.Errorf("读取CSV文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("读取CSV文件失败: %v", err)})
		return
	}

	// 处理编码问题
	processedContent, err := h.processCSVEncoding(fileContent)
	if err != nil {
		logger.Errorf("处理CSV文件编码失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("处理CSV文件编码失败: %v", err)})
		return
	}

	// 解析CSV文件
	hosts, parseErrors, err := h.parseCSVFile(bytes.NewReader(processedContent))
	if err != nil {
		logger.Errorf("解析CSV文件失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("CSV文件解析失败: %v", err)})
		return
	}

	if len(hosts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV文件中没有有效的主机数据"})
		return
	}

	logger.Infof("CSV文件解析完成，有效主机数量: %d, 解析错误数量: %d", len(hosts), len(parseErrors))

	// 批量导入主机
	var successHosts []models.Host
	var failedHosts []models.BatchImportError
	successCount := 0
	failedCount := 0

	// 添加解析错误到失败列表
	for _, parseError := range parseErrors {
		failedHosts = append(failedHosts, parseError)
		failedCount++
	}

	for i, hostReq := range hosts {
		// 创建主机对象
		host := models.Host{
			Name:        hostReq.Name,
			IP:          hostReq.IP,
			Port:        hostReq.Port,
			OS:          hostReq.OS,
			Description: hostReq.Description,
			Tags:        hostReq.Tags,
			AuthType:    hostReq.AuthType,
			Username:    hostReq.Username,
			Password:    hostReq.Password,
			PrivateKey:  hostReq.PrivateKey,
			Passphrase:  hostReq.Passphrase,
		}

		// 设置默认值
		if host.Port == 0 {
			host.Port = 22
		}
		if host.AuthType == "" {
			host.AuthType = "password"
		}

		// 验证必填字段
		if host.Name == "" || host.IP == "" {
			failedHosts = append(failedHosts, models.BatchImportError{
				Index: i,
				Host:  hostReq,
				Error: "主机名称和IP地址为必填项",
			})
			failedCount++
			continue
		}

		// 检查IP是否已存在
		var existingHost models.Host
		if err := h.db.Where("ip = ?", host.IP).First(&existingHost).Error; err == nil {
			failedHosts = append(failedHosts, models.BatchImportError{
				Index: i,
				Host:  hostReq,
				Error: fmt.Sprintf("IP地址 %s 已存在", host.IP),
			})
			failedCount++
			continue
		}

		// 创建主机
		if err := h.db.Create(&host).Error; err != nil {
			logger.Errorf("CSV批量导入主机失败: %v, IP: %s", err, host.IP)
			failedHosts = append(failedHosts, models.BatchImportError{
				Index: i,
				Host:  hostReq,
				Error: fmt.Sprintf("数据库创建失败: %v", err),
			})
			failedCount++
		} else {
			successHosts = append(successHosts, host)
			successCount++
			logger.Infof("CSV批量导入主机成功: %s (%s)", host.Name, host.IP)
		}
	}

	logger.Infof("CSV批量导入完成，成功: %d, 失败: %d", successCount, failedCount)
	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "batch_import_hosts_csv", "hosts", successCount > 0, fmt.Sprintf("CSV导入%d台主机，成功%d台", len(hosts)+len(parseErrors), successCount))

	response := models.BatchHostImportResponse{
		Success:      successCount,
		Failed:       failedCount,
		Total:        len(hosts) + len(parseErrors),
		SuccessHosts: successHosts,
		FailedHosts:  failedHosts,
	}

	c.JSON(http.StatusOK, response)
}

// 解析CSV文件
func (h *HostHandler) parseCSVFile(file io.Reader) ([]models.HostRequest, []models.BatchImportError, error) {
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允许不同行有不同的字段数量

	var hosts []models.HostRequest
	var parseErrors []models.BatchImportError
	lineNumber := 0

	// 读取所有记录
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("读取CSV文件失败: %v", err)
	}

	if len(records) == 0 {
		return nil, nil, fmt.Errorf("CSV文件为空")
	}

	// 解析表头
	headers := records[0]
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.TrimSpace(strings.ToLower(header))] = i
	}

	// 验证必需的列
	requiredColumns := []string{"name", "ip"}
	for _, col := range requiredColumns {
		if _, exists := headerMap[col]; !exists {
			return nil, nil, fmt.Errorf("CSV文件缺少必需的列: %s", col)
		}
	}

	// 解析数据行
	for i := 1; i < len(records); i++ {
		lineNumber = i + 1
		record := records[i]

		// 跳过空行
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		host, err := h.parseCSVRecord(record, headerMap, lineNumber)
		if err != nil {
			parseErrors = append(parseErrors, models.BatchImportError{
				Index: i - 1,
				Host:  models.HostRequest{},
				Error: fmt.Sprintf("第%d行解析错误: %v", lineNumber, err),
			})
			continue
		}

		hosts = append(hosts, host)
	}

	return hosts, parseErrors, nil
}

// 解析CSV记录
func (h *HostHandler) parseCSVRecord(record []string, headerMap map[string]int, lineNumber int) (models.HostRequest, error) {
	var host models.HostRequest

	// 获取字段值的辅助函数
	getField := func(fieldName string) string {
		if index, exists := headerMap[fieldName]; exists && index < len(record) {
			return strings.TrimSpace(record[index])
		}
		return ""
	}

	// 解析必填字段
	host.Name = getField("name")
	host.IP = getField("ip")

	if host.Name == "" {
		return host, fmt.Errorf("主机名称不能为空")
	}
	if host.IP == "" {
		return host, fmt.Errorf("IP地址不能为空")
	}

	// 解析端口
	portStr := getField("port")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return host, fmt.Errorf("端口格式错误: %s", portStr)
		}
		host.Port = port
	} else {
		host.Port = 22 // 默认端口
	}

	// 解析其他字段
	host.OS = getField("os")
	host.Description = getField("description")
	host.Tags = getField("tags")
	host.AuthType = getField("auth_type")
	host.Username = getField("username")
	host.Password = getField("password")
	host.PrivateKey = getField("private_key")
	host.Passphrase = getField("passphrase")

	// 设置默认认证类型
	if host.AuthType == "" {
		host.AuthType = "password"
	}

	// 验证认证类型
	validAuthTypes := map[string]bool{"password": true, "key": true}
	if !validAuthTypes[host.AuthType] {
		return host, fmt.Errorf("无效的认证类型: %s，支持的类型: password, key", host.AuthType)
	}

	return host, nil
}

// 处理CSV文件编码
func (h *HostHandler) processCSVEncoding(content []byte) ([]byte, error) {
	// 检测并移除BOM
	if len(content) >= 3 && content[0] == 0xEF && content[1] == 0xBB && content[2] == 0xBF {
		// UTF-8 BOM，直接移除
		return content[3:], nil
	}

	// 检查是否为有效的UTF-8
	if utf8.Valid(content) {
		return content, nil
	}

	// 尝试从GBK转换为UTF-8
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8Content, _, err := transform.Bytes(decoder, content)
	if err != nil {
		// 如果GBK转换失败，尝试GB18030
		decoder = simplifiedchinese.GB18030.NewDecoder()
		utf8Content, _, err = transform.Bytes(decoder, content)
		if err != nil {
			// 如果都失败了，返回原内容
			logger.Warnf("CSV文件编码转换失败，使用原始内容: %v", err)
			return content, nil
		}
	}

	logger.Infof("CSV文件编码已转换为UTF-8")
	return utf8Content, nil
}

// 下载CSV模板
func (h *HostHandler) DownloadCSVTemplate(c *gin.Context) {
	// CSV模板内容
	template := `name,ip,port,os,description,tags,auth_type,username,password
Web服务器-01,192.168.100.10,22,Ubuntu 20.04,生产环境Web服务器,web;production,password,root,your_password
数据库服务器-01,192.168.100.11,22,CentOS 8,MySQL主数据库,database;mysql,password,mysql,your_password
应用服务器-01,192.168.100.12,2222,Ubuntu 18.04,开发环境应用服务器,app;development,password,app,your_password`

	// 添加BOM以支持Excel正确显示中文
	bom := []byte{0xEF, 0xBB, 0xBF}
	content := append(bom, []byte(template)...)

	// 设置响应头
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=host_import_template.csv")
	c.Header("Content-Length", strconv.Itoa(len(content)))

	// 写入内容
	c.Data(http.StatusOK, "text/csv", content)

	logger.Infof("用户下载CSV导入模板")
	logger.LogUserAction(c.GetUint("user_id"), c.GetString("username"), "download_csv_template", "hosts", true, "下载CSV导入模板")
}
