package repository

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"workflow-engine/internal/model"
	"workflow-engine/internal/pkg/errors"
)

// ProcessDefinitionRepository 流程定义仓库接口
type ProcessDefinitionRepository interface {
	Create(ctx context.Context, def *model.ProcessDefinition) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessDefinition, error)
	GetByNameAndVersion(ctx context.Context, name string, version int) (*model.ProcessDefinition, error)
	List(ctx context.Context, status *model.DefStatus, name string, page, size int) ([]*model.ProcessDefinition, int, error)
	Update(ctx context.Context, def *model.ProcessDefinition) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProcessInstanceRepository 流程实例仓库接口
type ProcessInstanceRepository interface {
	Create(ctx context.Context, instance *model.ProcessInstance) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error)
	GetByBusinessKey(ctx context.Context, businessKey string) (*model.ProcessInstance, error)
	List(ctx context.Context, query *InstanceQuery) ([]*model.ProcessInstance, int, error)
	Update(ctx context.Context, instance *model.ProcessInstance) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.ProcessStatus) error
	UpdateCurrentStep(ctx context.Context, id uuid.UUID, stepIndex float64) error
}

// ApprovalStepRepository 审批步骤仓库接口
type ApprovalStepRepository interface {
	Create(ctx context.Context, step *model.ApprovalStep) error
	CreateBatch(ctx context.Context, steps []*model.ApprovalStep) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ApprovalStep, error)
	ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.ApprovalStep, error)
	GetNextPendingStep(ctx context.Context, instanceID uuid.UUID, currentIndex float64) (*model.ApprovalStep, error)
	Update(ctx context.Context, step *model.ApprovalStep) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status model.StepStatus) error
}

// ApproverListModificationRepository 审批列表修改记录仓库接口
type ApproverListModificationRepository interface {
	Create(ctx context.Context, mod *model.ApproverListModification) error
	ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.ApproverListModification, error)
}

// InstanceQuery 实例查询条件
type InstanceQuery struct {
	SubmittedBy string
	Status      *model.ProcessStatus
	BusinessKey string
	Page        int
	Size        int
}

// ==================== 内存实现 ====================

// MemoryProcessDefinitionRepository 内存流程定义仓库
type MemoryProcessDefinitionRepository struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*model.ProcessDefinition
}

// NewMemoryProcessDefinitionRepository 创建内存流程定义仓库
func NewMemoryProcessDefinitionRepository() ProcessDefinitionRepository {
	return &MemoryProcessDefinitionRepository{
		data: make(map[uuid.UUID]*model.ProcessDefinition),
	}
}

func (r *MemoryProcessDefinitionRepository) Create(ctx context.Context, def *model.ProcessDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	def.InitBaseModel()
	r.data[def.ID] = def
	return nil
}

func (r *MemoryProcessDefinitionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.data[id]
	if !ok {
		return nil, errors.New(errors.ErrDefinitionNotFound, "")
	}
	return def, nil
}

func (r *MemoryProcessDefinitionRepository) GetByNameAndVersion(ctx context.Context, name string, version int) (*model.ProcessDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, def := range r.data {
		if def.Name == name && def.Version == version {
			return def, nil
		}
	}
	return nil, errors.New(errors.ErrDefinitionNotFound, "")
}

func (r *MemoryProcessDefinitionRepository) List(ctx context.Context, status *model.DefStatus, name string, page, size int) ([]*model.ProcessDefinition, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*model.ProcessDefinition
	for _, def := range r.data {
		if status != nil && def.Status != *status {
			continue
		}
		if name != "" && !strings.Contains(def.Name, name) {
			continue
		}
		list = append(list, def)
	}

	// 按创建时间倒序
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	total := len(list)
	start, end := paginate(total, page, size)
	return list[start:end], total, nil
}

func (r *MemoryProcessDefinitionRepository) Update(ctx context.Context, def *model.ProcessDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[def.ID]; !ok {
		return errors.New(errors.ErrDefinitionNotFound, "")
	}

	def.UpdatedAt = time.Now()
	r.data[def.ID] = def
	return nil
}

func (r *MemoryProcessDefinitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.data, id)
	return nil
}

// MemoryProcessInstanceRepository 内存流程实例仓库
type MemoryProcessInstanceRepository struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*model.ProcessInstance
}

// NewMemoryProcessInstanceRepository 创建内存流程实例仓库
func NewMemoryProcessInstanceRepository() ProcessInstanceRepository {
	return &MemoryProcessInstanceRepository{
		data: make(map[uuid.UUID]*model.ProcessInstance),
	}
}

func (r *MemoryProcessInstanceRepository) Create(ctx context.Context, instance *model.ProcessInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance.InitBaseModel()
	r.data[instance.ID] = instance
	return nil
}

func (r *MemoryProcessInstanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ProcessInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instance, ok := r.data[id]
	if !ok {
		return nil, errors.New(errors.ErrInstanceNotFound, "")
	}
	return instance, nil
}

func (r *MemoryProcessInstanceRepository) GetByBusinessKey(ctx context.Context, businessKey string) (*model.ProcessInstance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, instance := range r.data {
		if instance.BusinessKey == businessKey {
			return instance, nil
		}
	}
	return nil, errors.New(errors.ErrInstanceNotFound, "")
}

func (r *MemoryProcessInstanceRepository) List(ctx context.Context, query *InstanceQuery) ([]*model.ProcessInstance, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*model.ProcessInstance
	for _, instance := range r.data {
		if query.SubmittedBy != "" && instance.SubmittedBy != query.SubmittedBy {
			continue
		}
		if query.Status != nil && instance.Status != *query.Status {
			continue
		}
		if query.BusinessKey != "" && !strings.Contains(instance.BusinessKey, query.BusinessKey) {
			continue
		}
		list = append(list, instance)
	}

	// 按创建时间倒序
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.After(list[j].CreatedAt)
	})

	total := len(list)
	start, end := paginate(total, query.Page, query.Size)
	return list[start:end], total, nil
}

func (r *MemoryProcessInstanceRepository) Update(ctx context.Context, instance *model.ProcessInstance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[instance.ID]; !ok {
		return errors.New(errors.ErrInstanceNotFound, "")
	}

	instance.UpdatedAt = time.Now()
	r.data[instance.ID] = instance
	return nil
}

func (r *MemoryProcessInstanceRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.ProcessStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, ok := r.data[id]
	if !ok {
		return errors.New(errors.ErrInstanceNotFound, "")
	}

	instance.Status = status
	instance.UpdatedAt = time.Now()
	if status == model.ProcessStatusCompleted || status == model.ProcessStatusRejected {
		now := time.Now()
		instance.CompletedAt = &now
	}
	return nil
}

func (r *MemoryProcessInstanceRepository) UpdateCurrentStep(ctx context.Context, id uuid.UUID, stepIndex float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instance, ok := r.data[id]
	if !ok {
		return errors.New(errors.ErrInstanceNotFound, "")
	}

	instance.CurrentStepIndex = stepIndex
	instance.UpdatedAt = time.Now()
	return nil
}

// MemoryApprovalStepRepository 内存审批步骤仓库
type MemoryApprovalStepRepository struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*model.ApprovalStep
}

// NewMemoryApprovalStepRepository 创建内存审批步骤仓库
func NewMemoryApprovalStepRepository() ApprovalStepRepository {
	return &MemoryApprovalStepRepository{
		data: make(map[uuid.UUID]*model.ApprovalStep),
	}
}

func (r *MemoryApprovalStepRepository) Create(ctx context.Context, step *model.ApprovalStep) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	step.InitBaseModel()
	r.data[step.ID] = step
	return nil
}

func (r *MemoryApprovalStepRepository) CreateBatch(ctx context.Context, steps []*model.ApprovalStep) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, step := range steps {
		step.InitBaseModel()
		r.data[step.ID] = step
	}
	return nil
}

func (r *MemoryApprovalStepRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ApprovalStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	step, ok := r.data[id]
	if !ok {
		return nil, fmt.Errorf("step not found")
	}
	return step, nil
}

func (r *MemoryApprovalStepRepository) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.ApprovalStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var steps []model.ApprovalStep
	for _, step := range r.data {
		if step.InstanceID == instanceID {
			steps = append(steps, *step)
		}
	}

	// 按 stepIndex 排序
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].StepIndex < steps[j].StepIndex
	})

	return steps, nil
}

func (r *MemoryApprovalStepRepository) GetNextPendingStep(ctx context.Context, instanceID uuid.UUID, currentIndex float64) (*model.ApprovalStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var candidates []*model.ApprovalStep
	for _, step := range r.data {
		if step.InstanceID == instanceID && step.StepIndex > currentIndex && step.Status == model.StepStatusPending {
			candidates = append(candidates, step)
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	// 找到最小的 stepIndex
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].StepIndex < candidates[j].StepIndex
	})

	return candidates[0], nil
}

func (r *MemoryApprovalStepRepository) Update(ctx context.Context, step *model.ApprovalStep) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[step.ID]; !ok {
		return fmt.Errorf("step not found")
	}

	step.UpdatedAt = time.Now()
	r.data[step.ID] = step
	return nil
}

func (r *MemoryApprovalStepRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.StepStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	step, ok := r.data[id]
	if !ok {
		return fmt.Errorf("step not found")
	}

	step.Status = status
	step.UpdatedAt = time.Now()
	return nil
}

// MemoryApproverListModificationRepository 内存审批列表修改记录仓库
type MemoryApproverListModificationRepository struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*model.ApproverListModification
}

// NewMemoryApproverListModificationRepository 创建内存审批列表修改记录仓库
func NewMemoryApproverListModificationRepository() ApproverListModificationRepository {
	return &MemoryApproverListModificationRepository{
		data: make(map[uuid.UUID]*model.ApproverListModification),
	}
}

func (r *MemoryApproverListModificationRepository) Create(ctx context.Context, mod *model.ApproverListModification) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mod.InitBaseModel()
	r.data[mod.ID] = mod
	return nil
}

func (r *MemoryApproverListModificationRepository) ListByInstance(ctx context.Context, instanceID uuid.UUID) ([]model.ApproverListModification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var mods []model.ApproverListModification
	for _, mod := range r.data {
		if mod.InstanceID == instanceID {
			mods = append(mods, *mod)
		}
	}

	// 按创建时间倒序
	sort.Slice(mods, func(i, j int) bool {
		return mods[i].CreatedAt.After(mods[j].CreatedAt)
	})

	return mods, nil
}

// paginate 分页辅助函数
func paginate(total, page, size int) (start, end int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	start = (page - 1) * size
	end = start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	return
}
