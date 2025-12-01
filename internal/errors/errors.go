// Package errors 提供统一的错误定义和处理机制.
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode type    错误码.
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

// Error struct    项目错误.
type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error method    实现 error 接口.
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap method    返回原始错误.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Wrap function    包装错误.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		Message: message,
		Cause:   err,
	}
}

// WrapWithCode function    包装错误并指定错误码.
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

// New function    创建新错误.
func New(code ErrorCode, message string) error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf function    创建格式化错误.
func Newf(code ErrorCode, format string, args ...interface{}) error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Is function    判断错误是否为指定类型.
func Is(err error, target error) bool {
	if err == nil || target == nil {
		return errors.Is(err, target)
	}

	var e1 *Error
	ok1 := errors.As(err, &e1)
	var e2 *Error
	ok2 := errors.As(target, &e2)

	if ok1 && ok2 {
		return e1.Code == e2.Code
	}

	return errors.Is(err, target)
}

// HasCode function    判断错误是否包含指定错误码.
func HasCode(err error, code ErrorCode) bool {
	if err == nil {
		return false
	}

	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}

	return false
}

// GetCode function    获取错误码.
func GetCode(err error) ErrorCode {
	if err == nil {
		return 0
	}

	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}

	return 0
}

// IsConfigError function    判断是否为配置错误.
func IsConfigError(err error) bool {
	return HasCode(err, ErrCodeConfig)
}

// IsParseError function    判断是否为解析错误.
func IsParseError(err error) bool {
	return HasCode(err, ErrCodeParse)
}

// IsGenerateError function    判断是否为生成错误.
func IsGenerateError(err error) bool {
	return HasCode(err, ErrCodeGenerate)
}

// IsDatabaseError function    判断是否为数据库错误.
func IsDatabaseError(err error) bool {
	return HasCode(err, ErrCodeDatabase)
}

// IsTemplateError function    判断是否为模板错误.
func IsTemplateError(err error) bool {
	return HasCode(err, ErrCodeTemplate)
}

// IsFileError function    判断是否为文件错误.
func IsFileError(err error) bool {
	return HasCode(err, ErrCodeFile)
}

// WithContext function    为错误添加上下文信息.
func WithContext(err error, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	var e *Error
	ok := errors.As(err, &e)
	if !ok {
		e = &Error{
			Message: err.Error(),
			Cause:   err,
		}
	}

	// 将上下文信息添加到错误消息中
	if len(context) > 0 {
		contextStr := ""
		for k, v := range context {
			contextStr += fmt.Sprintf(" %s=%v", k, v)
		}
		e.Message = e.Message + " |" + contextStr
	}

	return e
}

// Recover function    从 panic 中恢复并转换为错误.
func Recover() error {
	if r := recover(); r != nil {
		switch v := r.(type) {
		case error:
			return Wrap(v, "panic recovered")
		case string:
			return New(0, v)
		default:
			return Newf(0, "panic recovered: %v", r)
		}
	}
	return nil
}
