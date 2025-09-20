package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"

	"go.uber.org/zap"
)

// IsFPVResponse 判断 fullLine 是否是响应
func IsFPVResponse(fullLine []byte) bool {
	return bytes.HasPrefix(fullLine, []byte("AT+OK")) ||
		bytes.HasPrefix(fullLine, []byte("SET+OK")) ||
		bytes.HasPrefix(fullLine, []byte("OK")) ||
		bytes.HasPrefix(fullLine, []byte("AT+DEFAULT OK"))
}

// IsFPVWaringData 判断是否是 FPV 警告数据
func IsFPVWaringData(fullLine []byte) bool {
	return bytes.HasPrefix(fullLine, []byte("Waring,Freq")) && bytes.Contains(fullLine, []byte("RSSI"))
}

// ParseFPVWarningData 解析 FPV 警告数据
func ParseFPVWarningData(fullLine []byte, ip string, time int64, DetectionID int) (*dto.FPVWarningData, error) {
	// 解析格式: "Waring,Freq 5025,RSSI 0.60"
	parts := bytes.Split(fullLine, []byte(","))
	if len(parts) != 3 {
		return nil, fmt.Errorf("无效的FPV警告数据格式")
	}

	// 解析频率部分
	freqPart := bytes.TrimSpace(parts[1])
	if !bytes.HasPrefix(freqPart, []byte("Freq ")) {
		return nil, fmt.Errorf("频率数据格式错误")
	}
	freq := string(bytes.TrimSpace(freqPart[5:])) // 跳过 "Freq "

	// 解析RSSI部分
	rssiPart := bytes.TrimSpace(parts[2])
	if !bytes.HasPrefix(rssiPart, []byte("RSSI ")) {
		return nil, fmt.Errorf("RSSI数据格式错误")
	}
	rssi := string(bytes.TrimSpace(rssiPart[5:])) // 跳过 "RSSI "

	return &dto.FPVWarningData{
		DetectionID: DetectionID,
		Freq:        freq,
		RSSI:        rssi,
		IP:          ip,
		Time:        time,
	}, nil
}

// UpdateMediaMtxConfigPaths 以文本流方式插入 stream_X 配置
func UpdateMediaMtxConfigPaths(devices []models.DeviceModel) {
	wd, _ := os.Getwd()
	filePath := path.Join(wd, "config", "mediamtx.yml")

	inputFile, err := os.Open(filePath)
	if err != nil {
		global.Logger.Error("打开配置文件失败", zap.Error(err))
		return
	}
	defer inputFile.Close()
	scanner := bufio.NewScanner(inputFile)
	var lines []string
	var afterPaths []string
	var pathsIndex int = -1
	streamIndex := 0

	// 扫描所有行，记录第一个 paths: 的索引，跳过后续所有 paths:
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "paths:") {
			if pathsIndex == -1 {
				pathsIndex = len(lines)
				lines = append(lines, line)
			}
			// 跳过多余的 paths:
			continue
		}
		// paths: 之后的内容，遇到下一个顶级节点（无缩进）则开始收集 afterPaths
		if pathsIndex != -1 && (len(line) > 0 && line[0] != ' ' && !strings.HasPrefix(strings.TrimSpace(line), "paths:")) {
			afterPaths = append(afterPaths, line)
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		global.Logger.Error("扫描文件失败", zap.Error(err))
		return
	}

	// 构造新的 stream_X 配置
	var streamLines []string
	for _, device := range devices {
		streamIndex++
		newStream := fmt.Sprintf("%sstream_%02d:", strings.Repeat(" ", 4), device.DetectionID)
		newSource := fmt.Sprintf("%ssource: rtsp://%s:554/live/1_1", strings.Repeat(" ", 6), device.RTSPIP)
		streamLines = append(streamLines, newStream, newSource)
	}

	// 组装最终内容
	var finalLines []string
	if pathsIndex != -1 {
		// 只保留第一个 paths: 节点，插入 streamXX
		finalLines = append(finalLines, lines[:pathsIndex+1]...)
		finalLines = append(finalLines, streamLines...)
		finalLines = append(finalLines, afterPaths...)
	} else {
		// 没有 paths: 节点则追加
		finalLines = append(lines, "paths:")
		finalLines = append(finalLines, streamLines...)
		finalLines = append(finalLines, afterPaths...)
	}

	// 写回文件
	outputFile, err := os.Create(filePath)
	if err != nil {
		global.Logger.Error("创建配置文件失败", zap.Error(err))
		return
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range finalLines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			global.Logger.Error("写入配置失败", zap.Error(err))
			return
		}
	}
	writer.Flush()
}

// // RemoveMediaMtxConfigPath 删除 MediaMtx 配置中的路径
// func RemoveMediaMtxConfigPath(detectionID int) {
// 	viper.SetConfigName("mediamtx")
// 	paths := viper.GetStringMap("paths")
// 	key := fmt.Sprintf("stream_%d", detectionID)
// 	if _, exists := paths[key]; exists {
// 		delete(paths, key)
// 		global.Logger.Info("已删除 MediaMtx 配置中的路径:", zap.Int("detectionID", detectionID))
// 	} else {
// 		global.Logger.Warn("MediaMtx 配置中不存在该路径, 无需删除:", zap.Int("detectionID", detectionID))
// 	}
// 	viper.Set("paths", paths)

// 	global.Logger.Info("当前 MediaMtx 路径配置:", zap.Any("paths", paths))
// 	// 写回文件
// 	if err := viper.WriteConfig(); err != nil {
// 		global.Logger.Error("写回 MediaMtx 配置文件失败", zap.Error(err))
// 	}
// }
