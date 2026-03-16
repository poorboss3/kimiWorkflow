package repository

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"workflow-engine/internal/model"
)

// ApprovalRuleRepository 审批规则仓库接口
type ApprovalRuleRepository interface {
	Create(ctx context.Context, rule *model.ApprovalRule) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ApprovalRule, error)
	List(ctx context.Context, processDefID *uuid.UUID, isActive *bool) ([]*model.ApprovalRule, error)
	Update(ctx context.Context, rule *model.ApprovalRule) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProxyConfigRepository 代理配置仓库接口
type ProxyConfigRepository interface {
	Create(ctx context.Context, config *model.ProxyConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.ProxyConfig, error)
	List(ctx context.Context, principalID, agentID string, isActive *bool) ([]*model.ProxyConfig, error)
	ValidateProxy(ctx context.Context, agentID, principalID, processType string) (bool, error)
	ListPrincipalsByAgent(ctx context.Context, agentID string) ([]string, error)
	Update(ctx context.Context, config *model.ProxyConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// DelegationConfigRepository 委托配置仓库接口
type DelegationConfigRepository interface {
	Create(ctx context.Context, config *model.DelegationConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.DelegationConfig, error)
	List(ctx context.Context, delegatorID, delegateeID string, isActive *bool) ([]*model.DelegationConfig, error)
	GetEffective(ctx context.Context, delegatorID, processType string) (*model.DelegationConfig, error)
	Update(ctx context.Context, config *model.DelegationConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	Expire(ctx context.Context) error // 清理过期委托
}

// MemoryApprovalRuleRepository 内存审批规则仓库
type MemoryApprovalRuleRepository struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*model.ApprovalRule
}

// NewMemoryApprovalRuleRepository 创建内存审批规则仓库
func NewMemoryApprovalRuleRepository() ApprovalRuleRepository {
	return &MemoryApprovalRuleRepository{
		data: make(map[uuid.UUID]*model.ApprovalRule),
	}
}

func (r *MemoryApprovalRuleRepository) Create(ctx context.Context, rule *model.ApprovalRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	rule.InitBaseModel()
	r.data[rule.ID] = rule
	return nil
}

func (r *MemoryApprovalRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ApprovalRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, ok := r.data[id]
	if !ok {
		return nil, nil
	}
	return rule, nil
}

func (r *MemoryApprovalRuleRepository) List(ctx context.Context, processDefID *uuid.UUID, isActive *bool) ([]*model.ApprovalRule, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*model.ApprovalRule
	for _, rule := range r.data {
		if processDefID != nil {
			if rule.ProcessDefinitionID == nil || *rule.ProcessDefinitionID != *processDefID {
				continue
			}
		}
		if isActive != nil && rule.IsActive != *isActive {
			continue
		}
		list = append(list, rule)
	}

	// 按优先级倒序
	sort.Slice(list, func(i, j int) bool {
		return list[i].Priority > list[j].Priority
	})

	return list, nil
}

func (r *MemoryApprovalRuleRepository) Update(ctx context.Context, rule *model.ApprovalRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[rule.ID]; !ok {
		return nil
	}

	rule.UpdatedAt = time.Now()
	r.data[rule.ID] = rule
	return nil
}

func (r *MemoryApprovalRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.data, id)
	return nil
}

// MemoryProxyConfigRepository 内存代理配置仓库
type MemoryProxyConfigRepository struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*model.ProxyConfig
}

// NewMemoryProxyConfigRepository 创建内存代理配置仓库
func NewMemoryProxyConfigRepository() ProxyConfigRepository {
	return &MemoryProxyConfigRepository{
		data: make(map[uuid.UUID]*model.ProxyConfig),
	}
}

func (r *MemoryProxyConfigRepository) Create(ctx context.Context, config *model.ProxyConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	config.InitBaseModel()
	r.data[config.ID] = config
	return nil
}

func (r *MemoryProxyConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.ProxyConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.data[id]
	if !ok {
		return nil, nil
	}
	return config, nil
}

func (r *MemoryProxyConfigRepository) List(ctx context.Context, principalID, agentID string, isActive *bool) ([]*model.ProxyConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*model.ProxyConfig
	for _, config := range r.data {
		if principalID != "" && config.PrincipalID != principalID {
			continue
		}
		if agentID != "" && config.AgentID != agentID {
			continue
		}
		if isActive != nil && config.IsActive != *isActive {
			continue
		}
		list = append(list, config)
	}

	return list, nil
}

func (r *MemoryProxyConfigRepository) ValidateProxy(ctx context.Context, agentID, principalID, processType string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	for _, config := range r.data {
		if !config.IsActive {
			continue
		}
		if config.AgentID != agentID || config.PrincipalID != principalID {
			continue
		}
		if now.Before(config.ValidFrom) {
			continue
		}
		if config.ValidTo != nil && now.After(*config.ValidTo) {
			continue
		}

		// 检查流程类型
		if config.AllowedProcessTypes != nil {
			var types []string
			json.Unmarshal([]byte(*config.AllowedProcessTypes), &types)
			if len(types) > 0 {
				found := false
				for _, t := range types {
					if t == processType {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}

		return true, nil
	}

	return false, nil
}

func (r *MemoryProxyConfigRepository) ListPrincipalsByAgent(ctx context.Context, agentID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	principalMap := make(map[string]bool)
	for _, config := range r.data {
		if !config.IsActive || config.AgentID != agentID {
			continue
		}
		if now.Before(config.ValidFrom) {
			continue
		}
		if config.ValidTo != nil && now.After(*config.ValidTo) {
			continue
		}
		principalMap[config.PrincipalID] = true
	}

	var principals []string
	for p := range principalMap {
		principals = append(principals, p)
	}
	return principals, nil
}

func (r *MemoryProxyConfigRepository) Update(ctx context.Context, config *model.ProxyConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[config.ID]; !ok {
		return nil
	}

	config.UpdatedAt = time.Now()
	r.data[config.ID] = config
	return nil
}

func (r *MemoryProxyConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.data, id)
	return nil
}

// MemoryDelegationConfigRepository 内存委托配置仓库
type MemoryDelegationConfigRepository struct {
	mu   sync.RWMutex
	data map[uuid.UUID]*model.DelegationConfig
}

// NewMemoryDelegationConfigRepository 创建内存委托配置仓库
func NewMemoryDelegationConfigRepository() DelegationConfigRepository {
	return &MemoryDelegationConfigRepository{
		data: make(map[uuid.UUID]*model.DelegationConfig),
	}
}

func (r *MemoryDelegationConfigRepository) Create(ctx context.Context, config *model.DelegationConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	config.InitBaseModel()
	r.data[config.ID] = config
	return nil
}

func (r *MemoryDelegationConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.DelegationConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, ok := r.data[id]
	if !ok {
		return nil, nil
	}
	return config, nil
}

func (r *MemoryDelegationConfigRepository) List(ctx context.Context, delegatorID, delegateeID string, isActive *bool) ([]*model.DelegationConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*model.DelegationConfig
	for _, config := range r.data {
		if delegatorID != "" && config.DelegatorID != delegatorID {
			continue
		}
		if delegateeID != "" && config.DelegateeID != delegateeID {
			continue
		}
		if isActive != nil && config.IsActive != *isActive {
			continue
		}
		list = append(list, config)
	}

	return list, nil
}

func (r *MemoryDelegationConfigRepository) GetEffective(ctx context.Context, delegatorID, processType string) (*model.DelegationConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	for _, config := range r.data {
		if !config.IsActive || config.DelegatorID != delegatorID {
			continue
		}
		if now.Before(config.ValidFrom) {
			continue
		}
		if config.ValidTo != nil && now.After(*config.ValidTo) {
			continue
		}

		// 检查流程类型
		if config.AllowedProcessTypes != nil {
			var types []string
			json.Unmarshal([]byte(*config.AllowedProcessTypes), &types)
			if len(types) > 0 {
				found := false
				for _, t := range types {
					if t == processType {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}

		return config, nil
	}

	return nil, nil
}

func (r *MemoryDelegationConfigRepository) Update(ctx context.Context, config *model.DelegationConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.data[config.ID]; !ok {
		return nil
	}

	config.UpdatedAt = time.Now()
	r.data[config.ID] = config
	return nil
}

func (r *MemoryDelegationConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.data, id)
	return nil
}

func (r *MemoryDelegationConfigRepository) Expire(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, config := range r.data {
		if config.ValidTo != nil && now.After(*config.ValidTo) {
			config.IsActive = false
			config.UpdatedAt = now
			r.data[id] = config
		}
	}
	return nil
}
