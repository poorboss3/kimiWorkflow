package repository

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"workflow-engine/internal/model"
	"workflow-engine/internal/pkg/errors"
)

// TaskRepository 任务仓库接口
type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) (*model.Task, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error)
	ListByStep(ctx context.Context, stepID uuid.UUID) ([]model.Task, error)
	ListByAssignee(ctx context.Context, assigneeID string, status *model.TaskStatus, page, size int) ([]model.Task, int, error)
	ListCompletedByAssignee(ctx context.Context, assigneeID string, page, size int) ([]model.Task, int, error)
	Update(ctx context.Context, task *model.Task) error
	CountPendingByAssignee(ctx context.Context, assigneeID string) (int, error)
	CountUrgentByAssignee(ctx context.Context, assigneeID string) (int, error)
}

// MemoryTaskRepository 内存任务仓库
type MemoryTaskRepository struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*model.Task
}

// NewMemoryTaskRepository 创建内存任务仓库
func NewMemoryTaskRepository() TaskRepository {
	return &MemoryTaskRepository{
		data: make(map[uuid.UUID]*model.Task),
	}
}

func (r *MemoryTaskRepository) Create(ctx context.Context, task *model.Task) (*model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.InitBaseModel()
	r.data[task.ID] = task
	return task, nil
}

func (r *MemoryTaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.data[id]
	if !ok {
		return nil, errors.New(errors.ErrTaskNotFound, "")
	}
	return task, nil
}

func (r *MemoryTaskRepository) ListByStep(ctx context.Context, stepID uuid.UUID) ([]model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []model.Task
	for _, task := range r.data {
		if task.StepID == stepID {
			tasks = append(tasks, *task)
		}
	}
	return tasks, nil
}

func (r *MemoryTaskRepository) ListByAssignee(ctx context.Context, assigneeID string, status *model.TaskStatus, page, size int) ([]model.Task, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []model.Task
	for _, task := range r.data {
		if task.AssigneeID != assigneeID {
			continue
		}
		if status != nil && task.Status != *status {
			continue
		}
		tasks = append(tasks, *task)
	}

	// 排序：加急优先，然后按创建时间
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].IsUrgent != tasks[j].IsUrgent {
			return tasks[i].IsUrgent
		}
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})

	total := len(tasks)
	start, end := paginate(total, page, size)
	return tasks[start:end], total, nil
}

func (r *MemoryTaskRepository) ListCompletedByAssignee(ctx context.Context, assigneeID string, page, size int) ([]model.Task, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tasks []model.Task
	for _, task := range r.data {
		// 当前处理人或原始处理人
		if task.AssigneeID != assigneeID && (task.OriginalAssigneeID == nil || *task.OriginalAssigneeID != assigneeID) {
			continue
		}
		// 已完成状态
		if task.Status != model.TaskStatusCompleted && task.Status != model.TaskStatusReturned && task.Status != model.TaskStatusRejected {
			continue
		}
		tasks = append(tasks, *task)
	}

	// 按完成时间倒序
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].CompletedAt == nil || tasks[j].CompletedAt == nil {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		}
		return tasks[i].CompletedAt.After(*tasks[j].CompletedAt)
	})

	total := len(tasks)
	start, end := paginate(total, page, size)
	return tasks[start:end], total, nil
}

func (r *MemoryTaskRepository) Update(ctx context.Context, task *model.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[task.ID]; !ok {
		return errors.New(errors.ErrTaskNotFound, "")
	}

	task.UpdatedAt = time.Now()
	r.data[task.ID] = task
	return nil
}

func (r *MemoryTaskRepository) CountPendingByAssignee(ctx context.Context, assigneeID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, task := range r.data {
		if task.AssigneeID == assigneeID && task.Status == model.TaskStatusPending {
			count++
		}
	}
	return count, nil
}

func (r *MemoryTaskRepository) CountUrgentByAssignee(ctx context.Context, assigneeID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, task := range r.data {
		if task.AssigneeID == assigneeID && task.Status == model.TaskStatusPending && task.IsUrgent {
			count++
		}
	}
	return count, nil
}
