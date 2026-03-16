# 工作流引擎接口实现详细设计文档

> 技术栈：Go 1.21+ / Gin / GORM / PostgreSQL / Redis / RabbitMQ  
> 版本：1.0 | 日期：2026-03-16

---

## 目录

1. [技术架构](#一技术架构)
2. [项目结构](#二项目结构)
3. [数据模型定义](#三数据模型定义)
4. [数据库表设计](#四数据库表设计)
5. [API接口设计](#五api接口设计)
6. [服务层接口定义](#六服务层接口定义)
7. [扩展点实现](#七扩展点实现)
8. [核心业务流程实现](#八核心业务流程实现)
9. [并发控制方案](#九并发控制方案)
10. [缓存策略](#十缓存策略)
11. [错误码与响应规范](#十一错误码与响应规范)
12. [配置与部署](#十二配置与部署)

---

## 一、技术架构

### 1.1 技术选型

| 层次 | 选型 | 版本 | 说明 |
|------|------|------|------|
| Web框架 | Gin | v1.9+ | 高性能、中间件丰富 |
| ORM | GORM | v1.25+ | 功能完善，支持 PostgreSQL |
| 数据库 | PostgreSQL | 14+ | JSONB支持，事务可靠 |
| 缓存 | go-redis | v9+ | Redis 7.0+ 兼容 |
| 消息队列 | RabbitMQ | 3.12+ | amqp091-go 客户端 |
| 配置管理 | Viper | v1.18+ | 多格式配置支持 |
| 日志 | Zap | v1.26+ | 高性能结构化日志 |
| 校验 | go-playground/validator | v10+ | 结构体验证 |
| 任务调度 | go-co-op/gocron | v2+ | 定时任务（委托过期检查） |
| 文档 | swaggo/swag | v1.16+ | Swagger API 文档 |

### 1.2 架构分层

```
┌─────────────────────────────────────────────────────────────┐
│                        API Layer (Gin)                       │
│  ProcessHandler / TaskHandler / AdminHandler / CallbackHandler│
├─────────────────────────────────────────────────────────────┤
│                      Service Layer                           │
│  ProcessService / TaskService / RuleEngine / NotifyService   │
├─────────────────────────────────────────────────────────────┤
│                      Repository Layer (GORM)                 │
│  ProcessRepo / TaskRepo / RuleRepo / ProxyRepo / DelegationRepo│
├─────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                      │
│  PostgreSQL    Redis    RabbitMQ    HTTP Client    Scheduler │
└─────────────────────────────────────────────────────────────┘
```

---

## 二、项目结构

```
workflow-engine/
├── api/
│   ├── handler/              # HTTP 处理器
│   │   ├── process.go        # 流程相关接口
│   │   ├── task.go           # 任务相关接口
│   │   ├── admin.go          # 管理后台接口
│   │   └── callback.go       # 扩展点回调处理
│   ├── middleware/           # 中间件
│   │   ├── auth.go           # 认证中间件
│   │   ├── logger.go         # 日志中间件
│   │   └── recovery.go       # 恢复中间件
│   └── router.go             # 路由注册
├── internal/
│   ├── service/              # 业务服务层
│   │   ├── process.go        # 流程服务
│   │   ├── task.go           # 任务服务
│   │   ├── rule.go           # 规则引擎服务
│   │   ├── notify.go         # 通知服务
│   │   ├── proxy.go          # 代理服务
│   │   └── delegation.go     # 委托服务
│   ├── repository/           # 数据访问层
│   │   ├── process.go
│   │   ├── task.go
│   │   ├── rule.go
│   │   ├── proxy.go
│   │   └── delegation.go
│   ├── model/                # 领域模型
│   │   ├── process.go        # 流程相关模型
│   │   ├── task.go           # 任务相关模型
│   │   ├── rule.go           # 规则相关模型
│   │   └── common.go         # 通用模型
│   ├── entity/               # 数据库实体（GORM模型）
│   │   ├── process.go
│   │   ├── task.go
│   │   └── config.go
│   ├── extension/            # 扩展点接口
│   │   ├── resolver.go       # 审批人解析器
│   │   ├── validator.go      # 权限验证器
│   │   └── client.go         # HTTP 扩展点客户端
│   ├── pkg/
│   │   ├── redis/            # Redis 封装
│   │   ├── mq/               # 消息队列封装
│   │   ├── locker/           # 分布式锁
│   │   ├── errors/           # 错误码定义
│   │   └── utils/            # 工具函数
│   └── config/               # 配置定义
├── pkg/
│   └── sdk/                  # 对外SDK（业务系统集成用）
├── scripts/
│   ├── init_db.sql           # 数据库初始化脚本
│   └── migrate/              # 迁移脚本
├── docs/                     # Swagger 文档
├── configs/
│   └── config.yaml           # 配置文件模板
├── go.mod
└── main.go
```

---

## 三、数据模型定义

### 3.1 基础类型定义

```go
// internal/model/common.go

package model

import (
    "time"
    "github.com/google/uuid"
)

// BaseModel 基础模型
type BaseModel struct {
    ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    CreatedAt time.Time `json:"createdAt" gorm:"index"`
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
    TaskActionApprove    TaskAction = "approve"
    TaskActionReject     TaskAction = "reject"
    TaskActionReturn     TaskAction = "return"
    TaskActionDelegate   TaskAction = "delegate"
    TaskActionCountersign TaskAction = "countersign"
    TaskActionNotifyRead TaskAction = "notify_read"
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
```

### 3.2 流程定义模型

```go
// internal/model/process.go

package model

import (
    "encoding/json"
    "time"
)

// ProcessDefinition 流程定义
type ProcessDefinition struct {
    BaseModel
    Name            string          `json:"name" gorm:"size:100;not null;index"`
    Version         int             `json:"version" gorm:"not null;default:1"`
    Status          DefStatus       `json:"status" gorm:"size:20;not null;default:'draft'"`
    NodeTemplates   json.RawMessage `json:"nodeTemplates" gorm:"type:jsonb"`
    RuleSetID       *uuid.UUID      `json:"ruleSetId" gorm:"type:uuid;index"`
    ExtensionPoints json.RawMessage `json:"extensionPoints" gorm:"type:jsonb"`
}

// DefStatus 定义状态
type DefStatus string

const (
    DefStatusDraft    DefStatus = "draft"
    DefStatusActive   DefStatus = "active"
    DefStatusArchived DefStatus = "archived"
)

// ExtensionPointsConfig 扩展点配置
type ExtensionPointsConfig struct {
    ApproverResolverURL    string `json:"approverResolverUrl"`
    PermissionValidatorURL string `json:"permissionValidatorUrl"`
    TimeoutSeconds         int    `json:"timeoutSeconds"`
}

// ProcessInstance 流程实例
type ProcessInstance struct {
    BaseModel
    DefinitionID        uuid.UUID       `json:"definitionId" gorm:"type:uuid;not null;index"`
    DefinitionVersion   int             `json:"definitionVersion" gorm:"not null"`
    BusinessKey         string          `json:"businessKey" gorm:"size:100;not null;index"`
    FormDataSnapshot    json.RawMessage `json:"formDataSnapshot" gorm:"type:jsonb"`
    SubmittedBy         string          `json:"submittedBy" gorm:"size:64;not null;index"`
    OnBehalfOf          *string         `json:"onBehalfOf" gorm:"size:64;index"`
    Status              ProcessStatus   `json:"status" gorm:"size:20;not null;default:'running';index"`
    IsUrgent            bool            `json:"isUrgent" gorm:"default:false"`
    CurrentStepIndex    float64         `json:"currentStepIndex" gorm:"default:0"`
    CompletedAt         *time.Time      `json:"completedAt"`
    
    // 关联
    Steps []ApprovalStep `json:"steps,omitempty" gorm:"foreignKey:InstanceID"`
}

// ApprovalStep 审批步骤
type ApprovalStep struct {
    BaseModel
    InstanceID         uuid.UUID       `json:"instanceId" gorm:"type:uuid;not null;index:idx_instance_step"`
    StepIndex          float64         `json:"stepIndex" gorm:"not null;index:idx_instance_step"`
    Type               StepType        `json:"type" gorm:"size:20;not null"`
    Assignees          json.RawMessage `json:"assignees" gorm:"type:jsonb"` // []ApproverRef
    JointSignPolicy    JointSignPolicy `json:"jointSignPolicy" gorm:"size:20"`
    Status             StepStatus      `json:"status" gorm:"size:20;not null;default:'pending';index"`
    Source             StepSource      `json:"source" gorm:"size:20;default:'original'"`
    AddedByUserID      *string         `json:"addedByUserId" gorm:"size:64"`
    CompletedAt        *time.Time      `json:"completedAt"`
    CompletionResult   *string         `json:"completionResult" gorm:"size:20"` // approve | reject | return
    
    // 关联
    Tasks []Task `json:"tasks,omitempty" gorm:"foreignKey:StepID"`
}

// ApproverRef 审批人引用
type ApproverRef struct {
    Type  string `json:"type"`  // user | role | position | department_head | direct_supervisor
    Value string `json:"value"` // 对应值
    Name  string `json:"name,omitempty"` // 显示名称
}

// ApproverListModification 审批列表修改记录
type ApproverListModification struct {
    BaseModel
    InstanceID     uuid.UUID       `json:"instanceId" gorm:"type:uuid;not null;index"`
    ModifiedBy     string          `json:"modifiedBy" gorm:"size:64;not null"`
    OriginalSteps  json.RawMessage `json:"originalSteps" gorm:"type:jsonb"`
    FinalSteps     json.RawMessage `json:"finalSteps" gorm:"type:jsonb"`
    DiffSummary    json.RawMessage `json:"diffSummary" gorm:"type:jsonb"`
}

// DiffItem 差异项
type DiffItem struct {
    Action     string  `json:"action"`     // added | removed | replaced
    StepIndex  float64 `json:"stepIndex"`
    AssigneeID string  `json:"assigneeId,omitempty"`
    From       string  `json:"from,omitempty"`
    To         string  `json:"to,omitempty"`
}
```

### 3.3 任务模型

```go
// internal/model/task.go

package model

import (
    "time"
    "github.com/google/uuid"
)

// Task 审批任务
type Task struct {
    BaseModel
    InstanceID         uuid.UUID   `json:"instanceId" gorm:"type:uuid;not null;index"`
    StepID             uuid.UUID   `json:"stepId" gorm:"type:uuid;not null;index:idx_step_assignee"`
    AssigneeID         string      `json:"assigneeId" gorm:"size:64;not null;index:idx_step_assignee"`
    OriginalAssigneeID *string     `json:"originalAssigneeId" gorm:"size:64;index"`
    IsDelegated        bool        `json:"isDelegated" gorm:"default:false"`
    Status             TaskStatus  `json:"status" gorm:"size:20;not null;default:'pending';index"`
    IsUrgent           bool        `json:"isUrgent" gorm:"default:false"`
    Action             *TaskAction `json:"action" gorm:"size:20"`
    Comment            *string     `json:"comment" gorm:"type:text"`
    CompletedAt        *time.Time  `json:"completedAt" gorm:"index"`
    
    // 关联（不持久化，仅用于查询返回）
    Instance   *ProcessInstance `json:"instance,omitempty" gorm:"-"`
    Step       *ApprovalStep    `json:"step,omitempty" gorm:"-"`
}

// TaskListItem 任务列表项（聚合查询结果）
type TaskListItem struct {
    Task
    ProcessName    string          `json:"processName"`
    InitiatorID    string          `json:"initiatorId"`
    InitiatorName  string          `json:"initiatorName"`
    SubmittedAt    time.Time       `json:"submittedAt"`
    PendingHours   int             `json:"pendingHours"`
    FormSummary    json.RawMessage `json:"formSummary"` // 表单摘要
}

// TaskActionRequest 任务操作请求
type TaskActionRequest struct {
    Action      TaskAction `json:"action" binding:"required,oneof=approve reject return delegate countersign notify_read"`
    Comment     string     `json:"comment" binding:"max=500"`
    ReturnToStep *float64  `json:"returnToStep,omitempty"` // 退回时指定步骤
    CountersignData *CountersignData `json:"countersignData,omitempty"` // 加签数据
}

// CountersignData 加签数据
type CountersignData struct {
    Assignees       []ApproverRef   `json:"assignees" binding:"required,min=1"`
    Type            StepType        `json:"type" binding:"omitempty,oneof=approval joint_sign"`
    JointSignPolicy JointSignPolicy `json:"jointSignPolicy,omitempty"`
}
```

### 3.4 规则与配置模型

```go
// internal/model/rule.go

package model

import (
    "encoding/json"
    "time"
    "github.com/google/uuid"
)

// ApprovalRule 审批规则
type ApprovalRule struct {
    BaseModel
    Name                string          `json:"name" gorm:"size:100;not null"`
    Priority            int             `json:"priority" gorm:"not null;default:0"`
    ProcessDefinitionID *uuid.UUID      `json:"processDefinitionId" gorm:"type:uuid;index"`
    Conditions          json.RawMessage `json:"conditions" gorm:"type:jsonb"` // []Condition
    ConditionLogic      string          `json:"conditionLogic" gorm:"size:10;default:'AND'"`
    Result              json.RawMessage `json:"result" gorm:"type:jsonb"` // RuleResult
    IsActive            bool            `json:"isActive" gorm:"default:true"`
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

// StepConfig 步骤配置
type StepConfig struct {
    Type            StepType      `json:"type" binding:"required"`
    Assignees       []ApproverRef `json:"assignees" binding:"required"`
    JointSignPolicy JointSignPolicy `json:"jointSignPolicy,omitempty"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
    BaseModel
    PrincipalID         string    `json:"principalId" gorm:"size:64;not null;index:idx_principal_agent"`
    AgentID             string    `json:"agentId" gorm:"size:64;not null;index:idx_principal_agent"`
    AllowedProcessTypes *string   `json:"allowedProcessTypes" gorm:"type:text"` // JSON 数组
    ValidFrom           time.Time `json:"validFrom"`
    ValidTo             *time.Time `json:"validTo"`
    IsActive            bool      `json:"isActive" gorm:"default:true;index"`
}

// DelegationConfig 委托配置
type DelegationConfig struct {
    BaseModel
    DelegatorID         string    `json:"delegatorId" gorm:"size:64;not null;index:idx_delegator"`
    DelegateeID         string    `json:"delegateeId" gorm:"size:64;not null;index:idx_delegatee"`
    AllowedProcessTypes *string   `json:"allowedProcessTypes" gorm:"type:text"`
    ValidFrom           time.Time `json:"validFrom"`
    ValidTo             *time.Time `json:"validTo"`
    IsActive            bool      `json:"isActive" gorm:"default:true;index"`
    Reason              *string   `json:"reason" gorm:"size:200"`
}
```

### 3.5 扩展点请求/响应模型

```go
// internal/extension/types.go

package extension

import (
    "encoding/json"
    "github.com/google/uuid"
    "workflow-engine/internal/model"
)

// ResolveRequest 审批人解析请求
type ResolveRequest struct {
    ProcessType   string          `json:"processType" binding:"required"`
    FormData      json.RawMessage `json:"formData" binding:"required"`
    SubmittedBy   string          `json:"submittedBy" binding:"required"`
    OnBehalfOf    *string         `json:"onBehalfOf"`
    BusinessKey   string          `json:"businessKey,omitempty"`
    RequestID     string          `json:"requestId" binding:"required"`
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
    ProcessType   string               `json:"processType" binding:"required"`
    FormData      json.RawMessage      `json:"formData" binding:"required"`
    SubmittedBy   string               `json:"submittedBy" binding:"required"`
    OriginalSteps []model.StepConfig   `json:"originalSteps" binding:"required"`
    FinalSteps    []model.StepConfig   `json:"finalSteps" binding:"required"`
    IsModified    bool                 `json:"isModified"`
    RequestID     string               `json:"requestId" binding:"required"`
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
    EventType     string          `json:"eventType"` // submit | approve | reject | return | complete | urgent | delegate
    InstanceID    uuid.UUID       `json:"instanceId"`
    TaskID        *uuid.UUID      `json:"taskId,omitempty"`
    StepID        *uuid.UUID      `json:"stepId,omitempty"`
    RecipientID   string          `json:"recipientId"`
    ProcessType   string          `json:"processType"`
    BusinessKey   string          `json:"businessKey"`
    FormData      json.RawMessage `json:"formData,omitempty"`
    IsUrgent      bool            `json:"isUrgent"`
    Action        *string         `json:"action,omitempty"`
    Comment       *string         `json:"comment,omitempty"`
    Timestamp     int64           `json:"timestamp"`
}
```

---

## 四、数据库表设计

### 4.1 初始化 SQL

```sql
-- scripts/init_db.sql

-- 扩展 uuid-ossp（如果未启用）
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==================== 流程定义 ====================
CREATE TABLE process_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    version INT NOT NULL DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'archived')),
    node_templates JSONB,
    rule_set_id UUID,
    extension_points JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(name, version)
);

CREATE INDEX idx_pd_status ON process_definitions(status);
CREATE INDEX idx_pd_rule_set ON process_definitions(rule_set_id);

-- ==================== 流程实例 ====================
CREATE TABLE process_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    definition_id UUID NOT NULL REFERENCES process_definitions(id),
    definition_version INT NOT NULL,
    business_key VARCHAR(100) NOT NULL,
    form_data_snapshot JSONB,
    submitted_by VARCHAR(64) NOT NULL,
    on_behalf_of VARCHAR(64),
    status VARCHAR(20) NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'completed', 'rejected', 'withdrawn')),
    is_urgent BOOLEAN DEFAULT FALSE,
    current_step_index NUMERIC(10,4) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_pi_definition ON process_instances(definition_id);
CREATE INDEX idx_pi_business_key ON process_instances(business_key);
CREATE INDEX idx_pi_submitted_by ON process_instances(submitted_by);
CREATE INDEX idx_pi_on_behalf_of ON process_instances(on_behalf_of);
CREATE INDEX idx_pi_status ON process_instances(status);
CREATE INDEX idx_pi_created_at ON process_instances(created_at);
CREATE INDEX idx_pi_urgent ON process_instances(is_urgent, status);

-- ==================== 审批步骤 ====================
CREATE TABLE approval_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES process_instances(id) ON DELETE CASCADE,
    step_index NUMERIC(10,4) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('approval', 'joint_sign', 'notify')),
    assignees JSONB NOT NULL,
    joint_sign_policy VARCHAR(20) CHECK (joint_sign_policy IN ('ALL_PASS', 'ANY_ONE', 'MAJORITY')),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'active', 'completed', 'rejected', 'returned', 'skipped')),
    source VARCHAR(20) DEFAULT 'original' CHECK (source IN ('original', 'countersign', 'dynamic_added')),
    added_by_user_id VARCHAR(64),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    completion_result VARCHAR(20),
    UNIQUE(instance_id, step_index)
);

CREATE INDEX idx_as_instance ON approval_steps(instance_id);
CREATE INDEX idx_as_status ON approval_steps(status);
CREATE INDEX idx_as_instance_status ON approval_steps(instance_id, status);

-- ==================== 任务表 ====================
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES process_instances(id),
    step_id UUID NOT NULL REFERENCES approval_steps(id),
    assignee_id VARCHAR(64) NOT NULL,
    original_assignee_id VARCHAR(64),
    is_delegated BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'returned', 'rejected')),
    is_urgent BOOLEAN DEFAULT FALSE,
    action VARCHAR(20) CHECK (action IN ('approve', 'reject', 'return', 'delegate', 'countersign', 'notify_read')),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_task_assignee ON tasks(assignee_id);
CREATE INDEX idx_task_assignee_status ON tasks(assignee_id, status);
CREATE INDEX idx_task_step ON tasks(step_id);
CREATE INDEX idx_task_instance ON tasks(instance_id);
CREATE INDEX idx_task_completed ON tasks(completed_at);
CREATE INDEX idx_task_original ON tasks(original_assignee_id);

-- ==================== 审批规则 ====================
CREATE TABLE approval_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    priority INT NOT NULL DEFAULT 0,
    process_definition_id UUID REFERENCES process_definitions(id),
    conditions JSONB NOT NULL,
    condition_logic VARCHAR(10) DEFAULT 'AND',
    result JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ar_definition ON approval_rules(process_definition_id);
CREATE INDEX idx_ar_active ON approval_rules(is_active);

-- ==================== 审批列表修改记录 ====================
CREATE TABLE approver_list_modifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES process_instances(id),
    modified_by VARCHAR(64) NOT NULL,
    original_steps JSONB NOT NULL,
    final_steps JSONB NOT NULL,
    diff_summary JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_alm_instance ON approver_list_modifications(instance_id);

-- ==================== 代理配置 ====================
CREATE TABLE proxy_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    principal_id VARCHAR(64) NOT NULL,
    agent_id VARCHAR(64) NOT NULL,
    allowed_process_types TEXT, -- JSON array
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_to TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(principal_id, agent_id, valid_from)
);

CREATE INDEX idx_pc_principal ON proxy_configs(principal_id);
CREATE INDEX idx_pc_agent ON proxy_configs(agent_id);
CREATE INDEX idx_pc_active ON proxy_configs(is_active, valid_from, valid_to);

-- ==================== 委托配置 ====================
CREATE TABLE delegation_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    delegator_id VARCHAR(64) NOT NULL,
    delegatee_id VARCHAR(64) NOT NULL,
    allowed_process_types TEXT,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_to TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    reason VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_dc_delegator ON delegation_configs(delegator_id);
CREATE INDEX idx_dc_delegatee ON delegation_configs(delegatee_id);
CREATE INDEX idx_dc_active ON delegation_configs(is_active, valid_from, valid_to);

-- 更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_pd_updated_at BEFORE UPDATE ON process_definitions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pi_updated_at BEFORE UPDATE ON process_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_as_updated_at BEFORE UPDATE ON approval_steps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_task_updated_at BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ar_updated_at BEFORE UPDATE ON approval_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pc_updated_at BEFORE UPDATE ON proxy_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dc_updated_at BEFORE UPDATE ON delegation_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

---

## 五、API接口设计

### 5.1 RESTful API 概览

| 分组 | 路径前缀 | 说明 |
|------|----------|------|
| 流程管理 | `/api/v1/processes` | 流程定义、实例操作 |
| 任务管理 | `/api/v1/tasks` | 待办/已办、任务操作 |
| 管理后台 | `/api/v1/admin` | 规则配置、代理委托管理 |
| 扩展点 | `/internal/callback` | 业务系统回调（内部） |

### 5.2 流程定义接口

```go
// api/handler/process.go

// CreateDefinitionRequest 创建流程定义请求
type CreateDefinitionRequest struct {
    Name            string                     `json:"name" binding:"required,max=100"`
    NodeTemplates   []model.NodeTemplate       `json:"nodeTemplates" binding:"dive"`
    RuleSetID       *uuid.UUID                 `json:"ruleSetId"`
    ExtensionPoints model.ExtensionPointsConfig `json:"extensionPoints" binding:"required"`
}

// UpdateDefinitionRequest 更新流程定义请求
type UpdateDefinitionRequest struct {
    Name            string                     `json:"name" binding:"max=100"`
    NodeTemplates   []model.NodeTemplate       `json:"nodeTemplates" binding:"dive"`
    RuleSetID       *uuid.UUID                 `json:"ruleSetId"`
    ExtensionPoints *model.ExtensionPointsConfig `json:"extensionPoints"`
}

// ActivateDefinitionRequest 激活流程定义
type ActivateDefinitionRequest struct {
    Version int `json:"version" binding:"required,min=1"`
}
```

**接口定义：**

```yaml
# 创建流程定义
POST /api/v1/processes/definitions
Request: CreateDefinitionRequest
Response: model.ProcessDefinition

# 获取流程定义列表
GET /api/v1/processes/definitions
Query: page, size, status, name
Response: PageResult<ProcessDefinition>

# 获取单个流程定义
GET /api/v1/processes/definitions/:id
Response: model.ProcessDefinition

# 更新流程定义（仅草稿状态）
PUT /api/v1/processes/definitions/:id
Request: UpdateDefinitionRequest
Response: model.ProcessDefinition

# 激活流程定义
POST /api/v1/processes/definitions/:id/activate
Request: ActivateDefinitionRequest
Response: model.ProcessDefinition

# 归档流程定义
POST /api/v1/processes/definitions/:id/archive
Response: model.ProcessDefinition
```

### 5.3 流程实例接口

```go
// SubmitProcessRequest 提交流程请求
type SubmitProcessRequest struct {
    DefinitionID string                 `json:"definitionId" binding:"required,uuid"`
    BusinessKey  string                 `json:"businessKey" binding:"required,max=100"`
    FormData     map[string]interface{} `json:"formData" binding:"required"`
    FinalSteps   []model.StepConfig     `json:"finalSteps" binding:"required,min=1"`
    OnBehalfOf   *string                `json:"onBehalfOf" binding:"omitempty,max=64"`
    IsUrgent     bool                   `json:"isUrgent"`
}

// WithdrawProcessRequest 撤回流程请求
type WithdrawProcessRequest struct {
    Reason string `json:"reason" binding:"max=500"`
}

// ModifyStepsRequest 修改步骤请求
type ModifyStepsRequest struct {
    FinalSteps []model.StepConfig `json:"finalSteps" binding:"required"`
    Reason     string             `json:"reason" binding:"max=500"`
}
```

**接口定义：**

```yaml
# 提交流程
POST /api/v1/processes/instances
Request: SubmitProcessRequest
Response: model.ProcessInstance

# 获取流程实例详情
GET /api/v1/processes/instances/:id
Response: ProcessInstanceDetail

# 获取流程实例列表
GET /api/v1/processes/instances
Query: submittedBy, status, businessKey, page, size
Response: PageResult<ProcessInstance>

# 撤回流程（仅发起人可以撤回，且当前步骤未处理）
POST /api/v1/processes/instances/:id/withdraw
Request: WithdrawProcessRequest
Response: model.ProcessInstance

# 获取流程审批历史
GET /api/v1/processes/instances/:id/history
Response: []ApprovalHistoryItem

# 动态修改审批步骤
PUT /api/v1/processes/instances/:id/steps
Request: ModifyStepsRequest
Response: model.ProcessInstance
```

### 5.4 任务接口

```go
// GetPendingTasksRequest 获取待办任务请求
type GetPendingTasksRequest struct {
    Page   int    `form:"page" binding:"min=1"`
    Size   int    `form:"size" binding:"min=1,max=100"`
    IsUrgent *bool `form:"isUrgent"`
}

// GetCompletedTasksRequest 获取已办任务请求
type GetCompletedTasksRequest struct {
    Page   int    `form:"page" binding:"min=1"`
    Size   int    `form:"size" binding:"min=1,max=100"`
    StartTime *time.Time `form:"startTime" time_format:"2006-01-02"`
    EndTime   *time.Time `form:"endTime" time_format:"2006-01-02"`
}

// ProcessTaskResponse 处理任务响应
type ProcessTaskResponse struct {
    TaskID       uuid.UUID         `json:"taskId"`
    InstanceID   uuid.UUID         `json:"instanceId"`
    Action       string            `json:"action"`
    NextTaskIDs  []uuid.UUID       `json:"nextTaskIds,omitempty"`
    IsCompleted  bool              `json:"isCompleted"`
}
```

**接口定义：**

```yaml
# 获取我的待办任务
GET /api/v1/tasks/pending
Query: page, size, isUrgent
Response: PageResult<TaskListItem>

# 获取我的已办任务
GET /api/v1/tasks/completed
Query: page, size, startTime, endTime
Response: PageResult<TaskListItem>

# 获取任务详情
GET /api/v1/tasks/:id
Response: TaskDetail

# 处理任务（通过/驳回/退回/加签等）
POST /api/v1/tasks/:id/action
Request: model.TaskActionRequest
Response: ProcessTaskResponse

# 批量获取任务统计
GET /api/v1/tasks/statistics
Response: TaskStatistics
```

### 5.5 管理后台接口

```yaml
# ========== 审批规则管理 ==========

# 创建审批规则
POST /api/v1/admin/rules
Request: CreateRuleRequest
Response: model.ApprovalRule

# 获取审批规则列表
GET /api/v1/admin/rules
Query: processDefinitionId, isActive
Response: []model.ApprovalRule

# 更新审批规则
PUT /api/v1/admin/rules/:id
Request: UpdateRuleRequest
Response: model.ApprovalRule

# 删除审批规则
DELETE /api/v1/admin/rules/:id
Response: Empty

# ========== 代理配置管理 ==========

# 创建代理配置
POST /api/v1/admin/proxies
Request: CreateProxyRequest
Response: model.ProxyConfig

# 获取代理配置列表
GET /api/v1/admin/proxies
Query: principalId, agentId, isActive
Response: []model.ProxyConfig

# 更新代理配置
PUT /api/v1/admin/proxies/:id
Request: UpdateProxyRequest
Response: model.ProxyConfig

# 删除代理配置
DELETE /api/v1/admin/proxies/:id
Response: Empty

# ========== 委托配置管理 ==========

# 创建委托配置
POST /api/v1/admin/delegations
Request: CreateDelegationRequest
Response: model.DelegationConfig

# 获取委托配置列表
GET /api/v1/admin/delegations
Query: delegatorId, delegateeId, isActive
Response: []model.DelegationConfig

# 更新委托配置
PUT /api/v1/admin/delegations/:id
Request: UpdateDelegationRequest
Response: model.DelegationConfig

# 删除委托配置
DELETE /api/v1/admin/delegations/:id
Response: Empty

# 获取我的委托人列表（用于代提交时选择）
GET /api/v1/admin/proxies/my-principals
Response: []PrincipalInfo
```

---

## 六、服务层接口定义

### 6.1 流程服务接口

```go
// internal/service/process.go

package service

import (
    "context"
    "github.com/google/uuid"
    "workflow-engine/internal/model"
)

// ProcessService 流程服务接口
type ProcessService interface {
    // 流程定义管理
    CreateDefinition(ctx context.Context, req *CreateDefinitionRequest) (*model.ProcessDefinition, error)
    GetDefinition(ctx context.Context, id uuid.UUID) (*model.ProcessDefinition, error)
    ListDefinitions(ctx context.Context, query *DefinitionQuery) (*PageResult[*model.ProcessDefinition], error)
    UpdateDefinition(ctx context.Context, id uuid.UUID, req *UpdateDefinitionRequest) (*model.ProcessDefinition, error)
    ActivateDefinition(ctx context.Context, id uuid.UUID, version int) (*model.ProcessDefinition, error)
    ArchiveDefinition(ctx context.Context, id uuid.UUID) error
    
    // 流程实例管理
    SubmitProcess(ctx context.Context, req *SubmitProcessRequest) (*model.ProcessInstance, error)
    GetInstance(ctx context.Context, id uuid.UUID) (*ProcessInstanceDetail, error)
    ListInstances(ctx context.Context, query *InstanceQuery) (*PageResult[*model.ProcessInstance], error)
    WithdrawInstance(ctx context.Context, id uuid.UUID, userID, reason string) error
    GetInstanceHistory(ctx context.Context, id uuid.UUID) ([]*ApprovalHistoryItem, error)
    ModifySteps(ctx context.Context, id uuid.UUID, userID string, req *ModifyStepsRequest) error
    
    // 扩展点调用
    ResolveApprovers(ctx context.Context, definition *model.ProcessDefinition, req *ResolveApproverRequest) ([]model.StepConfig, error)
    ValidatePermissions(ctx context.Context, definition *model.ProcessDefinition, req *ValidatePermissionRequest) (*ValidateResult, error)
}

// ProcessServiceImpl 实现
type ProcessServiceImpl struct {
    defRepo         repository.ProcessDefinitionRepository
    instanceRepo    repository.ProcessInstanceRepository
    stepRepo        repository.ApprovalStepRepository
    taskRepo        repository.TaskRepository
    ruleRepo        repository.ApprovalRuleRepository
    proxyRepo       repository.ProxyConfigRepository
    delegationRepo  repository.DelegationConfigRepository
    modRepo         repository.ApproverListModificationRepository
    extensionClient extension.Client
    locker          locker.DistributedLocker
    mq              mq.MessageQueue
    notifyService   NotificationService
}
```

### 6.2 任务服务接口

```go
// internal/service/task.go

package service

import (
    "context"
    "github.com/google/uuid"
    "workflow-engine/internal/model"
)

// TaskService 任务服务接口
type TaskService interface {
    // 任务查询
    GetPendingTasks(ctx context.Context, userID string, query *TaskQuery) (*PageResult[*model.TaskListItem], error)
    GetCompletedTasks(ctx context.Context, userID string, query *TaskQuery) (*PageResult[*model.TaskListItem], error)
    GetTaskDetail(ctx context.Context, taskID uuid.UUID, userID string) (*TaskDetail, error)
    GetTaskStatistics(ctx context.Context, userID string) (*TaskStatistics, error)
    
    // 任务操作
    Approve(ctx context.Context, taskID uuid.UUID, userID string, req *ActionRequest) (*ActionResult, error)
    Reject(ctx context.Context, taskID uuid.UUID, userID string, req *ActionRequest) (*ActionResult, error)
    Return(ctx context.Context, taskID uuid.UUID, userID string, req *ReturnRequest) (*ActionResult, error)
    Countersign(ctx context.Context, taskID uuid.UUID, userID string, req *CountersignRequest) (*ActionResult, error)
    MarkNotifyRead(ctx context.Context, taskID uuid.UUID, userID string) error
    
    // 加急
    MarkUrgent(ctx context.Context, instanceID uuid.UUID, userID string) error
}

// TaskServiceImpl 实现
type TaskServiceImpl struct {
    taskRepo       repository.TaskRepository
    stepRepo       repository.ApprovalStepRepository
    instanceRepo   repository.ProcessInstanceRepository
    delegationRepo repository.DelegationConfigRepository
    locker         locker.DistributedLocker
    mq             mq.MessageQueue
    notifyService  NotificationService
}
```

### 6.3 规则引擎服务接口

```go
// internal/service/rule.go

package service

import (
    "context"
    "github.com/google/uuid"
    "workflow-engine/internal/model"
)

// RuleEngine 规则引擎接口
type RuleEngine interface {
    // Evaluate 根据表单数据评估规则，返回匹配的步骤配置
    Evaluate(ctx context.Context, processDefID *uuid.UUID, formData map[string]interface{}) ([]model.StepConfig, error)
    
    // CRUD
    CreateRule(ctx context.Context, req *CreateRuleRequest) (*model.ApprovalRule, error)
    GetRule(ctx context.Context, id uuid.UUID) (*model.ApprovalRule, error)
    ListRules(ctx context.Context, query *RuleQuery) ([]*model.ApprovalRule, error)
    UpdateRule(ctx context.Context, id uuid.UUID, req *UpdateRuleRequest) (*model.ApprovalRule, error)
    DeleteRule(ctx context.Context, id uuid.UUID) error
}

// RuleEngineImpl 实现
type RuleEngineImpl struct {
    ruleRepo repository.ApprovalRuleRepository
}

// evaluateCondition 评估单个条件
func (e *RuleEngineImpl) evaluateCondition(cond model.Condition, formData map[string]interface{}) bool {
    // 实现各种操作符：eq, neq, gt, gte, lt, lte, in, contains, regex
    // ...
}
```

### 6.4 代理委托服务接口

```go
// internal/service/proxy.go 和 delegation.go

// ProxyService 代理服务接口
type ProxyService interface {
    CreateProxy(ctx context.Context, req *CreateProxyRequest) (*model.ProxyConfig, error)
    ListProxies(ctx context.Context, query *ProxyQuery) ([]*model.ProxyConfig, error)
    ValidateProxy(ctx context.Context, agentID, principalID string, processType string) (bool, error)
    GetMyPrincipals(ctx context.Context, agentID string) ([]*PrincipalInfo, error)
}

// DelegationService 委托服务接口
type DelegationService interface {
    CreateDelegation(ctx context.Context, req *CreateDelegationRequest) (*model.DelegationConfig, error)
    ListDelegations(ctx context.Context, query *DelegationQuery) ([]*model.DelegationConfig, error)
    GetEffectiveDelegation(ctx context.Context, delegatorID string, processType string) (*model.DelegationConfig, error)
    ExpireDelegations(ctx context.Context) error // 定时任务：清理过期委托
}
```

### 6.5 通知服务接口

```go
// internal/service/notify.go

package service

import (
    "context"
    "workflow-engine/internal/extension"
)

// NotificationService 通知服务接口
type NotificationService interface {
    // 发送各类通知
    SendTaskNotification(ctx context.Context, event *extension.NotifyEvent) error
    SendProcessCompleteNotification(ctx context.Context, event *extension.NotifyEvent) error
    SendRejectNotification(ctx context.Context, event *extension.NotifyEvent) error
    SendUrgentNotification(ctx context.Context, event *extension.NotifyEvent) error
    
    // MQ 消费者处理
    HandleNotifyEvent(ctx context.Context, event *extension.NotifyEvent) error
}

// NotificationServiceImpl 实现
type NotificationServiceImpl struct {
    // 各种通知渠道客户端
    emailClient    EmailClient
    smsClient      SMSClient
    imClient       IMClient // 企业微信/钉钉
    internalClient InternalMessageClient // 站内信
}
```

---

## 七、扩展点实现

### 7.1 HTTP 扩展点客户端

```go
// internal/extension/client.go

package extension

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/google/uuid"
)

// Client 扩展点客户端接口
type Client interface {
    // ResolveApprovers 调用业务系统解析审批人
    ResolveApprovers(ctx context.Context, url string, timeoutSecs int, req *ResolveRequest) (*ResolveResult, error)
    
    // ValidatePermissions 调用业务系统验证权限
    ValidatePermissions(ctx context.Context, url string, timeoutSecs int, req *ValidateRequest) (*ValidateResponse, error)
}

// HTTPClient HTTP扩展点客户端
type HTTPClient struct {
    httpClient *http.Client
}

// NewHTTPClient 创建HTTP客户端
func NewHTTPClient() Client {
    return &HTTPClient{
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

// ResolveApprovers 实现
func (c *HTTPClient) ResolveApprovers(ctx context.Context, url string, timeoutSecs int, req *ResolveRequest) (*ResolveResult, error) {
    if url == "" {
        return nil, fmt.Errorf("approver resolver URL not configured")
    }
    
    if req.RequestID == "" {
        req.RequestID = uuid.New().String()
    }
    
    jsonBody, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("X-Workflow-Version", "1")
    httpReq.Header.Set("X-Request-ID", req.RequestID)
    
    // 设置超时
    if timeoutSecs > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
        defer cancel()
        httpReq = httpReq.WithContext(ctx)
    }
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to call approver resolver: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("approver resolver returned status %d", resp.StatusCode)
    }
    
    var result ResolveResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    if result.Code != 0 && result.Code != 200 {
        return nil, fmt.Errorf("approver resolver error: %s", result.Message)
    }
    
    return result.Data, nil
}

// ValidatePermissions 实现
func (c *HTTPClient) ValidatePermissions(ctx context.Context, url string, timeoutSecs int, req *ValidateRequest) (*ValidateResponse, error) {
    if url == "" {
        // 未配置验证器，默认通过
        return &ValidateResponse{Passed: true}, nil
    }
    
    if req.RequestID == "" {
        req.RequestID = uuid.New().String()
    }
    
    jsonBody, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("X-Workflow-Version", "1")
    httpReq.Header.Set("X-Request-ID", req.RequestID)
    
    if timeoutSecs > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSecs)*time.Second)
        defer cancel()
        httpReq = httpReq.WithContext(ctx)
    }
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("failed to call permission validator: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("permission validator returned status %d", resp.StatusCode)
    }
    
    var result ValidateResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &result, nil
}
```

### 7.2 SPI 插件接口（嵌入式）

```go
// internal/extension/spi.go

package extension

import (
    "context"
    "workflow-engine/internal/model"
)

// ApproverResolver SPI 审批人解析器接口
type ApproverResolver interface {
    Resolve(ctx context.Context, req *ResolveRequest) (*ResolveResult, error)
}

// PermissionValidator SPI 权限验证器接口
type PermissionValidator interface {
    Validate(ctx context.Context, req *ValidateRequest) (*ValidateResponse, error)
}

// Registry SPI 注册表
type Registry struct {
    resolvers  map[string]ApproverResolver   // key: processType
    validators map[string]PermissionValidator // key: processType
}

// RegisterResolver 注册解析器
func (r *Registry) RegisterResolver(processType string, resolver ApproverResolver) {
    r.resolvers[processType] = resolver
}

// RegisterValidator 注册验证器
func (r *Registry) RegisterValidator(processType string, validator PermissionValidator) {
    r.validators[processType] = validator
}
```

---

## 八、核心业务流程实现

### 8.1 提交流程实现

```go
// internal/service/process_submit.go

package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "workflow-engine/internal/model"
    "workflow-engine/internal/extension"
    "workflow-engine/internal/pkg/errors"
    "workflow-engine/internal/pkg/locker"
)

// SubmitProcess 提交流程核心实现
func (s *ProcessServiceImpl) SubmitProcess(ctx context.Context, req *SubmitProcessRequest) (*model.ProcessInstance, error) {
    // 1. 获取流程定义
    defID, _ := uuid.Parse(req.DefinitionID)
    definition, err := s.defRepo.GetByID(ctx, defID)
    if err != nil {
        return nil, err
    }
    if definition.Status != model.DefStatusActive {
        return nil, errors.New(errors.ErrDefinitionNotActive, "流程定义未激活")
    }
    
    submittedBy := req.SubmittedBy
    onBehalfOf := req.OnBehalfOf
    
    // 2. 代提交校验
    if onBehalfOf != nil && *onBehalfOf != submittedBy {
        valid, err := s.proxyRepo.ValidateProxy(ctx, submittedBy, *onBehalfOf, definition.Name)
        if err != nil {
            return nil, err
        }
        if !valid {
            return nil, errors.New(errors.ErrNoProxyPermission, "无代提交权限")
        }
    }
    
    // 3. 调用扩展点解析审批人（如果finalSteps为空或未提供）
    var originalSteps []model.StepConfig
    if len(req.FinalSteps) == 0 {
        resolveReq := &extension.ResolveRequest{
            ProcessType: definition.Name,
            FormData:    mustMarshal(req.FormData),
            SubmittedBy: submittedBy,
            OnBehalfOf:  onBehalfOf,
            BusinessKey: req.BusinessKey,
        }
        
        extConfig := &model.ExtensionPointsConfig{}
        _ = json.Unmarshal(definition.ExtensionPoints, extConfig)
        
        result, err := s.extensionClient.ResolveApprovers(ctx, extConfig.ApproverResolverURL, extConfig.TimeoutSeconds, resolveReq)
        if err != nil {
            return nil, fmt.Errorf("解析审批人失败: %w", err)
        }
        originalSteps = result.Steps
    } else {
        originalSteps = req.FinalSteps
    }
    
    // 4. 对比原始步骤和用户确认步骤，生成 diff
    isModified := !stepsEqual(originalSteps, req.FinalSteps)
    finalSteps := req.FinalSteps
    if finalSteps == nil {
        finalSteps = originalSteps
    }
    
    diffSummary := calculateDiff(originalSteps, finalSteps)
    
    // 5. 调用权限验证扩展点
    extConfig := &model.ExtensionPointsConfig{}
    _ = json.Unmarshal(definition.ExtensionPoints, extConfig)
    
    validateReq := &extension.ValidateRequest{
        ProcessType:   definition.Name,
        FormData:      mustMarshal(req.FormData),
        SubmittedBy:   submittedBy,
        OriginalSteps: originalSteps,
        FinalSteps:    finalSteps,
        IsModified:    isModified,
    }
    
    validateResp, err := s.extensionClient.ValidatePermissions(ctx, extConfig.PermissionValidatorURL, extConfig.TimeoutSeconds, validateReq)
    if err != nil {
        return nil, fmt.Errorf("权限验证失败: %w", err)
    }
    if !validateResp.Passed {
        return nil, errors.NewWithData(errors.ErrPermissionDenied, validateResp.Message, validateResp.FailedItems)
    }
    
    // 6. 创建流程实例（使用分布式锁防止重复提交）
    lockKey := fmt.Sprintf("submit:%s:%s", definition.Name, req.BusinessKey)
    lock := s.locker.Acquire(ctx, lockKey, 10*time.Second)
    if lock == nil {
        return nil, errors.New(errors.ErrDuplicateSubmit, "请勿重复提交")
    }
    defer lock.Release()
    
    // 7. 委托透明替换
    stepsWithDelegation := s.applyDelegation(ctx, finalSteps, definition.Name)
    
    // 8. 数据库事务创建实例
    instance, err := s.createInstanceWithSteps(ctx, &CreateInstanceParams{
        Definition:       definition,
        BusinessKey:      req.BusinessKey,
        FormData:         req.FormData,
        SubmittedBy:      submittedBy,
        OnBehalfOf:       onBehalfOf,
        IsUrgent:         req.IsUrgent,
        Steps:            stepsWithDelegation,
        OriginalSteps:    originalSteps,
        FinalSteps:       finalSteps,
        DiffSummary:      diffSummary,
    })
    if err != nil {
        return nil, err
    }
    
    // 9. 激活第一个步骤，创建任务
    if err := s.activateFirstStep(ctx, instance); err != nil {
        return nil, err
    }
    
    // 10. 发送通知（异步）
    s.sendSubmitNotifications(ctx, instance)
    
    return instance, nil
}

// applyDelegation 应用委托配置，透明替换审批人
func (s *ProcessServiceImpl) applyDelegation(ctx context.Context, steps []model.StepConfig, processType string) []model.StepConfig {
    result := make([]model.StepConfig, len(steps))
    for i, step := range steps {
        result[i] = step
        result[i].Assignees = make([]model.ApproverRef, len(step.Assignees))
        
        for j, assignee := range step.Assignees {
            result[i].Assignees[j] = assignee
            if assignee.Type == "user" {
                // 查询有效委托
                delegation, _ := s.delegationRepo.GetEffective(ctx, assignee.Value, processType)
                if delegation != nil {
                    // 替换为受托人，但保留原始信息在任务创建时使用
                    result[i].Assignees[j].OriginalValue = assignee.Value
                    result[i].Assignees[j].Value = delegation.DelegateeID
                    result[i].Assignees[j].IsDelegated = true
                }
            }
        }
    }
    return result
}

// createInstanceWithSteps 创建实例和步骤
func (s *ProcessServiceImpl) createInstanceWithSteps(ctx context.Context, params *CreateInstanceParams) (*model.ProcessInstance, error) {
    return s.instanceRepo.CreateWithSteps(ctx, params)
}

// activateFirstStep 激活第一个步骤
func (s *ProcessServiceImpl) activateFirstStep(ctx context.Context, instance *model.ProcessInstance) error {
    steps, err := s.stepRepo.ListByInstance(ctx, instance.ID)
    if err != nil {
        return err
    }
    
    if len(steps) == 0 {
        return nil
    }
    
    // 找到第一个 pending 步骤
    var firstStep *model.ApprovalStep
    for i := range steps {
        if steps[i].Status == model.StepStatusPending {
            firstStep = &steps[i]
            break
        }
    }
    
    if firstStep == nil {
        return nil
    }
    
    // 更新步骤状态为 active
    if err := s.stepRepo.UpdateStatus(ctx, firstStep.ID, model.StepStatusActive); err != nil {
        return err
    }
    
    // 为每个 assignee 创建任务
    var assignees []model.ApproverRef
    _ = json.Unmarshal(firstStep.Assignees, &assignees)
    
    for _, assignee := range assignees {
        task := &model.Task{
            InstanceID:         instance.ID,
            StepID:             firstStep.ID,
            AssigneeID:         assignee.Value,
            OriginalAssigneeID: &assignee.OriginalValue,
            IsDelegated:        assignee.IsDelegated,
            IsUrgent:           instance.IsUrgent,
        }
        if !assignee.IsDelegated {
            task.OriginalAssigneeID = nil
        }
        if _, err := s.taskRepo.Create(ctx, task); err != nil {
            return err
        }
    }
    
    // 更新实例当前步骤索引
    return s.instanceRepo.UpdateCurrentStep(ctx, instance.ID, firstStep.StepIndex)
}
```

### 8.2 任务审批实现

```go
// internal/service/task_action.go

package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "workflow-engine/internal/model"
    "workflow-engine/internal/pkg/errors"
)

// Approve 审批通过
func (s *TaskServiceImpl) Approve(ctx context.Context, taskID uuid.UUID, userID string, req *ActionRequest) (*ActionResult, error) {
    // 1. 获取任务并校验权限
    task, err := s.taskRepo.GetByID(ctx, taskID)
    if err != nil {
        return nil, err
    }
    if task.AssigneeID != userID {
        return nil, errors.New(errors.ErrForbidden, "无权处理此任务")
    }
    if task.Status != model.TaskStatusPending {
        return nil, errors.New(errors.ErrTaskNotPending, "任务状态非待处理")
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
        return nil, errors.New(errors.ErrConcurrentOperation, "操作过于频繁，请稍后重试")
    }
    defer lock.Release()
    
    // 4. 再次检查任务状态（双重检查）
    task, err = s.taskRepo.GetByID(ctx, taskID)
    if err != nil || task.Status != model.TaskStatusPending {
        return nil, errors.New(errors.ErrTaskAlreadyProcessed, "任务已被处理")
    }
    
    // 5. 完成任务
    now := time.Now()
    task.Status = model.TaskStatusCompleted
    task.Action = (*model.TaskAction)(strPtr("approve"))
    task.Comment = &req.Comment
    task.CompletedAt = &now
    
    if err := s.taskRepo.Update(ctx, task); err != nil {
        return nil, err
    }
    
    // 6. 检查步骤推进条件
    shouldProceed, err := s.checkStepCompletion(ctx, step)
    if err != nil {
        return nil, err
    }
    
    result := &ActionResult{
        TaskID:     taskID,
        InstanceID: task.InstanceID,
        Action:     "approve",
    }
    
    if shouldProceed {
        // 7. 推进到下一步或完成流程
        nextTasks, isCompleted, err := s.proceedToNextStep(ctx, step)
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

// checkStepCompletion 检查步骤是否满足完成条件
func (s *TaskServiceImpl) checkStepCompletion(ctx context.Context, step *model.ApprovalStep) (bool, error) {
    // 获取步骤下所有任务
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

// proceedToNextStep 推进到下一步
func (s *TaskServiceImpl) proceedToNextStep(ctx context.Context, currentStep *model.ApprovalStep) ([]uuid.UUID, bool, error) {
    // 标记当前步骤完成
    now := time.Now()
    currentStep.Status = model.StepStatusCompleted
    currentStep.CompletedAt = &now
    result := "approve"
    currentStep.CompletionResult = &result
    if err := s.stepRepo.Update(ctx, currentStep); err != nil {
        return nil, false, err
    }
    
    // 查找下一个 pending 步骤
    nextStep, err := s.stepRepo.GetNextPendingStep(ctx, currentStep.InstanceID, currentStep.StepIndex)
    if err != nil {
        return nil, false, err
    }
    
    // 没有下一步，流程完成
    if nextStep == nil {
        if err := s.completeProcess(ctx, currentStep.InstanceID); err != nil {
            return nil, false, err
        }
        return nil, true, nil
    }
    
    // 激活下一步
    if err := s.stepRepo.UpdateStatus(ctx, nextStep.ID, model.StepStatusActive); err != nil {
        return nil, false, err
    }
    
    // 创建新任务
    var assignees []model.ApproverRef
    _ = json.Unmarshal(nextStep.Assignees, &assignees)
    
    var nextTaskIDs []uuid.UUID
    for _, assignee := range assignees {
        task := &model.Task{
            InstanceID:         currentStep.InstanceID,
            StepID:             nextStep.ID,
            AssigneeID:         assignee.Value,
            OriginalAssigneeID: strPtr(assignee.OriginalValue),
            IsDelegated:        assignee.IsDelegated,
        }
        if !assignee.IsDelegated {
            task.OriginalAssigneeID = nil
        }
        
        created, err := s.taskRepo.Create(ctx, task)
        if err != nil {
            return nil, false, err
        }
        nextTaskIDs = append(nextTaskIDs, created.ID)
    }
    
    // 更新实例当前步骤
    if err := s.instanceRepo.UpdateCurrentStep(ctx, currentStep.InstanceID, nextStep.StepIndex); err != nil {
        return nil, false, err
    }
    
    return nextTaskIDs, false, nil
}

// Reject 驳回
func (s *TaskServiceImpl) Reject(ctx context.Context, taskID uuid.UUID, userID string, req *ActionRequest) (*ActionResult, error) {
    // 类似 Approve，但流程终止
    // ...
    // 标记所有 pending/active 步骤为 skipped
    // 实例状态更新为 rejected
    // 发送驳回通知
}

// Return 退回
func (s *TaskServiceImpl) Return(ctx context.Context, taskID uuid.UUID, userID string, req *ReturnRequest) (*ActionResult, error) {
    // 获取任务和步骤
    // 根据 ReturnToStep 决定退回目标
    // 标记当前步骤为 returned
    // 重置目标步骤为 active，重新生成任务
    // 发送退回通知
}

// Countersign 加签
func (s *TaskServiceImpl) Countersign(ctx context.Context, taskID uuid.UUID, userID string, req *CountersignRequest) (*ActionResult, error) {
    // 获取当前任务和步骤
    // 在当前步骤后插入新步骤（stepIndex 取中间值）
    // source = countersign
    // addedByUserId = userID
}

// helper functions
func strPtr(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}
```

---

## 九、并发控制方案

### 9.1 Redis 分布式锁实现

```go
// internal/pkg/locker/redis_locker.go

package locker

import (
    "context"
    "fmt"
    "time"
    
    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
)

// DistributedLocker 分布式锁接口
type DistributedLocker interface {
    Acquire(ctx context.Context, key string, ttl time.Duration) Lock
    AcquireWithRetry(ctx context.Context, key string, ttl time.Duration, retry int, interval time.Duration) Lock
}

// Lock 锁接口
type Lock interface {
    Release()
    Extend(ctx context.Context, ttl time.Duration) bool
}

// RedisLocker Redis分布式锁实现
type RedisLocker struct {
    client *redis.Client
}

// NewRedisLocker 创建Redis锁
func NewRedisLocker(client *redis.Client) DistributedLocker {
    return &RedisLocker{client: client}
}

// redisLock 锁实现
type redisLock struct {
    client    *redis.Client
    key       string
    token     string
    released  bool
}

// Acquire 获取锁
func (l *RedisLocker) Acquire(ctx context.Context, key string, ttl time.Duration) Lock {
    token := uuid.New().String()
    fullKey := fmt.Sprintf("lock:%s", key)
    
    ok, err := l.client.SetNX(ctx, fullKey, token, ttl).Result()
    if err != nil || !ok {
        return nil
    }
    
    return &redisLock{
        client: l.client,
        key:    fullKey,
        token:  token,
    }
}

// AcquireWithRetry 带重试的获取锁
func (l *RedisLocker) AcquireWithRetry(ctx context.Context, key string, ttl time.Duration, retry int, interval time.Duration) Lock {
    for i := 0; i < retry; i++ {
        lock := l.Acquire(ctx, key, ttl)
        if lock != nil {
            return lock
        }
        time.Sleep(interval)
    }
    return nil
}

// Release 释放锁
func (l *redisLock) Release() {
    if l.released {
        return
    }
    
    // Lua脚本确保原子性释放
    script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
    l.client.Eval(context.Background(), script, []string{l.key}, l.token)
    l.released = true
}

// Extend 延长锁
func (l *redisLock) Extend(ctx context.Context, ttl time.Duration) bool {
    script := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("expire", KEYS[1], ARGV[2])
        else
            return 0
        end
    `
    result, err := l.client.Eval(ctx, script, []string{l.key}, l.token, int(ttl.Seconds())).Result()
    return err == nil && result.(int64) == 1
}
```

### 9.2 数据库乐观锁（备用方案）

```go
// internal/entity/process.go 添加版本字段

type ProcessInstance struct {
    // ... 其他字段
    Version int `json:"version" gorm:"default:0"` // 乐观锁版本号
}

// UpdateWithVersion 带版本检查的更新
func (r *processInstanceRepo) UpdateWithVersion(ctx context.Context, instance *entity.ProcessInstance) error {
    result := r.db.WithContext(ctx).Model(instance).
        Where("id = ? AND version = ?", instance.ID, instance.Version).
        Updates(map[string]interface{}{
            "status":           instance.Status,
            "current_step_index": instance.CurrentStepIndex,
            "version":          instance.Version + 1,
        })
    
    if result.Error != nil {
        return result.Error
    }
    if result.RowsAffected == 0 {
        return errors.New(errors.ErrConcurrentUpdate, "并发更新冲突，请重试")
    }
    return nil
}
```

---

## 十、缓存策略

### 10.1 Redis 缓存配置

```go
// internal/pkg/redis/client.go

package redis

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/redis/go-redis/v9"
)

// Client Redis客户端封装
type Client struct {
    client *redis.Client
}

// NewClient 创建客户端
func NewClient(addr, password string, db int) *Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    return &Client{client: rdb}
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}

// Set 设置值
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    return c.client.Set(ctx, key, value, ttl).Err()
}

// GetJSON 获取JSON对象
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
    data, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(data), dest)
}

// SetJSON 设置JSON对象
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete 删除键
func (c *Client) Delete(ctx context.Context, keys ...string) error {
    return c.client.Del(ctx, keys...).Err()
}
```

### 10.2 缓存策略设计

```go
// internal/service/cache_strategy.go

package service

import (
    "context"
    "fmt"
    "time"
    
    "workflow-engine/internal/pkg/redis"
)

// CacheKey 缓存键常量
const (
    // 流程定义缓存
    CacheKeyProcessDef    = "wf:def:%s"         // wf:def:{definitionID}
    CacheKeyProcessDefByName = "wf:def:name:%s:v%d" // wf:def:name:{name}:v{version}
    
    // 待办任务列表缓存（分页）
    CacheKeyPendingTasks  = "wf:tasks:pending:%s:p%d:s%d" // wf:tasks:pending:{userID}:p{page}:s{size}
    
    // 委托配置缓存
    CacheKeyDelegation    = "wf:delegation:%s:%s" // wf:delegation:{delegatorID}:{processType}
    
    // 代理配置缓存
    CacheKeyProxy         = "wf:proxy:%s:%s"      // wf:proxy:{agentID}:{principalID}
    
    // 规则缓存
    CacheKeyRules         = "wf:rules:%s"         // wf:rules:{processDefinitionID}
)

// CacheService 缓存服务
type CacheService struct {
    redis *redis.Client
}

// InvalidateProcessDef 清除流程定义缓存
func (s *CacheService) InvalidateProcessDef(ctx context.Context, defID string) {
    s.redis.Delete(ctx, fmt.Sprintf(CacheKeyProcessDef, defID))
}

// InvalidatePendingTasks 清除用户待办缓存
func (s *CacheService) InvalidatePendingTasks(ctx context.Context, userID string) {
    // 使用模式删除所有分页缓存
    pattern := fmt.Sprintf("wf:tasks:pending:%s:*", userID)
    // redis 使用 scan + delete
}

// 缓存时间配置
const (
    // 流程定义缓存 30 分钟
    TTLProcessDef = 30 * time.Minute
    
    // 任务列表缓存 5 分钟
    TTLTaskList   = 5 * time.Minute
    
    // 委托配置缓存 10 分钟
    TTLDelegation = 10 * time.Minute
    
    // 规则缓存 1 小时（规则变更不频繁）
    TTLRules      = 1 * time.Hour
)
```

---

## 十一、错误码与响应规范

### 11.1 错误码定义

```go
// internal/pkg/errors/codes.go

package errors

// ErrorCode 错误码
type ErrorCode string

const (
    // 系统错误 1xxxx
    ErrInternal           ErrorCode = "10001"
    ErrDatabase           ErrorCode = "10002"
    ErrCache              ErrorCode = "10003"
    ErrMQ                 ErrorCode = "10004"
    ErrConcurrentUpdate   ErrorCode = "10005"
    
    // 参数错误 2xxxx
    ErrInvalidParam       ErrorCode = "20001"
    ErrMissingParam       ErrorCode = "20002"
    ErrInvalidFormat      ErrorCode = "20003"
    
    // 流程定义错误 3xxxx
    ErrDefinitionNotFound ErrorCode = "30001"
    ErrDefinitionNotActive ErrorCode = "30002"
    ErrDefinitionExists   ErrorCode = "30003"
    
    // 流程实例错误 4xxxx
    ErrInstanceNotFound   ErrorCode = "40001"
    ErrInstanceNotRunning ErrorCode = "40002"
    ErrDuplicateSubmit    ErrorCode = "40003"
    ErrWithdrawFailed     ErrorCode = "40004"
    
    // 任务错误 5xxxx
    ErrTaskNotFound       ErrorCode = "50001"
    ErrTaskNotPending     ErrorCode = "50002"
    ErrTaskAlreadyProcessed ErrorCode = "50003"
    
    // 权限错误 6xxxx
    ErrForbidden          ErrorCode = "60001"
    ErrNoProxyPermission  ErrorCode = "60002"
    ErrPermissionDenied   ErrorCode = "60003"
    
    // 扩展点错误 7xxxx
    ErrExtensionFailed    ErrorCode = "70001"
    ErrExtensionTimeout   ErrorCode = "70002"
    ErrValidationFailed   ErrorCode = "70003"
    
    // 并发错误 8xxxx
    ErrConcurrentOperation ErrorCode = "80001"
)

// ErrorMessages 错误码映射
var ErrorMessages = map[ErrorCode]string{
    ErrInternal:           "系统内部错误",
    ErrDatabase:           "数据库操作失败",
    ErrCache:              "缓存操作失败",
    ErrMQ:                 "消息队列操作失败",
    ErrConcurrentUpdate:   "数据已被修改，请刷新后重试",
    ErrInvalidParam:       "参数错误",
    ErrMissingParam:       "缺少必要参数",
    ErrInvalidFormat:      "参数格式错误",
    ErrDefinitionNotFound: "流程定义不存在",
    ErrDefinitionNotActive: "流程定义未激活",
    ErrDefinitionExists:   "流程定义已存在",
    ErrInstanceNotFound:   "流程实例不存在",
    ErrInstanceNotRunning: "流程不在运行中",
    ErrDuplicateSubmit:    "重复提交",
    ErrWithdrawFailed:     "撤回失败，流程已开始处理",
    ErrTaskNotFound:       "任务不存在",
    ErrTaskNotPending:     "任务状态非待处理",
    ErrTaskAlreadyProcessed: "任务已被处理",
    ErrForbidden:          "无权访问",
    ErrNoProxyPermission:  "无代提交权限",
    ErrPermissionDenied:   "权限验证失败",
    ErrExtensionFailed:    "扩展点调用失败",
    ErrExtensionTimeout:   "扩展点调用超时",
    ErrValidationFailed:   "数据验证失败",
    ErrConcurrentOperation: "操作过于频繁，请稍后重试",
}
```

### 11.2 统一响应结构

```go
// api/response/response.go

package response

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "workflow-engine/internal/pkg/errors"
)

// Response 统一响应结构
type Response struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// PageResult 分页结果
type PageResult struct {
    List     interface{} `json:"list"`
    Total    int64       `json:"total"`
    Page     int         `json:"page"`
    Size     int         `json:"size"`
    HasMore  bool        `json:"hasMore"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    "0",
        Message: "success",
        Data:    data,
    })
}

// SuccessWithPage 成功响应（分页）
func SuccessWithPage(c *gin.Context, list interface{}, total int64, page, size int) {
    c.JSON(http.StatusOK, Response{
        Code:    "0",
        Message: "success",
        Data: PageResult{
            List:    list,
            Total:   total,
            Page:    page,
            Size:    size,
            HasMore: total > int64(page*size),
        },
    })
}

// Error 错误响应
func Error(c *gin.Context, code errors.ErrorCode, message string) {
    if message == "" {
        message = errors.ErrorMessages[code]
    }
    c.JSON(http.StatusOK, Response{
        Code:    string(code),
        Message: message,
    })
}

// ErrorWithData 带数据的错误响应（用于验证失败场景）
func ErrorWithData(c *gin.Context, code errors.ErrorCode, message string, data interface{}) {
    if message == "" {
        message = errors.ErrorMessages[code]
    }
    c.JSON(http.StatusOK, Response{
        Code:    string(code),
        Message: message,
        Data:    data,
    })
}

// HTTPError HTTP状态码错误
func HTTPError(c *gin.Context, status int, message string) {
    c.JSON(status, Response{
        Code:    string(errors.ErrInternal),
        Message: message,
    })
}
```

---

## 十二、配置与部署

### 12.1 配置文件模板

```yaml
# configs/config.yaml

server:
  port: 8080
  mode: release  # debug | release
  read_timeout: 30s
  write_timeout: 30s

database:
  host: localhost
  port: 5432
  user: workflow
  password: workflow_pass
  dbname: workflow_engine
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 10
  conn_max_lifetime: 1h

redis:
  addr: localhost:6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5

rabbitmq:
  url: amqp://guest:guest@localhost:5672/
  exchange: workflow.events
  queue: workflow.notifications
  routing_key: workflow.notify

extension:
  default_timeout: 3s
  max_retry: 3
  retry_interval: 1s

lock:
  default_ttl: 30s

log:
  level: info
  format: json
  output: stdout
  file_path: ./logs/workflow.log

notification:
  channels:
    - internal  # 站内信，必须
    - email     # 邮件
    - sms       # 短信
    - im        # 企业IM
```

### 12.2 部署架构建议

```
                              ┌─────────────┐
                              │   Nginx     │
                              │  (LB/SSL)   │
                              └──────┬──────┘
                                     │
                    ┌────────────────┼────────────────┐
                    │                │                │
              ┌─────▼─────┐    ┌─────▼─────┐    ┌─────▼─────┐
              │ Workflow  │    │ Workflow  │    │ Workflow  │
              │ Engine-1  │    │ Engine-2  │    │ Engine-3  │
              │  (Go App) │    │  (Go App) │    │  (Go App) │
              └─────┬─────┘    └─────┬─────┘    └─────┬─────┘
                    │                │                │
                    └────────────────┼────────────────┘
                                     │
            ┌────────────────────────┼────────────────────────┐
            │                        │                        │
      ┌─────▼─────┐           ┌─────▼─────┐           ┌──────▼────┐
      │PostgreSQL │           │  Redis    │           │ RabbitMQ  │
      │ (Master)  │           │ Cluster   │           │  Cluster  │
      └───────────┘           └───────────┘           └───────────┘
```

### 12.3 Docker Compose 示例

```yaml
# docker-compose.yml

version: '3.8'

services:
  workflow-api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CONFIG_FILE=/app/configs/config.yaml
    volumes:
      - ./configs:/app/configs
    depends_on:
      - postgres
      - redis
      - rabbitmq
    deploy:
      replicas: 3

  postgres:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: workflow
      POSTGRES_PASSWORD: workflow_pass
      POSTGRES_DB: workflow_engine
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init_db.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

  rabbitmq:
    image: rabbitmq:3.12-management-alpine
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: admin123
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:
```

---

## 附录：核心接口汇总

### API 路由表

| 方法 | 路径 | 说明 | 鉴权 |
|------|------|------|------|
| POST | /api/v1/processes/definitions | 创建流程定义 | 管理员 |
| GET | /api/v1/processes/definitions | 流程定义列表 | 登录 |
| GET | /api/v1/processes/definitions/:id | 流程定义详情 | 登录 |
| PUT | /api/v1/processes/definitions/:id | 更新流程定义 | 管理员 |
| POST | /api/v1/processes/definitions/:id/activate | 激活流程定义 | 管理员 |
| POST | /api/v1/processes/definitions/:id/archive | 归档流程定义 | 管理员 |
| POST | /api/v1/processes/instances | 提交流程 | 登录 |
| GET | /api/v1/processes/instances | 流程实例列表 | 登录 |
| GET | /api/v1/processes/instances/:id | 流程实例详情 | 登录 |
| POST | /api/v1/processes/instances/:id/withdraw | 撤回流程 | 发起人 |
| GET | /api/v1/processes/instances/:id/history | 审批历史 | 登录 |
| PUT | /api/v1/processes/instances/:id/steps | 修改步骤 | 发起人/管理员 |
| GET | /api/v1/tasks/pending | 我的待办 | 登录 |
| GET | /api/v1/tasks/completed | 我的已办 | 登录 |
| GET | /api/v1/tasks/:id | 任务详情 | 登录 |
| POST | /api/v1/tasks/:id/action | 处理任务 | 登录 |
| GET | /api/v1/tasks/statistics | 任务统计 | 登录 |
| POST | /api/v1/admin/rules | 创建规则 | 管理员 |
| GET | /api/v1/admin/rules | 规则列表 | 管理员 |
| PUT | /api/v1/admin/rules/:id | 更新规则 | 管理员 |
| DELETE | /api/v1/admin/rules/:id | 删除规则 | 管理员 |
| POST | /api/v1/admin/proxies | 创建代理配置 | 登录 |
| GET | /api/v1/admin/proxies | 代理配置列表 | 登录 |
| GET | /api/v1/admin/proxies/my-principals | 我的委托人 | 登录 |
| POST | /api/v1/admin/delegations | 创建委托配置 | 登录 |
| GET | /api/v1/admin/delegations | 委托配置列表 | 登录 |

---

*文档结束*
