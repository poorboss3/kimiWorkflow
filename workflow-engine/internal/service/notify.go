package service

import (
	"context"
	"encoding/json"
	"fmt"

	"workflow-engine/internal/extension"
	"workflow-engine/internal/pkg/memory"
)

// NotificationService 通知服务接口
type NotificationService interface {
	Start(ctx context.Context)
	HandleNotifyEvent(ctx context.Context, event *extension.NotifyEvent) error
}

// NotificationServiceImpl 通知服务实现
type NotificationServiceImpl struct {
	mq *memory.Queue
}

// NewNotificationService 创建通知服务
func NewNotificationService(mq *memory.Queue) NotificationService {
	return &NotificationServiceImpl{
		mq: mq,
	}
}

// Start 启动通知服务
func (s *NotificationServiceImpl) Start(ctx context.Context) {
	// 订阅通知事件
	s.mq.Subscribe("workflow.notifications", func(ctx context.Context, msg *memory.Message) error {
		var event extension.NotifyEvent
		if err := json.Unmarshal(msg.Payload, &event); err != nil {
			return err
		}
		return s.HandleNotifyEvent(ctx, &event)
	})
}

// HandleNotifyEvent 处理通知事件
func (s *NotificationServiceImpl) HandleNotifyEvent(ctx context.Context, event *extension.NotifyEvent) error {
	switch event.EventType {
	case "submit":
		s.handleSubmit(ctx, event)
	case "approve":
		s.handleApprove(ctx, event)
	case "reject":
		s.handleReject(ctx, event)
	case "return":
		s.handleReturn(ctx, event)
	case "complete":
		s.handleComplete(ctx, event)
	case "urgent":
		s.handleUrgent(ctx, event)
	case "countersign":
		s.handleCountersign(ctx, event)
	default:
		fmt.Printf("[通知] 未知事件类型: %s\n", event.EventType)
	}
	return nil
}

func (s *NotificationServiceImpl) handleSubmit(ctx context.Context, event *extension.NotifyEvent) {
	fmt.Printf("[通知] 新任务待处理 | 流程: %s | 接收人: %s\n", event.ProcessType, event.RecipientID)
}

func (s *NotificationServiceImpl) handleApprove(ctx context.Context, event *extension.NotifyEvent) {
	fmt.Printf("[通知] 任务已通过 | 流程: %s | 处理人: %s\n", event.ProcessType, event.RecipientID)
}

func (s *NotificationServiceImpl) handleReject(ctx context.Context, event *extension.NotifyEvent) {
	fmt.Printf("[通知] 流程被驳回 | 流程: %s | 接收人: %s\n", event.ProcessType, event.RecipientID)
}

func (s *NotificationServiceImpl) handleReturn(ctx context.Context, event *extension.NotifyEvent) {
	fmt.Printf("[通知] 任务被退回 | 流程: %s | 接收人: %s\n", event.ProcessType, event.RecipientID)
}

func (s *NotificationServiceImpl) handleComplete(ctx context.Context, event *extension.NotifyEvent) {
	fmt.Printf("[通知] 流程已完成 | 流程: %s | 接收人: %s\n", event.ProcessType, event.RecipientID)
}

func (s *NotificationServiceImpl) handleUrgent(ctx context.Context, event *extension.NotifyEvent) {
	fmt.Printf("[通知-加急] 任务已加急 | 流程: %s | 接收人: %s\n", event.ProcessType, event.RecipientID)
}

func (s *NotificationServiceImpl) handleCountersign(ctx context.Context, event *extension.NotifyEvent) {
	fmt.Printf("[通知] 加签任务待处理 | 流程: %s | 接收人: %s\n", event.ProcessType, event.RecipientID)
}
