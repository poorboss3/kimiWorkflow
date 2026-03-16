package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ApprovalRule 审批规则
type ApprovalRule struct {
	BaseModel
	Name                string          `json:"name"`
	Priority            int             `json:"priority"`
	ProcessDefinitionID *uuid.UUID      `json:"processDefinitionId,omitempty"`
	Conditions          json.RawMessage `json:"conditions"` // []Condition
	ConditionLogic      string          `json:"conditionLogic"`
	Result              json.RawMessage `json:"result"` // RuleResult
	IsActive            bool            `json:"isActive"`
}

// Condition 规则条件
type Condition struct {
	Field    string      `json:"field" binding:"required"`
	Operator string      `json:"operator" binding:"required,oneof=eq neq gt gte lt lte in contains regex"`
	Value    interface{} `json:"value" binding:"required"`
}

// RuleResult 规则结果
type RuleResult struct {
	Steps []StepConfig `json:"steps" binding:"required"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	BaseModel
	PrincipalID         string    `json:"principalId"`
	AgentID             string    `json:"agentId"`
	AllowedProcessTypes *string   `json:"allowedProcessTypes,omitempty"` // JSON 数组
	ValidFrom           time.Time `json:"validFrom"`
	ValidTo             *time.Time `json:"validTo,omitempty"`
	IsActive            bool      `json:"isActive"`
}

// DelegationConfig 委托配置
type DelegationConfig struct {
	BaseModel
	DelegatorID         string    `json:"delegatorId"`
	DelegateeID         string    `json:"delegateeId"`
	AllowedProcessTypes *string   `json:"allowedProcessTypes,omitempty"`
	ValidFrom           time.Time `json:"validFrom"`
	ValidTo             *time.Time `json:"validTo,omitempty"`
	IsActive            bool      `json:"isActive"`
	Reason              *string   `json:"reason,omitempty"`
}

// PrincipalInfo 委托人信息
type PrincipalInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
