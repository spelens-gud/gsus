# 项目信息
PROJECT_NAME := gsus
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# 构建标志
LDFLAGS := -ldflags "-X 'github.com/spelens-gud/gsus/internal/version.Version=$(VERSION)' \
                      -X 'github.com/spelens-gud/gsus/internal/version.GitCommit=$(GIT_COMMIT)' \
                      -X 'github.com/spelens-gud/gsus/internal/version.BuildTime=$(BUILD_TIME)'"

.PHONY: all build test clean install lint fmt help

# 默认目标
all: test build

# 构建项目
build:
	@echo "Building $(PROJECT_NAME)..."
	go build $(LDFLAGS) -o $(PROJECT_NAME) .
	@echo "Build complete: ./$(PROJECT_NAME)"

# 运行测试
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 清理构建产物
clean:
	@echo "Cleaning..."
	rm -f $(PROJECT_NAME)
	rm -f coverage.out coverage.html
	go clean
	@echo "Clean complete"

# 安装到 GOPATH/bin
install:
	@echo "Installing $(PROJECT_NAME)..."
	go install $(LDFLAGS) .
	@echo "Install complete"

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .
	@echo "Format complete"

# 运行项目
run:
	@echo "Running $(PROJECT_NAME)..."
	go run $(LDFLAGS) . $(ARGS)

# 显示帮助信息
help:
	@echo "Available targets:"
	@echo "  all            - Run tests and build (default)"
	@echo "  build          - Build the project"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install to GOPATH/bin"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  run            - Run the project (use ARGS='...' for arguments)"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make run ARGS='db2struct users'"
	@echo "  make install"
