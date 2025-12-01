// Package logger 提供统一的日志管理功能.
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// Level type    日志级别.
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
	// LevelFatal 致命错误级别.
	LevelFatal
)

// String method    返回日志级别字符串.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel function    解析日志级别字符串.
func ParseLevel(s string) Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Logger interface    日志接口.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Fatal(msg string, args ...interface{})
	SetLevel(level Level)
	SetOutput(w io.Writer)
	WithPrefix(prefix string) Logger
}

// defaultLogger struct    默认日志实现.
type defaultLogger struct {
	level      Level
	logger     *log.Logger
	prefix     string
	timeFormat string
	colorize   bool
}

var std Logger

func init() {
	std = &defaultLogger{
		level:      LevelInfo,
		logger:     log.New(os.Stdout, "", 0),
		prefix:     "[gsus]",
		timeFormat: "2006-01-02 15:04:05",
		colorize:   true,
	}
}

// New function    创建新的日志器.
func New(prefix string, level Level) Logger {
	return &defaultLogger{
		level:      level,
		logger:     log.New(os.Stdout, "", 0),
		prefix:     prefix,
		timeFormat: "2006-01-02 15:04:05",
		colorize:   true,
	}
}

// Debug function    输出调试日志.
func Debug(msg string, args ...interface{}) {
	std.Debug(msg, args...)
}

// Info function    输出信息日志.
func Info(msg string, args ...interface{}) {
	std.Info(msg, args...)
}

// Warn function    输出警告日志.
func Warn(msg string, args ...interface{}) {
	std.Warn(msg, args...)
}

// Error function    输出错误日志.
func Error(msg string, args ...interface{}) {
	std.Error(msg, args...)
}

// Fatal function    输出致命错误日志并退出.
func Fatal(msg string, args ...interface{}) {
	std.Fatal(msg, args...)
}

// SetLevel function    设置日志级别.
func SetLevel(level Level) {
	std.SetLevel(level)
}

// SetOutput function    设置日志输出.
func SetOutput(w io.Writer) {
	std.SetOutput(w)
}

// Default function    获取默认日志器.
func Default() Logger {
	return std
}

// WithPrefix function    创建带前缀的日志器.
func WithPrefix(prefix string) Logger {
	return std.WithPrefix(prefix)
}

// formatMessage method    格式化日志消息.
func (l *defaultLogger) formatMessage(level Level, msg string, args ...interface{}) string {
	timestamp := time.Now().Format(l.timeFormat)
	message := msg
	if len(args) > 0 {
		message = fmt.Sprintf(msg, args...)
	}

	levelStr := level.String()
	if l.colorize {
		levelStr = l.colorizeLevel(level)
	}

	return fmt.Sprintf("%s %s %s %s", timestamp, l.prefix, levelStr, message)
}

// colorizeLevel method    为日志级别添加颜色.
func (l *defaultLogger) colorizeLevel(level Level) string {
	const (
		colorReset  = "\033[0m"
		colorGray   = "\033[90m"
		colorBlue   = "\033[34m"
		colorYellow = "\033[33m"
		colorRed    = "\033[31m"
		colorPurple = "\033[35m"
	)

	switch level {
	case LevelDebug:
		return colorGray + "[DEBUG]" + colorReset
	case LevelInfo:
		return colorBlue + "[INFO]" + colorReset
	case LevelWarn:
		return colorYellow + "[WARN]" + colorReset
	case LevelError:
		return colorRed + "[ERROR]" + colorReset
	case LevelFatal:
		return colorPurple + "[FATAL]" + colorReset
	default:
		return "[UNKNOWN]"
	}
}

// Debug method    输出调试日志.
func (l *defaultLogger) Debug(msg string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.logger.Println(l.formatMessage(LevelDebug, msg, args...))
	}
}

// Info method    输出信息日志.
func (l *defaultLogger) Info(msg string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.logger.Println(l.formatMessage(LevelInfo, msg, args...))
	}
}

// Warn method    输出警告日志.
func (l *defaultLogger) Warn(msg string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.logger.Println(l.formatMessage(LevelWarn, msg, args...))
	}
}

// Error method    输出错误日志.
func (l *defaultLogger) Error(msg string, args ...interface{}) {
	if l.level <= LevelError {
		l.logger.Println(l.formatMessage(LevelError, msg, args...))
	}
}

// Fatal method    输出致命错误日志并退出.
func (l *defaultLogger) Fatal(msg string, args ...interface{}) {
	if l.level <= LevelFatal {
		l.logger.Println(l.formatMessage(LevelFatal, msg, args...))
		os.Exit(1)
	}
}

// SetLevel method    设置日志级别.
func (l *defaultLogger) SetLevel(level Level) {
	l.level = level
}

// SetOutput method    设置日志输出.
func (l *defaultLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

// WithPrefix method    创建带新前缀的日志器.
func (l *defaultLogger) WithPrefix(prefix string) Logger {
	return &defaultLogger{
		level:      l.level,
		logger:     l.logger,
		prefix:     prefix,
		timeFormat: l.timeFormat,
		colorize:   l.colorize,
	}
}

// Printf function    格式化输出信息日志.
func Printf(format string, args ...interface{}) {
	std.Info(format, args...)
}
