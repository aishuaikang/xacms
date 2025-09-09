package store

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"
	"time"
	"xacms/internal/pkg/config"
	"xacms/internal/utils"

	"github.com/gofiber/fiber/v2/log"
)

type FPVStore interface {
	// 获取当前的 FPV 警告数据列表
	GetFPVWaringDataList() []*utils.FPVWaringData
}

type fpvStore struct {
	ctx    context.Context
	config *config.Config

	fpvWaringDataList []*utils.FPVWaringData
	mu                sync.RWMutex
}

func NewFPVStore(ctx context.Context, config *config.Config) FPVStore {
	fpvStore := &fpvStore{
		ctx:               ctx,
		config:            config,
		fpvWaringDataList: make([]*utils.FPVWaringData, 0),
		mu:                sync.RWMutex{},
	}
	go fpvStore.startFPVServer()
	return fpvStore
}

func (s *fpvStore) startFPVServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Configuration.FpvPort))
	if err != nil {
		log.Fatalf("无法启动fpv tcp服务器: %v", err)
	}
	defer listener.Close()
	log.Infof("fpv tcp服务器已启动，监听端口: %d", s.config.Configuration.FpvPort)

	// 用一个 goroutine 监听 ctx.Done()，在取消时关闭 listener
	go func() {
		<-s.ctx.Done()
		log.Info("关闭 fpv tcp 服务器...")
		listener.Close() // 会导致 Accept 返回错误，从而退出主循环
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				log.Info("fpv tcp 服务器已关闭")
				return
			default:
				log.Errorf("接受连接失败: %v", err)
				continue
			}
		}
		go s.handleConnection(conn)
	}
}

// handleConnection 处理每个连接
func (s *fpvStore) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Infof("新的FPV连接来自: %s", conn.RemoteAddr().String())

	// 发送给客户端AT 指令
	at := []byte{0x41, 0x54, 0x0D, 0x0A} // 对应 "AT\r\n"

	if _, err := conn.Write(at); err != nil {
		log.Errorf("发送失败: %v", err)
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
				log.Infof("连接超时，关闭连接: %s", conn.RemoteAddr().String())
			} else {
				log.Errorf("读取数据失败: %v", err)
			}
			break
		}

		// 判断 buffer 是否超过 10kB，防止内存耗尽攻击
		if buffer.Len() > 10*1024 {
			log.Warnf("FPV 数据超过 10kB，关闭连接: %s", conn.RemoteAddr().String())
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
			// log.Infof("[%s] 接收到FPV数据: %s", conn.RemoteAddr().String(), string(fullLine))

			if utils.IsFPVResponse(fullLine) {
				// TODO: 浏览器进行操作后这里会响应我需要通知前端
				log.Infof("[%s]FPV 响应数据: %s", conn.RemoteAddr().String(), string(fullLine))
				// 继续处理下一行
				continue
			}

			if utils.IsFPVWaringData(fullLine) {
				waringData, err := utils.ParseFPVWaringData(fullLine)
				if err != nil {
					log.Errorf("解析FPV警告数据失败: %v", err)
					continue
				}
				// log.Infof("[%s]FPV 警告数据 - 频率: %s MHz, RSSI: %s", conn.RemoteAddr().String(), waringData.Freq, waringData.RSSI)
				waringData.IP = conn.RemoteAddr().String()
				waringData.Time = time.Now().Unix()

				// 将新的警告数据添加到列表中
				s.pushFPVWaringDataList(waringData)

			}

		}

	}

}

// GetFPVWaringDataList 获取当前的 FPV 警告数据列表
func (s *fpvStore) GetFPVWaringDataList() []*utils.FPVWaringData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fpvWaringDataList
}

// pushFPVWaringDataList 添加 FPV 警告数据到列表
func (s *fpvStore) pushFPVWaringDataList(data *utils.FPVWaringData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 过滤掉 time 与 当前时间差超过 2 秒的数据
	var filteredList []*utils.FPVWaringData

	for _, item := range s.fpvWaringDataList {
		if data.Time-item.Time <= 2 {
			filteredList = append(filteredList, item)
		}
	}

	s.fpvWaringDataList = append(filteredList, data)
}
