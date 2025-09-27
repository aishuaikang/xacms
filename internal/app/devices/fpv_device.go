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
	fpv_fsm "uav_defender/internal/app/devices/fms/fpv"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"go.uber.org/zap"
)

type FPVDevice struct {
	ctx                 context.Context
	fpvWarningDataCache cache.FPVWarningDataCache
	devicesCache        cache.DevicesCache
}

func NewFPVDevice(ctx context.Context, fpvWarningDataCache cache.FPVWarningDataCache, devicesCache cache.DevicesCache) *FPVDevice {
	return &FPVDevice{
		ctx:                 ctx,
		fpvWarningDataCache: fpvWarningDataCache,
		devicesCache:        devicesCache,
	}
}

func (s *FPVDevice) Start() {
	utils.BuildTcpServer(s.ctx, "FPV模块", fmt.Sprintf(":%d", config.AppConfig.Configuration.FPVPort), s.handleConnection)
}

// handleConnection 处理每个连接
func (s *FPVDevice) handleConnection(module string, conn net.Conn) {
	isFirstResponse := false

	defer conn.Close()
	addr := conn.RemoteAddr().String()
	global.Logger.Info("新的FPV连接来自", zap.String("module", module), zap.String("address", addr))

	// 根据连接的IP地址查找对应的设备
	fpvIP := strings.Split(addr, ":")[0]
	device, ok := s.devicesCache.GetDeviceByFPVIP(fpvIP)
	if !ok {
		global.Logger.Warn("未找到匹配的设备，关闭连接", zap.String("module", module), zap.String("address", addr))
		return
	}

	// 如果状态机在离线状态，尝试切换到扫描状态
	if device.FPVFsm.FSM.Is(string(fpv_fsm.StateOffline)) {
		if err := device.FPVFsm.FSM.Event(s.ctx, string(fpv_fsm.EventToScanning)); err != nil {
			global.Logger.Error("状态机切换到扫描状态失败，关闭连接", zap.String("module", module), zap.String("address", addr), zap.Error(err))
			return
		}
	}
	defer func() {
		// 连接关闭时，切换状态机到离线状态
		if !device.FPVFsm.FSM.Is(string(fpv_fsm.StateOffline)) {
			device.FPVFsm.FSM.Event(s.ctx, string(fpv_fsm.EventToOffline))
		}
	}()

	c := conn_.NewConn(device.ID, conn)
	conn_.FPVConnPool.AddConnection(c)
	defer conn_.FPVConnPool.RemoveConnection(c)

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
				if !isFirstResponse {
					isFirstResponse = true
					// 首次响应，忽略
					continue
				}
				select {
				case c.GetResponseChannel() <- string(fullLine):
				case <-time.After(2 * time.Second):
					global.Logger.Warn("写入 Response 超时(响应):", zap.String("module", module), zap.String("address", addr))
				}
				continue
			}

			// 处理 FPV 警告数据
			if utils.IsFPVWaringData(fullLine) {
				ip := conn.RemoteAddr().String()

				time := time.Now()

				warningData, err := utils.ParseFPVWarningData(fullLine, ip, time.Unix(), device.ID)
				if err != nil {
					global.Logger.Error("解析 FPV 警告数据失败", zap.String("module", module), zap.String("address", addr), zap.Error(err))
					continue
				}

				// 将新的警告数据添加到列表中
				s.fpvWarningDataCache.PushFPVWarning(warningData)

				// 设置最后更新时间
				s.fpvWarningDataCache.SetLastUpdated(time)
			}

			// 移除已处理部分（包括 \r\n）
			buffer.Next(index + 2)
		}

	}

}
