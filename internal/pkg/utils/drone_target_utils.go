package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"

	"github.com/bytedance/sonic"
)

func GenerateDroneTargetTableCSVByLang(droneTargets []models.DroneTargetModel, lang dto.Lang) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	// Write CSV header
	if err := writer.Write(GetDroneTargetTableHeadersByLang(lang)); err != nil {
		return nil, err
	}

	// Write CSV rows
	for _, target := range droneTargets {
		trajectory, err := sonic.Marshal(target.Trajectories)
		if err != nil {
			return nil, err
		}

		row := []string{
			target.Model,
			ConvertDetectionType(target.DetectionType, lang),
			fmt.Sprintf("%.0f", target.Frequency), // 不保留小数，例如：
			target.Serial,
			target.CreatedAt.Time().Format(time.DateTime),
			fmt.Sprintf("%d", target.Device),
			fmt.Sprintf("%.2f", target.Height),
			fmt.Sprintf("%.2f", target.Distance),
			fmt.Sprintf("%.2f", target.DroneLng),
			fmt.Sprintf("%.2f", target.DroneLat),
			string(trajectory),
		}

		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buffer.Bytes(), nil
}

func GetDroneTargetTableHeadersByLang(lang dto.Lang) []string {
	switch lang {
	case dto.LangZH:
		return []string{
			"型号",
			"探测类型",
			"频点",
			"ID",
			"入侵时间",
			"传感器",
			"高度",
			"距离",
			"经度",
			"纬度",
			"轨迹",
		}
	case dto.LangEN:
		return []string{
			"Model",          // 型号
			"Detection Type", // 探测类型
			"Frequency",      // 频点
			"ID",             // ID
			"Intrusion Time", // 入侵时间
			"Sensor",         // 传感器
			"Altitude",       // 高度
			"Distance",       // 距离
			"Longitude",      // 经度
			"Latitude",       // 纬度
			"Trajectory",     // 轨迹
		}
	case dto.LangRU:
		return []string{
			"Модель",          // 型号
			"Тип обнаружения", // 探测类型
			"Частота",         // 频点
			"ID",              // ID
			"Время вторжения", // 入侵时间
			"Датчик",          // 传感器
			"Высота",          // 高度
			"Дистанция",       // 距离
			"Долгота",         // 经度
			"Широта",          // 纬度
			"Траектория",      // 轨迹
		}
	case dto.LangPT:
		return []string{
			"Modelo",           // 型号
			"Tipo de Detecção", // 探测类型
			"Frequência",       // 频点
			"ID",               // ID
			"Hora da Intrusão", // 入侵时间
			"Sensor",           // 传感器
			"Altitude",         // 高度
			"Distância",        // 距离
			"Longitude",        // 经度
			"Latitude",         // 纬度
			"Trajetória",       // 轨迹
		}
	default:
		return []string{
			"型号",
			"探测类型",
			"频点",
			"ID",
			"入侵时间",
			"传感器",
			"高度",
			"距离",
			"经度",
			"纬度",
			"轨迹",
		}
	}
}

// 根据语言获取侦测类型枚举对应的字符串
func ConvertDetectionType(detectionType models.DetectionType, lang dto.Lang) string {
	switch lang {
	case dto.LangZH:
		switch detectionType {
		case models.DetectionTypeDetection:
			return "探测"
		case models.DetectionTypeParse:
			return "定位"
		}
	case dto.LangEN:
		switch detectionType {
		case models.DetectionTypeDetection:
			return "Drone Detect"
		case models.DetectionTypeParse:
			return "Drone Locate"
		}
	case dto.LangRU:
		switch detectionType {
		case models.DetectionTypeDetection:
			return "Обнаружение"
		case models.DetectionTypeParse:
			return "Локализация"
		}
	case dto.LangPT:
		switch detectionType {
		case models.DetectionTypeDetection:
			return "Detecção"
		case models.DetectionTypeParse:
			return "Localização"
		}
	default:
		switch detectionType {
		case models.DetectionTypeDetection:
			return "探测"
		case models.DetectionTypeParse:
			return "定位"
		}
	}
	return "Unknown"
}
