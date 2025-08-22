package scheduler

import (
	"time"
	"go-devops/internal/models"
	"go-devops/internal/ssh"
	"go-devops/internal/logger"
	"gorm.io/gorm"
)

type Scheduler struct {
	db       *gorm.DB
	stopChan chan bool
	running  bool
}

func NewScheduler(db *gorm.DB) *Scheduler {
	return &Scheduler{
		db:       db,
		stopChan: make(chan bool),
		running:  false,
	}
}

// 启动定时任务调度器
func (s *Scheduler) Start() {
	if s.running {
		return
	}
	
	s.running = true
	logger.Infof("定时任务调度器启动")
	
	// 启动主机状态检查定时任务（每5分钟执行一次）
	go s.startHostStatusChecker()
}

// 停止定时任务调度器
func (s *Scheduler) Stop() {
	if !s.running {
		return
	}
	
	s.running = false
	s.stopChan <- true
	logger.Infof("定时任务调度器停止")
}

// 主机状态检查定时任务
func (s *Scheduler) startHostStatusChecker() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
	defer ticker.Stop()
	
	// 立即执行一次检查
	s.checkAllHostsStatus()
	
	for {
		select {
		case <-ticker.C:
			s.checkAllHostsStatus()
		case <-s.stopChan:
			logger.Infof("主机状态检查定时任务停止")
			return
		}
	}
}

// 检查所有主机状态
func (s *Scheduler) checkAllHostsStatus() {
	logger.Infof("开始定时检查所有主机状态")
	
	var hosts []models.Host
	if err := s.db.Find(&hosts).Error; err != nil {
		logger.Errorf("获取主机列表失败: %v", err)
		return
	}
	
	onlineCount := 0
	offlineCount := 0
	unknownCount := 0
	
	for _, host := range hosts {
		status := s.checkSingleHostStatus(host)
		
		// 只有状态发生变化时才更新数据库
		if host.Status != status {
			if err := s.db.Model(&host).Update("status", status).Error; err != nil {
				logger.Errorf("更新主机 %s 状态失败: %v", host.Name, err)
			} else {
				logger.Infof("主机 %s 状态从 %s 变更为 %s", host.Name, host.Status, status)
			}
		}
		
		// 统计状态
		switch status {
		case "online":
			onlineCount++
		case "offline":
			offlineCount++
		default:
			unknownCount++
		}
	}
	
	logger.Infof("定时检查完成 - 在线: %d台，离线: %d台，未知: %d台", 
		onlineCount, offlineCount, unknownCount)
}

// 检查单个主机状态
func (s *Scheduler) checkSingleHostStatus(host models.Host) string {
	if host.Username == "" && host.Password == "" && host.PrivateKey == "" {
		return "unknown"
	}
	
	testResult, err := ssh.TestSSHConnection(&host)
	if err != nil {
		logger.Debugf("主机 %s SSH连接测试出错: %v", host.Name, err)
		return "offline"
	}
	
	if testResult.Success {
		return "online"
	}
	
	return "offline"
}

// 设置检查间隔（分钟）
func (s *Scheduler) SetHostCheckInterval(minutes int) {
	if minutes < 1 {
		minutes = 5 // 最小间隔5分钟
	}
	
	logger.Infof("主机状态检查间隔设置为 %d 分钟", minutes)
	// 这里可以实现动态调整检查间隔的逻辑
}
