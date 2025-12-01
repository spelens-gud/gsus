package validator

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateRequired function    测试必填字段验证.
func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "非空值通过验证",
			value:     "test",
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "空值验证失败",
			value:     "",
			fieldName: "field",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequired(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequired() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidatePort function    测试端口号验证.
func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{
			name:    "有效端口 80",
			port:    80,
			wantErr: false,
		},
		{
			name:    "有效端口 8080",
			port:    8080,
			wantErr: false,
		},
		{
			name:    "有效端口 65535",
			port:    65535,
			wantErr: false,
		},
		{
			name:    "无效端口 0",
			port:    0,
			wantErr: true,
		},
		{
			name:    "无效端口 -1",
			port:    -1,
			wantErr: true,
		},
		{
			name:    "无效端口 65536",
			port:    65536,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateRange function    测试数值范围验证.
func TestValidateRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min       int
		max       int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "值在范围内",
			value:     5,
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "值等于最小值",
			value:     1,
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "值等于最大值",
			value:     10,
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "值小于最小值",
			value:     0,
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   true,
		},
		{
			name:      "值大于最大值",
			value:     11,
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRange(tt.value, tt.min, tt.max, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateURL function    测试 URL 验证.
func TestValidateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "有效的 HTTP URL",
			url:       "http://example.com",
			fieldName: "url",
			wantErr:   false,
		},
		{
			name:      "有效的 HTTPS URL",
			url:       "https://example.com/path",
			fieldName: "url",
			wantErr:   false,
		},
		{
			name:      "空 URL",
			url:       "",
			fieldName: "url",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateEmail function    测试邮箱验证.
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "有效邮箱",
			email:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "有效邮箱带数字",
			email:   "user123@test.com",
			wantErr: false,
		},
		{
			name:    "有效邮箱带点号",
			email:   "user.name@example.com",
			wantErr: false,
		},
		{
			name:    "无效邮箱缺少@",
			email:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "无效邮箱缺少域名",
			email:   "test@",
			wantErr: true,
		},
		{
			name:    "空邮箱",
			email:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidatePattern function    测试正则表达式验证.
func TestValidatePattern(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		pattern   string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "匹配数字模式",
			value:     "123",
			pattern:   `^\d+$`,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "不匹配数字模式",
			value:     "abc",
			pattern:   `^\d+$`,
			fieldName: "field",
			wantErr:   true,
		},
		{
			name:      "匹配字母模式",
			value:     "abc",
			pattern:   `^[a-z]+$`,
			fieldName: "field",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePattern(tt.value, tt.pattern, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePattern() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateLength function    测试字符串长度验证.
func TestValidateLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		min       int
		max       int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "长度在范围内",
			value:     "test",
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "长度等于最小值",
			value:     "a",
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "长度小于最小值",
			value:     "",
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   true,
		},
		{
			name:      "长度大于最大值",
			value:     "12345678901",
			min:       1,
			max:       10,
			fieldName: "field",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLength(tt.value, tt.min, tt.max, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLength() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateOneOf function    测试枚举值验证.
func TestValidateOneOf(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		allowed   []string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "值在允许列表中",
			value:     "option1",
			allowed:   []string{"option1", "option2", "option3"},
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "值不在允许列表中",
			value:     "option4",
			allowed:   []string{"option1", "option2", "option3"},
			fieldName: "field",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOneOf(tt.value, tt.allowed, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOneOf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateGoPackageName function    测试 Go 包名验证.
func TestValidateGoPackageName(t *testing.T) {
	tests := []struct {
		name    string
		pkgName string
		wantErr bool
	}{
		{
			name:    "有效包名",
			pkgName: "mypackage",
			wantErr: false,
		},
		{
			name:    "有效包名带下划线",
			pkgName: "my_package",
			wantErr: false,
		},
		{
			name:    "有效包名带数字",
			pkgName: "package123",
			wantErr: false,
		},
		{
			name:    "无效包名以数字开头",
			pkgName: "123package",
			wantErr: true,
		},
		{
			name:    "无效包名包含大写字母",
			pkgName: "MyPackage",
			wantErr: true,
		},
		{
			name:    "无效包名包含连字符",
			pkgName: "my-package",
			wantErr: true,
		},
		{
			name:    "空包名",
			pkgName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGoPackageName(tt.pkgName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGoPackageName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidatePositive function    测试正整数验证.
func TestValidatePositive(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "正整数",
			value:     1,
			fieldName: "field",
			wantErr:   false,
		},
		{
			name:      "零不是正整数",
			value:     0,
			fieldName: "field",
			wantErr:   true,
		},
		{
			name:      "负数不是正整数",
			value:     -1,
			fieldName: "field",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositive(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePositive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateFile function    测试文件验证.
func TestValidateFile(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	// nolint: errcheck
	defer os.Remove(tmpFile.Name())
	_ = tmpFile.Close()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatal(err)
	}
	// nolint: errcheck
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{
			name:    "有效文件",
			file:    tmpFile.Name(),
			wantErr: false,
		},
		{
			name:    "文件不存在",
			file:    "/nonexistent/file.txt",
			wantErr: true,
		},
		{
			name:    "路径是目录不是文件",
			file:    tmpDir,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFile(tt.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateExtension function    测试文件扩展名验证.
func TestValidateExtension(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		allowedExts []string
		wantErr     bool
	}{
		{
			name:        "有效扩展名 .go",
			file:        "test.go",
			allowedExts: []string{".go", ".txt"},
			wantErr:     false,
		},
		{
			name:        "有效扩展名 .txt",
			file:        "test.txt",
			allowedExts: []string{".go", ".txt"},
			wantErr:     false,
		},
		{
			name:        "无效扩展名",
			file:        "test.md",
			allowedExts: []string{".go", ".txt"},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExtension(tt.file, tt.allowedExts)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExtension() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateDBConfig function    测试数据库配置验证.
func TestValidateDBConfig(t *testing.T) {
	tests := []struct {
		name     string
		user     string
		password string
		host     string
		db       string
		port     int
		wantErr  bool
	}{
		{
			name:     "有效配置",
			user:     "root",
			password: "password",
			host:     "localhost",
			db:       "testdb",
			port:     3306,
			wantErr:  false,
		},
		{
			name:     "缺少用户名",
			user:     "",
			password: "password",
			host:     "localhost",
			db:       "testdb",
			port:     3306,
			wantErr:  true,
		},
		{
			name:     "缺少主机",
			user:     "root",
			password: "password",
			host:     "",
			db:       "testdb",
			port:     3306,
			wantErr:  true,
		},
		{
			name:     "无效端口",
			user:     "root",
			password: "password",
			host:     "localhost",
			db:       "testdb",
			port:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDBConfig(tt.user, tt.password, tt.host, tt.db, tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDBConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidatePathWritable function    测试路径可写验证.
func TestValidatePathWritable(t *testing.T) {
	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "test_writable_*")
	if err != nil {
		t.Fatal(err)
	}
	// nolint: errcheck
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "可写目录",
			path:    tmpDir,
			wantErr: false,
		},
		{
			name:    "可写路径（新文件）",
			path:    filepath.Join(tmpDir, "newfile.txt"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePathWritable(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePathWritable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
