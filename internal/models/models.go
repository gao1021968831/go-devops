package models

import (
	"time"
)

// 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	Role      string    `json:"role" gorm:"default:user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 主机模型
type Host struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"not null"`
	IP          string `json:"ip" gorm:"not null"`
	Port        int    `json:"port" gorm:"default:22"`
	OS          string `json:"os"`
	Status      string `json:"status" gorm:"default:unknown"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	// SSH认证相关字段
	AuthType   string    `json:"auth_type" gorm:"default:password"` // password, key
	Username   string    `json:"username"`
	Password   string    `json:"-" gorm:"column:password"` // 不在JSON中显示密码
	PrivateKey string    `json:"-" gorm:"type:text"`       // SSH私钥，不在JSON中显示
	Passphrase string    `json:"-"`                        // 私钥密码短语
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// 脚本模型
type Script struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Content     string    `json:"content" gorm:"type:text"`
	Type        string    `json:"type" gorm:"default:shell"`
	Description string    `json:"description"`
	CreatedBy   uint      `json:"created_by"`
	User        User      `json:"user" gorm:"foreignKey:CreatedBy"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 作业模型
type Job struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	ScriptID    uint      `json:"script_id"`
	Script      Script    `json:"script" gorm:"foreignKey:ScriptID"`
	HostIDs     string    `json:"host_ids"` // JSON数组字符串
	Status      string    `json:"status" gorm:"default:pending"`
	IsTemporary bool      `json:"is_temporary" gorm:"default:false"` // 是否为临时作业
	CreatedBy   uint      `json:"created_by"`
	User        User      `json:"user" gorm:"foreignKey:CreatedBy"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 作业执行记录
type JobExecution struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	JobID       uint       `json:"job_id"`
	Job         Job        `json:"job" gorm:"foreignKey:JobID"`
	HostID      uint       `json:"host_id"`
	Host        Host       `json:"host" gorm:"foreignKey:HostID"`
	Status      string     `json:"status" gorm:"default:running"`
	Output      string     `json:"output" gorm:"type:text"`
	Error       string     `json:"error" gorm:"type:text"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	ExecutedBy  uint       `json:"executed_by"`
	ExecutedUser User      `json:"executed_user" gorm:"foreignKey:ExecutedBy"`
	// 冗余字段，用于临时作业删除后仍能显示信息
	JobName       string `json:"job_name"`
	ScriptName    string `json:"script_name"`
	ScriptContent string `json:"script_content" gorm:"type:text"`
	ScriptType    string `json:"script_type"`
	IsQuickExec   bool   `json:"is_quick_exec" gorm:"default:false"` // 标记是否为快速执行
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}



// 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// JWT响应
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// 主机创建/更新请求
type HostRequest struct {
	Name        string `json:"name" binding:"required"`
	IP          string `json:"ip" binding:"required"`
	Port        int    `json:"port"`
	OS          string `json:"os"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	AuthType    string `json:"auth_type"` // password, key
	Username    string `json:"username"`
	Password    string `json:"password"`
	PrivateKey  string `json:"private_key"`
	Passphrase  string `json:"passphrase"`
}

// SSH连接测试响应
type SSHTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency string `json:"latency,omitempty"`
}

// 批量主机导入请求
type BatchHostImportRequest struct {
	Hosts []HostRequest `json:"hosts" binding:"required"`
}

// 批量主机导入响应
type BatchHostImportResponse struct {
	Success      int                `json:"success"`
	Failed       int                `json:"failed"`
	Total        int                `json:"total"`
	SuccessHosts []Host             `json:"success_hosts,omitempty"`
	FailedHosts  []BatchImportError `json:"failed_hosts,omitempty"`
}

// 批量导入错误信息
type BatchImportError struct {
	Index int         `json:"index"`
	Host  HostRequest `json:"host"`
	Error string      `json:"error"`
}

// 主机密码更新请求
type HostPasswordUpdateRequest struct {
	AuthType   string `json:"auth_type,omitempty"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
	Passphrase string `json:"passphrase,omitempty"`
}

// 批量主机操作请求
type BatchHostOperationRequest struct {
	HostIDs   []uint      `json:"host_ids" binding:"required"`
	Operation string      `json:"operation" binding:"required"` // update_status, test_connection, update_auth
	Data      interface{} `json:"data,omitempty"`
}

// 批量操作响应
type BatchOperationResponse struct {
	Success int                    `json:"success"`
	Failed  int                    `json:"failed"`
	Total   int                    `json:"total"`
	Results []BatchOperationResult `json:"results"`
}

// 批量操作结果
type BatchOperationResult struct {
	HostID  uint        `json:"host_id"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 业务模型
type Business struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;unique"`
	Code        string    `json:"code" gorm:"not null;unique"`
	Description string    `json:"description"`
	Owner       string    `json:"owner"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 环境模型
type Environment struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Code        string    `json:"code" gorm:"not null"`
	BusinessID  uint      `json:"business_id"`
	Business    Business  `json:"business" gorm:"foreignKey:BusinessID"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 集群模型
type Cluster struct {
	ID            uint        `json:"id" gorm:"primaryKey"`
	Name          string      `json:"name" gorm:"not null"`
	Code          string      `json:"code" gorm:"not null"`
	EnvironmentID uint        `json:"environment_id"`
	Environment   Environment `json:"environment" gorm:"foreignKey:EnvironmentID"`
	Description   string      `json:"description"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// 扩展主机模型，添加拓扑关联
type HostTopology struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	HostID    uint      `json:"host_id"`
	Host      Host      `json:"host" gorm:"foreignKey:HostID"`
	ClusterID uint      `json:"cluster_id"`
	Cluster   Cluster   `json:"cluster" gorm:"foreignKey:ClusterID"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 拓扑树节点
type TopologyNode struct {
	ID       uint           `json:"id"`
	UniqueID string         `json:"unique_id"`
	Name     string         `json:"name"`
	Code     string         `json:"code,omitempty"`
	Type     string         `json:"type"` // business, environment, cluster, host
	ParentID *uint          `json:"parent_id,omitempty"`
	Children []TopologyNode `json:"children,omitempty"`
	HostInfo *Host          `json:"host_info,omitempty"`
	Stats    *NodeStats     `json:"stats,omitempty"`
}

// 节点统计信息
type NodeStats struct {
	TotalHosts   int `json:"total_hosts"`
	OnlineHosts  int `json:"online_hosts"`
	OfflineHosts int `json:"offline_hosts"`
}

// 拓扑请求模型
type BusinessRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code"` // 编码由后端自动生成，不再必需
	Description string `json:"description"`
	Owner       string `json:"owner"`
}

type EnvironmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code"` // 编码由后端自动生成
	BusinessID  uint   `json:"business_id" binding:"required"`
	Description string `json:"description"`
}

type ClusterRequest struct {
	Name          string `json:"name" binding:"required"`
	Code          string `json:"code"` // 编码由后端自动生成
	EnvironmentID uint   `json:"environment_id" binding:"required"`
	Description   string `json:"description"`
}

type HostTopologyRequest struct {
	HostID    uint `json:"host_id" binding:"required"`
	ClusterID uint `json:"cluster_id" binding:"required"`
}
