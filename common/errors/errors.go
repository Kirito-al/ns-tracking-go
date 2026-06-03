// Package errors 统一错误定义
package errors

import (
	"fmt"
)

// CodeError 业务错误结构
type CodeError struct {
	Code    int
	Message string
}

// NewCodeError 创建业务错误
func NewCodeError(code int, message string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: message,
	}
}

// Error 实现error接口
func (e *CodeError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// GetCode 获取错误码
func (e *CodeError) GetCode() int {
	return e.Code
}

// GetMessage 获取错误消息
func (e *CodeError) GetMessage() string {
	return e.Message
}

// ===================== 预定义错误 =====================

var (
	// ErrSignatureInvalid 签名验证失败
	ErrSignatureInvalid = NewCodeError(401, "invalid HMAC signature")
	// ErrTrackingNotFound 轨迹数据不存在
	ErrTrackingNotFound = NewCodeError(404, "tracking data not found")
	// ErrDatabaseError 数据库错误
	ErrDatabaseError = NewCodeError(500, "database operation failed")
	// ErrRedisError Redis错误
	ErrRedisError = NewCodeError(500, "redis operation failed")
	// ErrRpcError RPC调用错误
	ErrRpcError = NewCodeError(500, "rpc call failed")
	// ErrInvalidParams 参数错误
	ErrInvalidParams = NewCodeError(400, "invalid request parameters")
)