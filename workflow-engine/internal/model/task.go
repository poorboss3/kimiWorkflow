package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Task 审批任务
type Task struct {
	BaseModel
	InstanceID         uuid.UUID   `json:"instanceId"`
	StepID             uuid.UUID   `json:"stepId"`
	AssigneeID         string      `json:"assigneeId"`
	OriginalAssigneeID *string     `json:"originalAssigneeId,omitempty"`
	IsDelegated        bool        `json:"isDelegated"`
	Status             TaskStatus  `json:"status"`
	IsUrgent           bool        `json:"isUrgent"`
	Action             *TaskAction `json:"action,omitempty"`
	Comment            *string     `json:"comment,omitempty"`
	CompletedAt        *time.Time  `json:"completedAt,omitempty"`
	
	// 关联（非持久化）
	Instance *ProcessInstance `json:"instance,omitempty"`
	Step     *ApprovalStep    `json:"step,omitempty"`
}

// TaskListItem 任务列表项（聚合查询结果）
type TaskListItem struct {
	Task
	ProcessName   string          `json:"processName"`
	InitiatorID   string          `json:"initiatorId"`
	InitiatorName string          `json:"initiatorName"`
	SubmittedAt   time.Time       `json:"submittedAt"`
	PendingHours  int             `json:"pendingHours"`
	FormSummary   json.RawMessage `json:"formSummary,omitempty"`
}

// TaskDetail 任务详情
type TaskDetail struct {
	Task
	ProcessName    string                 `json:"processName"`
	DefinitionID   uuid.UUID              `json:"definitionId"`
	BusinessKey    string                 `json:"businessKey"`
	FormData       map[string]interface{} `json:"formData,omitempty"`
	SubmittedBy    string                 `json:"submittedBy"`
	OnBehalfOf     *string                `json:"onBehalfOf,omitempty"`
	CanReturn      bool                   `json:"canReturn"`
	CanReject      bool                   `json:"canReject"`
	Steps          []StepInfo             `json:"steps,omitempty"`
}

// StepInfo 步骤信息
type StepInfo struct {
	StepIndex float64    `json:"stepIndex"`
	Type      StepType   `json:"type"`
	Status    StepStatus `json:"status"`
	Assignees []string   `json:"assignees"`
}

// TaskActionRequest 任务操作请求
type TaskActionRequest struct {
	Action          TaskAction       `json:"action" binding:"required,oneof=approve reject return delegate countersign notify_read"`
	Comment         string           `json:"comment" binding:"max=500"`
	ReturnToStep    *float64         `json:"returnToStep,omitempty"`
	CountersignData *CountersignData `json:"countersignData,omitempty"`
}

// CountersignData 加签数据
type CountersignData struct {
	Assignees       []ApproverRef   `json:"assignees" binding:"required,min=1"`
	Type            StepType        `json:"type,omitempty" binding:"omitempty,oneof=approval joint_sign"`
	JointSignPolicy JointSignPolicy `json:"jointSignPolicy,omitempty"`
}

// TaskStatistics 任务统计
type TaskStatistics struct {
	PendingCount   int `json:"pendingCount"`
	CompletedCount int `json:"completedCount"`
	UrgentCount    int `json:"urgentCount"`
}

// ActionResult 操作结果
type ActionResult struct {
	TaskID      uuid.UUID   `json:"taskId"`
	InstanceID  uuid.UUID   `json:"instanceId"`
	Action      string      `json:"action"`
	NextTaskIDs []uuid.UUID `json:"nextTaskIds,omitempty"`
	IsCompleted bool        `json:"isCompleted"`
}
