package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"

	"github.com/google/uuid"
	ffmpeg "github.com/u2takey/ffmpeg-go"
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
func ParseFPVWarningData(fullLine []byte, ip string, time int64, deviceId uuid.UUID) (*dto.FPVWarningData, error) {
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
		DeviceID: deviceId,
		Freq:     freq,
		RSSI:     rssi,
		Time:     time,
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
		newStream := fmt.Sprintf("%sstream_%s:", strings.Repeat(" ", 4), device.ID)
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

// ffmpegPath := "ffmpeg"
// // 启动录制
// // 构建FFmpeg命令
// RtspCmd = exec.Command(
// 	ffmpegPath,
// 	"-y",
// 	"-rtsp_transport", "tcp",
// 	"-i", rtspURL,

// 	// 视频处理参数

// 	"-c:v", "libx264",
// 	"-profile:v", "main", // 强制Main Profile
// 	"-level:v", "4.0", // 兼容性Level
// 	"-pix_fmt", "yuv420p", // 修正色彩格式（原流为yuvj420p）
// 	"-vsync", "1", // 防止帧率波动
// 	"-x264-params", "colorprim=bt709:transfer=bt709:colormatrix=bt709", // 明确色彩标准

// 	// 音频处理参数
// 	"-c:a", "aac",
// 	"-ar", "44100", // 重采样到标准频率（原流16000Hz非常规）
// 	"-ac", "2", // 强制双声道（原流为单声道）

//	// 封装参数
//	"-f", "mp4", // 不使用 -movflags +faststart
//	outputFile,
//
// )
// BuildFPVCommand 构建FPV设备命令
func BuildFPVCommand(frequency int, command int) (fpvCommand string, expectedResponse string) {
	switch command {
	// case 1, 2, 3, 4, 5:
	// 	if fpv.StartFreq <= 0 || fpv.StopFreq <= 0 {
	// 		return "", "", fmt.Errorf("invalid frequency range: start/stop must > 0")
	// 	}
	// 	comm = fmt.Sprintf("AT+FREQ_GROUP_%d=%d,%d\r\n", command, fpv.StartFreq, fpv.StopFreq)
	// 	expected = fmt.Sprintf("OK\nFREQ_GROUP_%d = %d--%d", command, fpv.StartFreq, fpv.StopFreq)
	// case 6:
	// 	if fpv.StartStep < 1 || fpv.StartStep > 100 {
	// 		return "", "", fmt.Errorf("invalid step value: 1 <= start_step <= 100")
	// 	}
	// 	comm = fmt.Sprintf("AT+FREQ_STEP=%d\r\n", fpv.StartStep)
	// 	expected = fmt.Sprintf("SET+OK\nFREQ_STEP=%d", fpv.StartStep)
	// case 7:
	// 	if fpv.DeteTime < 1 || fpv.DeteTime > 100 {
	// 		return "", "", fmt.Errorf("invalid detection time: 1 <= dete_time <= 100")
	// 	}
	// 	comm = fmt.Sprintf("AT+DETE_TIME=%d\r\n", fpv.DeteTime)
	// 	expected = fmt.Sprintf("SET+OK\nDETE_STEP=%d", fpv.DeteTime)
	// case 8:
	// 	comm = "AT+GET_FREQ?\r\n"
	case 9:
		fpvCommand = fmt.Sprintf("AT+POINT_FREQ=%d\r\n", frequency)
		expectedResponse = "SET+OK\n"
	case 10:
		fpvCommand = "AT+DEFAULT\r\n"
		expectedResponse = "SET+OK\n"
	default:
		return
	}
	return fpvCommand, strings.TrimSpace(expectedResponse)
}

// IsExpectedResponse 验证设备响应是否符合预期
func IsExpectedResponse(response, expected string) bool {
	// 验证响应
	if expected != "" && !strings.Contains(response, expected) {
		return false
	}
	return true
}

type RtspRecorder struct {
	cmd       *exec.Cmd
	lock      sync.Mutex
	stdinPipe io.WriteCloser

	videosDir string
	filename  *string
}

func NewRtspRecorder() (*RtspRecorder, error) {
	// 在工作目录下videos文件夹中保存录像
	// 如果不存在则创建
	wd, err := os.Getwd()
	if err != nil {
		global.Logger.Error("获取工作目录失败", zap.Error(err))
		return nil, fmt.Errorf("获取当前工作目录失败: %v", err)
	}

	global.Logger.Info("当前工作目录", zap.String("wd", wd))

	videosDir := path.Join(wd, "videos")
	if err := os.MkdirAll(videosDir, 0777); err != nil {
		global.Logger.Error("创建videos文件夹失败", zap.Error(err))
		return nil, fmt.Errorf("创建videos文件夹失败: %v", err)
	}

	global.Logger.Info("录像保存目录", zap.String("videosDir", videosDir))

	return &RtspRecorder{videosDir: videosDir, filename: nil}, nil
}

func (r *RtspRecorder) Start(streamKey, filename string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	if r.cmd != nil {
		return fmt.Errorf("录像已在进行")
	}

	global.Logger.Info("开始录像", zap.String("streamKey", streamKey), zap.String("filename", filename))

	r.filename = &filename

	outputFile := path.Join(r.videosDir, filename)

	global.Logger.Info("录像保存路径", zap.String("outputFile", outputFile))

	outArgs := ffmpeg.KwArgs{
		"y":           "",
		"c:v":         "libx264",
		"profile:v":   "main",
		"level:v":     "4.0",
		"pix_fmt":     "yuv420p",
		"vsync":       "1",
		"x264-params": "colorprim=bt709:transfer=bt709:colormatrix=bt709",
		"c:a":         "aac",
		"ar":          "44100",
		"ac":          "2",
		"f":           "mp4",
	}

	rtspURL, exists := config.MediaMtxAppConfig.Paths[streamKey]
	if !exists {
		return fmt.Errorf("streamKey 不存在于配置中")
	}

	global.Logger.Info("使用RTSP地址", zap.String("rtspURL", rtspURL.Source))

	ffCmd := ffmpeg.Input(rtspURL.Source, ffmpeg.KwArgs{"rtsp_transport": "tcp"}).
		Output(outputFile, outArgs).OverWriteOutput().Compile()
	r.cmd = ffCmd

	stdinPipe, err := ffCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("获取进程输入管道失败: %v", err)
	}
	r.stdinPipe = stdinPipe

	return ffCmd.Start()
}

func (r *RtspRecorder) Stop() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.cmd == nil || r.cmd.Process == nil {
		return fmt.Errorf("没有正在运行的录像进程")
	}

	// 优雅停止
	if r.stdinPipe != nil {
		_, _ = r.stdinPipe.Write([]byte("q\n"))
		r.stdinPipe.Close()
	}

	done := make(chan error, 1)
	go func() {
		done <- r.cmd.Wait()
	}()

	timeout := time.After(2 * time.Second)
	var err error
	select {
	case err = <-done:
	case <-timeout:
		if r.cmd.Process != nil {
			_ = r.cmd.Process.Kill()
			global.Logger.Warn("录像进程超时未退出，已强制kill")
		}
		err = fmt.Errorf("录像进程超时未优雅退出，已强制kill")
	}

	// 清理资源
	r.cmd = nil
	r.stdinPipe = nil

	if err != nil && r.filename != nil {
		outputFile := path.Join(r.videosDir, *r.filename)
		if removeErr := os.Remove(outputFile); removeErr != nil {
			global.Logger.Error("删除录像文件失败", zap.Error(removeErr))
		}
		r.filename = nil
	}

	if err != nil && err.Error() != "signal: killed" {
		global.Logger.Error("录像进程退出异常", zap.Error(err))
		return fmt.Errorf("录像进程退出异常: %v", err)
	}

	return nil
}
