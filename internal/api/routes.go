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
	dashboardHandler := handlers.NewDashboardHandler(db)
	topologyHandler := handlers.NewTopologyHandler(db)

	// 公开路由
	public := router.Group("/")
	{
		public.POST("/login", authHandler.Login)
		public.POST("/register", authHandler.Register)
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

		// 主机管理
		protected.GET("/hosts", hostHandler.GetHosts)
		protected.POST("/hosts", hostHandler.CreateHost)
		protected.GET("/hosts/:id", hostHandler.GetHost)
		protected.PUT("/hosts/:id", hostHandler.UpdateHost)
		protected.DELETE("/hosts/:id", hostHandler.DeleteHost)
		protected.POST("/hosts/:id/check", hostHandler.CheckHostStatus)
		protected.POST("/hosts/:id/test-ssh", hostHandler.TestSSHConnection)
		protected.POST("/hosts/check-all", hostHandler.CheckAllHostsStatus)
		protected.GET("/hosts/schedule/config", hostHandler.GetScheduleConfig)
		protected.PUT("/hosts/schedule/config", hostHandler.UpdateScheduleConfig)
		protected.PUT("/hosts/:id/auth", hostHandler.UpdateHostAuth)

		// 批量主机操作
		protected.POST("/hosts/batch/import", hostHandler.BatchImportHosts)
		protected.POST("/hosts/batch/import-csv", hostHandler.BatchImportHostsFromCSV)
		protected.GET("/hosts/csv-template", hostHandler.DownloadCSVTemplate)
		protected.POST("/hosts/batch/operation", hostHandler.BatchHostOperation)

		// 脚本管理
		protected.GET("/scripts", scriptHandler.GetScripts)
		protected.POST("/scripts", scriptHandler.CreateScript)
		protected.GET("/scripts/:id", scriptHandler.GetScript)
		protected.PUT("/scripts/:id", scriptHandler.UpdateScript)
		protected.DELETE("/scripts/:id", scriptHandler.DeleteScript)

		// 作业管理
		protected.GET("/jobs", jobHandler.GetJobs)
		protected.POST("/jobs", jobHandler.CreateJob)
		protected.GET("/jobs/:id", jobHandler.GetJob)
		protected.PUT("/jobs/:id", jobHandler.UpdateJob)
		protected.DELETE("/jobs/:id", jobHandler.DeleteJob)
		protected.POST("/jobs/:id/execute", jobHandler.ExecuteJob)
		protected.GET("/jobs/:id/executions", jobHandler.GetJobExecutions)

		// 快速脚本执行
		protected.POST("/scripts/quick-execute", jobHandler.QuickExecuteScript)

		// 执行记录管理
		protected.GET("/executions", jobHandler.GetAllExecutions)
		protected.GET("/executions/:id", jobHandler.GetExecutionDetail)


		// 仪表盘API
		protected.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
		protected.GET("/dashboard/activities", dashboardHandler.GetRecentActivities)
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
	}

	// 管理员路由
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuth())
	admin.Use(middleware.AdminRequired())
	{
		admin.GET("/users", authHandler.GetUsers)
		admin.DELETE("/users/:id", authHandler.DeleteUser)
		admin.PUT("/users/:id/role", authHandler.UpdateUserRole)
	}
}
