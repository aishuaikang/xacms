# TCP 服务器重构总结

## 重构内容

本次重构将 TCP 连接处理逻辑完全封装到 `utils.TCPServer` 中，支持自定义协议解析器，开发者只需关注业务逻辑。

## 文件结构

```
internal/pkg/
├── utils/
│   └── tcp_server.go           # 通用TCP服务器封装
├── protocol/
│   ├── strike_protocol.go      # 打击设备协议定义
│   └── strike_parser.go        # 打击设备协议解析器
├── devices/
│   ├── parse_device.go         # 解析设备（使用换行符协议）
│   └── strike_device.go        # 打击设备（使用二进制协议）
docs/
├── TCP服务器使用说明.md
└── 打击设备协议说明.md
examples/
└── strike_client_test.go       # 测试客户端示例
```

## 核心特性

### 1. 协议解析器接口

```go
type ProtocolParser interface {
    Parse(buffer []byte) ([]byte, int, error)
    Build(data []byte) []byte
}
```

-   **Parse**: 从缓冲区解析完整消息，返回消息数据和已消费字节数
-   **Build**: 构造符合协议的消息包

### 2. 自动处理的功能

✅ **粘包**: 自动识别并分割多个连续的数据包  
✅ **分包**: 自动等待数据接收完整再处理  
✅ **心跳**: 可选的应用层心跳检测  
✅ **并发**: 每个连接独立的 goroutine  
✅ **错误恢复**: 解析失败时自动恢复

### 3. 支持的协议类型

#### 换行符协议（默认）

-   适用场景：文本数据、调试、简单协议
-   分隔符：`\n`（可自定义）
-   示例：`parse_device.go`

#### 二进制协议（自定义）

-   适用场景：嵌入式设备、硬件通信
-   格式：Header(1) + ID(2) + Length(1) + Cmd(1) + Data(N) + Tail(1)
-   示例：`strike_device.go`

## 使用方法

### 步骤 1: 实现 MessageHandler 接口

```go
type YourDevice struct {
    tcpServer *utils.TCPServer
}

func (d *YourDevice) OnMessage(data []byte, conn net.Conn, remoteAddr string) []byte {
    // 处理接收到的消息
    // 返回响应数据，或返回nil不回复
}

func (d *YourDevice) OnConnect(conn net.Conn, remoteAddr string) {
    // 客户端连接时的处理
}

func (d *YourDevice) OnDisconnect(conn net.Conn, remoteAddr string) {
    // 客户端断开时的处理
}
```

### 步骤 2: 配置并启动服务器

```go
func (d *YourDevice) Run() {
    config := utils.TCPServerConfig{
        Address:         "0.0.0.0:8080",
        EnableHeartbeat: true,
        // 使用默认的换行符解析器
        // 或者指定自定义解析器：
        // Parser: protocol.NewStrikeProtocolParser(),
    }

    d.tcpServer = utils.NewTCPServer(config, d)
    d.tcpServer.Start()
}
```

### 步骤 3: 使用工具方法

```go
// 发送消息到指定客户端
d.tcpServer.SendToClient(remoteAddr, data)

// 广播消息到所有客户端
d.tcpServer.Broadcast(data)

// 获取连接数
count := d.tcpServer.GetClientCount()
```

## 打击设备协议示例

### 协议格式

```
AA 00 01 08 04 01 64 55
│  │  │  │  │  │  │  │
│  │  │  │  │  │  │  └─ 协议尾 (0x55)
│  │  │  │  │  │  └──── 功率值 (100)
│  │  │  │  │  └─────── 通道号 (1)
│  │  │  │  └────────── 命令 (设置功率)
│  │  │  └───────────── 总长度 (8字节)
│  └──└──────────────── 设备ID (1, 大端序)
└────────────────────── 协议头 (0xAA)
```

### 发送命令

```go
// 方法1: 使用协议包构造函数
packet := protocol.BuildStrikePacket(1, protocol.CmdSetPower, []byte{0x01, 0x64})
strikeDevice.SendToDevice(remoteAddr, packet)

// 方法2: 使用便捷方法
strikeDevice.SendPacketToDevice(remoteAddr, 1, protocol.CmdSetPower, []byte{0x01, 0x64})
```

### 处理命令

在 `strike_device.go` 中，协议已自动解析，只需实现业务逻辑：

```go
func (s *StrikeDevice) handleSetPower(packet *protocol.StrikePacket, remoteAddr string) []byte {
    channel := packet.Data[0]
    power := packet.Data[1]

    // 执行业务逻辑
    // ...

    // 返回响应
    return protocol.BuildStrikePacket(packet.ID, protocol.CmdSetPower, []byte{0x01})
}
```

## 测试

### 编译并运行测试客户端

```bash
cd examples
go run strike_client_test.go
```

### 测试内容

1. ✅ 发送设置功率命令
2. ✅ 发送查询模块信息命令
3. ✅ 粘包测试（连续发送 3 个包）
4. ✅ 分包测试（分 3 次发送 1 个包）

## 优势

### 代码简洁

-   业务代码只需关注数据处理
-   无需处理 TCP 连接、粘包、分包等底层细节

### 协议灵活

-   支持任意自定义协议
-   通过实现 `ProtocolParser` 接口即可

### 易于扩展

-   新增设备只需实现 `MessageHandler` 接口
-   可复用 `TCPServer` 和现有协议解析器

### 健壮可靠

-   完善的错误处理
-   自动恢复机制
-   心跳检测保证连接有效性

## 配置项说明

| 配置项                 | 类型           | 默认值 | 说明                       |
| ---------------------- | -------------- | ------ | -------------------------- |
| Address                | string         | -      | **必填**，监听地址         |
| HeartbeatTimeout       | time.Duration  | 30s    | 心跳超时时间               |
| HeartbeatCheckInterval | time.Duration  | 10s    | 心跳检测间隔               |
| ReadBufferSize         | int            | 4096   | 读取缓冲区大小             |
| MessageDelimiter       | byte           | '\n'   | 消息分隔符（仅默认解析器） |
| EnableHeartbeat        | bool           | true   | 是否启用心跳               |
| Parser                 | ProtocolParser | nil    | 自定义协议解析器           |

## 扩展指南

### 创建自定义协议

1. 定义协议格式（参考 `strike_protocol.go`）
2. 实现 `ProtocolParser` 接口（参考 `strike_parser.go`）
3. 在设备初始化时指定自定义解析器

```go
serverConfig := utils.TCPServerConfig{
    Parser: &YourProtocolParser{},
    // ... 其他配置
}
```

### 支持其他设备

1. 创建设备结构体
2. 实现 `MessageHandler` 接口
3. 配置并启动 TCP 服务器

参考现有的 `parse_device.go` 和 `strike_device.go` 实现。

## 总结

本次重构实现了：

1. ✅ TCP 连接处理完全封装
2. ✅ 自动处理粘包和分包
3. ✅ 支持自定义协议解析器
4. ✅ 打击设备二进制协议支持
5. ✅ 应用层心跳检测
6. ✅ 完善的错误处理和日志
7. ✅ 易于测试和扩展

现在你可以：

-   专注于业务逻辑，无需关心 TCP 底层细节
-   轻松支持多种协议（文本、二进制）
-   快速开发新的设备接入功能
