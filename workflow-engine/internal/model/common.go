package model

import (
	"time"

	"github.com/google/uuid"
)

// BaseModel 基础模型
type BaseModel struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ProcessStatus 流程状态
type ProcessStatus string

const (
	ProcessStatusRunning   ProcessStatus = "running"
	ProcessStatusCompleted ProcessStatus = "completed"
	ProcessStatusRejected  ProcessStatus = "rejected"
	ProcessStatusWithdrawn ProcessStatus = "withdrawn"
)

// StepStatus 步骤状态
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusActive    StepStatus = "active"
	StepStatusCompleted StepStatus = "completed"
	StepStatusRejected  StepStatus = "rejected"
	StepStatusReturned  StepStatus = "returned"
	StepStatusSkipped   StepStatus = "skipped"
)

// StepType 步骤类型
type StepType string

const (
	StepTypeApproval  StepType = "approval"
	StepTypeJointSign StepType = "joint_sign"
	StepTypeNotify    StepType = "notify"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusReturned  TaskStatus = "returned"
	TaskStatusRejected  TaskStatus = "rejected"
)

// TaskAction 任务操作
type TaskAction string

const (
	TaskActionApprove     TaskAction = "approve"
	TaskActionReject      TaskAction = "reject"
	TaskActionReturn      TaskAction = "return"
	TaskActionDelegate    TaskAction = "delegate"
	TaskActionCountersign TaskAction = "countersign"
	TaskActionNotifyRead  TaskAction = "notify_read"
)

// JointSignPolicy 会签策略
type JointSignPolicy string

const (
	JointSignAllPass  JointSignPolicy = "ALL_PASS"
	JointSignAnyOne   JointSignPolicy = "ANY_ONE"
	JointSignMajority JointSignPolicy = "MAJORITY"
)

// StepSource 步骤来源
type StepSource string

const (
	StepSourceOriginal     StepSource = "original"
	StepSourceCountersign  StepSource = "countersign"
	StepSourceDynamicAdded StepSource = "dynamic_added"
)

// DefStatus 定义状态
type DefStatus string

const (
	DefStatusDraft    DefStatus = "draft"
	DefStatusActive   DefStatus = "active"
	DefStatusArchived DefStatus = "archived"
)

// InitBaseModel 初始化基础模型
func (m *BaseModel) InitBaseModel() {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	now := time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = now
	}
	m.UpdatedAt = now
}
