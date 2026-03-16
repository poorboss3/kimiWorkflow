// memory 内存队列实现
package memory

import (
	"context"
	"encoding/json"
	"math/rand"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Message 消息
type Message struct {
	ID      string          `json:"id"`
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

// Queue 内存队列
type Queue struct {
	mu       sync.RWMutex
	channels map[string]chan *Message
	handlers map[string][]func(context.Context, *Message) error
	closed   bool
}

// NewQueue 创建内存队列
func NewQueue() *Queue {
	q := &Queue{
		channels: make(map[string]chan *Message),
		handlers: make(map[string][]func(context.Context, *Message) error),
	}
	return q
}

// Publish 发布消息
func (q *Queue) Publish(ctx context.Context, topic string, payload interface{}) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if q.closed {
		return nil
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	msg := &Message{
		ID:      generateMsgID(),
		Topic:   topic,
		Payload: data,
	}
	
	// 获取或创建channel
	ch, ok := q.channels[topic]
	if !ok {
		ch = make(chan *Message, 1000)
		q.channels[topic] = ch
		// 启动消费者
		go q.consume(topic, ch)
	}
	
	select {
	case ch <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Subscribe 订阅消息
func (q *Queue) Subscribe(topic string, handler func(context.Context, *Message) error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	q.handlers[topic] = append(q.handlers[topic], handler)
}

// consume 消费消息
func (q *Queue) consume(topic string, ch chan *Message) {
	for msg := range ch {
		q.mu.RLock()
		handlers := q.handlers[topic]
		q.mu.RUnlock()
		
		for _, handler := range handlers {
			go func(h func(context.Context, *Message) error) {
				ctx := context.Background()
				_ = h(ctx, msg) // 忽略错误，实际应该记录日志
			}(handler)
		}
	}
}

// Close 关闭队列
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	q.closed = true
	for _, ch := range q.channels {
		close(ch)
	}
}

func generateMsgID() string {
	return time.Now().Format("20060102150405") + randString(6)
}

func randString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
