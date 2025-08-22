package database

import (
	"go-devops/internal/logger"
	"go-devops/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedData 初始化种子数据
func SeedData(db *gorm.DB) error {
	// 检查是否已有管理员用户
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)

	if adminCount == 0 {
		logger.Info("创建默认管理员账户")
		// 创建默认管理员账户
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			logger.Errorf("管理员密码加密失败: %v", err)
			return err
		}

		admin := models.User{
			Username: "admin",
			Email:    "admin@example.com",
			Password: string(hashedPassword),
			Role:     "admin",
		}

		if err := db.Create(&admin).Error; err != nil {
			logger.Errorf("创建管理员账户失败: %v", err)
			logger.LogDBOperation("create", "users", false, err.Error())
			return err
		}

		logger.Info("默认管理员账户创建成功")
		logger.LogDBOperation("create", "users", true, "")
	}

	// 创建示例主机
	var hostCount int64
	db.Model(&models.Host{}).Count(&hostCount)

	if hostCount == 0 {
		logger.Info("创建示例主机数据")
		hosts := []models.Host{
			{
				Name:        "Web服务器-01",
				IP:          "192.168.1.100",
				Port:        22,
				OS:          "Linux",
				Status:      "online",
				Description: "Web应用服务器",
				Tags:        "web,production",
				AuthType:    "password",
				Username:    "root",
				Password:    "123456",
			},
			{
				Name:        "数据库",
				IP:          "192.168.200.182",
				Port:        22,
				OS:          "Rocky Linux",
				Status:      "online",
				Description: "MySQL数据库服务器",
				Tags:        "database,production",
				AuthType:    "password",
				Username:    "root",
				Password:    "123456",
			},
			{
				Name:        "测试服务器",
				IP:          "192.168.1.101",
				Port:        22,
				OS:          "Linux",
				Status:      "offline",
				Description: "测试环境服务器",
				Tags:        "test,development",
				AuthType:    "key",
				Username:    "ubuntu",
				PrivateKey:  "-----BEGIN OPENSSH PRIVATE KEY-----\n示例密钥内容\n-----END OPENSSH PRIVATE KEY-----",
			},
		}

		for _, host := range hosts {
			if err := db.Create(&host).Error; err != nil {
				logger.Errorf("创建示例主机失败: %v", err)
				logger.LogDBOperation("create", "hosts", false, err.Error())
			} else {
				logger.Infof("创建示例主机成功: %s", host.Name)
				logger.LogDBOperation("create", "hosts", true, "")
			}
		}
	}

	// 创建示例脚本
	var scriptCount int64
	db.Model(&models.Script{}).Count(&scriptCount)

	if scriptCount == 0 {
		// 获取管理员用户ID
		var admin models.User
		db.Where("username = ?", "admin").First(&admin)

		scripts := []models.Script{
			{
				Name:        "系统信息检查",
				Type:        "shell",
				Description: "获取系统基本信息",
				Content: `#!/bin/bash
echo "=== 系统信息 ==="
uname -a
echo ""
echo "=== 内存使用情况 ==="
free -h
echo ""
echo "=== 磁盘使用情况 ==="
df -h
echo ""
echo "=== CPU信息 ==="
cat /proc/cpuinfo | grep "model name" | head -1`,
				CreatedBy: admin.ID,
			},
			{
				Name:        "服务状态检查",
				Type:        "shell",
				Description: "检查常用服务状态",
				Content: `#!/bin/bash
echo "=== 服务状态检查 ==="
services=("nginx" "mysql" "redis" "docker")

for service in "${services[@]}"; do
    if systemctl is-active --quiet $service; then
        echo "$service: 运行中"
    else
        echo "$service: 已停止"
    fi
done`,
				CreatedBy: admin.ID,
			},
			{
				Name:        "日志清理",
				Type:        "shell",
				Description: "清理系统日志文件",
				Content: `#!/bin/bash
echo "开始清理日志文件..."

# 清理7天前的日志
find /var/log -name "*.log" -mtime +7 -exec rm -f {} \;
find /var/log -name "*.log.*" -mtime +7 -exec rm -f {} \;

echo "日志清理完成"
du -sh /var/log`,
				CreatedBy: admin.ID,
			},
		}

		logger.Info("创建示例脚本数据")
		for _, script := range scripts {
			if err := db.Create(&script).Error; err != nil {
				logger.Errorf("创建示例脚本失败: %v, 脚本: %s", err, script.Name)
				logger.LogDBOperation("create", "scripts", false, err.Error())
			} else {
				logger.Infof("创建示例脚本成功: %s", script.Name)
				logger.LogDBOperation("create", "scripts", true, "")
			}
		}
	}

	// 创建拓扑示例数据
	var businessCount int64
	db.Model(&models.Business{}).Count(&businessCount)

	if businessCount == 0 {
		logger.Info("创建拓扑示例数据")

		// 创建业务
		business := models.Business{
			Name:        "电商平台",
			Code:        "ecommerce",
			Description: "在线电商业务平台",
			Owner:       "张三",
		}
		if err := db.Create(&business).Error; err != nil {
			logger.Errorf("创建示例业务失败: %v", err)
		} else {
			logger.Info("创建示例业务成功: 电商平台")

			// 创建环境
			environments := []models.Environment{
				{
					Name:        "生产环境",
					Code:        "prod",
					BusinessID:  business.ID,
					Description: "生产环境",
				},
				{
					Name:        "测试环境",
					Code:        "test",
					BusinessID:  business.ID,
					Description: "测试环境",
				},
			}

			for _, env := range environments {
				if err := db.Create(&env).Error; err != nil {
					logger.Errorf("创建示例环境失败: %v", err)
				} else {
					logger.Infof("创建示例环境成功: %s", env.Name)

					// 创建集群
					clusters := []models.Cluster{
						{
							Name:          "Web集群",
							Code:          "web-cluster",
							EnvironmentID: env.ID,
							Description:   "Web应用服务器集群",
						},
						{
							Name:          "数据库集群",
							Code:          "db-cluster",
							EnvironmentID: env.ID,
							Description:   "数据库服务器集群",
						},
					}

					for _, cluster := range clusters {
						if err := db.Create(&cluster).Error; err != nil {
							logger.Errorf("创建示例集群失败: %v", err)
						} else {
							logger.Infof("创建示例集群成功: %s", cluster.Name)

							// 将示例主机分配到对应集群
							if env.Code == "prod" {
								if cluster.Code == "web-cluster" {
									// Web服务器分配到Web集群
									var webHost models.Host
									if err := db.Where("name = ?", "Web服务器-01").First(&webHost).Error; err == nil {
										hostTopology := models.HostTopology{
											HostID:    webHost.ID,
											ClusterID: cluster.ID,
										}
										if err := db.Create(&hostTopology).Error; err != nil {
											logger.Errorf("创建主机拓扑关联失败: %v", err)
										} else {
											logger.Infof("主机 %s 已分配到集群 %s", webHost.Name, cluster.Name)
										}
									}
								} else if cluster.Code == "db-cluster" {
									// 数据库服务器分配到数据库集群
									var dbHost models.Host
									if err := db.Where("name = ?", "数据库").First(&dbHost).Error; err == nil {
										hostTopology := models.HostTopology{
											HostID:    dbHost.ID,
											ClusterID: cluster.ID,
										}
										if err := db.Create(&hostTopology).Error; err != nil {
											logger.Errorf("创建主机拓扑关联失败: %v", err)
										} else {
											logger.Infof("主机 %s 已分配到集群 %s", dbHost.Name, cluster.Name)
										}
									}
								}
							} else if env.Code == "test" && cluster.Code == "web-cluster" {
								// 测试服务器分配到测试环境的Web集群
								var testHost models.Host
								if err := db.Where("name = ?", "测试服务器").First(&testHost).Error; err == nil {
									hostTopology := models.HostTopology{
										HostID:    testHost.ID,
										ClusterID: cluster.ID,
									}
									if err := db.Create(&hostTopology).Error; err != nil {
										logger.Errorf("创建主机拓扑关联失败: %v", err)
									} else {
										logger.Infof("主机 %s 已分配到集群 %s", testHost.Name, cluster.Name)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}
