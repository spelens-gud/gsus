// Package logger 提供统一的日志管理功能.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Level type    日志级别.
type Level int

const (
	// LevelDebug 调试级别.
	LevelDebug Level = iota
	// LevelInfo 信息级别.
	LevelInfo
	// LevelWarn 警告级别.
	LevelWarn
	// LevelError 错误级别.
	LevelError
)

// Logger interface    日志接口.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	SetLevel(level Level)
	SetOutput(w io.Writer)
}

// defaultLogger struct    默认日志实现.
type defaultLogger struct {
	level  Level
	logger *log.Logger
}

var std Logger

func init() {
	std = &defaultLogger{
		level:  LevelInfo,
		logger: log.New(os.Stdout, "[gsus] ", 0),
	}
}

// Debug function    输出调试日志.
func Debug(msg string, args ...interface{}) {
	std.Debug(msg, args...)
}

// Info function    输出信息日志.
func Info(msg string, args ...interface{}) {
	std.Info(msg, args...)
}

// Warn function    输出警告日志.
func Warn(msg string, args ...interface{}) {
	std.Warn(msg, args...)
}

// Error function    输出错误日志.
func Error(msg string, args ...interface{}) {
	std.Error(msg, args...)
}

// SetLevel function    设置日志级别.
func SetLevel(level Level) {
	std.SetLevel(level)
}

// SetOutput function    设置日志输出.
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// Default function    获取默认日志器.
func Default() Logger {
	return std
}

// Debug method    输出调试日志.
func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.logger.Printf("[DEBUG] "+msg, args...)
	}
}

// Info method    输出信息日志.
func (l *defaultLogger) Info(msg string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.logger.Printf("[INFO] "+msg, args...)
	}
}

// Warn method    输出警告日志.
func (l *defaultLogger) Warn(msg string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.logger.Printf("[WARN] "+msg, args...)
	}
}

// Error method    输出错误日志.
func (l *defaultLogger) Error(msg string, args ...interface{}) {
	if l.level <= LevelError {
		l.logger.Printf("[ERROR] "+msg, args...)
	}
}

// SetLevel method    设置日志级别.
func (l *defaultLogger) SetLevel(level Level) {
	l.level = level
}

// SetOutput method    设置日志输出.
func (l *defaultLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

func Printf(format string, args ...interface{}) {
	std.Info(fmt.Sprintf(format, args...))
}
