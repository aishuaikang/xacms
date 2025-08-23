# 跨平台开发指南

本项目支持 Windows、macOS 和 Linux 平台的开发。

## 系统要求

-   Go 1.21 或更高版本
-   Make (Windows 用户可以通过 `winget install GnuWin32.Make` 安装)
-   Docker (可选，用于数据库容器)

## 平台检测

Makefile 会自动检测当前操作系统并使用相应的命令：

-   **Windows**: 生成 `main.exe`
-   **macOS/Linux**: 生成 `main`

查看当前平台信息：

```bash
make info
```

## 可用命令

查看所有可用命令：

```bash
make help
```

### 基本命令

```bash
# 构建项目
make build

# 运行项目
make run

# 清理构建文件
make clean

# 运行测试
make test

# 运行集成测试
make itest
```

### 开发命令

```bash
# 启动热重载开发模式
make watch

# 设置air配置（自动根据OS选择）
make setup-air
```

### Docker 命令

```bash
# 启动数据库容器
make docker-run

# 停止数据库容器
make docker-down
```

## 开发环境设置

### Windows

1. 安装 Go: https://golang.org/dl/
2. 安装 Make: `winget install GnuWin32.Make`
3. 克隆项目并运行：
    ```cmd
    git clone <repo-url>
    cd xacms
    make info
    make build
    make watch
    ```

### macOS

1. 安装 Go: `brew install go`
2. Make 通常已预装
3. 克隆项目并运行：
    ```bash
    git clone <repo-url>
    cd xacms
    make info
    make build
    make watch
    ```

### Linux

1. 安装 Go: 参考 https://golang.org/dl/
2. 安装 Make: `sudo apt install make` (Ubuntu/Debian)
3. 克隆项目并运行：
    ```bash
    git clone <repo-url>
    cd xacms
    make info
    make build
    make watch
    ```

## 热重载开发

项目使用 [air](https://github.com/air-verse/air) 进行热重载开发：

```bash
# 首次使用会自动安装air
make watch
```

air 配置会根据当前操作系统自动调整：

-   Windows: 监控 `main.exe`
-   macOS/Linux: 监控 `main`

## 项目结构

```
├── cmd/api/          # 应用程序入口点
├── internal/         # 内部包
│   ├── database/     # 数据库连接
│   ├── models/       # 数据模型
│   ├── routes/       # 路由定义
│   ├── server/       # 服务器配置
│   ├── services/     # 业务逻辑
│   └── utils/        # 工具函数
├── scripts/          # 构建脚本
├── tmp/              # 临时文件（air使用）
├── .air.toml         # air配置文件
├── Makefile          # 跨平台构建配置
└── docker-compose.yml # Docker配置
```

## 故障排除

### Windows 问题

1. **Make 命令未找到**: 安装 GnuWin32 Make
2. **权限问题**: 以管理员身份运行终端
3. **路径问题**: 确保 Go 在 PATH 中

### macOS/Linux 问题

1. **权限问题**: 运行 `chmod +x scripts/setup-air.sh`
2. **Go 未找到**: 确保 Go 正确安装并在 PATH 中

### 常用调试命令

```bash
# 检查Go版本
go version

# 检查依赖
go mod tidy

# 手动构建
go build -v ./cmd/api

# 运行测试（详细输出）
go test -v ./...
```
