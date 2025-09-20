package dto

import (
	"uav_defender/internal/models"
)

// DroneTargetQueryRequest 无人机目标查询请求结构
type DroneTargetQueryRequest struct {
	BaseQueryRequest

	Model         *string               `form:"model" validate:"omitempty,max=64"`             // 无人机型号
	DetectionType *models.DetectionType `form:"detection_type" validate:"omitempty,oneof=1 2"` // 探测类型
	StartTime     *int64                `form:"start_time" validate:"omitempty,min=0"`         // 入侵时间起始 (时间戳/秒)
	EndTime       *int64                `form:"end_time" validate:"omitempty,min=0"`           // 入侵时间结束 (时间戳/秒)
}

type Lang string

const (
	LangZH Lang = "zh"
	LangEN Lang = "en"
	LangRU Lang = "ru"
	LangPT Lang = "pt"
)

// DroneTargetExportRequest 无人机目标导出请求结构
type DroneTargetExportRequest struct {
	Model         *string               `form:"model" validate:"omitempty,max=64"`             // 无人机型号
	DetectionType *models.DetectionType `form:"detection_type" validate:"omitempty,oneof=1 2"` // 探测类型
	StartTime     *int64                `form:"start_time" validate:"omitempty,min=0"`         // 入侵时间起始 (时间戳/秒)
	EndTime       *int64                `form:"end_time" validate:"omitempty,min=0"`           // 入侵时间结束 (时间戳/秒)
	Lang          Lang                  `form:"lang" validate:"omitempty,oneof=zh en ru pt"`   // 语言
}
