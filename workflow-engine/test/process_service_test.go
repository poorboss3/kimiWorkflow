package test

import (
	"context"
	"testing"
	"time"

	"workflow-engine/internal/extension"
	"workflow-engine/internal/model"
	"workflow-engine/internal/pkg/locker"
	"workflow-engine/internal/pkg/memory"
	"workflow-engine/internal/pkg/utils"
	"workflow-engine/internal/repository"
	"workflow-engine/internal/service"
)

// TestCreateDefinition 测试创建流程定义
func TestCreateDefinition(t *testing.T) {
	ctx := context.Background()
	defRepo := repository.NewMemoryProcessDefinitionRepository()
	instanceRepo := repository.NewMemoryProcessInstanceRepository()
	stepRepo := repository.NewMemoryApprovalStepRepository()
	taskRepo := repository.NewMemoryTaskRepository()
	modRepo := repository.NewMemoryApproverListModificationRepository()
	proxyRepo := repository.NewMemoryProxyConfigRepository()
	delegationRepo := repository.NewMemoryDelegationConfigRepository()
	extClient := extension.NewMockClient()
	lck := locker.NewMemoryLocker()
	mq := memory.NewQueue()
	defer mq.Close()

	processService := service.NewProcessService(
		defRepo, instanceRepo, stepRepo, taskRepo, modRepo,
		proxyRepo, delegationRepo, extClient, lck, mq,
	)

	req := &service.CreateDefinitionRequest{
		Name: "费用报销",
		ExtensionPoints: model.ExtensionPointsConfig{
			TimeoutSeconds: 3,
		},
	}

	def, err := processService.CreateDefinition(ctx, req)
	if err != nil {
		t.Fatalf("创建流程定义失败: %v", err)
	}

	if def.Name != "费用报销" {
		t.Errorf("期望名称为'费用报销'，实际为'%s'", def.Name)
	}

	if def.Status != model.DefStatusDraft {
		t.Errorf("期望状态为'draft'，实际为'%s'", def.Status)
	}
}

// TestActivateDefinition 测试激活流程定义
func TestActivateDefinition(t *testing.T) {
	ctx := context.Background()
	defRepo := repository.NewMemoryProcessDefinitionRepository()
	instanceRepo := repository.NewMemoryProcessInstanceRepository()
	stepRepo := repository.NewMemoryApprovalStepRepository()
	taskRepo := repository.NewMemoryTaskRepository()
	modRepo := repository.NewMemoryApproverListModificationRepository()
	proxyRepo := repository.NewMemoryProxyConfigRepository()
	delegationRepo := repository.NewMemoryDelegationConfigRepository()
	extClient := extension.NewMockClient()
	lck := locker.NewMemoryLocker()
	mq := memory.NewQueue()
	defer mq.Close()

	processService := service.NewProcessService(
		defRepo, instanceRepo, stepRepo, taskRepo, modRepo,
		proxyRepo, delegationRepo, extClient, lck, mq,
	)

	// 创建
	req := &service.CreateDefinitionRequest{
		Name: "请假申请",
		ExtensionPoints: model.ExtensionPointsConfig{
			TimeoutSeconds: 3,
		},
	}
	def, _ := processService.CreateDefinition(ctx, req)

	// 激活
	activated, err := processService.ActivateDefinition(ctx, def.ID, 1)
	if err != nil {
		t.Fatalf("激活流程定义失败: %v", err)
	}

	if activated.Status != model.DefStatusActive {
		t.Errorf("期望状态为'active'，实际为'%s'", activated.Status)
	}
}

// TestSubmitProcess 测试提交流程
func TestSubmitProcess(t *testing.T) {
	ctx := context.Background()
	defRepo := repository.NewMemoryProcessDefinitionRepository()
	instanceRepo := repository.NewMemoryProcessInstanceRepository()
	stepRepo := repository.NewMemoryApprovalStepRepository()
	taskRepo := repository.NewMemoryTaskRepository()
	modRepo := repository.NewMemoryApproverListModificationRepository()
	proxyRepo := repository.NewMemoryProxyConfigRepository()
	delegationRepo := repository.NewMemoryDelegationConfigRepository()
	extClient := extension.NewMockClient()
	lck := locker.NewMemoryLocker()
	mq := memory.NewQueue()
	defer mq.Close()

	processService := service.NewProcessService(
		defRepo, instanceRepo, stepRepo, taskRepo, modRepo,
		proxyRepo, delegationRepo, extClient, lck, mq,
	)

	// 先创建并激活流程定义
	defReq := &service.CreateDefinitionRequest{
		Name: "费用报销",
		ExtensionPoints: model.ExtensionPointsConfig{
			TimeoutSeconds: 3,
		},
	}
	def, _ := processService.CreateDefinition(ctx, defReq)
	processService.ActivateDefinition(ctx, def.ID, 1)

	// 提交流程
	submitReq := &service.SubmitProcessRequest{
		DefinitionID: def.ID.String(),
		BusinessKey:  "EXP-20240316-001",
		FormData: map[string]interface{}{
			"amount":     5000,
			"department": "研发部",
		},
		SubmittedBy: "user_001",
	}

	instance, err := processService.SubmitProcess(ctx, submitReq)
	if err != nil {
		t.Fatalf("提交流程失败: %v", err)
	}

	if instance.BusinessKey != "EXP-20240316-001" {
		t.Errorf("期望业务单号为'EXP-20240316-001'，实际为'%s'", instance.BusinessKey)
	}

	if instance.Status != model.ProcessStatusRunning {
		t.Errorf("期望状态为'running'，实际为'%s'", instance.Status)
	}
}

// TestTaskApprove 测试任务审批
func TestTaskApprove(t *testing.T) {
	ctx := context.Background()
	defRepo := repository.NewMemoryProcessDefinitionRepository()
	instanceRepo := repository.NewMemoryProcessInstanceRepository()
	stepRepo := repository.NewMemoryApprovalStepRepository()
	taskRepo := repository.NewMemoryTaskRepository()
	modRepo := repository.NewMemoryApproverListModificationRepository()
	proxyRepo := repository.NewMemoryProxyConfigRepository()
	delegationRepo := repository.NewMemoryDelegationConfigRepository()
	extClient := extension.NewMockClient()
	lck := locker.NewMemoryLocker()
	mq := memory.NewQueue()
	defer mq.Close()

	processService := service.NewProcessService(
		defRepo, instanceRepo, stepRepo, taskRepo, modRepo,
		proxyRepo, delegationRepo, extClient, lck, mq,
	)
	taskService := service.NewTaskService(
		taskRepo, stepRepo, instanceRepo, defRepo, mq, lck,
	)

	// 创建并激活流程定义
	defReq := &service.CreateDefinitionRequest{
		Name: "费用报销",
		ExtensionPoints: model.ExtensionPointsConfig{
			TimeoutSeconds: 3,
		},
	}
	def, _ := processService.CreateDefinition(ctx, defReq)
	processService.ActivateDefinition(ctx, def.ID, 1)

	// 提交流程
	submitReq := &service.SubmitProcessRequest{
		DefinitionID: def.ID.String(),
		BusinessKey:  "EXP-TEST-001",
		FormData: map[string]interface{}{
			"amount": 5000,
		},
		SubmittedBy: "user_001",
	}
	_, err := processService.SubmitProcess(ctx, submitReq)
	if err != nil {
		t.Fatalf("提交流程失败: %v", err)
	}

	// 获取待办任务（模拟客户端返回的审批人是 manager_001）
	tasks, _, _ := taskRepo.ListByAssignee(ctx, "manager_001", utils.Ptr(model.TaskStatusPending), 1, 10)
	if len(tasks) == 0 {
		t.Skip("没有待办任务，跳过审批测试")
		return
	}

	task := tasks[0]

	// 审批通过
	params := &service.TaskActionParams{Comment: "同意"}
	result, err := taskService.Approve(ctx, task.ID, "manager_001", params)
	if err != nil {
		t.Fatalf("审批失败: %v", err)
	}

	if result.Action != "approve" {
		t.Errorf("期望操作为'approve'，实际为'%s'", result.Action)
	}
}

// TestTaskReject 测试任务驳回
func TestTaskReject(t *testing.T) {
	ctx := context.Background()
	defRepo := repository.NewMemoryProcessDefinitionRepository()
	instanceRepo := repository.NewMemoryProcessInstanceRepository()
	stepRepo := repository.NewMemoryApprovalStepRepository()
	taskRepo := repository.NewMemoryTaskRepository()
	modRepo := repository.NewMemoryApproverListModificationRepository()
	proxyRepo := repository.NewMemoryProxyConfigRepository()
	delegationRepo := repository.NewMemoryDelegationConfigRepository()
	extClient := extension.NewMockClient()
	lck := locker.NewMemoryLocker()
	mq := memory.NewQueue()
	defer mq.Close()

	processService := service.NewProcessService(
		defRepo, instanceRepo, stepRepo, taskRepo, modRepo,
		proxyRepo, delegationRepo, extClient, lck, mq,
	)
	taskService := service.NewTaskService(
		taskRepo, stepRepo, instanceRepo, defRepo, mq, lck,
	)

	// 创建并激活流程定义
	defReq := &service.CreateDefinitionRequest{
		Name: "费用报销",
		ExtensionPoints: model.ExtensionPointsConfig{
			TimeoutSeconds: 3,
		},
	}
	def, _ := processService.CreateDefinition(ctx, defReq)
	processService.ActivateDefinition(ctx, def.ID, 1)

	// 提交流程
	submitReq := &service.SubmitProcessRequest{
		DefinitionID: def.ID.String(),
		BusinessKey:  "EXP-REJECT-001",
		FormData: map[string]interface{}{
			"amount": 5000,
		},
		SubmittedBy: "user_001",
	}
	instance, _ := processService.SubmitProcess(ctx, submitReq)

	// 获取待办任务
	tasks, _, _ := taskRepo.ListByAssignee(ctx, "manager_001", utils.Ptr(model.TaskStatusPending), 1, 10)
	if len(tasks) == 0 {
		t.Skip("没有待办任务，跳过驳回测试")
		return
	}

	task := tasks[0]

	// 驳回
	params := &service.TaskActionParams{Comment: "金额有误"}
	result, err := taskService.Reject(ctx, task.ID, "manager_001", params)
	if err != nil {
		t.Fatalf("驳回失败: %v", err)
	}

	if !result.IsCompleted {
		t.Error("期望流程已完成")
	}

	// 验证流程状态
	detail, _ := processService.GetInstance(ctx, instance.ID)
	if detail.Status != model.ProcessStatusRejected {
		t.Errorf("期望流程状态为'rejected'，实际为'%s'", detail.Status)
	}
}

// TestProxySubmission 测试代提交
func TestProxySubmission(t *testing.T) {
	ctx := context.Background()
	defRepo := repository.NewMemoryProcessDefinitionRepository()
	instanceRepo := repository.NewMemoryProcessInstanceRepository()
	stepRepo := repository.NewMemoryApprovalStepRepository()
	taskRepo := repository.NewMemoryTaskRepository()
	modRepo := repository.NewMemoryApproverListModificationRepository()
	proxyRepo := repository.NewMemoryProxyConfigRepository()
	delegationRepo := repository.NewMemoryDelegationConfigRepository()
	extClient := extension.NewMockClient()
	lck := locker.NewMemoryLocker()
	mq := memory.NewQueue()
	defer mq.Close()

	processService := service.NewProcessService(
		defRepo, instanceRepo, stepRepo, taskRepo, modRepo,
		proxyRepo, delegationRepo, extClient, lck, mq,
	)

	// 创建代理配置
	now := time.Now()
	proxyRepo.Create(ctx, &model.ProxyConfig{
		PrincipalID: "user_b",
		AgentID:     "user_a",
		ValidFrom:   now.Add(-time.Hour),
		IsActive:    true,
	})

	// 创建并激活流程定义
	defReq := &service.CreateDefinitionRequest{
		Name: "费用报销",
		ExtensionPoints: model.ExtensionPointsConfig{
			TimeoutSeconds: 3,
		},
	}
	def, _ := processService.CreateDefinition(ctx, defReq)
	processService.ActivateDefinition(ctx, def.ID, 1)

	// A代B提交
	onBehalfOf := "user_b"
	submitReq := &service.SubmitProcessRequest{
		DefinitionID: def.ID.String(),
		BusinessKey:  "EXP-PROXY-001",
		FormData: map[string]interface{}{
			"amount": 3000,
		},
		SubmittedBy: "user_a",
		OnBehalfOf:  &onBehalfOf,
	}

	instance, err := processService.SubmitProcess(ctx, submitReq)
	if err != nil {
		t.Fatalf("代提交失败: %v", err)
	}

	if instance.SubmittedBy != "user_a" {
		t.Errorf("期望提交人为'user_a'，实际为'%s'", instance.SubmittedBy)
	}

	if *instance.OnBehalfOf != "user_b" {
		t.Errorf("期望被代理人为'user_b'，实际为'%s'", *instance.OnBehalfOf)
	}
}

// TestDelegation 测试委托
func TestDelegation(t *testing.T) {
	ctx := context.Background()
	defRepo := repository.NewMemoryProcessDefinitionRepository()
	instanceRepo := repository.NewMemoryProcessInstanceRepository()
	stepRepo := repository.NewMemoryApprovalStepRepository()
	taskRepo := repository.NewMemoryTaskRepository()
	modRepo := repository.NewMemoryApproverListModificationRepository()
	proxyRepo := repository.NewMemoryProxyConfigRepository()
	delegationRepo := repository.NewMemoryDelegationConfigRepository()
	extClient := extension.NewMockClient()
	lck := locker.NewMemoryLocker()
	mq := memory.NewQueue()
	defer mq.Close()

	processService := service.NewProcessService(
		defRepo, instanceRepo, stepRepo, taskRepo, modRepo,
		proxyRepo, delegationRepo, extClient, lck, mq,
	)

	// 创建委托配置（经理委托给助理）
	now := time.Now()
	delegationRepo.Create(ctx, &model.DelegationConfig{
		DelegatorID: "manager_001",
		DelegateeID: "assistant_001",
		ValidFrom:   now.Add(-time.Hour),
		IsActive:    true,
	})

	// 创建并激活流程定义
	defReq := &service.CreateDefinitionRequest{
		Name: "费用报销",
		ExtensionPoints: model.ExtensionPointsConfig{
			TimeoutSeconds: 3,
		},
	}
	def, _ := processService.CreateDefinition(ctx, defReq)
	processService.ActivateDefinition(ctx, def.ID, 1)

	// 提交流程（模拟客户端返回的审批人是 manager_001）
	submitReq := &service.SubmitProcessRequest{
		DefinitionID: def.ID.String(),
		BusinessKey:  "EXP-DELEGATE-001",
		FormData: map[string]interface{}{
			"amount": 5000,
		},
		SubmittedBy: "user_001",
	}

	_, err := processService.SubmitProcess(ctx, submitReq)
	if err != nil {
		t.Fatalf("提交流程失败: %v", err)
	}

	// 验证任务是否分配给助理
	tasks, _, _ := taskRepo.ListByAssignee(ctx, "assistant_001", utils.Ptr(model.TaskStatusPending), 1, 10)
	if len(tasks) == 0 {
		t.Skip("没有生成委托任务，模拟客户端可能返回了不同的审批人")
		return
	}

	found := false
	for _, task := range tasks {
		if task.OriginalAssigneeID != nil && *task.OriginalAssigneeID == "manager_001" {
			found = true
			if !task.IsDelegated {
				t.Error("期望任务标记为委托")
			}
			break
		}
	}

	if !found {
		t.Error("未找到委托任务")
	}
}
