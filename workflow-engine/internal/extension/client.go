package extension

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"workflow-engine/internal/model"
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

// MockClient 模拟扩展点客户端（用于本地开发测试）
type MockClient struct {
	defaultSteps []model.StepConfig
}

// NewMockClient 创建模拟客户端
func NewMockClient() Client {
	return &MockClient{
		defaultSteps: []model.StepConfig{
			{
				Type: model.StepTypeApproval,
				Assignees: []model.ApproverRef{
					{Type: "user", Value: "manager_001", Name: "部门经理"},
				},
			},
			{
				Type: model.StepTypeApproval,
				Assignees: []model.ApproverRef{
					{Type: "user", Value: "director_001", Name: "总监"},
				},
			},
		},
	}
}

// ResolveApprovers 模拟实现
func (c *MockClient) ResolveApprovers(ctx context.Context, url string, timeoutSecs int, req *ResolveRequest) (*ResolveResult, error) {
	// 模拟：根据表单金额返回不同的审批流程
	var formData map[string]interface{}
	json.Unmarshal(req.FormData, &formData)

	steps := make([]model.StepConfig, len(c.defaultSteps))
	copy(steps, c.defaultSteps)

	// 如果金额大于10000，添加财务审批
	if amount, ok := formData["amount"].(float64); ok && amount > 10000 {
		steps = append(steps, model.StepConfig{
			Type: model.StepTypeApproval,
			Assignees: []model.ApproverRef{
				{Type: "user", Value: "finance_001", Name: "财务经理"},
			},
		})
	}

	return &ResolveResult{Steps: steps}, nil
}

// ValidatePermissions 模拟实现（默认通过）
func (c *MockClient) ValidatePermissions(ctx context.Context, url string, timeoutSecs int, req *ValidateRequest) (*ValidateResponse, error) {
	// 模拟：如果修改了审批人，返回警告但不阻止
	if req.IsModified {
		return &ValidateResponse{
			Passed:  true,
			Message: "审批列表已修改，请注意核对",
		}, nil
	}
	return &ValidateResponse{Passed: true}, nil
}
