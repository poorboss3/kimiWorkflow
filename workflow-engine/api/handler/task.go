package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"workflow-engine/api/middleware"
	"workflow-engine/api/response"
	"workflow-engine/internal/model"
	"workflow-engine/internal/service"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// RegisterRoutes 注册路由
func (h *TaskHandler) RegisterRoutes(r *gin.RouterGroup) {
	tasks := r.Group("/tasks")
	{
		tasks.GET("/pending", h.GetPendingTasks)
		tasks.GET("/completed", h.GetCompletedTasks)
		tasks.GET("/statistics", h.GetTaskStatistics)
		tasks.GET("/:id", h.GetTaskDetail)
		tasks.POST("/:id/action", h.ProcessTask)
	}
}

// GetPendingTasks godoc
// @Summary 获取待办任务
// @Description 获取当前用户的待办任务列表
// @Tags 任务
// @Produce json
// @Param isUrgent query bool false "是否只显示加急"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResult}
// @Router /api/v1/tasks/pending [get]
func (h *TaskHandler) GetPendingTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	result, err := h.taskService.GetPendingTasks(c, userID, page, size)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// GetCompletedTasks godoc
// @Summary 获取已办任务
// @Description 获取当前用户的已办任务列表
// @Tags 任务
// @Produce json
// @Param startTime query string false "开始时间(YYYY-MM-DD)"
// @Param endTime query string false "结束时间(YYYY-MM-DD)"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResult}
// @Router /api/v1/tasks/completed [get]
func (h *TaskHandler) GetCompletedTasks(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	result, err := h.taskService.GetCompletedTasks(c, userID, page, size)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// GetTaskStatistics godoc
// @Summary 获取任务统计
// @Description 获取当前用户的任务统计数据
// @Tags 任务
// @Produce json
// @Success 200 {object} response.Response{data=model.TaskStatistics}
// @Router /api/v1/tasks/statistics [get]
func (h *TaskHandler) GetTaskStatistics(c *gin.Context) {
	userID := middleware.GetUserID(c)

	stats, err := h.taskService.GetTaskStatistics(c, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, stats)
}

// GetTaskDetail godoc
// @Summary 获取任务详情
// @Description 获取任务详细信息，包括流程数据和审批步骤
// @Tags 任务
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} response.Response{data=model.TaskDetail}
// @Router /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetTaskDetail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	userID := middleware.GetUserID(c)
	detail, err := h.taskService.GetTaskDetail(c, id, userID)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, detail)
}

// ProcessTask godoc
// @Summary 处理任务
// @Description 执行任务操作（通过/驳回/退回/加签等）
// @Tags 任务
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param request body model.TaskActionRequest true "操作参数"
// @Success 200 {object} response.Response{data=model.ActionResult}
// @Router /api/v1/tasks/{id}/action [post]
func (h *TaskHandler) ProcessTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	var req model.TaskActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "20001", err.Error())
		return
	}

	userID := middleware.GetUserID(c)

	var result *model.ActionResult
	switch req.Action {
	case model.TaskActionApprove:
		params := &service.TaskActionParams{Comment: req.Comment}
		result, err = h.taskService.Approve(c, id, userID, params)
	case model.TaskActionReject:
		params := &service.TaskActionParams{Comment: req.Comment}
		result, err = h.taskService.Reject(c, id, userID, params)
	case model.TaskActionReturn:
		params := &service.ReturnParams{
			Comment:      req.Comment,
			ReturnToStep: req.ReturnToStep,
		}
		result, err = h.taskService.Return(c, id, userID, params)
	case model.TaskActionCountersign:
		if req.CountersignData == nil {
			response.Error(c, "20001", "加签数据不能为空")
			return
		}
		params := &service.CountersignParams{
			Comment:         req.Comment,
			Assignees:       req.CountersignData.Assignees,
			Type:            req.CountersignData.Type,
			JointSignPolicy: req.CountersignData.JointSignPolicy,
		}
		result, err = h.taskService.Countersign(c, id, userID, params)
	case model.TaskActionNotifyRead:
		err = h.taskService.MarkNotifyRead(c, id, userID)
	default:
		response.Error(c, "20001", "不支持的操作类型")
		return
	}

	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}
