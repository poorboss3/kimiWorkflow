package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ProcessDefinition 流程定义
type ProcessDefinition struct {
	BaseModel
	Name            string          `json:"name"`
	Version         int             `json:"version"`
	Status          DefStatus       `json:"status"`
	NodeTemplates   json.RawMessage `json:"nodeTemplates,omitempty"`
	RuleSetID       *uuid.UUID      `json:"ruleSetId,omitempty"`
	ExtensionPoints json.RawMessage `json:"extensionPoints,omitempty"`
}

// ExtensionPointsConfig 扩展点配置
type ExtensionPointsConfig struct {
	ApproverResolverURL    string `json:"approverResolverUrl"`
	PermissionValidatorURL string `json:"permissionValidatorUrl"`
	TimeoutSeconds         int    `json:"timeoutSeconds"`
}

// ProcessInstance 流程实例
type ProcessInstance struct {
	BaseModel
	DefinitionID        uuid.UUID       `json:"definitionId"`
	DefinitionVersion   int             `json:"definitionVersion"`
	BusinessKey         string          `json:"businessKey"`
	FormDataSnapshot    json.RawMessage `json:"formDataSnapshot,omitempty"`
	SubmittedBy         string          `json:"submittedBy"`
	OnBehalfOf          *string         `json:"onBehalfOf,omitempty"`
	Status              ProcessStatus   `json:"status"`
	IsUrgent            bool            `json:"isUrgent"`
	CurrentStepIndex    float64         `json:"currentStepIndex"`
	CompletedAt         *time.Time      `json:"completedAt,omitempty"`
	
	// 关联（非持久化）
	Steps []ApprovalStep `json:"steps,omitempty"`
}

// ApprovalStep 审批步骤
type ApprovalStep struct {
	BaseModel
	InstanceID         uuid.UUID       `json:"instanceId"`
	StepIndex          float64         `json:"stepIndex"`
	Type               StepType        `json:"type"`
	Assignees          json.RawMessage `json:"assignees"` // []ApproverRef
	JointSignPolicy    JointSignPolicy `json:"jointSignPolicy,omitempty"`
	Status             StepStatus      `json:"status"`
	Source             StepSource      `json:"source"`
	AddedByUserID      *string         `json:"addedByUserId,omitempty"`
	CompletedAt        *time.Time      `json:"completedAt,omitempty"`
	CompletionResult   *string         `json:"completionResult,omitempty"`
	
	// 关联（非持久化）
	Tasks []Task `json:"tasks,omitempty"`
}

// ApproverRef 审批人引用
type ApproverRef struct {
	Type          string `json:"type"`                    // user | role | position | department_head | direct_supervisor
	Value         string `json:"value"`                   // 对应值
	Name          string `json:"name,omitempty"`          // 显示名称
	OriginalValue string `json:"originalValue,omitempty"` // 原始值（委托场景）
	IsDelegated   bool   `json:"isDelegated,omitempty"`   // 是否委托
}

// ApproverListModification 审批列表修改记录
type ApproverListModification struct {
	BaseModel
	InstanceID     uuid.UUID       `json:"instanceId"`
	ModifiedBy     string          `json:"modifiedBy"`
	OriginalSteps  json.RawMessage `json:"originalSteps"`
	FinalSteps     json.RawMessage `json:"finalSteps"`
	DiffSummary    json.RawMessage `json:"diffSummary"`
}

// DiffItem 差异项
type DiffItem struct {
	Action     string  `json:"action"`     // added | removed | replaced
	StepIndex  float64 `json:"stepIndex"`
	AssigneeID string  `json:"assigneeId,omitempty"`
	From       string  `json:"from,omitempty"`
	To         string  `json:"to,omitempty"`
}

// StepConfig 步骤配置（用于API）
type StepConfig struct {
	Type            StepType      `json:"type" binding:"required,oneof=approval joint_sign notify"`
	Assignees       []ApproverRef `json:"assignees" binding:"required,min=1,dive"`
	JointSignPolicy JointSignPolicy `json:"jointSignPolicy,omitempty" binding:"omitempty,oneof=ALL_PASS ANY_ONE MAJORITY"`
}

// NodeTemplate 节点模板
type NodeTemplate struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// ProcessInstanceDetail 流程实例详情
type ProcessInstanceDetail struct {
	ProcessInstance
	DefinitionName string                  `json:"definitionName"`
	History        []*ApprovalHistoryItem  `json:"history"`
}

// ApprovalHistoryItem 审批历史项
type ApprovalHistoryItem struct {
	StepIndex       float64     `json:"stepIndex"`
	StepType        StepType    `json:"stepType"`
	AssigneeID      string      `json:"assigneeId"`
	AssigneeName    string      `json:"assigneeName"`
	OriginalAssigneeID string   `json:"originalAssigneeId,omitempty"`
	IsDelegated     bool        `json:"isDelegated"`
	Action          string      `json:"action"`
	Comment         string      `json:"comment,omitempty"`
	CompletedAt     *time.Time  `json:"completedAt,omitempty"`
}
