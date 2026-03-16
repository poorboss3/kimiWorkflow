package errors

import "fmt"

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
	ErrDefinitionNotFound  ErrorCode = "30001"
	ErrDefinitionNotActive ErrorCode = "30002"
	ErrDefinitionExists    ErrorCode = "30003"

	// 流程实例错误 4xxxx
	ErrInstanceNotFound       ErrorCode = "40001"
	ErrInstanceNotRunning     ErrorCode = "40002"
	ErrDuplicateSubmit        ErrorCode = "40003"
	ErrWithdrawFailed         ErrorCode = "40004"
	ErrNoPermissionToModify   ErrorCode = "40005"

	// 任务错误 5xxxx
	ErrTaskNotFound         ErrorCode = "50001"
	ErrTaskNotPending       ErrorCode = "50002"
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
	ErrInternal:            "系统内部错误",
	ErrDatabase:            "数据库操作失败",
	ErrCache:               "缓存操作失败",
	ErrMQ:                  "消息队列操作失败",
	ErrConcurrentUpdate:    "数据已被修改，请刷新后重试",
	ErrInvalidParam:        "参数错误",
	ErrMissingParam:        "缺少必要参数",
	ErrInvalidFormat:       "参数格式错误",
	ErrDefinitionNotFound:  "流程定义不存在",
	ErrDefinitionNotActive: "流程定义未激活",
	ErrDefinitionExists:    "流程定义已存在",
	ErrInstanceNotFound:    "流程实例不存在",
	ErrInstanceNotRunning:  "流程不在运行中",
	ErrDuplicateSubmit:     "重复提交",
	ErrWithdrawFailed:      "撤回失败，流程已开始处理",
	ErrNoPermissionToModify:"无权修改审批步骤",
	ErrTaskNotFound:        "任务不存在",
	ErrTaskNotPending:      "任务状态非待处理",
	ErrTaskAlreadyProcessed:"任务已被处理",
	ErrForbidden:           "无权访问",
	ErrNoProxyPermission:   "无代提交权限",
	ErrPermissionDenied:    "权限验证失败",
	ErrExtensionFailed:     "扩展点调用失败",
	ErrExtensionTimeout:    "扩展点调用超时",
	ErrValidationFailed:    "数据验证失败",
	ErrConcurrentOperation: "操作过于频繁，请稍后重试",
}

// AppError 应用错误
type AppError struct {
	Code    ErrorCode
	Message string
	Data    interface{}
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New 创建错误
func New(code ErrorCode, message string) *AppError {
	if message == "" {
		message = ErrorMessages[code]
	}
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// NewWithData 创建带数据的错误
func NewWithData(code ErrorCode, message string, data interface{}) *AppError {
	if message == "" {
		message = ErrorMessages[code]
	}
	return &AppError{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// Is 判断错误码
func Is(err error, code ErrorCode) bool {
	if e, ok := err.(*AppError); ok {
		return e.Code == code
	}
	return false
}
