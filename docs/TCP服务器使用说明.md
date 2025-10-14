# TCP 服务器使用说明

## 概述

我们已经将 TCP 连接和处理逻辑封装到了 `utils.TCPServer` 中，现在你只需要关注业务数据的处理和回复即可。

## 核心概念

### 1. MessageHandler 接口

任何需要处理 TCP 消息的服务都需要实现这个接口：

```go
type MessageHandler interface {
    // 处理接收到的消息
    OnMessage(data []byte, conn net.Conn, remoteAddr string) []byte

    // 客户端连接时调用（可选）
    OnConnect(conn net.Conn, remoteAddr string)

    // 客户端断开时调用（可选）
    OnDisconnect(conn net.Conn, remoteAddr string)
}
```

### 2. 自动处理的功能

TCP 服务器已经自动处理了以下功能，你无需关心：

✅ **粘包和分包问题**：使用消息分隔符（默认 `\n`）自动分割消息
✅ **心跳检测**：自动检测客户端是否在线，超时自动断开
✅ **并发连接**：每个客户端连接独立的 goroutine 处理
✅ **连接管理**：自动管理所有客户端连接
✅ **错误处理**：自动处理网络错误和超时

## 使用步骤

### 步骤 1: 实现 MessageHandler 接口

```go
type YourService struct {
    tcpServer *utils.TCPServer
}

// OnMessage 处理接收到的消息
func (s *YourService) OnMessage(data []byte, conn net.Conn, remoteAddr string) []byte {
    // 1. 解析数据
    // 2. 执行业务逻辑
    // 3. 返回响应（如果需要），返回nil则不回复

    // 示例：
    if string(data) == "PING" {
        return []byte("PONG") // 自动发送给客户端
    }

    return nil // 不回复
}

// OnConnect 客户端连接时调用
func (s *YourService) OnConnect(conn net.Conn, remoteAddr string) {
    // 可选：连接建立时的初始化操作
    // 例如：记录连接日志、验证客户端身份等
}

// OnDisconnect 客户端断开时调用
func (s *YourService) OnDisconnect(conn net.Conn, remoteAddr string) {
    // 可选：连接断开时的清理操作
    // 例如：清除缓存、通知其他模块等
}
```

### 步骤 2: 创建并启动 TCP 服务器

```go
func (s *YourService) Run() {
    // 配置TCP服务器
    config := utils.TCPServerConfig{
        Address:                "0.0.0.0:8080",    // 监听地址
        HeartbeatTimeout:       30 * time.Second,  // 心跳超时（可选）
        HeartbeatCheckInterval: 10 * time.Second,  // 心跳检测间隔（可选）
        ReadBufferSize:         4096,              // 读取缓冲区大小（可选）
        MessageDelimiter:       '\n',              // 消息分隔符（可选）
        EnableHeartbeat:        true,              // 启用心跳（可选）
    }

    // 创建服务器，传入自己作为消息处理器
    s.tcpServer = utils.NewTCPServer(config, s)

    // 启动服务器
    err := s.tcpServer.Start()
    if err != nil {
        log.Fatal(err)
    }
}
```

### 步骤 3: 使用工具方法（可选）

```go
// 发送消息到指定客户端
func (s *YourService) SendToClient(remoteAddr string, message []byte) error {
    return s.tcpServer.SendToClient(remoteAddr, message)
}

// 广播消息到所有客户端
func (s *YourService) BroadcastMessage(message []byte) {
    s.tcpServer.Broadcast(message)
}

// 获取当前连接数
func (s *YourService) GetConnectionCount() int {
    return s.tcpServer.GetClientCount()
}

// 停止服务器
func (s *YourService) Stop() {
    s.tcpServer.Stop()
}
```

## 配置说明

### TCPServerConfig 配置项

| 字段                   | 类型          | 默认值 | 说明                                  |
| ---------------------- | ------------- | ------ | ------------------------------------- |
| Address                | string        | -      | **必填**，监听地址，如 "0.0.0.0:8080" |
| HeartbeatTimeout       | time.Duration | 30s    | 心跳超时时间，超过此时间无数据则断开  |
| HeartbeatCheckInterval | time.Duration | 10s    | 心跳检测间隔                          |
| ReadBufferSize         | int           | 4096   | 读取缓冲区大小（字节）                |
| MessageDelimiter       | byte          | '\n'   | 消息分隔符，用于分割消息              |
| EnableHeartbeat        | bool          | true   | 是否启用心跳检测                      |

## 消息协议

### 默认协议（换行符分隔）

客户端发送的数据应该以 `\n` 结尾：

```
消息1\n
消息2\n
消息3\n
```

服务器会自动处理粘包和分包，将每条完整的消息传递给 `OnMessage` 方法。

### 自定义分隔符

你可以修改 `MessageDelimiter` 来使用其他分隔符：

```go
config := utils.TCPServerConfig{
    MessageDelimiter: '\r',  // 使用回车符
    // 或者
    MessageDelimiter: '|',   // 使用管道符
}
```

## 完整示例

参考 `parse_device.go` 的实现：

```go
type ParseDevice struct {
    tcpServer *utils.TCPServer
}

func (p *ParseDevice) Run() {
    serverConfig := utils.TCPServerConfig{
        Address:                config.AppConfig.Configuration.ParserTcpServer,
        HeartbeatTimeout:       30 * time.Second,
        HeartbeatCheckInterval: 10 * time.Second,
        ReadBufferSize:         4096,
        MessageDelimiter:       '\n',
        EnableHeartbeat:        true,
    }

    p.tcpServer = utils.NewTCPServer(serverConfig, p)
    p.tcpServer.Start()
}

func (p *ParseDevice) OnMessage(data []byte, conn net.Conn, remoteAddr string) []byte {
    // 只关注业务逻辑
    p.parseData(data, remoteAddr)

    // 如果需要回复，返回数据
    return []byte("OK")
}

func (p *ParseDevice) OnConnect(conn net.Conn, remoteAddr string) {
    // 客户端连接时的处理
}

func (p *ParseDevice) OnDisconnect(conn net.Conn, remoteAddr string) {
    // 客户端断开时的处理
}
```

## 测试客户端

使用 `test_client.go` 中的测试客户端进行测试：

```go
// 连接到服务器
client, err := NewTestClient("localhost:8080")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// 发送单条消息
client.SendMessage([]byte("Hello, Server!"))

// 测试粘包
messages := [][]byte{
    []byte("Message 1"),
    []byte("Message 2"),
    []byte("Message 3"),
}
client.SendMultipleMessages(messages)

// 测试分包
largeMessage := make([]byte, 1000)
client.SendMessageInChunks(largeMessage, 100)
```

## 常见问题

### Q: 如何处理二进制数据？

A: 可以修改 `MessageDelimiter` 为特殊字节，或者实现自定义的消息头（包含长度）协议。

### Q: 如何实现客户端主动发送心跳？

A: 客户端定时发送心跳包（如"PING"），服务器在 `OnMessage` 中识别并回复"PONG"。

### Q: 如何禁用心跳？

A: 设置 `EnableHeartbeat: false`。

### Q: 如何获取所有连接的客户端地址？

A: 可以在 `OnConnect` 中记录，在 `OnDisconnect` 中清除。

## 优势

✅ **简单**：只需实现 3 个方法，关注业务逻辑
✅ **可靠**：自动处理粘包、分包、心跳
✅ **高效**：并发处理，支持大量连接
✅ **灵活**：可配置的超时、缓冲区、分隔符
✅ **可复用**：其他 TCP 服务可以直接使用

## 下一步

1. 在 `OnMessage` 中实现你的数据解析逻辑
2. 在 `parseData` 中实现具体的业务处理
3. 根据需要返回响应数据
4. 使用工具方法主动发送消息或广播
