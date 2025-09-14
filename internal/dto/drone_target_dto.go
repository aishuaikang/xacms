package dto

import (
	"xacms/internal/models"
)

// DroneTargetQueryRequest 无人机目标查询请求结构
type DroneTargetQueryRequest struct {
	BaseQueryRequest

	Model         *string               `query:"model" validate:"omitempty,max=64"`             // 无人机型号
	DetectionType *models.DetectionType `query:"detection_type" validate:"omitempty,oneof=1 2"` // 探测类型
	StartTime     *int64                `query:"start_time" validate:"omitempty,min=0"`         // 入侵时间起始 (时间戳/秒)
	EndTime       *int64                `query:"end_time" validate:"omitempty,min=0"`           // 入侵时间结束 (时间戳/秒)
}
