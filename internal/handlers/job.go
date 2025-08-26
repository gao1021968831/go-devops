package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JobHandler struct {
	db          *gorm.DB
	crudHandler *JobCRUDHandler
	execHandler *JobExecutionHandler
}

func NewJobHandler(db *gorm.DB) *JobHandler {
	return &JobHandler{
		db:          db,
		crudHandler: NewJobCRUDHandler(db),
		execHandler: NewJobExecutionHandler(db),
	}
}

// 委托给CRUD处理器的方法
func (h *JobHandler) GetJobs(c *gin.Context) {
	h.crudHandler.GetJobs(c)
}

func (h *JobHandler) CreateJob(c *gin.Context) {
	h.crudHandler.CreateJob(c)
}

func (h *JobHandler) GetJob(c *gin.Context) {
	h.crudHandler.GetJob(c)
}

func (h *JobHandler) UpdateJob(c *gin.Context) {
	h.crudHandler.UpdateJob(c)
}

func (h *JobHandler) DeleteJob(c *gin.Context) {
	h.crudHandler.DeleteJob(c)
}

// 委托给执行处理器的方法
func (h *JobHandler) ExecuteJob(c *gin.Context) {
	h.execHandler.ExecuteJob(c)
}

func (h *JobHandler) QuickExecuteScript(c *gin.Context) {
	h.execHandler.QuickExecuteScript(c)
}

// 批量删除作业
func (h *JobHandler) BatchDeleteJobs(c *gin.Context) {
	h.crudHandler.BatchDeleteJobs(c)
}

// 导出作业
func (h *JobHandler) ExportJobs(c *gin.Context) {
	h.crudHandler.ExportJobs(c)
}

func (h *JobHandler) GetJobExecutions(c *gin.Context) {
	h.execHandler.GetJobExecutions(c)
}

func (h *JobHandler) GetAllExecutions(c *gin.Context) {
	h.execHandler.GetAllExecutions(c)
}

func (h *JobHandler) GetExecutionDetail(c *gin.Context) {
	h.execHandler.GetExecutionDetail(c)
}
