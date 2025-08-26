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
	Name        string    `json:"name" gorm:"not null;type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Content     string    `json:"content" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Type        string    `json:"type" gorm:"default:shell;type:varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	Description string    `json:"description" gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"`
	CreatedBy   uint      `json:"created_by"`
	User        User      `json:"user" gorm:"foreignKey:CreatedBy"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 作业
type Job struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	ScriptID    uint      `json:"script_id"`
	Script      Script    `json:"script" gorm:"foreignKey:ScriptID"`
	HostIDs     string    `json:"host_ids" gorm:"type:text"` // JSON数组存储主机ID列表
	Parameters  string    `json:"parameters" gorm:"type:text"` // 脚本参数
	Timeout     int       `json:"timeout" gorm:"default:300"` // 超时时间（秒）
	Status      string    `json:"status" gorm:"default:pending"` // 作业状态：pending, running, completed, failed
	// 文件关联字段
	InputFileIDs    string `json:"input_file_ids" gorm:"type:text"`     // 输入文件ID列表（JSON数组）
	SaveOutput      bool   `json:"save_output" gorm:"default:false"`    // 是否保存输出为文件
	SaveError       bool   `json:"save_error" gorm:"default:false"`     // 是否保存错误日志为文件
	OutputCategory  string `json:"output_category" gorm:"default:script_output"` // 输出文件分类
	CreatedBy   uint      `json:"created_by"`
	User        User      `json:"user" gorm:"foreignKey:CreatedBy"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 作业执行记录
type JobExecution struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	JobID       *uint      `json:"job_id"` // 允许为NULL，快速执行时不关联作业
	Job         *Job       `json:"job" gorm:"foreignKey:JobID"`
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
	// 文件关联字段
	OutputFileID  *uint `json:"output_file_id"`                            // 输出文件ID
	OutputFile    *File `json:"output_file" gorm:"foreignKey:OutputFileID"` // 输出文件
	ErrorFileID   *uint `json:"error_file_id"`                             // 错误日志文件ID
	ErrorFile     *File `json:"error_file" gorm:"foreignKey:ErrorFileID"`   // 错误日志文件
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

// 用户活动记录
type UserActivity struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	User        User      `json:"user" gorm:"foreignKey:UserID"`
	Action      string    `json:"action"`      // 操作类型：create, update, delete, login, logout等
	Resource    string    `json:"resource"`    // 资源类型：user, host, job, script等  
	ResourceID  *uint     `json:"resource_id"` // 资源ID，可为空
	Description string    `json:"description"` // 操作描述
	IPAddress   string    `json:"ip_address"`  // 操作IP地址
	UserAgent   string    `json:"user_agent"`  // 用户代理
	Status      string    `json:"status" gorm:"default:success"` // success, failed
	Details     string    `json:"details" gorm:"type:text"`      // 详细信息或错误信息
	CreatedAt   time.Time `json:"created_at"`
}

// 文件模型
type File struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`                    // 文件名
	OriginalName string   `json:"original_name" gorm:"not null"`           // 原始文件名
	Path        string    `json:"path" gorm:"not null"`                    // 文件存储路径
	Size        int64     `json:"size"`                                    // 文件大小（字节）
	MimeType    string    `json:"mime_type"`                               // MIME类型
	MD5Hash     string    `json:"md5_hash" gorm:"index"`                   // MD5哈希值
	Category    string    `json:"category" gorm:"default:general"`         // 文件分类：script, config, package, general
	Description string    `json:"description"`                             // 文件描述
	IsPublic    bool      `json:"is_public" gorm:"default:false"`          // 是否公开
	UploadedBy  uint      `json:"uploaded_by"`                             // 上传者ID
	User        User      `json:"user" gorm:"foreignKey:UploadedBy"`       // 上传者信息
	DownloadCount int     `json:"download_count" gorm:"default:0"`         // 下载次数
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 文件分发记录
type FileDistribution struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FileID      uint      `json:"file_id"`                                 // 文件ID
	File        File      `json:"file" gorm:"foreignKey:FileID"`           // 文件信息
	HostIDs     string    `json:"host_ids" gorm:"type:text"`               // 目标主机ID列表（JSON数组）
	TargetPath  string    `json:"target_path" gorm:"not null"`             // 目标路径
	Status      string    `json:"status" gorm:"default:pending"`           // 分发状态：pending, running, completed, failed
	Progress    int       `json:"progress" gorm:"default:0"`               // 分发进度（0-100）
	StartTime   *time.Time `json:"start_time"`                             // 开始时间
	EndTime     *time.Time `json:"end_time"`                               // 结束时间
	CreatedBy   uint      `json:"created_by"`                              // 创建者ID
	User        User      `json:"user" gorm:"foreignKey:CreatedBy"`        // 创建者信息
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 文件分发详情
type FileDistributionDetail struct {
	ID             uint               `json:"id" gorm:"primaryKey"`
	DistributionID uint               `json:"distribution_id"`                    // 分发记录ID
	Distribution   FileDistribution   `json:"distribution" gorm:"foreignKey:DistributionID"`
	HostID         uint               `json:"host_id"`                            // 主机ID
	Host           Host               `json:"host" gorm:"foreignKey:HostID"`      // 主机信息
	Status         string             `json:"status" gorm:"default:pending"`      // 状态：pending, running, completed, failed
	Output         string             `json:"output" gorm:"type:text"`            // 执行输出
	Error          string             `json:"error" gorm:"type:text"`             // 错误信息
	StartTime      *time.Time         `json:"start_time"`                         // 开始时间
	EndTime        *time.Time         `json:"end_time"`                           // 结束时间
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// 文件上传请求
type FileUploadRequest struct {
	Category    string `form:"category"`
	Description string `form:"description"`
	IsPublic    bool   `form:"is_public"`
}

// 文件分发请求
type FileDistributionRequest struct {
	FileID     uint   `json:"file_id" binding:"required"`
	HostIDs    []uint `json:"host_ids" binding:"required"`
	TargetPath string `json:"target_path" binding:"required"`
}

// 文件更新请求
type FileUpdateRequest struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// 作业创建/更新请求（扩展支持文件关联）
type JobRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	ScriptID       uint   `json:"script_id" binding:"required"`
	HostIDs        []uint `json:"host_ids" binding:"required"`
	Parameters     string `json:"parameters"`
	Timeout        int    `json:"timeout"`
	InputFileIDs   []uint `json:"input_file_ids"`   // 输入文件ID列表
	SaveOutput     bool   `json:"save_output"`      // 是否保存输出为文件
	SaveError      bool   `json:"save_error"`       // 是否保存错误日志为文件
	OutputCategory string `json:"output_category"`  // 输出文件分类
}

// 脚本创建/更新请求
type ScriptRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
	Type        string `json:"type" binding:"required"`
}

// 脚本执行结果文件保存请求
type SaveExecutionResultRequest struct {
	ExecutionID    uint   `json:"execution_id" binding:"required"`
	SaveOutput     bool   `json:"save_output"`
	SaveError      bool   `json:"save_error"`
	OutputCategory string `json:"output_category"`
}
