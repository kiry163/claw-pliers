.PHONY: help dev dev-up dev-down build run clean test lint

# 帮助信息
help:
	@echo "Claw Pliers Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  make dev-up      - 启动 Docker 开发环境"
	@echo "  make dev-down    - 停止 Docker 开发环境"
	@echo "  make dev-logs    - 查看 Docker 日志"
	@echo "  make run         - 运行服务端"
	@echo "  make build       - 构建二进制文件"
	@echo "  make build-cli   - 构建 CLI 工具"
	@echo "  make clean       - 清理构建文件"
	@echo "  make test        - 运行测试"
	@echo "  make lint        - 代码检查"

# Docker 开发环境
dev-up:
	docker-compose -f docker-compose.dev.yaml up -d
	@echo "开发环境已启动: http://localhost:8080"

dev-down:
	docker-compose -f docker-compose.dev.yaml down

dev-logs:
	docker-compose -f docker-compose.dev.yaml logs -f

dev-restart: dev-down dev-up

# 构建
build:
	go build -o claw-pliers ./cmd/claw-pliers

build-cli:
	go build -o claw-pliers-cli ./cli

build-all: build build-cli

# 运行
run:
	go run ./cmd/claw-pliers

# 清理
clean:
	rm -f claw-pliers claw-pliers-cli
	rm -rf logs/*

# 测试
test:
	go test -v ./...

# 代码检查
lint:
	go vet ./...

# 格式化
fmt:
	gofmt -w .
	goimports -w .

# 创建必要的目录
init:
	mkdir -p data logs
