package api

import (
	"go-devops/internal/handlers"
	"go-devops/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// 初始化处理器
	authHandler := handlers.NewAuthHandler(db)
	hostHandler := handlers.NewHostHandler(db)
	scriptHandler := handlers.NewScriptHandler(db)
	jobHandler := handlers.NewJobHandler(db)
	jobExecutionHandler := handlers.NewJobExecutionHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)
	topologyHandler := handlers.NewTopologyHandler(db)
	systemHandler := handlers.NewSystemHandler()
	fileHandler := handlers.NewFileHandler(db)

	// 公开路由
	public := router.Group("/")
	{
		public.POST("/login", authHandler.Login)
		public.POST("/register", authHandler.Register)
		public.GET("/system/info", systemHandler.GetSystemInfo)
		public.GET("/system/health", systemHandler.GetHealthCheck)
	}

	// 需要认证的路由
	protected := router.Group("/")
	protected.Use(middleware.JWTAuth())
	{
		// 用户相关
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)
		protected.PUT("/profile/password", authHandler.ChangePassword)
		protected.GET("/profile/stats", authHandler.GetUserStats)

		// 主机管理 - 查看权限对所有用户开放
		protected.GET("/hosts", hostHandler.GetHosts)
		protected.GET("/hosts/:id", hostHandler.GetHost)

		// 脚本管理
		scripts := protected.Group("/scripts")
		{
			scripts.GET("", scriptHandler.GetScripts)
			scripts.POST("", scriptHandler.CreateScript)
			scripts.GET("/:id", scriptHandler.GetScript)
			scripts.PUT("/:id", scriptHandler.UpdateScript)
			scripts.DELETE("/:id", scriptHandler.DeleteScript)
			scripts.POST("/batch/delete", scriptHandler.BatchDeleteScripts)
			scripts.GET("/export", scriptHandler.ExportScripts)
			scripts.POST("/batch/import", scriptHandler.BatchImportScripts)
			scripts.GET("/import/template", scriptHandler.DownloadImportTemplate)
		}

		// 作业执行
		protected.POST("/jobs/:id/execute", jobExecutionHandler.ExecuteJob)
		protected.POST("/scripts/quick-execute", jobExecutionHandler.QuickExecuteScript)
		// 脚本文件关联功能
		protected.POST("/job-executions/save-result", jobExecutionHandler.SaveExecutionResult)

		// 作业管理
		protected.GET("/jobs", jobHandler.GetJobs)
		protected.POST("/jobs", jobHandler.CreateJob)
		protected.GET("/jobs/:id", jobHandler.GetJob)
		protected.PUT("/jobs/:id", jobHandler.UpdateJob)
		protected.DELETE("/jobs/:id", jobHandler.DeleteJob)
		protected.POST("/jobs/batch/delete", jobHandler.BatchDeleteJobs)
		protected.GET("/jobs/export", jobHandler.ExportJobs)
		protected.GET("/jobs/:id/executions", jobHandler.GetJobExecutions)

		// 执行记录管理
		protected.GET("/executions", jobHandler.GetAllExecutions)
		protected.GET("/executions/:id", jobHandler.GetExecutionDetail)

		// 仪表盘API
		protected.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
		protected.GET("/dashboard/activities", dashboardHandler.GetAllActivities)
		protected.GET("/dashboard/recent-activities", dashboardHandler.GetRecentActivities)
		protected.GET("/dashboard/job-trend", dashboardHandler.GetJobTrend)
		protected.GET("/dashboard/host-status", dashboardHandler.GetHostStatusDistribution)

		// 拓扑管理
		protected.GET("/topology/tree", topologyHandler.GetTopologyTree)
		protected.GET("/topology/businesses", topologyHandler.GetBusinesses)
		protected.POST("/topology/businesses", topologyHandler.CreateBusiness)
		protected.PUT("/topology/businesses/:id", topologyHandler.UpdateBusiness)
		protected.DELETE("/topology/businesses/:id", topologyHandler.DeleteBusiness)

		protected.GET("/topology/environments", topologyHandler.GetEnvironments)
		protected.POST("/topology/environments", topologyHandler.CreateEnvironment)
		protected.PUT("/topology/environments/:id", topologyHandler.UpdateEnvironment)
		protected.DELETE("/topology/environments/:id", topologyHandler.DeleteEnvironment)

		protected.GET("/topology/clusters", topologyHandler.GetClusters)
		protected.POST("/topology/clusters", topologyHandler.CreateCluster)
		protected.PUT("/topology/clusters/:id", topologyHandler.UpdateCluster)
		protected.DELETE("/topology/clusters/:id", topologyHandler.DeleteCluster)

		protected.POST("/topology/hosts/assign", topologyHandler.AssignHostToCluster)
		protected.DELETE("/topology/hosts/:hostId/remove", topologyHandler.RemoveHostFromCluster)
		protected.GET("/topology/hosts/unassigned", topologyHandler.GetUnassignedHosts)
		protected.GET("/topology/clusters/:clusterId/hosts", topologyHandler.GetHostsByCluster)

		// 文件管理
		protected.POST("/files/upload", fileHandler.UploadFile)
		protected.GET("/files/:id/download", fileHandler.DownloadFile)
		protected.GET("/files", fileHandler.GetFiles)
		protected.GET("/files/:id", fileHandler.GetFile)
		protected.PUT("/files/:id", fileHandler.UpdateFile)
		protected.DELETE("/files/:id", fileHandler.DeleteFile)
		protected.POST("/files/:id/distribute", fileHandler.DistributeFile)
		protected.GET("/file-distributions", fileHandler.GetDistributions)
		protected.GET("/file-distributions/:id", fileHandler.GetDistributionDetail)
		protected.DELETE("/file-distributions/:id", fileHandler.DeleteDistribution)
	}

	// 管理员路由
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuth())
	admin.Use(middleware.AdminRequired())
	{
		// 用户管理
		admin.GET("/users", authHandler.GetUsers)
		admin.DELETE("/users/:id", authHandler.DeleteUser)
		admin.PUT("/users/:id/role", authHandler.UpdateUserRole)

		// 主机管理操作（仅管理员）
		admin.POST("/hosts", hostHandler.CreateHost)
		admin.PUT("/hosts/:id", hostHandler.UpdateHost)
		admin.DELETE("/hosts/:id", hostHandler.DeleteHost)
		admin.POST("/hosts/:id/check", hostHandler.CheckHostStatus)
		admin.POST("/hosts/:id/test-ssh", hostHandler.TestSSHConnection)
		admin.POST("/hosts/check-all", hostHandler.CheckAllHostsStatus)
		admin.GET("/hosts/schedule/config", hostHandler.GetScheduleConfig)
		admin.PUT("/hosts/schedule/config", hostHandler.UpdateScheduleConfig)
		admin.PUT("/hosts/:id/auth", hostHandler.UpdateHostAuth)

		// 批量主机操作（仅管理员）
		admin.POST("/hosts/batch/import", hostHandler.BatchImportHosts)
		admin.POST("/hosts/batch/import-csv", hostHandler.BatchImportHostsFromCSV)
		admin.GET("/hosts/csv-template", hostHandler.DownloadCSVTemplate)
		admin.POST("/hosts/batch/operation", hostHandler.BatchHostOperation)

		// 作业执行记录管理（仅管理员）
		admin.DELETE("/executions/:id", jobExecutionHandler.DeleteJobExecution)
		admin.POST("/executions/batch/delete", jobExecutionHandler.BatchDeleteJobExecutions)
	}
}
