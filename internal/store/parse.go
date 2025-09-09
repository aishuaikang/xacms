package store

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"xacms/internal/pkg/config"
	"xacms/internal/utils"

	"github.com/gofiber/fiber/v2/log"
)

type ParseStore interface {
	GetParseDataList() []*utils.ParseData
}

type parseStore struct {
	ctx           context.Context
	config        *config.Config
	commonStore   CommonStore
	mu            sync.RWMutex
	parseDataList []*utils.ParseData
}

func NewParseStore(ctx context.Context, config *config.Config, commonStore CommonStore) ParseStore {
	parseStore := &parseStore{
		ctx:         ctx,
		config:      config,
		commonStore: commonStore,
		mu:          sync.RWMutex{},
	}

	go parseStore.startParseServer()

	return parseStore
}

func (s *parseStore) startParseServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Configuration.ParsePort))
	if err != nil {
		log.Fatalf("无法启动parse tcp服务器: %v", err)
	}
	defer listener.Close()
	log.Infof("parse tcp服务器已启动，监听端口: %d", s.config.Configuration.ParsePort)

	// 用一个 goroutine 监听 ctx.Done()，在取消时关闭 listener
	go func() {
		<-s.ctx.Done()
		log.Info("关闭 parse tcp 服务器...")
		listener.Close() // 会导致 Accept 返回错误，从而退出主循环
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				log.Info("parse tcp 服务器已关闭")
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
func (s *parseStore) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Infof("新的Parse连接来自: %s", conn.RemoteAddr().String())

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

		buffer.Write(line)

		for {
			index := bytes.Index(buffer.Bytes(), []byte("\r\n"))
			if index == -1 {
				// 没有找到完整的行，继续读取
				break
			}

			// 提取完整的行
			fullLine := bytes.TrimSpace(buffer.Next(index + 2)) // 包括 \r\n
			// log.Infof("[%s] 接收到Parse数据: %s", conn.RemoteAddr().String(), string(fullLine))

			var parseData utils.ParseData

			isHasSerial := false

			if utils.IsRID(fullLine) {
				if err := utils.ParseRID(fullLine, &parseData); err != nil {
					log.Errorf("解析 RID 数据失败: %v", err)
					continue
				}

				ip := strings.Split(conn.RemoteAddr().String(), ":")[0]
				parseId, ok := s.commonStore.GetDeviceParseIDByParseIP(ip)
				if !ok {
					log.Warnf("未找到匹配的设备，忽略该报文: %s, IP: %s", string(fullLine), ip)
					continue
				}

				parseData.Device = strconv.Itoa(parseId)

				isHasSerial = true
			} else if utils.IsEncryption(fullLine) {
				if err := utils.ParseEncryption(fullLine, &parseData, s.commonStore.GetDecryptToken(), &isHasSerial); err != nil {
					log.Errorf("解析 Encryption 数据失败: %v", err)
					continue
				}

			} else if utils.IsDID(fullLine) {
				if err := utils.ParseDID(fullLine, &parseData); err != nil {
					log.Errorf("解析 DID 数据失败: %v", err)
					continue
				}

				if parseData.Serial != "" {
					isHasSerial = true
				}

			}

			// 这里进行报文内容校验，确保数据 hasSerial 是否存在Serial字段
			if !isHasSerial {
				log.Warnf("报文内容无效，缺少 Serial 字段，忽略该报文: %s", string(fullLine))
				continue
			}

			// gps 与解析出来的提供的都是 wgs84

			// 验证了这条告警是不是完整的
			if parseData.Serial != "" || parseData.Model != "" {
				// 根据targeId 也就是设备的serial 序列号进行判断，这里能保证 model.DroneTarget
				// 表里面的target设备唯一
				// if err := service.AddDroneFromParseAlert(parseData); err != nil {
				// 	global.Logger.Errorf("添加警报信息失败: %v", err)
				// 	return nil
				// }

				// TODO: 这里可以把数据存储到数据库或者发送到消息队列
			}

			// TODO: 根据设置的map类型设置进行坐标转换

			if parseData.Model == "" {
				continue
			}

			s.updateParseDataList(&parseData)

			// data, _ := sonic.Marshal(parseData)
			// log.Infof("解析结果: %s", data)
		}
	}
}

func (s *parseStore) updateParseDataList(data *utils.ParseData) {
	merged := false

	for i, item := range s.GetParseDataList() {
		if item.Serial == data.Serial && item.Device == data.Device {
			// 如果序列号相同，则合并数据
			mergedData, err := utils.MergeParseData(*item, *data)
			if err != nil {
				log.Errorf("合并定位数据失败: %v", err)
				return
			}

			// 1. 通过设备编号获取设备注册信息
			parseId, err := strconv.Atoi(item.Device) // 直接转换，不需要循环
			if err != nil {
				log.Info("未找到定位设备")
				return
			}
			detectionID, ok := s.commonStore.GetDeviceDetectionIDByParseID(parseId)
			if !ok {
				log.Info("未找到定位设备对应的侦测设备")
				return
			}

			device, ok := s.commonStore.GetDeviceByDetectionID(detectionID)
			if !ok {
				log.Info("未找到定位设备对应的侦测设备信息")
				return
			}

			if device.Latitude != 0 && device.Longitude != 0 {
				distance := utils.Haversine(
					device.Latitude,               // 设备纬度
					device.Longitude,              // 设备经度
					mergedData.DroneGPS.Latitude,  // 无人机纬度
					mergedData.DroneGPS.Longitude, // 无人机经度
				)

				if device.Latitude != 0 && device.Longitude != 0 && mergedData.DroneGPS.Longitude > 0.1 {
					azimuth := utils.CalculateBearing(
						device.Latitude,               // 设备纬度
						device.Longitude,              // 设备经度
						mergedData.DroneGPS.Latitude,  // 无人机纬度
						mergedData.DroneGPS.Longitude, // 无人机经度
					)

					// 4. 更新 LdResult（无需方位角）
					mergedData.LdResult = utils.LdResult{
						SensorId:  fmt.Sprintf("%d", detectionID), // 直接使用字符串类型ID
						Distance:  distance,
						Azimuth:   azimuth,
						DeviceLon: device.Longitude,
						DeviceLat: device.Latitude,
					}
				} else {
					// 如果未计算出有效的方位角
					mergedData.LdResult = utils.LdResult{
						SensorId: fmt.Sprintf("%d", detectionID), // 直接使用字符串类型ID
						Distance: 0,
						Azimuth:  499, // 明确标注未计算方位角
					}
				}

			} else {

				mergedData.LdResult = utils.LdResult{
					SensorId: fmt.Sprintf("%d", detectionID), // 直接使用字符串类型ID
					Distance: 0,
					Azimuth:  500, // 明确标注未计算方位角
				}

			}

			s.updateParseData(i, &mergedData)

			merged = true
			break
		}
	}

	if !merged && data.Serial != "" && data.Model != "" {
		s.pushParseData(data)
		log.Infof("添加新定位数据: %s", data.Serial)
	}

	// 只对非DJI-Drone模型排序
	// if data.Model != "DJI-Drone" && len(s.parseDataList) >= 2 {
	// 	sort.Slice(s.parseDataList, func(i, j int) bool {
	// 		return s.parseDataList[i].Expires < s.parseDataList[j].Expires
	// 	})
	// }
}

func (s *parseStore) GetParseDataList() []*utils.ParseData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.parseDataList
}

func (s *parseStore) pushParseData(data *utils.ParseData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.parseDataList = append(s.parseDataList, data)
}

func (s *parseStore) updateParseData(index int, data *utils.ParseData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.parseDataList[index] = data
}
