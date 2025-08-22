package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-devops/internal/models"
)

type TopologyHandler struct {
	db *gorm.DB
}

func NewTopologyHandler(db *gorm.DB) *TopologyHandler {
	return &TopologyHandler{db: db}
}

// 获取完整拓扑树
func (h *TopologyHandler) GetTopologyTree(c *gin.Context) {
	var businesses []models.Business
	if err := h.db.Find(&businesses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取业务列表失败"})
		return
	}

	var topologyTree []models.TopologyNode
	for _, business := range businesses {
		businessNode := models.TopologyNode{
			ID:       business.ID,
			UniqueID: fmt.Sprintf("business-%d", business.ID),
			Name:     business.Name,
			Code:     business.Code,
			Type:     "business",
		}

		// 获取业务下的环境
		var environments []models.Environment
		h.db.Where("business_id = ?", business.ID).Find(&environments)

		for _, env := range environments {
			envNode := models.TopologyNode{
				ID:       env.ID,
				UniqueID: fmt.Sprintf("environment-%d", env.ID),
				Name:     env.Name,
				Code:     env.Code,
				Type:     "environment",
				ParentID: &business.ID,
			}

			// 获取环境下的集群
			var clusters []models.Cluster
			h.db.Where("environment_id = ?", env.ID).Find(&clusters)

			for _, cluster := range clusters {
				clusterNode := models.TopologyNode{
					ID:       cluster.ID,
					UniqueID: fmt.Sprintf("cluster-%d", cluster.ID),
					Name:     cluster.Name,
					Code:     cluster.Code,
					Type:     "cluster",
					ParentID: &env.ID,
				}

				// 获取集群下的主机
				var hostTopologies []models.HostTopology
				h.db.Preload("Host").Where("cluster_id = ?", cluster.ID).Find(&hostTopologies)

				var hostNodes []models.TopologyNode
				onlineCount := 0
				totalCount := len(hostTopologies)

				for _, hostTopo := range hostTopologies {
					hostNode := models.TopologyNode{
						ID:       hostTopo.Host.ID,
						UniqueID: fmt.Sprintf("host-%d", hostTopo.Host.ID),
						Name:     hostTopo.Host.Name,
						Type:     "host",
						ParentID: &cluster.ID,
						HostInfo: &hostTopo.Host,
					}
					hostNodes = append(hostNodes, hostNode)

					if hostTopo.Host.Status == "online" {
						onlineCount++
					}
				}

				clusterNode.Children = hostNodes
				clusterNode.Stats = &models.NodeStats{
					TotalHosts:   totalCount,
					OnlineHosts:  onlineCount,
					OfflineHosts: totalCount - onlineCount,
				}

				envNode.Children = append(envNode.Children, clusterNode)
			}

			businessNode.Children = append(businessNode.Children, envNode)
		}

		topologyTree = append(topologyTree, businessNode)
	}

	c.JSON(http.StatusOK, gin.H{"data": topologyTree})
}

// 业务管理
func (h *TopologyHandler) CreateBusiness(c *gin.Context) {
	var req models.BusinessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 自动生成唯一编码
	code := h.generateUniqueBusinessCode(req.Name)

	business := models.Business{
		Name:        req.Name,
		Code:        code,
		Description: req.Description,
		Owner:       req.Owner,
	}

	if err := h.db.Create(&business).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建业务失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": business})
}

// generateUniqueBusinessCode 生成唯一的业务编码
func (h *TopologyHandler) generateUniqueBusinessCode(name string) string {
	// 基于业务名称生成编码前缀
	var codePrefix string
	runes := []rune(name)
	
	// 提取中文首字母或英文首字母
	for _, r := range runes {
		if r >= 'A' && r <= 'Z' {
			codePrefix += string(r)
		} else if r >= 'a' && r <= 'z' {
			codePrefix += strings.ToUpper(string(r))
		} else if r >= 0x4e00 && r <= 0x9fff { // 中文字符范围
			// 简单处理：取中文字符的拼音首字母（这里简化处理）
			codePrefix += "B" // Business的首字母
			break
		}
		if len(codePrefix) >= 3 {
			break
		}
	}
	
	// 如果没有提取到字母，使用默认前缀
	if codePrefix == "" {
		codePrefix = "BIZ"
	} else if len(codePrefix) < 3 {
		codePrefix = fmt.Sprintf("BIZ%s", codePrefix)
		if len(codePrefix) > 3 {
			codePrefix = codePrefix[:3]
		}
	}
	
	// 生成唯一编码
	for i := 1; i <= 999; i++ {
		code := fmt.Sprintf("%s%03d", codePrefix, i)
		
		// 检查编码是否已存在
		var count int64
		h.db.Model(&models.Business{}).Where("code = ?", code).Count(&count)
		if count == 0 {
			return code
		}
	}
	
	// 如果前999个都被占用，使用时间戳
	return fmt.Sprintf("%s%d", codePrefix, time.Now().Unix()%10000)
}

// generateUniqueEnvironmentCode 生成唯一的环境编码
func (h *TopologyHandler) generateUniqueEnvironmentCode(name string, businessID uint) string {
	// 基于环境名称生成编码前缀
	var codePrefix string
	runes := []rune(name)
	
	// 提取中文首字母或英文首字母
	for _, r := range runes {
		if r >= 'A' && r <= 'Z' {
			codePrefix += string(r)
		} else if r >= 'a' && r <= 'z' {
			codePrefix += strings.ToUpper(string(r))
		} else if r >= 0x4e00 && r <= 0x9fff { // 中文字符范围
			// 简单处理：取中文字符的拼音首字母（这里简化处理）
			codePrefix += "E" // Environment的首字母
			break
		}
		if len(codePrefix) >= 3 {
			break
		}
	}
	
	// 如果没有提取到字母，使用默认前缀
	if codePrefix == "" {
		codePrefix = "ENV"
	} else if len(codePrefix) < 3 {
		codePrefix = fmt.Sprintf("ENV%s", codePrefix)
		if len(codePrefix) > 3 {
			codePrefix = codePrefix[:3]
		}
	}
	
	// 生成唯一编码
	for i := 1; i <= 999; i++ {
		code := fmt.Sprintf("%s%03d", codePrefix, i)
		
		// 检查编码是否已存在（在同一业务下）
		var count int64
		h.db.Model(&models.Environment{}).Where("code = ? AND business_id = ?", code, businessID).Count(&count)
		if count == 0 {
			return code
		}
	}
	
	// 如果前999个都被占用，使用时间戳
	return fmt.Sprintf("%s%d", codePrefix, time.Now().Unix()%10000)
}

// generateUniqueClusterCode 生成唯一的集群编码
func (h *TopologyHandler) generateUniqueClusterCode(name string, environmentID uint) string {
	// 基于集群名称生成编码前缀
	var codePrefix string
	runes := []rune(name)
	
	// 提取中文首字母或英文首字母
	for _, r := range runes {
		if r >= 'A' && r <= 'Z' {
			codePrefix += string(r)
		} else if r >= 'a' && r <= 'z' {
			codePrefix += strings.ToUpper(string(r))
		} else if r >= 0x4e00 && r <= 0x9fff { // 中文字符范围
			// 简单处理：取中文字符的拼音首字母（这里简化处理）
			codePrefix += "C" // Cluster的首字母
			break
		}
		if len(codePrefix) >= 3 {
			break
		}
	}
	
	// 如果没有提取到字母，使用默认前缀
	if codePrefix == "" {
		codePrefix = "CLS"
	} else if len(codePrefix) < 3 {
		codePrefix = fmt.Sprintf("CLS%s", codePrefix)
		if len(codePrefix) > 3 {
			codePrefix = codePrefix[:3]
		}
	}
	
	// 生成唯一编码
	for i := 1; i <= 999; i++ {
		code := fmt.Sprintf("%s%03d", codePrefix, i)
		
		// 检查编码是否已存在（在同一环境下）
		var count int64
		h.db.Model(&models.Cluster{}).Where("code = ? AND environment_id = ?", code, environmentID).Count(&count)
		if count == 0 {
			return code
		}
	}
	
	// 如果前999个都被占用，使用时间戳
	return fmt.Sprintf("%s%d", codePrefix, time.Now().Unix()%10000)
}

func (h *TopologyHandler) GetBusinesses(c *gin.Context) {
	var businesses []models.Business
	if err := h.db.Find(&businesses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取业务列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": businesses})
}

func (h *TopologyHandler) UpdateBusiness(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的业务ID"})
		return
	}

	var req models.BusinessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var business models.Business
	if err := h.db.First(&business, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "业务不存在"})
		return
	}

	// 更新业务信息，但不更新code字段（保持原有的code）
	business.Name = req.Name
	business.Description = req.Description
	business.Owner = req.Owner
	// 不更新 business.Code，保持数据库中的原值

	if err := h.db.Save(&business).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新业务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": business})
}

func (h *TopologyHandler) DeleteBusiness(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的业务ID"})
		return
	}

	// 检查是否有关联的环境
	var envCount int64
	h.db.Model(&models.Environment{}).Where("business_id = ?", id).Count(&envCount)
	if envCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该业务下还有环境，无法删除"})
		return
	}

	if err := h.db.Delete(&models.Business{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除业务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "业务删除成功"})
}

// 环境管理
func (h *TopologyHandler) CreateEnvironment(c *gin.Context) {
	var req models.EnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 自动生成唯一编码
	code := h.generateUniqueEnvironmentCode(req.Name, req.BusinessID)

	environment := models.Environment{
		Name:        req.Name,
		Code:        code,
		BusinessID:  req.BusinessID,
		Description: req.Description,
	}

	if err := h.db.Create(&environment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建环境失败"})
		return
	}

	// 预加载业务信息
	h.db.Preload("Business").First(&environment, environment.ID)

	c.JSON(http.StatusCreated, gin.H{"data": environment})
}

func (h *TopologyHandler) GetEnvironments(c *gin.Context) {
	businessID := c.Query("business_id")
	
	var environments []models.Environment
	query := h.db.Preload("Business")
	
	if businessID != "" {
		query = query.Where("business_id = ?", businessID)
	}
	
	if err := query.Find(&environments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取环境列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": environments})
}

func (h *TopologyHandler) UpdateEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的环境ID"})
		return
	}

	var req models.EnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var environment models.Environment
	if err := h.db.First(&environment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "环境不存在"})
		return
	}

	// 更新环境信息，但不更新code字段
	environment.Name = req.Name
	environment.BusinessID = req.BusinessID
	environment.Description = req.Description
	// 不更新 environment.Code

	if err := h.db.Save(&environment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新环境失败"})
		return
	}

	// 预加载业务信息
	h.db.Preload("Business").First(&environment, environment.ID)

	c.JSON(http.StatusOK, gin.H{"data": environment})
}

func (h *TopologyHandler) DeleteEnvironment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的环境ID"})
		return
	}

	// 检查是否有关联的集群
	var clusterCount int64
	h.db.Model(&models.Cluster{}).Where("environment_id = ?", id).Count(&clusterCount)
	if clusterCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该环境下还有集群，无法删除"})
		return
	}

	if err := h.db.Delete(&models.Environment{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除环境失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "环境删除成功"})
}

// 集群管理
func (h *TopologyHandler) CreateCluster(c *gin.Context) {
	var req models.ClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 自动生成唯一编码
	code := h.generateUniqueClusterCode(req.Name, req.EnvironmentID)

	cluster := models.Cluster{
		Name:          req.Name,
		Code:          code,
		EnvironmentID: req.EnvironmentID,
		Description:   req.Description,
	}

	if err := h.db.Create(&cluster).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建集群失败"})
		return
	}

	// 预加载环境和业务信息
	h.db.Preload("Environment.Business").First(&cluster, cluster.ID)

	c.JSON(http.StatusCreated, gin.H{"data": cluster})
}

func (h *TopologyHandler) GetClusters(c *gin.Context) {
	environmentID := c.Query("environment_id")
	
	var clusters []models.Cluster
	query := h.db.Preload("Environment.Business")
	
	if environmentID != "" {
		query = query.Where("environment_id = ?", environmentID)
	}
	
	if err := query.Find(&clusters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取集群列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": clusters})
}

func (h *TopologyHandler) UpdateCluster(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的集群ID"})
		return
	}

	var req models.ClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cluster models.Cluster
	if err := h.db.First(&cluster, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "集群不存在"})
		return
	}

	// 更新集群信息，但不更新code字段
	cluster.Name = req.Name
	cluster.EnvironmentID = req.EnvironmentID
	cluster.Description = req.Description
	// 不更新 cluster.Code

	if err := h.db.Save(&cluster).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新集群失败"})
		return
	}

	// 预加载环境和业务信息
	h.db.Preload("Environment.Business").First(&cluster, cluster.ID)

	c.JSON(http.StatusOK, gin.H{"data": cluster})
}

func (h *TopologyHandler) DeleteCluster(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的集群ID"})
		return
	}

	// 检查是否有关联的主机
	var hostCount int64
	h.db.Model(&models.HostTopology{}).Where("cluster_id = ?", id).Count(&hostCount)
	if hostCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该集群下还有主机，无法删除"})
		return
	}

	if err := h.db.Delete(&models.Cluster{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除集群失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "集群删除成功"})
}

// 主机拓扑管理
func (h *TopologyHandler) AssignHostToCluster(c *gin.Context) {
	var req models.HostTopologyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查主机是否已经分配到其他集群
	var existingTopology models.HostTopology
	if err := h.db.Where("host_id = ?", req.HostID).First(&existingTopology).Error; err == nil {
		// 更新现有分配
		existingTopology.ClusterID = req.ClusterID
		if err := h.db.Save(&existingTopology).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新主机分配失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": existingTopology, "message": "主机分配更新成功"})
		return
	}

	// 创建新的拓扑关联
	hostTopology := models.HostTopology{
		HostID:    req.HostID,
		ClusterID: req.ClusterID,
	}

	if err := h.db.Create(&hostTopology).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "分配主机到集群失败"})
		return
	}

	// 预加载相关信息
	h.db.Preload("Host").Preload("Cluster.Environment.Business").First(&hostTopology, hostTopology.ID)

	c.JSON(http.StatusCreated, gin.H{"data": hostTopology, "message": "主机分配成功"})
}

func (h *TopologyHandler) RemoveHostFromCluster(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("hostId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	if err := h.db.Where("host_id = ?", hostID).Delete(&models.HostTopology{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "移除主机失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "主机移除成功"})
}

func (h *TopologyHandler) GetUnassignedHosts(c *gin.Context) {
	var hosts []models.Host
	
	// 查找未分配到任何集群的主机
	if err := h.db.Where("id NOT IN (?)", 
		h.db.Table("host_topologies").Select("host_id")).Find(&hosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取未分配主机失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hosts})
}

func (h *TopologyHandler) GetHostsByCluster(c *gin.Context) {
	clusterID, err := strconv.ParseUint(c.Param("clusterId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的集群ID"})
		return
	}

	var hostTopologies []models.HostTopology
	if err := h.db.Preload("Host").Where("cluster_id = ?", clusterID).Find(&hostTopologies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取集群主机失败"})
		return
	}

	var hosts []models.Host
	for _, ht := range hostTopologies {
		hosts = append(hosts, ht.Host)
	}

	c.JSON(http.StatusOK, gin.H{"data": hosts})
}
