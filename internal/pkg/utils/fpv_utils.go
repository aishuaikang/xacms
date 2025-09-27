package utils

import (
	"bufio"
	"bytes"
	"context"
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
func ParseFPVWarningData(fullLine []byte, ip string, time int64, deviceId uint) (*dto.FPVWarningData, error) {
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
func UpdateMediaMtxConfigPaths(devices []models.Device) {
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
		newStream := fmt.Sprintf("%sstream_%d:", strings.Repeat(" ", 4), device.ID)
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

func NewRtspRecorder() *RtspRecorder {
	return &RtspRecorder{videosDir: GetVideosDir(), filename: nil}
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

	// 验证RTSP地址是否可以正常拉流
	if err := validateRtspStream(rtspURL.Source); err != nil {
		return fmt.Errorf("RTSP流验证失败: %v", err)
	}

	global.Logger.Info("RTSP流验证通过，开始录制")

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

// GetVideosDir 获取 videos 目录的绝对路径，确保目录存在
func GetVideosDir() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic("无法获取当前工作目录: " + err.Error())
	}

	// 确保 videos 目录存在
	videosDir := path.Join(workingDir, "static", "videos")
	if err := os.MkdirAll(videosDir, os.ModePerm); err != nil {
		panic("创建 videos 目录失败: " + err.Error())
	}
	return videosDir
}

// validateRtspStream 验证RTSP流是否可以正常连接和拉流
func validateRtspStream(rtspURL string) error {
	global.Logger.Info("验证RTSP流", zap.String("url", rtspURL))

	// 使用context控制超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用ffprobe验证流是否可访问，只读取前几帧来快速验证
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-rtsp_transport", "tcp",
		"-analyzeduration", "3000000", // 3秒分析时间
		"-probesize", "1000000", // 1MB探测大小
		"-show_entries", "stream=codec_type",
		"-of", "csv=p=0",
		rtspURL)

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("RTSP流验证超时: %s", rtspURL)
		}
		return fmt.Errorf("RTSP流不可访问: %v", err)
	}

	// 检查是否包含视频流
	outputStr := string(output)
	if !strings.Contains(outputStr, "video") {
		return fmt.Errorf("RTSP流中未发现视频轨道")
	}

	global.Logger.Info("RTSP流验证成功", zap.String("url", rtspURL))
	return nil
}
