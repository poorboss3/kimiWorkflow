// memory 内存存储实现
package memory

import (
	"sync"
	"time"
)

// Store 内存存储
type Store struct {
	mu    sync.RWMutex
	data  map[string]interface{}
	ttl   map[string]time.Time
}

// NewStore 创建内存存储
func NewStore() *Store {
	s := &Store{
		data: make(map[string]interface{}),
		ttl:  make(map[string]time.Time),
	}
	// 启动清理过期key的goroutine
	go s.cleanup()
	return s
}

// Set 设置值
func (s *Store) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// SetWithTTL 设置值带过期时间
func (s *Store) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	s.ttl[key] = time.Now().Add(ttl)
}

// Get 获取值
func (s *Store) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// 检查是否过期
	if expireTime, ok := s.ttl[key]; ok && time.Now().After(expireTime) {
		return nil, false
	}
	
	val, ok := s.data[key]
	return val, ok
}

// Delete 删除值
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	delete(s.ttl, key)
}

// Exists 检查key是否存在
func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if expireTime, ok := s.ttl[key]; ok && time.Now().After(expireTime) {
		return false
	}
	
	_, ok := s.data[key]
	return ok
}

// Keys 获取所有key（匹配前缀）
func (s *Store) Keys(prefix string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	now := time.Now()
	var keys []string
	for k := range s.data {
		// 检查过期
		if expireTime, ok := s.ttl[k]; ok && now.After(expireTime) {
			continue
		}
		if prefix == "" || len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k)
		}
	}
	return keys
}

// cleanup 清理过期key
func (s *Store) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, expireTime := range s.ttl {
			if now.After(expireTime) {
				delete(s.data, k)
				delete(s.ttl, k)
			}
		}
		s.mu.Unlock()
	}
}

// Clear 清空所有数据
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
	s.ttl = make(map[string]time.Time)
}

// Size 获取存储大小
func (s *Store) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}
