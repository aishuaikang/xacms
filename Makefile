# Go项目的简单Makefile

# 检测操作系统
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    BINARY_NAME := main.exe
    RM_CMD := del /f
    NULL_DEVICE := nul
    WHICH_CMD := where
    AND_OP := &
else
    DETECTED_OS := $(shell uname -s)
    BINARY_NAME := main
    RM_CMD := rm -f
    NULL_DEVICE := /dev/null
    WHICH_CMD := command -v
    AND_OP := &&
endif

# 构建应用程序
all: build test

build:
	@echo "正在为 $(DETECTED_OS) 构建..."
	@go build -o $(BINARY_NAME) ./cmd/api

# 运行应用程序
run:
	@go run ./cmd/api
# 创建数据库容器
docker-run:
ifeq ($(OS),Windows_NT)
	@docker compose up --build 2>nul || docker-compose up --build
else
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "回退到 Docker Compose V1"; \
		docker-compose up --build; \
	fi
endif

# 关闭数据库容器
docker-down:
ifeq ($(OS),Windows_NT)
	@docker compose down 2>nul || docker-compose down
else
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "回退到 Docker Compose V1"; \
		docker-compose down; \
	fi
endif

# 测试应用程序
test:
	@echo "正在测试..."
	@go test ./... -v
# 应用程序集成测试
itest:
	@echo "正在运行集成测试..."
	@go test ./internal/database -v

# 清理二进制文件
clean:
	@echo "正在清理 $(BINARY_NAME)..."
ifeq ($(OS),Windows_NT)
	@$(RM_CMD) $(BINARY_NAME) 2>$(NULL_DEVICE) || echo "没有需要清理的二进制文件"
else
	@$(RM_CMD) $(BINARY_NAME)
endif

# 为当前操作系统设置air配置
setup-air:
ifeq ($(OS),Windows_NT)
	@echo "正在为Windows设置air..."
	@scripts\setup-air.bat
else
	@echo "正在为类Unix系统设置air..."
	@chmod +x scripts/setup-air.sh
	@scripts/setup-air.sh
endif

# 热重载
watch: setup-air
ifeq ($(OS),Windows_NT)
	@$(WHICH_CMD) air >$(NULL_DEVICE) 2>&1 $(AND_OP) ( \
		air $(AND_OP) echo 正在监控... \
	) || ( \
		echo 您的机器上未安装Go的'air'工具。正在安装... $(AND_OP) \
		go install github.com/air-verse/air@latest $(AND_OP) \
		air $(AND_OP) echo 正在监控... \
	)
else
	@if $(WHICH_CMD) air > $(NULL_DEVICE); then \
		air; \
		echo "正在监控..."; \
	else \
		read -p "您的机器上未安装Go的'air'工具。您想要安装它吗？[Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/air-verse/air@latest; \
			air; \
			echo "正在监控..."; \
		else \
			echo "您选择不安装air。正在退出..."; \
			exit 1; \
		fi; \
	fi
endif

# 显示帮助信息
help:
	@echo "$(DETECTED_OS) 可用命令："
	@echo "  build      - 构建应用程序 (输出: $(BINARY_NAME))"
	@echo "  run        - 直接运行应用程序"
	@echo "  watch      - 启动热重载开发模式"
	@echo "  test       - 运行所有测试"
	@echo "  itest      - 运行集成测试"
	@echo "  clean      - 删除构建的二进制文件"
	@echo "  setup-air  - 为当前操作系统设置air配置"
	@echo "  docker-run - 启动数据库容器"
	@echo "  docker-down- 停止数据库容器"
	@echo "  help       - 显示此帮助信息"

# 显示操作系统信息
info:
	@echo "检测到的操作系统: $(DETECTED_OS)"
	@echo "二进制文件名: $(BINARY_NAME)"
	@echo "Go版本: $(shell go version)"

.PHONY: all build run test clean watch docker-run docker-down itest setup-air help info
