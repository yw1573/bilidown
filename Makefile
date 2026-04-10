.PHONY: all build frontend backend clean run dev

# 变量
BINARY_NAME := bilidown
FRONTEND_DIR := ui
BACKEND_DIR := server
BUILD_DIR := build

# 默认目标
all: build

# 构建所有（前端 -> 后端）
build: frontend backend

# 构建前端（输出到 server/internal/static/ui/）
frontend:
	@echo "Building frontend..."
	cd $(FRONTEND_DIR) && pnpm install && pnpm build

# 构建后端（静态资源已嵌入二进制，输出到 build 目录）
backend:
	@echo "Building backend..."
	mkdir -p $(BUILD_DIR)
	cd $(BACKEND_DIR) && go build -ldflags="-s -w" -o ../$(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/bilidown

# 开发模式（前端开发服务器 + 后端运行）
dev:
	@echo "Starting development servers..."
	@cd $(FRONTEND_DIR) && pnpm dev &
	@cd $(BACKEND_DIR) && go run ./cmd/bilidown

# 运行
run: build
	@echo "Running application..."
	./$(BUILD_DIR)/$(BINARY_NAME).exe

# 清理
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf $(BACKEND_DIR)/internal/static/ui/*

# 安装依赖
install:
	@echo "Installing dependencies..."
	cd $(FRONTEND_DIR) && pnpm install
	cd $(BACKEND_DIR) && go mod download

# 测试
test:
	@echo "Running tests..."
	cd $(BACKEND_DIR) && go test ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	cd $(BACKEND_DIR) && go fmt ./...

# 交叉编译 Linux (AMD64)
build-linux:
	@echo "Building for Linux AMD64..."
	mkdir -p $(BUILD_DIR)
	cd $(BACKEND_DIR) && GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o ../$(BUILD_DIR)/$(BINARY_NAME) ./cmd/bilidown

# 交叉编译 macOS (AMD64)
build-darwin-amd64:
	@echo "Building for macOS AMD64..."
	mkdir -p $(BUILD_DIR)
	cd $(BACKEND_DIR) && GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o ../$(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/bilidown

# 交叉编译 macOS (ARM64)
build-darwin-arm64:
	@echo "Building for macOS ARM64..."
	mkdir -p $(BUILD_DIR)
	cd $(BACKEND_DIR) && GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-s -w" -o ../$(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/bilidown

# 交叉编译 Windows
build-windows:
	@echo "Building for Windows..."
	mkdir -p $(BUILD_DIR)
	cd $(BACKEND_DIR) && GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w" -o ../$(BUILD_DIR)/$(BINARY_NAME).exe ./cmd/bilidown

# 构建所有平台
build-all: build-linux build-darwin-amd64 build-darwin-arm64 build-windows

# 帮助
help:
	@echo "Available targets:"
	@echo "  all              - Build frontend and backend (default)"
	@echo "  build            - Build frontend and backend"
	@echo "  frontend         - Build frontend only"
	@echo "  backend          - Build backend only"
	@echo "  dev              - Start development servers"
	@echo "  run              - Build and run the application"
	@echo "  clean            - Remove build artifacts"
	@echo "  install          - Install dependencies"
	@echo "  test             - Run tests"
	@echo "  fmt              - Format Go code"
	@echo "  build-linux      - Cross-compile for Linux AMD64"
	@echo "  build-darwin-amd64  - Cross-compile for macOS AMD64"
	@echo "  build-darwin-arm64  - Cross-compile for macOS ARM64"
	@echo "  build-windows    - Cross-compile for Windows"
	@echo "  build-all        - Build for all platforms"
	@echo "  help             - Show this help message"