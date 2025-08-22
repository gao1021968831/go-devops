# DevOps管理平台 Makefile

# 变量定义
APP_NAME := go-devops
VERSION := 1.0.0
BUILD_TIME := $(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go相关变量
GO_VERSION := 1.21
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# 构建标志
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "DevOps管理平台构建工具"
	@echo ""
	@echo "使用方法:"
	@echo "  make <target>"
	@echo ""
	@echo "可用目标:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 开发环境
.PHONY: dev
dev: ## 启动开发环境
	@echo "启动后端开发服务器..."
	@go run main.go &
	@echo "启动前端开发服务器..."
	@cd web && npm run dev

.PHONY: dev-backend
dev-backend: ## 仅启动后端开发服务器
	@echo "启动后端开发服务器..."
	@go run main.go

.PHONY: dev-frontend
dev-frontend: ## 仅启动前端开发服务器
	@echo "启动前端开发服务器..."
	@cd web && npm run dev

# 依赖管理
.PHONY: deps
deps: ## 安装所有依赖
	@echo "安装Go依赖..."
	@go mod tidy
	@echo "安装前端依赖..."
	@cd web && npm install

.PHONY: deps-update
deps-update: ## 更新所有依赖
	@echo "更新Go依赖..."
	@go get -u ./...
	@go mod tidy
	@echo "更新前端依赖..."
	@cd web && npm update

# 构建
.PHONY: build
build: build-frontend build-backend ## 构建完整应用

.PHONY: build-backend
build-backend: ## 构建后端应用
	@echo "构建后端应用..."
	@CGO_ENABLED=1 go build $(LDFLAGS) -o bin/$(APP_NAME) main.go
	@echo "后端构建完成: bin/$(APP_NAME)"

.PHONY: build-frontend
build-frontend: ## 构建前端应用
	@echo "构建前端应用..."
	@cd web && npm run build
	@echo "前端构建完成: web/dist/"

.PHONY: build-linux
build-linux: ## 构建Linux版本
	@echo "构建Linux版本..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-linux-amd64 main.go

.PHONY: build-windows
build-windows: ## 构建Windows版本
	@echo "构建Windows版本..."
	@CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-windows-amd64.exe main.go

.PHONY: build-darwin
build-darwin: ## 构建macOS版本
	@echo "构建macOS版本..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(APP_NAME)-darwin-amd64 main.go

.PHONY: build-all
build-all: build-frontend build-linux build-windows build-darwin ## 构建所有平台版本

# 测试
.PHONY: test
test: ## 运行所有测试
	@echo "运行Go测试..."
	@go test -v ./...
	@echo "运行前端测试..."
	@cd web && npm test

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试覆盖率..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告生成: coverage.html"

.PHONY: benchmark
benchmark: ## 运行性能测试
	@echo "运行性能测试..."
	@go test -bench=. -benchmem ./...

# 代码质量
.PHONY: lint
lint: ## 代码检查
	@echo "运行Go代码检查..."
	@golangci-lint run
	@echo "运行前端代码检查..."
	@cd web && npm run lint

.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化Go代码..."
	@go fmt ./...
	@echo "格式化前端代码..."
	@cd web && npm run format

.PHONY: vet
vet: ## Go代码静态分析
	@echo "运行Go静态分析..."
	@go vet ./...

# Docker
.PHONY: docker-build
docker-build: ## 构建Docker镜像
	@echo "构建Docker镜像..."
	@docker build -t $(APP_NAME):$(VERSION) .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

.PHONY: docker-run
docker-run: ## 运行Docker容器
	@echo "运行Docker容器..."
	@docker run -d -p 8080:8080 --name $(APP_NAME) $(APP_NAME):latest

.PHONY: docker-stop
docker-stop: ## 停止Docker容器
	@echo "停止Docker容器..."
	@docker stop $(APP_NAME) || true
	@docker rm $(APP_NAME) || true

.PHONY: docker-compose-up
docker-compose-up: ## 启动Docker Compose
	@echo "启动Docker Compose..."
	@docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## 停止Docker Compose
	@echo "停止Docker Compose..."
	@docker-compose down

# 清理
.PHONY: clean
clean: ## 清理构建文件
	@echo "清理构建文件..."
	@rm -rf bin/
	@rm -rf web/dist/
	@rm -f coverage.out coverage.html
	@go clean -cache
	@cd web && rm -rf node_modules/.cache

.PHONY: clean-all
clean-all: clean ## 清理所有文件包括依赖
	@echo "清理所有文件..."
	@cd web && rm -rf node_modules/
	@go clean -modcache

# 数据库
.PHONY: db-reset
db-reset: ## 重置数据库
	@echo "重置数据库..."
	@rm -f devops.db
	@echo "数据库已重置"

.PHONY: db-backup
db-backup: ## 备份数据库
	@echo "备份数据库..."
	@cp devops.db backups/devops-$(shell date +%Y%m%d_%H%M%S).db
	@echo "数据库备份完成"

# 日志
.PHONY: logs
logs: ## 查看应用日志
	@tail -f logs/app-$(shell date +%Y-%m-%d).log

.PHONY: logs-clean
logs-clean: ## 清理日志文件
	@echo "清理日志文件..."
	@find logs/ -name "*.log" -mtime +30 -delete
	@echo "日志清理完成"

# 发布
.PHONY: release
release: clean build-all ## 创建发布包
	@echo "创建发布包..."
	@mkdir -p release
	@tar -czf release/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C bin $(APP_NAME)-linux-amd64 -C ../web/dist .
	@zip -r release/$(APP_NAME)-$(VERSION)-windows-amd64.zip bin/$(APP_NAME)-windows-amd64.exe web/dist/
	@tar -czf release/$(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz -C bin $(APP_NAME)-darwin-amd64 -C ../web/dist .
	@echo "发布包创建完成: release/"

# 安装工具
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/air-verse/air@latest
	@echo "开发工具安装完成"

# 版本信息
.PHONY: version
version: ## 显示版本信息
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo "Go版本: $(shell go version)"
	@echo "平台: $(GOOS)/$(GOARCH)"
