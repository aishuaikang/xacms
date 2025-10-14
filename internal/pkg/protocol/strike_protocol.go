package protocol

import (
	"encoding/binary"
	"fmt"
	"uav_defender/internal/models"
)

// 协议常量
const (
	ProtocolHeader = 0xAA
	ProtocolTail   = 0x55

	// 固定头部长度: Header(1) + ID(2) + Length(1) + Cmd(1) + Tail(1) = 6
	MinPacketLength = 6

	// 命令类型
	CmdSetPower     = 0x04 // 设置N路开关功率
	CmdQueryModules = 0x05 // 查询N路模块信息
	CmdDownloadAddr = 0x10 // 下载地址列表
	CmdQueryAddr    = 0x11 // 查询地址列表
)

// StrikePacket 打击设备协议包
type StrikePacket struct {
	Header uint8  // 协议头 0xAA
	ID     uint16 // 设备ID
	Length uint8  // 数据包总长度
	Cmd    uint8  // 命令类型
	Data   []byte // 数据内容（可选）
	Tail   uint8  // 协议尾 0x55
}

// ParseStrikePacket 从字节数组解析打击设备协议包
func ParseStrikePacket(data []byte) (*StrikePacket, error) {
	// 固定头部长度: Header(1) + ID(2) + Length(1) + Cmd(1) + Tail(1) = 6
	if len(data) < MinPacketLength {
		return nil, fmt.Errorf("数据长度不足，需要至少%d字节，实际%d字节", MinPacketLength, len(data))
	}

	packet := &StrikePacket{}

	// 逐字段解析，避免因平台相关类型大小导致偏移错误
	packet.Header = data[0]
	if packet.Header != ProtocolHeader {
		return nil, fmt.Errorf("无效的协议头: 0x%02X，期望0x%02X", packet.Header, ProtocolHeader)
	}

	packet.ID = binary.BigEndian.Uint16(data[1:3])
	packet.Length = data[3]
	packet.Cmd = data[4]

	// 期望的总长度校验
	expectedTotal := int(packet.Length)
	if expectedTotal < MinPacketLength {
		return nil, fmt.Errorf("长度字段非法: %d，最小应为%d", packet.Length, MinPacketLength)
	}
	if len(data) != expectedTotal {
		return nil, fmt.Errorf("实际长度(%d)与长度字段(%d)不一致", len(data), expectedTotal)
	}

	// 解析数据部分
	dataLen := expectedTotal - MinPacketLength
	if dataLen > 0 {
		packet.Data = make([]byte, dataLen)
		copy(packet.Data, data[5:5+dataLen])
	} else {
		packet.Data = nil
	}

	// 解析协议尾
	packet.Tail = data[5+dataLen]
	if packet.Tail != ProtocolTail {
		return nil, fmt.Errorf("无效的协议尾: 0x%02X，期望0x%02X", packet.Tail, ProtocolTail)
	}

	return packet, nil
}

// // BuildStrikePacket 构造打击设备协议包
// func BuildStrikePacket(id uint16, cmd uint8, data []byte) []byte {
// 	dataLen := len(data)
// 	totalLen := MinPacketLength + dataLen

// 	packet := make([]byte, totalLen)

// 	// 协议头
// 	packet[0] = ProtocolHeader

// 	// 设备ID（大端序）
// 	binary.BigEndian.PutUint16(packet[1:3], id)

// 	// 总长度
// 	packet[3] = uint8(totalLen)

// 	// 命令类型
// 	packet[4] = cmd

// 	// 数据内容
// 	if dataLen > 0 {
// 		copy(packet[5:], data)
// 	}

// 	// 协议尾
// 	packet[5+dataLen] = ProtocolTail

// 	return packet
// }

// String 返回协议包的字符串表示
func (p *StrikePacket) String() string {
	return fmt.Sprintf("StrikePacket{Header:0x%02X, ID:%d, Length:%d, Cmd:0x%02X, DataLen:%d, Tail:0x%02X}",
		p.Header, p.ID, p.Length, p.Cmd, len(p.Data), p.Tail)
}

// GetCmdName 获取命令名称
func (p *StrikePacket) GetCmdName() string {
	switch p.Cmd {
	case CmdSetPower:
		return "设置N路开关功率"
	case CmdQueryModules:
		return "查询N路模块信息"
	case CmdDownloadAddr:
		return "下载地址列表"
	case CmdQueryAddr:
		return "查询地址列表"
	default:
		return fmt.Sprintf("未知命令(0x%02X)", p.Cmd)
	}
}

// 频段到索引的映射
var freqToIndexMap = map[string]int{
	"433M": 1,
	"915M": 2,
	"1.2G": 3,
	"1.4G": 4,
	"1.5G": 5,
	"1.8G": 6,
	"3.3G": 7,
	"5.2G": 8,
	"2.4G": 9,
	"5.8G": 10,
}

// 频段索引到功率的默认值
const (
	MaxChannels       = 10   // 最大通道数
	DefaultPower      = 0x3C // 默认功率60% (0-7通道)
	HighPower         = 0x64 // 高功率100% (8-9通道)
	LowPowerThreshold = 8    // 低功率通道阈值
)

// BuildStrikeCmd 根据频段列表构建打击命令
// frequencies: 频段列表，如 []string{"2.4G", "5.8G"}
// 返回: 完整的打击命令字节数组
func BuildStrikeCmd(frequencies models.StringSlice) ([]byte, error) {
	if len(frequencies) == 0 {
		return nil, fmt.Errorf("频段列表不能为空")
	}

	// 初始化通道开关和功率数组
	onOffArr := make([]byte, MaxChannels)
	powerArr := make([]byte, MaxChannels)

	// 默认所有通道关闭，功率设为默认值
	for i := 0; i < MaxChannels; i++ {
		onOffArr[i] = 0x00
		powerArr[i] = DefaultPower
	}

	// 过滤并处理有效频段
	validCount := 0
	for _, freq := range frequencies {
		index, ok := freqToIndexMap[freq]
		if !ok {
			// 跳过无效频段，不返回错误（更宽容的处理）
			continue
		}

		// 数组索引从0开始，映射值从1开始
		idx := index - 1

		// 再次校验索引范围（防御性编程）
		if idx < 0 || idx >= MaxChannels {
			return nil, fmt.Errorf("频段 %s 的索引 %d 超出范围 [0, %d)", freq, idx, MaxChannels)
		}

		// 设置通道开启
		onOffArr[idx] = 0x01

		// 根据通道设置功率
		if idx >= LowPowerThreshold {
			powerArr[idx] = HighPower // 通道8-9使用高功率100%
		} else {
			powerArr[idx] = DefaultPower // 通道0-7使用默认功率60%
		}

		validCount++
	}

	// 检查是否有有效频段
	if validCount == 0 {
		return nil, fmt.Errorf("没有有效的频段，支持的频段: %v", getSupportedFrequencies())
	}

	// 构建命令
	// 协议格式: AA FF FF 1A 04 [10个通道的开关+功率] 55
	// 总长度: 1(header) + 2(id) + 1(length) + 1(cmd) + 20(10*2数据) + 1(tail) = 26 (0x1A)
	const cmdLength = 0x1A
	cmd := make([]byte, 0, cmdLength)

	cmd = append(cmd, ProtocolHeader) // 0xAA
	cmd = append(cmd, 0xFF, 0xFF)     // 设备ID: 0xFFFF (广播)
	cmd = append(cmd, cmdLength)      // 总长度: 0x1A (26字节)
	cmd = append(cmd, CmdSetPower)    // 命令: 0x04 (设置功率)

	// 添加10个通道的开关和功率值
	for i := 0; i < MaxChannels; i++ {
		cmd = append(cmd, onOffArr[i], powerArr[i])
	}

	cmd = append(cmd, ProtocolTail) // 0x55

	return cmd, nil
}

// getSupportedFrequencies 获取支持的频段列表
func getSupportedFrequencies() []string {
	freqs := make([]string, 0, len(freqToIndexMap))
	for freq := range freqToIndexMap {
		freqs = append(freqs, freq)
	}
	return freqs
}

// // ParseStrikeCmd 解析打击命令数据
// // data: 命令的数据部分（不包括协议头、ID、长度、命令类型、协议尾）
// // 返回: 各通道的开关状态和功率值
// func ParseStrikeCmd(data []byte) (onOff []bool, power []uint8, error error) {
// 	// 打击命令数据长度应该是20字节（10个通道 * 2）
// 	if len(data) != MaxChannels*2 {
// 		return nil, nil, fmt.Errorf("打击命令数据长度错误，期望%d字节，实际%d字节", MaxChannels*2, len(data))
// 	}

// 	onOff = make([]bool, MaxChannels)
// 	power = make([]uint8, MaxChannels)

// 	for i := 0; i < MaxChannels; i++ {
// 		onOff[i] = data[i*2] == 0x01
// 		power[i] = data[i*2+1]
// 	}

// 	return onOff, power, nil
// }

// // GetFrequencyByIndex 根据索引获取频段名称
// func GetFrequencyByIndex(index int) (string, bool) {
// 	for freq, idx := range freqToIndexMap {
// 		if idx == index {
// 			return freq, true
// 		}
// 	}
// 	return "", false
// }

// // GetIndexByFrequency 根据频段名称获取索引
// func GetIndexByFrequency(freq string) (int, bool) {
// 	idx, ok := freqToIndexMap[freq]
// 	return idx, ok
// }
