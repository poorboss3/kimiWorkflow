package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateID 生成UUID
func GenerateID() uuid.UUID {
	return uuid.New()
}

// GenerateIDString 生成UUID字符串
func GenerateIDString() string {
	return uuid.New().String()
}

// Ptr 返回指针
func Ptr[T any](v T) *T {
	return &v
}

// MustMarshal JSON序列化（忽略错误）
func MustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

// MustUnmarshal JSON反序列化（忽略错误）
func MustUnmarshal(data []byte, v interface{}) {
	_ = json.Unmarshal(data, v)
}

// Copy 深度拷贝
func Copy(src, dst interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// FormatTime 格式化时间
func FormatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// NowPtr 返回当前时间的指针
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

// StrInSlice 判断字符串是否在切片中
func StrInSlice(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// ClampFloat 限制浮点数范围
func ClampFloat(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// GenerateStepIndex 生成步骤索引（支持中间插入）
// 在 prev 和 next 之间生成一个中间值
func GenerateStepIndex(prev, next float64) float64 {
	if prev >= next {
		next = prev + 1
	}
	return prev + (next-prev)/2
}

// LockKey 生成锁键
func LockKey(parts ...string) string {
	key := "lock"
	for _, p := range parts {
		key += ":" + p
	}
	return key
}

// PageResult 分页结果
type PageResult[T any] struct {
	List    []T  `json:"list"`
	Total   int  `json:"total"`
	Page    int  `json:"page"`
	Size    int  `json:"size"`
	HasMore bool `json:"hasMore"`
}

// NewPageResult 创建分页结果
func NewPageResult[T any](list []T, total, page, size int) *PageResult[T] {
	return &PageResult[T]{
		List:    list,
		Total:   total,
		Page:    page,
		Size:    size,
		HasMore: total > page*size,
	}
}

// Pagination 分页计算
func Pagination(total, page, size int) (start, end int) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	start = (page - 1) * size
	end = start + size
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	return
}

// FormatDuration 格式化持续时间
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d小时", int(d.Hours()))
	}
	return fmt.Sprintf("%d天", int(d.Hours()/24))
}
