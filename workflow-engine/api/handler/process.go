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

// ProcessHandler 流程处理器
type ProcessHandler struct {
	processService service.ProcessService
}

// NewProcessHandler 创建流程处理器
func NewProcessHandler(processService service.ProcessService) *ProcessHandler {
	return &ProcessHandler{
		processService: processService,
	}
}

// RegisterRoutes 注册路由
func (h *ProcessHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 流程定义
	defGroup := r.Group("/processes/definitions")
	{
		defGroup.POST("", h.CreateDefinition)
		defGroup.GET("", h.ListDefinitions)
		defGroup.GET("/:id", h.GetDefinition)
		defGroup.PUT("/:id", h.UpdateDefinition)
		defGroup.POST("/:id/activate", h.ActivateDefinition)
		defGroup.POST("/:id/archive", h.ArchiveDefinition)
	}

	// 流程实例
	instGroup := r.Group("/processes/instances")
	{
		instGroup.POST("", h.SubmitProcess)
		instGroup.GET("", h.ListInstances)
		instGroup.GET("/:id", h.GetInstance)
		instGroup.POST("/:id/withdraw", h.WithdrawInstance)
		instGroup.GET("/:id/history", h.GetInstanceHistory)
		instGroup.PUT("/:id/steps", h.ModifySteps)
		instGroup.POST("/:id/urgent", h.MarkUrgent)
	}
}

// CreateDefinition godoc
// @Summary 创建流程定义
// @Description 创建新的流程定义
// @Tags 流程定义
// @Accept json
// @Produce json
// @Param request body service.CreateDefinitionRequest true "创建参数"
// @Success 200 {object} response.Response{data=model.ProcessDefinition}
// @Router /api/v1/processes/definitions [post]
func (h *ProcessHandler) CreateDefinition(c *gin.Context) {
	var req service.CreateDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "20001", err.Error())
		return
	}

	def, err := h.processService.CreateDefinition(c, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, def)
}

// GetDefinition godoc
// @Summary 获取流程定义
// @Description 根据ID获取流程定义详情
// @Tags 流程定义
// @Produce json
// @Param id path string true "流程定义ID"
// @Success 200 {object} response.Response{data=model.ProcessDefinition}
// @Router /api/v1/processes/definitions/{id} [get]
func (h *ProcessHandler) GetDefinition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	def, err := h.processService.GetDefinition(c, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, def)
}

// ListDefinitions godoc
// @Summary 获取流程定义列表
// @Description 分页获取流程定义列表
// @Tags 流程定义
// @Produce json
// @Param status query string false "状态: draft|active|archived"
// @Param name query string false "名称模糊查询"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResult}
// @Router /api/v1/processes/definitions [get]
func (h *ProcessHandler) ListDefinitions(c *gin.Context) {
	var status *model.DefStatus
	if s := c.Query("status"); s != "" {
		st := model.DefStatus(s)
		status = &st
	}
	name := c.Query("name")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	result, err := h.processService.ListDefinitions(c, status, name, page, size)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// UpdateDefinition godoc
// @Summary 更新流程定义
// @Description 更新流程定义信息（仅草稿状态）
// @Tags 流程定义
// @Accept json
// @Produce json
// @Param id path string true "流程定义ID"
// @Param request body service.UpdateDefinitionRequest true "更新参数"
// @Success 200 {object} response.Response{data=model.ProcessDefinition}
// @Router /api/v1/processes/definitions/{id} [put]
func (h *ProcessHandler) UpdateDefinition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	var req service.UpdateDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "20001", err.Error())
		return
	}

	def, err := h.processService.UpdateDefinition(c, id, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, def)
}

// ActivateDefinition godoc
// @Summary 激活流程定义
// @Description 激活流程定义，使其可以创建实例
// @Tags 流程定义
// @Accept json
// @Produce json
// @Param id path string true "流程定义ID"
// @Param version body int true "版本号"
// @Success 200 {object} response.Response{data=model.ProcessDefinition}
// @Router /api/v1/processes/definitions/{id}/activate [post]
func (h *ProcessHandler) ActivateDefinition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	var req struct {
		Version int `json:"version" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "20001", err.Error())
		return
	}

	def, err := h.processService.ActivateDefinition(c, id, req.Version)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, def)
}

// ArchiveDefinition godoc
// @Summary 归档流程定义
// @Description 归档流程定义，不再允许创建实例
// @Tags 流程定义
// @Produce json
// @Param id path string true "流程定义ID"
// @Success 200 {object} response.Response
// @Router /api/v1/processes/definitions/{id}/archive [post]
func (h *ProcessHandler) ArchiveDefinition(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	if err := h.processService.ArchiveDefinition(c, id); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// SubmitProcess godoc
// @Summary 提交流程
// @Description 创建新的流程实例
// @Tags 流程实例
// @Accept json
// @Produce json
// @Param request body service.SubmitProcessRequest true "提交参数"
// @Success 200 {object} response.Response{data=model.ProcessInstance}
// @Router /api/v1/processes/instances [post]
func (h *ProcessHandler) SubmitProcess(c *gin.Context) {
	var req service.SubmitProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "20001", err.Error())
		return
	}

	// 获取当前用户
	req.SubmittedBy = middleware.GetUserID(c)

	instance, err := h.processService.SubmitProcess(c, &req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, instance)
}

// GetInstance godoc
// @Summary 获取流程实例详情
// @Description 获取流程实例详细信息，包括步骤和历史
// @Tags 流程实例
// @Produce json
// @Param id path string true "流程实例ID"
// @Success 200 {object} response.Response{data=model.ProcessInstanceDetail}
// @Router /api/v1/processes/instances/{id} [get]
func (h *ProcessHandler) GetInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	detail, err := h.processService.GetInstance(c, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, detail)
}

// ListInstances godoc
// @Summary 获取流程实例列表
// @Description 分页获取流程实例列表
// @Tags 流程实例
// @Produce json
// @Param status query string false "状态: running|completed|rejected|withdrawn"
// @Param businessKey query string false "业务单号"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.PageResult}
// @Router /api/v1/processes/instances [get]
func (h *ProcessHandler) ListInstances(c *gin.Context) {
	query := &service.InstanceQuery{
		SubmittedBy: middleware.GetUserID(c),
		Page:        1,
		Size:        10,
	}

	if status := c.Query("status"); status != "" {
		s := model.ProcessStatus(status)
		query.Status = &s
	}
	query.BusinessKey = c.Query("businessKey")
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Size, _ = strconv.Atoi(c.DefaultQuery("size", "10"))

	result, err := h.processService.ListInstances(c, query)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, result)
}

// WithdrawInstance godoc
// @Summary 撤回流程
// @Description 发起人撤回尚未处理的流程
// @Tags 流程实例
// @Accept json
// @Produce json
// @Param id path string true "流程实例ID"
// @Param reason body string false "撤回原因"
// @Success 200 {object} response.Response
// @Router /api/v1/processes/instances/{id}/withdraw [post]
func (h *ProcessHandler) WithdrawInstance(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	userID := middleware.GetUserID(c)
	if err := h.processService.WithdrawInstance(c, id, userID, req.Reason); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// GetInstanceHistory godoc
// @Summary 获取流程审批历史
// @Description 获取流程实例的审批历史记录
// @Tags 流程实例
// @Produce json
// @Param id path string true "流程实例ID"
// @Success 200 {object} response.Response{data=[]model.ApprovalHistoryItem}
// @Router /api/v1/processes/instances/{id}/history [get]
func (h *ProcessHandler) GetInstanceHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	history, err := h.processService.GetInstanceHistory(c, id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, history)
}

// ModifySteps godoc
// @Summary 修改审批步骤
// @Description 动态修改流程实例的审批步骤
// @Tags 流程实例
// @Accept json
// @Produce json
// @Param id path string true "流程实例ID"
// @Param request body service.ModifyStepsRequest true "修改参数"
// @Success 200 {object} response.Response
// @Router /api/v1/processes/instances/{id}/steps [put]
func (h *ProcessHandler) ModifySteps(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	var req service.ModifyStepsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, "20001", err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.processService.ModifySteps(c, id, userID, &req); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}

// MarkUrgent godoc
// @Summary 标记流程加急
// @Description 发起人将流程标记为加急
// @Tags 流程实例
// @Produce json
// @Param id path string true "流程实例ID"
// @Success 200 {object} response.Response
// @Router /api/v1/processes/instances/{id}/urgent [post]
func (h *ProcessHandler) MarkUrgent(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, "20001", "无效的ID")
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.processService.MarkUrgent(c, id, userID); err != nil {
		response.HandleError(c, err)
		return
	}

	response.Success(c, nil)
}
