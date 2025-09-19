package devices

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"time"
	conn_ "uav_defender/internal/app/devices/conn"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"go.uber.org/zap"
)

type FPVDevice struct {
	ctx                 context.Context
	config              *config.Config
	fpvWarningDataCache cache.FPVWarningDataCache
	devicesCache        cache.DevicesCache
	fpvConnection       *conn_.FpvConnection
}

func NewFPVDevice(ctx context.Context, config *config.Config, fpvWarningDataCache cache.FPVWarningDataCache, devicesCache cache.DevicesCache, fpvConnection *conn_.FpvConnection) *FPVDevice {
	return &FPVDevice{
		ctx:                 ctx,
		config:              config,
		fpvWarningDataCache: fpvWarningDataCache,
		devicesCache:        devicesCache,
		fpvConnection:       fpvConnection,
	}
}

func (s *FPVDevice) Start() {
	utils.BuildTcpServer(s.ctx, "FPV模块", fmt.Sprintf(":%d", s.config.Configuration.FPVPort), s.handleConnection)
}

// handleConnection 处理每个连接
func (s *FPVDevice) handleConnection(module string, conn net.Conn) {
	addr := conn.RemoteAddr().String()
	global.Logger.Info("新的FPV连接来自", zap.String("module", module), zap.String("address", addr))

	// 根据连接的IP地址查找对应的设备
	fpvIP := strings.Split(addr, ":")[0]
	device, ok := s.devicesCache.GetDeviceByFPVIP(fpvIP)
	if !ok {
		global.Logger.Warn("未找到匹配的设备，关闭连接", zap.String("module", module), zap.String("address", addr))
		conn.Close()
		return
	}

	c := conn_.NewConn(conn)
	s.fpvConnection.AddConnection(c)
	defer s.fpvConnection.RemoveConnection(c)

	// 发送给客户端AT 指令
	at := []byte{0x41, 0x54, 0x0D, 0x0A} // 对应 "AT\r\n"
	if _, err := conn.Write(at); err != nil {
		global.Logger.Error("发送AT指令失败", zap.String("module", module), zap.String("address", addr), zap.Error(err))
		return
	}

	scanner := bufio.NewReader(conn)
	var buffer bytes.Buffer

	for {
		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// 读取数据直到遇到换行符
		line, err := scanner.ReadBytes('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				global.Logger.Info("连接超时，关闭连接", zap.String("module", module), zap.String("address", addr))
			} else {
				global.Logger.Warn("读取数据失败", zap.String("module", module), zap.Error(err))
			}
			break
		}

		// 判断 buffer 是否超过 10kB，防止内存耗尽攻击
		if buffer.Len() > 10*1024 {
			global.Logger.Warn("缓冲区数据过大，关闭连接", zap.String("module", module), zap.String("address", addr))
			break
		}

		buffer.Write(line)

		for {
			index := bytes.Index(buffer.Bytes(), []byte("\r\n"))
			if index == -1 {
				// 没有找到完整的行，继续读取
				break
			}

			// 提取完整的行
			fullLine := bytes.TrimSpace(buffer.Next(index + 2)) // 包括 \r\n

			// 处理 FPV 响应数据
			if utils.IsFPVResponse(fullLine) {
				// TODO: 这里可以根据需要处理响应数据
				continue
			}

			// 处理 FPV 警告数据
			if utils.IsFPVWaringData(fullLine) {
				ip := conn.RemoteAddr().String()

				time := time.Now()

				warningData, err := utils.ParseFPVWarningData(fullLine, ip, time.Unix(), device.DetectionID)
				if err != nil {
					// log.Errorf("[%s] 解析 FPV 警告数据失败: %v", module, err)
					global.Logger.Error("解析 FPV 警告数据失败", zap.String("module", module), zap.String("address", addr), zap.Error(err))
					continue
				}

				// log.Infof("[%s] 接收到FPV警告数据: %+v", module, warningData)

				// 将新的警告数据添加到列表中
				s.fpvWarningDataCache.PushFPVWarning(warningData)

				s.fpvWarningDataCache.SetLastUpdated(time)
			}

			// 移除已处理部分（包括 \r\n）
			buffer.Next(index + 2)
		}

	}

}
