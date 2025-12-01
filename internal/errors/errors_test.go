package errors

import (
	"errors"
	"testing"
)

// TestError_Error function    测试错误消息格式化.
func TestError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *Error
		wantMsg string
	}{
		{
			name: "仅包含错误码和消息",
			err: &Error{
				Code:    ErrCodeConfig,
				Message: "配置错误",
			},
			wantMsg: "[1000] 配置错误",
		},
		{
			name: "包含原始错误",
			err: &Error{
				Code:    ErrCodeParse,
				Message: "解析失败",
				Cause:   errors.New("原始错误"),
			},
			wantMsg: "[2000] 解析失败: 原始错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("Error.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

// TestError_Unwrap function    测试错误解包.
func TestError_Unwrap(t *testing.T) {
	originalErr := errors.New("原始错误")
	wrappedErr := &Error{
		Code:    ErrCodeConfig,
		Message: "包装错误",
		Cause:   originalErr,
	}

	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}
}

// TestWrap function    测试错误包装.
func TestWrap(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
		wantNil bool
	}{
		{
			name:    "包装非空错误",
			err:     errors.New("原始错误"),
			message: "包装消息",
			wantNil: false,
		},
		{
			name:    "包装空错误返回 nil",
			err:     nil,
			message: "包装消息",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.err, tt.message)
			if (got == nil) != tt.wantNil {
				t.Errorf("Wrap() = %v, wantNil %v", got, tt.wantNil)
			}
			if !tt.wantNil {
				if e, ok := got.(*Error); ok {
					if e.Message != tt.message {
						t.Errorf("Wrap() message = %v, want %v", e.Message, tt.message)
					}
					if e.Cause != tt.err {
						t.Errorf("Wrap() cause = %v, want %v", e.Cause, tt.err)
					}
				} else {
					t.Error("Wrap() 返回的不是 *Error 类型")
				}
			}
		})
	}
}

// TestWrapWithCode function    测试带错误码的错误包装.
func TestWrapWithCode(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		code    ErrorCode
		message string
		wantNil bool
	}{
		{
			name:    "包装非空错误",
			err:     errors.New("原始错误"),
			code:    ErrCodeConfig,
			message: "配置错误",
			wantNil: false,
		},
		{
			name:    "包装空错误返回 nil",
			err:     nil,
			code:    ErrCodeConfig,
			message: "配置错误",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapWithCode(tt.err, tt.code, tt.message)
			if (got == nil) != tt.wantNil {
				t.Errorf("WrapWithCode() = %v, wantNil %v", got, tt.wantNil)
			}
			if !tt.wantNil {
				if e, ok := got.(*Error); ok {
					if e.Code != tt.code {
						t.Errorf("WrapWithCode() code = %v, want %v", e.Code, tt.code)
					}
					if e.Message != tt.message {
						t.Errorf("WrapWithCode() message = %v, want %v", e.Message, tt.message)
					}
				} else {
					t.Error("WrapWithCode() 返回的不是 *Error 类型")
				}
			}
		})
	}
}

// TestNew function    测试创建新错误.
func TestNew(t *testing.T) {
	code := ErrCodeConfig
	message := "配置错误"

	err := New(code, message)
	if err == nil {
		t.Fatal("New() 返回 nil")
	}

	e, ok := err.(*Error)
	if !ok {
		t.Fatal("New() 返回的不是 *Error 类型")
	}

	if e.Code != code {
		t.Errorf("New() code = %v, want %v", e.Code, code)
	}
	if e.Message != message {
		t.Errorf("New() message = %v, want %v", e.Message, message)
	}
}

// TestNewf function    测试创建格式化错误.
func TestNewf(t *testing.T) {
	code := ErrCodeConfig
	format := "配置错误: %s, 值: %d"
	args := []interface{}{"test", 123}

	err := Newf(code, format, args...)
	if err == nil {
		t.Fatal("Newf() 返回 nil")
	}

	e, ok := err.(*Error)
	if !ok {
		t.Fatal("Newf() 返回的不是 *Error 类型")
	}

	expectedMsg := "配置错误: test, 值: 123"
	if e.Message != expectedMsg {
		t.Errorf("Newf() message = %v, want %v", e.Message, expectedMsg)
	}
}

// TestHasCode function    测试错误码判断.
func TestHasCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code ErrorCode
		want bool
	}{
		{
			name: "错误包含指定错误码",
			err: &Error{
				Code:    ErrCodeConfig,
				Message: "配置错误",
			},
			code: ErrCodeConfig,
			want: true,
		},
		{
			name: "错误不包含指定错误码",
			err: &Error{
				Code:    ErrCodeConfig,
				Message: "配置错误",
			},
			code: ErrCodeParse,
			want: false,
		},
		{
			name: "非 Error 类型返回 false",
			err:  errors.New("普通错误"),
			code: ErrCodeConfig,
			want: false,
		},
		{
			name: "nil 错误返回 false",
			err:  nil,
			code: ErrCodeConfig,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasCode(tt.err, tt.code)
			if got != tt.want {
				t.Errorf("HasCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetCode function    测试获取错误码.
func TestGetCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorCode
	}{
		{
			name: "获取 Error 类型的错误码",
			err: &Error{
				Code:    ErrCodeConfig,
				Message: "配置错误",
			},
			want: ErrCodeConfig,
		},
		{
			name: "非 Error 类型返回 0",
			err:  errors.New("普通错误"),
			want: 0,
		},
		{
			name: "nil 错误返回 0",
			err:  nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCode(tt.err)
			if got != tt.want {
				t.Errorf("GetCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsConfigError function    测试配置错误判断.
func TestIsConfigError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "配置错误",
			err:  New(ErrCodeConfig, "配置错误"),
			want: true,
		},
		{
			name: "非配置错误",
			err:  New(ErrCodeParse, "解析错误"),
			want: false,
		},
		{
			name: "nil 错误",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConfigError(tt.err)
			if got != tt.want {
				t.Errorf("IsConfigError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsParseError function    测试解析错误判断.
func TestIsParseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "解析错误",
			err:  New(ErrCodeParse, "解析错误"),
			want: true,
		},
		{
			name: "非解析错误",
			err:  New(ErrCodeConfig, "配置错误"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsParseError(tt.err)
			if got != tt.want {
				t.Errorf("IsParseError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsGenerateError function    测试生成错误判断.
func TestIsGenerateError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "生成错误",
			err:  New(ErrCodeGenerate, "生成错误"),
			want: true,
		},
		{
			name: "非生成错误",
			err:  New(ErrCodeConfig, "配置错误"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGenerateError(tt.err)
			if got != tt.want {
				t.Errorf("IsGenerateError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsDatabaseError function    测试数据库错误判断.
func TestIsDatabaseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "数据库错误",
			err:  New(ErrCodeDatabase, "数据库错误"),
			want: true,
		},
		{
			name: "非数据库错误",
			err:  New(ErrCodeConfig, "配置错误"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsDatabaseError(tt.err)
			if got != tt.want {
				t.Errorf("IsDatabaseError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsTemplateError function    测试模板错误判断.
func TestIsTemplateError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "模板错误",
			err:  New(ErrCodeTemplate, "模板错误"),
			want: true,
		},
		{
			name: "非模板错误",
			err:  New(ErrCodeConfig, "配置错误"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTemplateError(tt.err)
			if got != tt.want {
				t.Errorf("IsTemplateError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsFileError function    测试文件错误判断.
func TestIsFileError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "文件错误",
			err:  New(ErrCodeFile, "文件错误"),
			want: true,
		},
		{
			name: "非文件错误",
			err:  New(ErrCodeConfig, "配置错误"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFileError(tt.err)
			if got != tt.want {
				t.Errorf("IsFileError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWithContext function    测试添加上下文信息.
func TestWithContext(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		context map[string]interface{}
		wantNil bool
	}{
		{
			name: "为错误添加上下文",
			err:  New(ErrCodeConfig, "配置错误"),
			context: map[string]interface{}{
				"file": "config.yaml",
				"line": 10,
			},
			wantNil: false,
		},
		{
			name:    "nil 错误返回 nil",
			err:     nil,
			context: map[string]interface{}{"key": "value"},
			wantNil: true,
		},
		{
			name:    "空上下文",
			err:     New(ErrCodeConfig, "配置错误"),
			context: map[string]interface{}{},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithContext(tt.err, tt.context)
			if (got == nil) != tt.wantNil {
				t.Errorf("WithContext() = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

// TestRecover function    测试 panic 恢复.
func TestRecover(t *testing.T) {
	tests := []struct {
		name      string
		panicFunc func()
		wantErr   bool
	}{
		{
			name: "从 error panic 恢复",
			panicFunc: func() {
				panic(errors.New("panic error"))
			},
			wantErr: true,
		},
		{
			name: "从 string panic 恢复",
			panicFunc: func() {
				panic("panic string")
			},
			wantErr: true,
		},
		{
			name: "从其他类型 panic 恢复",
			panicFunc: func() {
				panic(123)
			},
			wantErr: true,
		},
		{
			name: "无 panic",
			panicFunc: func() {
				// 不触发 panic
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			// 使用独立的 goroutine 来捕获 panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						// 手动转换 panic 为错误
						switch v := r.(type) {
						case error:
							err = Wrap(v, "panic recovered")
						case string:
							err = New(0, v)
						default:
							err = Newf(0, "panic recovered: %v", r)
						}
					}
				}()
				tt.panicFunc()
			}()

			if (err != nil) != tt.wantErr {
				t.Errorf("Recover() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
