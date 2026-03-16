package extension

import (
	"encoding/json"

	"github.com/google/uuid"
	"workflow-engine/internal/model"
)

// ResolveRequest 审批人解析请求
type ResolveRequest struct {
	ProcessType string          `json:"processType" binding:"required"`
	FormData    json.RawMessage `json:"formData" binding:"required"`
	SubmittedBy string          `json:"submittedBy" binding:"required"`
	OnBehalfOf  *string         `json:"onBehalfOf,omitempty"`
	BusinessKey string          `json:"businessKey,omitempty"`
	RequestID   string          `json:"requestId" binding:"required"`
}

// ResolveResult 解析结果
type ResolveResult struct {
	Steps []model.StepConfig `json:"steps" binding:"required"`
}

// ResolveResponse 扩展点响应
type ResolveResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    *ResolveResult `json:"data,omitempty"`
}

// ValidateRequest 权限验证请求
type ValidateRequest struct {
	ProcessType   string             `json:"processType" binding:"required"`
	FormData      json.RawMessage    `json:"formData" binding:"required"`
	SubmittedBy   string             `json:"submittedBy" binding:"required"`
	OriginalSteps []model.StepConfig `json:"originalSteps" binding:"required"`
	FinalSteps    []model.StepConfig `json:"finalSteps" binding:"required"`
	IsModified    bool               `json:"isModified"`
	RequestID     string             `json:"requestId" binding:"required"`
}

// ValidationFailedItem 验证失败项
type ValidationFailedItem struct {
	StepIndex  float64 `json:"stepIndex"`
	AssigneeID string  `json:"assigneeId"`
	Reason     string  `json:"reason"`
}

// ValidateResponse 验证响应
type ValidateResponse struct {
	Passed      bool                   `json:"passed"`
	FailedItems []ValidationFailedItem `json:"failedItems,omitempty"`
	Message     string                 `json:"message"`
}

// NotifyEvent 通知事件（MQ消息）
type NotifyEvent struct {
	EventType   string          `json:"eventType"` // submit | approve | reject | return | complete | urgent | delegate
	InstanceID  uuid.UUID       `json:"instanceId"`
	TaskID      *uuid.UUID      `json:"taskId,omitempty"`
	StepID      *uuid.UUID      `json:"stepId,omitempty"`
	RecipientID string          `json:"recipientId"`
	ProcessType string          `json:"processType"`
	BusinessKey string          `json:"businessKey"`
	FormData    json.RawMessage `json:"formData,omitempty"`
	IsUrgent    bool            `json:"isUrgent"`
	Action      *string         `json:"action,omitempty"`
	Comment     *string         `json:"comment,omitempty"`
	Timestamp   int64           `json:"timestamp"`
}
