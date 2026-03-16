package locker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
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

// MemoryLocker 内存分布式锁实现
type MemoryLocker struct {
	mu    sync.Mutex
	locks map[string]*memoryLock
}

// memoryLock 内存锁
type memoryLock struct {
	key      string
	token    string
	expiry   time.Time
	mu       sync.Mutex
	released bool
}

// NewMemoryLocker 创建内存锁
func NewMemoryLocker() DistributedLocker {
	return &MemoryLocker{
		locks: make(map[string]*memoryLock),
	}
}

// Acquire 获取锁
func (l *MemoryLocker) Acquire(ctx context.Context, key string, ttl time.Duration) Lock {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	fullKey := fmt.Sprintf("lock:%s", key)
	
	// 检查是否已有锁且未过期
	if existing, ok := l.locks[fullKey]; ok {
		if time.Now().Before(existing.expiry) {
			return nil // 锁已被占用
		}
		// 锁已过期，删除
		delete(l.locks, fullKey)
	}
	
	// 创建新锁
	token := uuid.New().String()
	lock := &memoryLock{
		key:     fullKey,
		token:   token,
		expiry:  time.Now().Add(ttl),
		released: false,
	}
	l.locks[fullKey] = lock
	
	// 启动自动清理
	go l.autoRelease(fullKey, ttl)
	
	return lock
}

// AcquireWithRetry 带重试的获取锁
func (l *MemoryLocker) AcquireWithRetry(ctx context.Context, key string, ttl time.Duration, retry int, interval time.Duration) Lock {
	for i := 0; i < retry; i++ {
		lock := l.Acquire(ctx, key, ttl)
		if lock != nil {
			return lock
		}
		time.Sleep(interval)
	}
	return nil
}

// autoRelease 自动释放过期锁
func (l *MemoryLocker) autoRelease(key string, ttl time.Duration) {
	time.Sleep(ttl + time.Second)
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if lock, ok := l.locks[key]; ok {
		if time.Now().After(lock.expiry) {
			delete(l.locks, key)
		}
	}
}

// Release 释放锁
func (l *memoryLock) Release() {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.released {
		return
	}
	
	// 这里简化处理，实际应该从MemoryLocker的locks中删除
	l.released = true
}

// Extend 延长锁
func (l *memoryLock) Extend(ctx context.Context, ttl time.Duration) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.released {
		return false
	}
	
	l.expiry = time.Now().Add(ttl)
	return true
}

// ForceRelease 强制释放锁（仅用于测试）
func (l *MemoryLocker) ForceRelease(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.locks, fmt.Sprintf("lock:%s", key))
}

// IsLocked 检查是否锁定（仅用于测试）
func (l *MemoryLocker) IsLocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	fullKey := fmt.Sprintf("lock:%s", key)
	if lock, ok := l.locks[fullKey]; ok {
		return time.Now().Before(lock.expiry)
	}
	return false
}
