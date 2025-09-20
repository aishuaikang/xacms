package fpv_fsm

import (
	"fmt"
	"strings"
)

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
