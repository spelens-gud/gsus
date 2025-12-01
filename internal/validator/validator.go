// Package validator 提供统一的验证功能.
package validator

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spelens-gud/gsus/internal/errors"
)

// Validator interface    验证器接口.
type Validator interface {
	Validate(data interface{}) error
}

// ValidationError struct    验证错误.
type ValidationError struct {
	Field   string
	Message string
}

// Error method    实现 error 接口.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// NewValidationError function    创建验证错误.
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// ValidateRequired function    验证必填字段.
func ValidateRequired(value, fieldName string) error {
	if value == "" {
		return errors.New(errors.ErrCodeConfig, fmt.Sprintf("%s is required", fieldName))
	}
	return nil
}

// ValidatePath function    验证路径是否存在.
func ValidatePath(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("path not found: %s", path))
		}
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("failed to access path: %s", path))
	}
	return nil
}

// ValidateDir function    验证目录是否存在.
func ValidateDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("directory not found: %s", dir))
		}
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("failed to access directory: %s", dir))
	}
	if !info.IsDir() {
		return errors.New(errors.ErrCodeFile, fmt.Sprintf("path is not a directory: %s", dir))
	}
	return nil
}

// ValidateFile function    验证文件是否存在.
func ValidateFile(file string) error {
	info, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("file not found: %s", file))
		}
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("failed to access file: %s", file))
	}
	if info.IsDir() {
		return errors.New(errors.ErrCodeFile, fmt.Sprintf("path is a directory, not a file: %s", file))
	}
	return nil
}

// ValidateExtension function    验证文件扩展名.
func ValidateExtension(file string, allowedExts []string) error {
	ext := filepath.Ext(file)
	for _, allowed := range allowedExts {
		if ext == allowed {
			return nil
		}
	}
	return errors.New(errors.ErrCodeFile, fmt.Sprintf("invalid file extension: %s, allowed: %v", ext, allowedExts))
}

// ValidateNotEmpty function    验证切片不为空.
func ValidateNotEmpty(slice []string, fieldName string) error {
	if len(slice) == 0 {
		return errors.New(errors.ErrCodeConfig, fmt.Sprintf("%s cannot be empty", fieldName))
	}
	return nil
}

// ValidateRange function    验证数值范围.
func ValidateRange(value, min, max int, fieldName string) error {
	if value < min || value > max {
		return NewValidationError(fieldName, fmt.Sprintf("value %d is out of range [%d, %d]", value, min, max))
	}
	return nil
}

// ValidatePort function    验证端口号.
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return NewValidationError("port", fmt.Sprintf("invalid port number: %d, must be between 1 and 65535", port))
	}
	return nil
}

// ValidateURL function    验证 URL 格式.
func ValidateURL(urlStr, fieldName string) error {
	if urlStr == "" {
		return NewValidationError(fieldName, "URL cannot be empty")
	}
	_, err := url.Parse(urlStr)
	if err != nil {
		return NewValidationError(fieldName, fmt.Sprintf("invalid URL format: %v", err))
	}
	return nil
}

// ValidateEmail function    验证邮箱格式.
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return NewValidationError("email", fmt.Sprintf("invalid email format: %s", email))
	}
	return nil
}

// ValidatePattern function    验证正则表达式匹配.
func ValidatePattern(value, pattern, fieldName string) error {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return NewValidationError(fieldName, fmt.Sprintf("invalid pattern: %v", err))
	}
	if !matched {
		return NewValidationError(fieldName, fmt.Sprintf("value '%s' does not match pattern '%s'", value, pattern))
	}
	return nil
}

// ValidateLength function    验证字符串长度.
func ValidateLength(value string, min, max int, fieldName string) error {
	length := len(value)
	if length < min || length > max {
		return NewValidationError(fieldName, fmt.Sprintf("length %d is out of range [%d, %d]", length, min, max))
	}
	return nil
}

// ValidateOneOf function    验证值是否在允许的列表中.
func ValidateOneOf(value string, allowed []string, fieldName string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return NewValidationError(fieldName, fmt.Sprintf("value '%s' is not in allowed list: %v", value, allowed))
}

// ValidateDBConfig function    验证数据库配置.
func ValidateDBConfig(user, password, host, db string, port int) error {
	if err := ValidateRequired(user, "database user"); err != nil {
		return err
	}
	if err := ValidateRequired(host, "database host"); err != nil {
		return err
	}
	if err := ValidateRequired(db, "database name"); err != nil {
		return err
	}
	if err := ValidatePort(port); err != nil {
		return err
	}
	return nil
}

// ValidatePathWritable function    验证路径是否可写.
func ValidatePathWritable(path string) error {
	// 如果路径不存在，检查父目录
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if err := ValidateDir(dir); err != nil {
			return err
		}
		path = dir
	}

	// 尝试创建临时文件测试写权限
	testFile := filepath.Join(path, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("path is not writable: %s", path))
	}
	_ = f.Close()
	_ = os.Remove(testFile)
	return nil
}

// ValidateGoPackageName function    验证 Go 包名格式.
func ValidateGoPackageName(name string) error {
	if name == "" {
		return NewValidationError("package name", "package name cannot be empty")
	}
	// Go 包名规则：小写字母、数字、下划线，不能以数字开头
	matched, _ := regexp.MatchString(`^[a-z_][a-z0-9_]*$`, name)
	if !matched {
		return NewValidationError("package name", fmt.Sprintf("invalid Go package name: %s", name))
	}
	return nil
}

// ValidatePositive function    验证正整数.
func ValidatePositive(value int, fieldName string) error {
	if value <= 0 {
		return NewValidationError(fieldName, fmt.Sprintf("value must be positive, got: %d", value))
	}
	return nil
}

// ValidateNonNegative function    验证非负整数.
func ValidateNonNegative(value int, fieldName string) error {
	if value < 0 {
		return NewValidationError(fieldName, fmt.Sprintf("value must be non-negative, got: %d", value))
	}
	return nil
}

// ValidateStringNotEmpty function    验证字符串非空（去除空白后）.
func ValidateStringNotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(fieldName, "value cannot be empty or whitespace only")
	}
	return nil
}

// ValidateNumericString function    验证数字字符串.
func ValidateNumericString(value, fieldName string) error {
	if _, err := strconv.Atoi(value); err != nil {
		return NewValidationError(fieldName, fmt.Sprintf("value '%s' is not a valid number", value))
	}
	return nil
}

// ValidateMapNotEmpty function    验证 map 不为空.
func ValidateMapNotEmpty(m map[string]string, fieldName string) error {
	if len(m) == 0 {
		return NewValidationError(fieldName, "map cannot be empty")
	}
	return nil
}
