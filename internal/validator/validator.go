// Package validator 提供统一的验证功能.
package validator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spelens-gud/gsus/internal/errors"
)

// Validator interface    验证器接口.
type Validator interface {
	Validate(data interface{}) error
}

// ValidateRequired function    验证必填字段.
func ValidateRequired(value, fieldName string) error {
	if value == "" {
		return errors.New(errors.ErrCodeConfig, fmt.Sprintf("%s is required", fieldName))
	}
	return nil
}

// ValidatePath function    验证路径是否存在.
func ValidatePath(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("path not found: %s", path))
		}
		return errors.WrapWithCode(err, errors.ErrCodeFile, fmt.Sprintf("failed to access path: %s", path))
	}
	return nil
}

// ValidateDir function    验证目录是否存在.
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

// ValidateFile function    验证文件是否存在.
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

// ValidateExtension function    验证文件扩展名.
func ValidateExtension(file string, allowedExts []string) error {
	ext := filepath.Ext(file)
	for _, allowed := range allowedExts {
		if ext == allowed {
			return nil
		}
	}
	return errors.New(errors.ErrCodeFile, fmt.Sprintf("invalid file extension: %s, allowed: %v", ext, allowedExts))
}

// ValidateNotEmpty function    验证切片不为空.
func ValidateNotEmpty(slice []string, fieldName string) error {
	if len(slice) == 0 {
		return errors.New(errors.ErrCodeConfig, fmt.Sprintf("%s cannot be empty", fieldName))
	}
	return nil
}
