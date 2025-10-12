# GCash DeepLink Generator Makefile
# 支持跨平台打包: Windows, Linux, macOS

# 变量定义
APP_NAME=gcash-deeplink
BUILD_DIR=build
PUBLIC_DIR=public
VERSION?=1.0.0
TIMESTAMP=$(shell date +%Y%m%d_%H%M%S)

# Go 编译参数
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION)"
GO_BUILD=go build $(LDFLAGS)

# 颜色输出
CYAN=\033[0;36m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

.PHONY: all clean build build-all build-linux build-darwin build-windows help

# 默认目标
all: build

# 显示帮助信息
help:
	@echo "$(CYAN)GCash DeepLink Generator - 构建命令$(NC)"
	@echo ""
	@echo "$(GREEN)可用命令:$(NC)"
	@echo "  make build         - 构建当前平台"
	@echo "  make build-all     - 构建所有平台 (Linux, macOS, Windows)"
	@echo "  make build-linux   - 构建 Linux 版本"
	@echo "  make build-darwin  - 构建 macOS 版本"
	@echo "  make build-windows - 构建 Windows 版本"
	@echo "  make clean         - 清理构建目录"
	@echo "  make run           - 运行程序"
	@echo "  make help          - 显示此帮助信息"
	@echo ""

# 清理构建目录
clean:
	@echo "$(YELLOW)清理构建目录...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)✓ 清理完成$(NC)"

# 构建当前平台
build: clean
	@echo "$(CYAN)构建当前平台版本...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GO_BUILD) -o $(BUILD_DIR)/$(APP_NAME)
	@cp -r $(PUBLIC_DIR) $(BUILD_DIR)/
	@echo "$(GREEN)✓ 构建完成: $(BUILD_DIR)/$(APP_NAME)$(NC)"
	@echo "$(GREEN)✓ 静态文件已复制到: $(BUILD_DIR)/$(PUBLIC_DIR)$(NC)"

# 构建所有平台
build-all: clean build-linux build-darwin build-windows
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)✓ 所有平台构建完成!$(NC)"
	@echo "$(GREEN)========================================$(NC)"
	@echo ""
	@echo "$(CYAN)构建产物:$(NC)"
	@ls -lh $(BUILD_DIR)/*/* 2>/dev/null | grep -E "$(APP_NAME)" || true
	@echo ""

# 构建 Linux (amd64)
build-linux:
	@echo "$(CYAN)构建 Linux (amd64)...$(NC)"
	@mkdir -p $(BUILD_DIR)/linux-amd64
	@GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/linux-amd64/$(APP_NAME)
	@cp -r $(PUBLIC_DIR) $(BUILD_DIR)/linux-amd64/
	@echo "$(GREEN)✓ Linux (amd64) 构建完成$(NC)"

# 构建 Linux (arm64)
build-linux-arm64:
	@echo "$(CYAN)构建 Linux (arm64)...$(NC)"
	@mkdir -p $(BUILD_DIR)/linux-arm64
	@GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(BUILD_DIR)/linux-arm64/$(APP_NAME)
	@cp -r $(PUBLIC_DIR) $(BUILD_DIR)/linux-arm64/
	@echo "$(GREEN)✓ Linux (arm64) 构建完成$(NC)"

# 构建 macOS (amd64 - Intel)
build-darwin:
	@echo "$(CYAN)构建 macOS (amd64 - Intel)...$(NC)"
	@mkdir -p $(BUILD_DIR)/darwin-amd64
	@GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/darwin-amd64/$(APP_NAME)
	@cp -r $(PUBLIC_DIR) $(BUILD_DIR)/darwin-amd64/
	@echo "$(GREEN)✓ macOS (amd64) 构建完成$(NC)"

# 构建 macOS (arm64 - Apple Silicon)
build-darwin-arm64:
	@echo "$(CYAN)构建 macOS (arm64 - Apple Silicon)...$(NC)"
	@mkdir -p $(BUILD_DIR)/darwin-arm64
	@GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(BUILD_DIR)/darwin-arm64/$(APP_NAME)
	@cp -r $(PUBLIC_DIR) $(BUILD_DIR)/darwin-arm64/
	@echo "$(GREEN)✓ macOS (arm64) 构建完成$(NC)"

# 构建 Windows (amd64)
build-windows:
	@echo "$(CYAN)构建 Windows (amd64)...$(NC)"
	@mkdir -p $(BUILD_DIR)/windows-amd64
	@GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/windows-amd64/$(APP_NAME).exe
	@cp -r $(PUBLIC_DIR) $(BUILD_DIR)/windows-amd64/
	@echo "$(GREEN)✓ Windows (amd64) 构建完成$(NC)"

# 构建 Windows (386)
build-windows-386:
	@echo "$(CYAN)构建 Windows (386)...$(NC)"
	@mkdir -p $(BUILD_DIR)/windows-386
	@GOOS=windows GOARCH=386 $(GO_BUILD) -o $(BUILD_DIR)/windows-386/$(APP_NAME).exe
	@cp -r $(PUBLIC_DIR) $(BUILD_DIR)/windows-386/
	@echo "$(GREEN)✓ Windows (386) 构建完成$(NC)"

# 构建完整版本（包含所有架构）
build-full: clean build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows build-windows-386
	@echo ""
	@echo "$(GREEN)========================================$(NC)"
	@echo "$(GREEN)✓ 完整构建完成（所有平台和架构）!$(NC)"
	@echo "$(GREEN)========================================$(NC)"

# 打包所有平台（创建压缩包）
package: build-all
	@echo "$(CYAN)创建发布包...$(NC)"
	@cd $(BUILD_DIR) && for dir in */; do \
		tar -czf "$${dir%/}-$(VERSION).tar.gz" "$$dir"; \
		echo "$(GREEN)✓ 已创建: $${dir%/}-$(VERSION).tar.gz$(NC)"; \
	done
	@echo "$(GREEN)✓ 所有发布包创建完成$(NC)"

# 运行程序
run:
	@echo "$(CYAN)启动程序...$(NC)"
	@go run main.go

# 运行示例
examples:
	@echo "$(CYAN)运行示例...$(NC)"
	@go run main.go examples

# 运行测试
test:
	@echo "$(CYAN)运行测试...$(NC)"
	@go test -v ./...

# 安装依赖
deps:
	@echo "$(CYAN)安装依赖...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)✓ 依赖安装完成$(NC)"

# 代码格式化
fmt:
	@echo "$(CYAN)格式化代码...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ 代码格式化完成$(NC)"

# 代码检查
lint:
	@echo "$(CYAN)代码检查...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ 代码检查完成$(NC)"

# 显示版本信息
version:
	@echo "$(CYAN)GCash DeepLink Generator$(NC)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(TIMESTAMP)"
