package protocol

import (
	"fmt"
	"uav_defender/internal/pkg/utils"
)

// StrikeProtocolParser 打击设备协议解析器
type StrikeProtocolParser struct{}

// Parse 实现 utils.ProtocolParser 接口 - 从缓冲区解析消息
func (p *StrikeProtocolParser) Parse(buffer []byte) ([]byte, int, error) {
	bufLen := len(buffer)

	// 数据长度不足最小包长度
	if bufLen < MinPacketLength {
		return nil, 0, nil // 等待更多数据
	}

	// 查找协议头
	headerIndex := -1
	for i := 0; i < bufLen; i++ {
		if buffer[i] == ProtocolHeader {
			headerIndex = i
			break
		}
	}

	// 没有找到协议头
	if headerIndex == -1 {
		// 丢弃所有数据
		return nil, bufLen, fmt.Errorf("未找到协议头")
	}

	// 协议头不在起始位置，丢弃之前的数据
	if headerIndex > 0 {
		return nil, headerIndex, fmt.Errorf("协议头之前有 %d 字节无效数据", headerIndex)
	}

	// 检查是否有足够的数据读取长度字段
	if bufLen < 4 {
		return nil, 0, nil // 等待更多数据
	}

	// 读取长度字段（第4个字节）
	packetLength := int(buffer[3])

	// 长度字段校验
	if packetLength < MinPacketLength {
		return nil, 1, fmt.Errorf("长度字段非法: %d，最小应为 %d", packetLength, MinPacketLength)
	}

	// 检查是否有足够的数据
	if bufLen < packetLength {
		// 数据不完整，等待更多数据
		return nil, 0, nil
	}

	// 提取完整的数据包
	packet := buffer[:packetLength]

	// 验证协议包（这会做完整性校验）
	_, err := ParseStrikePacket(packet)
	if err != nil {
		// 解析失败，丢弃这个包，从下一个字节继续查找
		return nil, 1, fmt.Errorf("解析协议包失败: %v", err)
	}

	// 返回完整的消息和已消费的字节数
	return packet, packetLength, nil
}

// Build 实现 utils.ProtocolParser 接口 - 构造消息包
// 注意：这里的data应该是完整的协议包（已经包含Header、ID、Length等字段）
// 如果需要构造新包，使用 BuildStrikePacket 函数
func (p *StrikeProtocolParser) Build(data []byte) []byte {
	// 直接返回数据（假设已经是完整的协议包）
	return data
}

// NewStrikeProtocolParser 创建打击设备协议解析器
func NewStrikeProtocolParser() utils.ProtocolParser {
	return &StrikeProtocolParser{}
}
