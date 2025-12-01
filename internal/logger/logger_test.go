package logger

import (
	"bytes"
	"strings"
	"testing"
)

// TestLevel_String function    测试日志级别字符串转换.
func TestLevel_String(t *testing.T) {
	tests := []struct {
		name  string
		level Level
		want  string
	}{
		{
			name:  "调试级别",
			level: LevelDebug,
			want:  "DEBUG",
		},
		{
			name:  "信息级别",
			level: LevelInfo,
			want:  "INFO",
		},
		{
			name:  "警告级别",
			level: LevelWarn,
			want:  "WARN",
		},
		{
			name:  "错误级别",
			level: LevelError,
			want:  "ERROR",
		},
		{
			name:  "致命错误级别",
			level: LevelFatal,
			want:  "FATAL",
		},
		{
			name:  "未知级别",
			level: Level(999),
			want:  "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseLevel function    测试解析日志级别.
func TestParseLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Level
	}{
		{
			name:  "解析 DEBUG",
			input: "DEBUG",
			want:  LevelDebug,
		},
		{
			name:  "解析 debug 小写",
			input: "debug",
			want:  LevelDebug,
		},
		{
			name:  "解析 INFO",
			input: "INFO",
			want:  LevelInfo,
		},
		{
			name:  "解析 WARN",
			input: "WARN",
			want:  LevelWarn,
		},
		{
			name:  "解析 WARNING",
			input: "WARNING",
			want:  LevelWarn,
		},
		{
			name:  "解析 ERROR",
			input: "ERROR",
			want:  LevelError,
		},
		{
			name:  "解析 FATAL",
			input: "FATAL",
			want:  LevelFatal,
		},
		{
			name:  "解析未知级别返回 INFO",
			input: "UNKNOWN",
			want:  LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.want {
				t.Errorf("ParseLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLogger_Output function    测试日志输出.
func TestLogger_Output(t *testing.T) {
	tests := []struct {
		name     string
		level    Level
		logFunc  func(Logger, string)
		message  string
		contains string
	}{
		{
			name:  "输出调试日志",
			level: LevelDebug,
			logFunc: func(l Logger, msg string) {
				l.Debug(msg)
			},
			message:  "debug message",
			contains: "DEBUG",
		},
		{
			name:  "输出信息日志",
			level: LevelInfo,
			logFunc: func(l Logger, msg string) {
				l.Info(msg)
			},
			message:  "info message",
			contains: "INFO",
		},
		{
			name:  "输出警告日志",
			level: LevelWarn,
			logFunc: func(l Logger, msg string) {
				l.Warn(msg)
			},
			message:  "warn message",
			contains: "WARN",
		},
		{
			name:  "输出错误日志",
			level: LevelError,
			logFunc: func(l Logger, msg string) {
				l.Error(msg)
			},
			message:  "error message",
			contains: "ERROR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New("[test]", tt.level)
			logger.SetOutput(&buf)

			tt.logFunc(logger, tt.message)
			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("日志输出不包含 %s, got: %s", tt.contains, output)
			}
			if !strings.Contains(output, tt.message) {
				t.Errorf("日志输出不包含消息 %s, got: %s", tt.message, output)
			}
		})
	}
}

// TestLogger_Level function    测试日志级别过滤.
func TestLogger_Level(t *testing.T) {
	tests := []struct {
		name       string
		setLevel   Level
		logLevel   Level
		shouldLog  bool
		logMessage string
	}{
		{
			name:       "INFO 级别不输出 DEBUG",
			setLevel:   LevelInfo,
			logLevel:   LevelDebug,
			shouldLog:  false,
			logMessage: "debug message",
		},
		{
			name:       "INFO 级别输出 INFO",
			setLevel:   LevelInfo,
			logLevel:   LevelInfo,
			shouldLog:  true,
			logMessage: "info message",
		},
		{
			name:       "WARN 级别不输出 INFO",
			setLevel:   LevelWarn,
			logLevel:   LevelInfo,
			shouldLog:  false,
			logMessage: "info message",
		},
		{
			name:       "ERROR 级别输出 ERROR",
			setLevel:   LevelError,
			logLevel:   LevelError,
			shouldLog:  true,
			logMessage: "error message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New("[test]", tt.setLevel)
			logger.SetOutput(&buf)

			switch tt.logLevel {
			case LevelDebug:
				logger.Debug(tt.logMessage)
			case LevelInfo:
				logger.Info(tt.logMessage)
			case LevelWarn:
				logger.Warn(tt.logMessage)
			case LevelError:
				logger.Error(tt.logMessage)
			}
			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.shouldLog {
				t.Errorf("期望输出=%v, 实际输出=%v, output: %s", tt.shouldLog, hasOutput, output)
			}
		})
	}
}

// TestLogger_WithPrefix function    测试带前缀的日志器.
func TestLogger_WithPrefix(t *testing.T) {
	var buf bytes.Buffer
	logger := New("[test]", LevelInfo)
	logger.SetOutput(&buf)

	prefixedLogger := logger.WithPrefix("[module]")
	prefixedLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[module]") {
		t.Errorf("日志输出不包含前缀 [module], got: %s", output)
	}
}

// TestLogger_FormatArgs function    测试格式化参数.
func TestLogger_FormatArgs(t *testing.T) {
	var buf bytes.Buffer
	logger := New("[test]", LevelInfo)
	logger.SetOutput(&buf)

	logger.Info("test %s %d", "message", 123)

	output := buf.String()
	if !strings.Contains(output, "test message 123") {
		t.Errorf("日志输出格式化错误, got: %s", output)
	}
}
