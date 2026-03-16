package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"workflow-engine/internal/extension"
	"workflow-engine/internal/model"
	"workflow-engine/internal/pkg/errors"
	"workflow-engine/internal/pkg/locker"
	"workflow-engine/internal/pkg/memory"
	"workflow-engine/internal/pkg/utils"
	"workflow-engine/internal/repository"
)

// TaskService 任务服务接口
type TaskService interface {
	// 任务查询
	GetPendingTasks(ctx context.Context, userID string, page, size int) (*utils.PageResult[*model.TaskListItem], error)
	GetCompletedTasks(ctx context.Context, userID string, page, size int) (*utils.PageResult[*model.TaskListItem], error)
	GetTaskDetail(ctx context.Context, taskID uuid.UUID, userID string) (*model.TaskDetail, error)
	GetTaskStatistics(ctx context.Context, userID string) (*model.TaskStatistics, error)

	// 任务操作
	Approve(ctx context.Context, taskID uuid.UUID, userID string, req *TaskActionParams) (*model.ActionResult, error)
	Reject(ctx context.Context, taskID uuid.UUID, userID string, req *TaskActionParams) (*model.ActionResult, error)
	Return(ctx context.Context, taskID uuid.UUID, userID string, req *ReturnParams) (*model.ActionResult, error)
	Countersign(ctx context.Context, taskID uuid.UUID, userID string, req *CountersignParams) (*model.ActionResult, error)
	MarkNotifyRead(ctx context.Context, taskID uuid.UUID, userID string) error
}

// TaskServiceImpl 任务服务实现
type TaskServiceImpl struct {
	taskRepo       repository.TaskRepository
	stepRepo       repository.ApprovalStepRepository
	instanceRepo   repository.ProcessInstanceRepository
	defRepo        repository.ProcessDefinitionRepository
	mq             *memory.Queue
	locker         locker.DistributedLocker
}

// NewTaskService 创建任务服务
func NewTaskService(
	taskRepo repository.TaskRepository,
	stepRepo repository.ApprovalStepRepository,
	instanceRepo repository.ProcessInstanceRepository,
	defRepo repository.ProcessDefinitionRepository,
	mq *memory.Queue,
	locker locker.DistributedLocker,
) TaskService {
	return &TaskServiceImpl{
		taskRepo:     taskRepo,
		stepRepo:     stepRepo,
		instanceRepo: instanceRepo,
		defRepo:      defRepo,
		mq:           mq,
		locker:       locker,
	}
}

// TaskActionParams 任务操作参数
type TaskActionParams struct {
	Comment string `json:"comment"`
}

// ReturnParams 退回参数
type ReturnParams struct {
	Comment      string   `json:"comment"`
	ReturnToStep *float64 `json:"returnToStep"`
}

// CountersignParams 加签参数
type CountersignParams struct {
	Comment     string             `json:"comment"`
	Assignees   []model.ApproverRef `json:"assignees"`
	Type        model.StepType      `json:"type"`
	JointSignPolicy model.JointSignPolicy `json:"jointSignPolicy"`
}

// ==================== 任务查询 ====================

func (s *TaskServiceImpl) GetPendingTasks(ctx context.Context, userID string, page, size int) (*utils.PageResult[*model.TaskListItem], error) {
	tasks, total, err := s.taskRepo.ListByAssignee(ctx, userID, utils.Ptr(model.TaskStatusPending), page, size)
	if err != nil {
		return nil, err
	}

	items := make([]*model.TaskListItem, len(tasks))
	for i, task := range tasks {
		item, err := s.buildTaskListItem(ctx, &task)
		if err != nil {
			continue
		}
		items[i] = item
	}

	return utils.NewPageResult(items, total, page, size), nil
}

func (s *TaskServiceImpl) GetCompletedTasks(ctx context.Context, userID string, page, size int) (*utils.PageResult[*model.TaskListItem], error) {
	tasks, total, err := s.taskRepo.ListCompletedByAssignee(ctx, userID, page, size)
	if err != nil {
		return nil, err
	}

	items := make([]*model.TaskListItem, len(tasks))
	for i, task := range tasks {
		item, err := s.buildTaskListItem(ctx, &task)
		if err != nil {
			continue
		}
		items[i] = item
	}

	return utils.NewPageResult(items, total, page, size), nil
}

func (s *TaskServiceImpl) buildTaskListItem(ctx context.Context, task *model.Task) (*model.TaskListItem, error) {
	instance, err := s.instanceRepo.GetByID(ctx, task.InstanceID)
	if err != nil {
		return nil, err
	}

	def, _ := s.defRepo.GetByID(ctx, instance.DefinitionID)
	processName := ""
	if def != nil {
		processName = def.Name
	}

	initiatorID := instance.SubmittedBy
	if instance.OnBehalfOf != nil {
		initiatorID = *instance.OnBehalfOf
	}

	// 计算待处理时长
	pendingHours := int(time.Since(task.CreatedAt).Hours())

	// 提取表单摘要
	var formData map[string]interface{}
	json.Unmarshal(instance.FormDataSnapshot, &formData)
	formSummary := extractFormSummary(formData)

	item := &model.TaskListItem{
		Task:          *task,
		ProcessName:   processName,
		InitiatorID:   initiatorID,
		InitiatorName: initiatorID,
		SubmittedAt:   instance.CreatedAt,
		PendingHours:  pendingHours,
		FormSummary:   formSummary,
	}

	return item, nil
}

func extractFormSummary(formData map[string]interface{}) []byte {
	// 提取关键字段作为摘要
	summary := make(map[string]interface{})
	keyFields := []string{"amount", "title", "type", "department"}
	for _, key := range keyFields {
		if val, ok := formData[key]; ok {
			summary[key] = val
		}
	}
	data, _ := json.Marshal(summary)
	return data
}

func (s *TaskServiceImpl) GetTaskDetail(ctx context.Context, taskID uuid.UUID, userID string) (*model.TaskDetail, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if task.AssigneeID != userID {
		return nil, errors.New(errors.ErrForbidden, "无权查看此任务")
	}

	instance, err := s.instanceRepo.GetByID(ctx, task.InstanceID)
	if err != nil {
		return nil, err
	}

	def, _ := s.defRepo.GetByID(ctx, instance.DefinitionID)
	processName := ""
	if def != nil {
		processName = def.Name
	}

	// 解析表单数据
	var formData map[string]interface{}
	json.Unmarshal(instance.FormDataSnapshot, &formData)

	// 获取流程步骤信息
	steps, _ := s.stepRepo.ListByInstance(ctx, instance.ID)
	stepInfos := make([]model.StepInfo, len(steps))
	for i, step := range steps {
		var assignees []model.ApproverRef
		json.Unmarshal(step.Assignees, &assignees)
		assigneeIDs := make([]string, len(assignees))
		for j, a := range assignees {
			assigneeIDs[j] = a.Value
		}

		stepInfos[i] = model.StepInfo{
			StepIndex: step.StepIndex,
			Type:      step.Type,
			Status:    step.Status,
			Assignees: assigneeIDs,
		}
	}

	detail := &model.TaskDetail{
		Task:         *task,
		ProcessName:  processName,
		DefinitionID: instance.DefinitionID,
		BusinessKey:  instance.BusinessKey,
		FormData:     formData,
		SubmittedBy:  instance.SubmittedBy,
		OnBehalfOf:   instance.OnBehalfOf,
		CanReturn:    task.Step != nil && task.Step.StepIndex > 1,
		CanReject:    true,
		Steps:        stepInfos,
	}

	return detail, nil
}

func (s *TaskServiceImpl) GetTaskStatistics(ctx context.Context, userID string) (*model.TaskStatistics, error) {
	pendingCount, _ := s.taskRepo.CountPendingByAssignee(ctx, userID)
	urgentCount, _ := s.taskRepo.CountUrgentByAssignee(ctx, userID)

	// 获取已办数量（近30天）
	_, completedTotal, _ := s.taskRepo.ListCompletedByAssignee(ctx, userID, 1, 1000)

	return &model.TaskStatistics{
		PendingCount:   pendingCount,
		CompletedCount: completedTotal,
		UrgentCount:    urgentCount,
	}, nil
}

// ==================== 任务操作 ====================

func (s *TaskServiceImpl) Approve(ctx context.Context, taskID uuid.UUID, userID string, req *TaskActionParams) (*model.ActionResult, error) {
	// 1. 获取任务并校验权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.AssigneeID != userID {
		return nil, errors.New(errors.ErrForbidden, "无权处理此任务")
	}
	if task.Status != model.TaskStatusPending {
		return nil, errors.New(errors.ErrTaskNotPending, "")
	}

	// 2. 获取步骤信息
	step, err := s.stepRepo.GetByID(ctx, task.StepID)
	if err != nil {
		return nil, err
	}

	// 3. 分布式锁防止会签并发问题
	lockKey := fmt.Sprintf("step:%s:complete", step.ID)
	lock := s.locker.Acquire(ctx, lockKey, 30*time.Second)
	if lock == nil {
		return nil, errors.New(errors.ErrConcurrentOperation, "")
	}
	defer lock.Release()

	// 4. 再次检查任务状态
	task, err = s.taskRepo.GetByID(ctx, taskID)
	if err != nil || task.Status != model.TaskStatusPending {
		return nil, errors.New(errors.ErrTaskAlreadyProcessed, "")
	}

	// 5. 完成任务
	now := time.Now()
	action := model.TaskActionApprove
	task.Status = model.TaskStatusCompleted
	task.Action = &action
	if req != nil {
		task.Comment = &req.Comment
	}
	task.CompletedAt = &now

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// 6. 检查步骤推进条件
	shouldProceed, err := s.checkStepCompletion(ctx, step)
	if err != nil {
		return nil, err
	}

	result := &model.ActionResult{
		TaskID:     taskID,
		InstanceID: task.InstanceID,
		Action:     "approve",
	}

	if shouldProceed {
		// 7. 推进到下一步或完成流程
		nextTasks, isCompleted, err := s.proceedToNextStep(ctx, step, "approve")
		if err != nil {
			return nil, err
		}
		result.NextTaskIDs = nextTasks
		result.IsCompleted = isCompleted
	}

	// 8. 发送通知
	s.sendActionNotification(ctx, task, step, "approve")

	return result, nil
}

func (s *TaskServiceImpl) Reject(ctx context.Context, taskID uuid.UUID, userID string, req *TaskActionParams) (*model.ActionResult, error) {
	// 1. 获取任务并校验权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.AssigneeID != userID {
		return nil, errors.New(errors.ErrForbidden, "无权处理此任务")
	}
	if task.Status != model.TaskStatusPending {
		return nil, errors.New(errors.ErrTaskNotPending, "")
	}

	// 2. 获取步骤信息
	step, err := s.stepRepo.GetByID(ctx, task.StepID)
	if err != nil {
		return nil, err
	}

	// 3. 完成任务
	now := time.Now()
	action := model.TaskActionReject
	task.Status = model.TaskStatusRejected
	task.Action = &action
	if req != nil {
		task.Comment = &req.Comment
	}
	task.CompletedAt = &now

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// 4. 标记步骤为 rejected
	step.Status = model.StepStatusRejected
	step.CompletionResult = utils.Ptr("reject")
	step.CompletedAt = &now
	s.stepRepo.Update(ctx, step)

	// 5. 终止流程
	instance, _ := s.instanceRepo.GetByID(ctx, task.InstanceID)
	if instance != nil {
		instance.Status = model.ProcessStatusRejected
		instance.CompletedAt = &now
		s.instanceRepo.Update(ctx, instance)
	}

	// 6. 跳过其他 pending 步骤
	steps, _ := s.stepRepo.ListByInstance(ctx, task.InstanceID)
	for _, step := range steps {
		if step.Status == model.StepStatusPending {
			step.Status = model.StepStatusSkipped
			s.stepRepo.Update(ctx, &step)
		}
	}

	result := &model.ActionResult{
		TaskID:      taskID,
		InstanceID:  task.InstanceID,
		Action:      "reject",
		IsCompleted: true,
	}

	// 7. 发送通知
	s.sendActionNotification(ctx, task, step, "reject")

	return result, nil
}

func (s *TaskServiceImpl) Return(ctx context.Context, taskID uuid.UUID, userID string, req *ReturnParams) (*model.ActionResult, error) {
	// 1. 获取任务并校验权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.AssigneeID != userID {
		return nil, errors.New(errors.ErrForbidden, "无权处理此任务")
	}
	if task.Status != model.TaskStatusPending {
		return nil, errors.New(errors.ErrTaskNotPending, "")
	}

	// 2. 获取步骤信息
	step, err := s.stepRepo.GetByID(ctx, task.StepID)
	if err != nil {
		return nil, err
	}

	// 3. 完成任务
	now := time.Now()
	action := model.TaskActionReturn
	task.Status = model.TaskStatusReturned
	task.Action = &action
	if req != nil {
		task.Comment = &req.Comment
	}
	task.CompletedAt = &now

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// 4. 标记当前步骤为 returned
	step.Status = model.StepStatusReturned
	step.CompletionResult = utils.Ptr("return")
	step.CompletedAt = &now
	s.stepRepo.Update(ctx, step)

	// 5. 确定退回目标
	var targetStep *model.ApprovalStep
	if req != nil && req.ReturnToStep != nil {
		// 退回到指定步骤
		steps, _ := s.stepRepo.ListByInstance(ctx, task.InstanceID)
		for _, s := range steps {
			if s.StepIndex == *req.ReturnToStep {
				targetStep = &s
				break
			}
		}
	} else {
		// 默认退回到上一步
		steps, _ := s.stepRepo.ListByInstance(ctx, task.InstanceID)
		for i := len(steps) - 1; i >= 0; i-- {
			if steps[i].StepIndex < step.StepIndex && steps[i].Status == model.StepStatusCompleted {
				targetStep = &steps[i]
				break
			}
		}
	}

	if targetStep == nil {
		// 没有可退回的步骤，退回到发起人
		instance, _ := s.instanceRepo.GetByID(ctx, task.InstanceID)
		if instance != nil {
			instance.Status = model.ProcessStatusRunning
			instance.CurrentStepIndex = 0
			s.instanceRepo.Update(ctx, instance)
		}
		return &model.ActionResult{
			TaskID:     taskID,
			InstanceID: task.InstanceID,
			Action:     "return",
		}, nil
	}

	// 6. 重置目标步骤
	targetStep.Status = model.StepStatusActive
	targetStep.CompletionResult = nil
	targetStep.CompletedAt = nil
	s.stepRepo.Update(ctx, targetStep)

	// 7. 为目标步骤创建新任务
	var assignees []model.ApproverRef
	json.Unmarshal(targetStep.Assignees, &assignees)

	var nextTaskIDs []uuid.UUID
	for _, assignee := range assignees {
		newTask := &model.Task{
			InstanceID:  task.InstanceID,
			StepID:      targetStep.ID,
			AssigneeID:  assignee.Value,
			IsDelegated: assignee.IsDelegated,
			Status:      model.TaskStatusPending,
		}
		if assignee.IsDelegated {
			newTask.OriginalAssigneeID = &assignee.OriginalValue
		}

		created, err := s.taskRepo.Create(ctx, newTask)
		if err != nil {
			continue
		}
		nextTaskIDs = append(nextTaskIDs, created.ID)
	}

	// 8. 更新实例当前步骤
	s.instanceRepo.UpdateCurrentStep(ctx, task.InstanceID, targetStep.StepIndex)

	result := &model.ActionResult{
		TaskID:      taskID,
		InstanceID:  task.InstanceID,
		Action:      "return",
		NextTaskIDs: nextTaskIDs,
	}

	// 9. 发送通知
	s.sendActionNotification(ctx, task, targetStep, "return")

	return result, nil
}

func (s *TaskServiceImpl) Countersign(ctx context.Context, taskID uuid.UUID, userID string, req *CountersignParams) (*model.ActionResult, error) {
	// 1. 获取任务并校验权限
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.AssigneeID != userID {
		return nil, errors.New(errors.ErrForbidden, "无权处理此任务")
	}
	if task.Status != model.TaskStatusPending {
		return nil, errors.New(errors.ErrTaskNotPending, "")
	}

	// 2. 获取当前步骤
	currentStep, err := s.stepRepo.GetByID(ctx, task.StepID)
	if err != nil {
		return nil, err
	}

	// 3. 查找下一个步骤的位置
	nextStep, _ := s.stepRepo.GetNextPendingStep(ctx, task.InstanceID, currentStep.StepIndex)

	// 4. 计算新步骤的索引
	newStepIndex := currentStep.StepIndex + 0.5
	if nextStep != nil {
		newStepIndex = (currentStep.StepIndex + nextStep.StepIndex) / 2
	}

	// 5. 创建新步骤
	assignees, _ := json.Marshal(req.Assignees)
	newStep := &model.ApprovalStep{
		InstanceID:      task.InstanceID,
		StepIndex:       newStepIndex,
		Type:            req.Type,
		Assignees:       assignees,
		JointSignPolicy: req.JointSignPolicy,
		Status:          model.StepStatusPending,
		Source:          model.StepSourceCountersign,
		AddedByUserID:   &userID,
	}

	if err := s.stepRepo.Create(ctx, newStep); err != nil {
		return nil, err
	}

	// 6. 创建新任务
	var nextTaskIDs []uuid.UUID
	for _, assignee := range req.Assignees {
		newTask := &model.Task{
			InstanceID: task.InstanceID,
			StepID:     newStep.ID,
			AssigneeID: assignee.Value,
			Status:     model.TaskStatusPending,
		}
		created, err := s.taskRepo.Create(ctx, newTask)
		if err != nil {
			continue
		}
		nextTaskIDs = append(nextTaskIDs, created.ID)
	}

	// 7. 完成当前任务（加签后仍需审批）
	action := model.TaskActionCountersign
	task.Action = &action
	task.Comment = &req.Comment
	s.taskRepo.Update(ctx, task)

	result := &model.ActionResult{
		TaskID:      taskID,
		InstanceID:  task.InstanceID,
		Action:      "countersign",
		NextTaskIDs: nextTaskIDs,
	}

	// 8. 发送通知
	s.sendActionNotification(ctx, task, newStep, "countersign")

	return result, nil
}

func (s *TaskServiceImpl) MarkNotifyRead(ctx context.Context, taskID uuid.UUID, userID string) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}
	if task.AssigneeID != userID {
		return errors.New(errors.ErrForbidden, "无权处理此任务")
	}
	if task.Status != model.TaskStatusPending {
		return errors.New(errors.ErrTaskNotPending, "")
	}

	// 获取步骤类型
	step, _ := s.stepRepo.GetByID(ctx, task.StepID)
	if step == nil || step.Type != model.StepTypeNotify {
		return errors.New(errors.ErrInvalidParam, "此任务不是通知类型")
	}

	// 标记为已完成
	now := time.Now()
	action := model.TaskActionNotifyRead
	task.Status = model.TaskStatusCompleted
	task.Action = &action
	task.CompletedAt = &now

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return err
	}

	// 检查是否需要推进
	shouldProceed, _ := s.checkStepCompletion(ctx, step)
	if shouldProceed {
		s.proceedToNextStep(ctx, step, "approve")
	}

	return nil
}

// ==================== 辅助方法 ====================

func (s *TaskServiceImpl) checkStepCompletion(ctx context.Context, step *model.ApprovalStep) (bool, error) {
	tasks, err := s.taskRepo.ListByStep(ctx, step.ID)
	if err != nil {
		return false, err
	}

	// 通知类型步骤自动完成
	if step.Type == model.StepTypeNotify {
		return true, nil
	}

	// 普通审批（单人）
	if step.Type == model.StepTypeApproval && len(tasks) == 1 {
		return tasks[0].Status == model.TaskStatusCompleted, nil
	}

	// 会签逻辑
	if step.Type == model.StepTypeJointSign {
		var completedCount, totalCount int
		for _, t := range tasks {
			totalCount++
			if t.Status == model.TaskStatusCompleted {
				completedCount++
			}
		}

		policy := step.JointSignPolicy
		switch policy {
		case model.JointSignAllPass:
			return completedCount == totalCount, nil
		case model.JointSignAnyOne:
			return completedCount >= 1, nil
		case model.JointSignMajority:
			return completedCount > totalCount/2, nil
		default:
			return completedCount == totalCount, nil
		}
	}

	return false, nil
}

func (s *TaskServiceImpl) proceedToNextStep(ctx context.Context, currentStep *model.ApprovalStep, result string) ([]uuid.UUID, bool, error) {
	// 标记当前步骤完成
	now := time.Now()
	currentStep.Status = model.StepStatusCompleted
	currentStep.CompletionResult = &result
	currentStep.CompletedAt = &now
	s.stepRepo.Update(ctx, currentStep)

	// 查找下一个 pending 步骤
	nextStep, _ := s.stepRepo.GetNextPendingStep(ctx, currentStep.InstanceID, currentStep.StepIndex)

	// 没有下一步，流程完成
	if nextStep == nil {
		instance, _ := s.instanceRepo.GetByID(ctx, currentStep.InstanceID)
		if instance != nil {
			instance.Status = model.ProcessStatusCompleted
			instance.CompletedAt = &now
			s.instanceRepo.Update(ctx, instance)

			// 发送完成通知
			s.sendCompleteNotification(ctx, instance)
		}
		return nil, true, nil
	}

	// 激活下一步
	nextStep.Status = model.StepStatusActive
	s.stepRepo.Update(ctx, nextStep)

	// 创建新任务
	var assignees []model.ApproverRef
	json.Unmarshal(nextStep.Assignees, &assignees)

	var nextTaskIDs []uuid.UUID
	for _, assignee := range assignees {
		task := &model.Task{
			InstanceID:  currentStep.InstanceID,
			StepID:      nextStep.ID,
			AssigneeID:  assignee.Value,
			IsDelegated: assignee.IsDelegated,
			Status:      model.TaskStatusPending,
		}
		if assignee.IsDelegated {
			task.OriginalAssigneeID = &assignee.OriginalValue
		}

		created, err := s.taskRepo.Create(ctx, task)
		if err != nil {
			continue
		}
		nextTaskIDs = append(nextTaskIDs, created.ID)
	}

	// 更新实例当前步骤
	s.instanceRepo.UpdateCurrentStep(ctx, currentStep.InstanceID, nextStep.StepIndex)

	return nextTaskIDs, false, nil
}

func (s *TaskServiceImpl) sendActionNotification(ctx context.Context, task *model.Task, step *model.ApprovalStep, action string) {
	event := &extension.NotifyEvent{
		EventType:   action,
		InstanceID:  task.InstanceID,
		TaskID:      &task.ID,
		StepID:      &task.StepID,
		RecipientID: task.AssigneeID,
		Action:      &action,
		Comment:     task.Comment,
		Timestamp:   time.Now().Unix(),
	}
	s.mq.Publish(ctx, "workflow.notifications", event)
}

func (s *TaskServiceImpl) sendCompleteNotification(ctx context.Context, instance *model.ProcessInstance) {
	event := &extension.NotifyEvent{
		EventType:   "complete",
		InstanceID:  instance.ID,
		RecipientID: instance.SubmittedBy,
		Timestamp:   time.Now().Unix(),
	}
	s.mq.Publish(ctx, "workflow.notifications", event)
}
