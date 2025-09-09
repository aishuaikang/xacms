package utils

import (
	"bytes"
	"fmt"
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

type FPVWaringData struct {
	Freq string `json:"freq"`
	RSSI string `json:"rssi"`
	IP   string `json:"ip"`
	Time int64  `json:"time"`
}

// ParseFPVWaringData 解析 FPV 警告数据
func ParseFPVWaringData(fullLine []byte) (*FPVWaringData, error) {
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

	return &FPVWaringData{
		Freq: freq,
		RSSI: rssi,
	}, nil
}
