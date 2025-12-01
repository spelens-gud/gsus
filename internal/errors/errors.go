// Package errors 提供统一的错误定义和处理机制.
package errors

import (
	"fmt"
)

// ErrorCode type    错误码.
type ErrorCode int

const (
	// ErrCodeConfig 配置错误.
	ErrCodeConfig ErrorCode = 1000
	// ErrCodeParse 解析错误.
	ErrCodeParse ErrorCode = 2000
	// ErrCodeGenerate 生成错误.
	ErrCodeGenerate ErrorCode = 3000
	// ErrCodeDatabase 数据库错误.
	ErrCodeDatabase ErrorCode = 4000
	// ErrCodeTemplate 模板错误.
	ErrCodeTemplate ErrorCode = 5000
	// ErrCodeFile 文件操作错误.
	ErrCodeFile ErrorCode = 6000
)

// Error struct    项目错误.
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error method    实现 error 接口.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap method    返回原始错误.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Wrap function    包装错误.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		Message: message,
		Cause:   err,
	}
}

// WrapWithCode function    包装错误并指定错误码.
func WrapWithCode(err error, code ErrorCode, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// New function    创建新错误.
func New(code ErrorCode, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf function    创建格式化错误.
func Newf(code ErrorCode, format string, args ...interface{}) error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}
