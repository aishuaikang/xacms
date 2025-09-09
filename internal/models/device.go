package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceModel struct {
	ID uuid.UUID `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"` // 唯一ID

	Name      string  `json:"name" gorm:"uniqueIndex;size:64;not null;comment:设备名称"` // 设备名称
	Longitude float64 `json:"longitude" gorm:"type:decimal(10,6);comment:设备经度"`      // 设备经度
	Latitude  float64 `json:"latitude" gorm:"type:decimal(10,6);comment:设备纬度"`       // 设备纬度

	// 侦测模块
	DetectionID   int    `json:"detection_id" gorm:"uniqueIndex;type:char(36);comment:侦测模块ID"` // 侦测模块ID
	DetectionIP   string `json:"detection_ip" gorm:"size:64;comment:侦测模块IP"`                   // 侦测模块IP
	DetectionPort int    `json:"detection_port" gorm:"comment:侦测模块端口"`                         // 侦测模块端口

	// 解析模块
	ParseID int    `json:"parse_id" gorm:"uniqueIndex;type:char(36);comment:解析模块ID"` // 解析模块ID
	ParseIP string `json:"parse_ip" gorm:"size:64;comment:解析模块IP"`                   // 解析模块IP

	// FPV模块
	FPVIP          string `json:"fpv_ip" gorm:"size:64;comment:FPV模块IP"`            // FPV模块IP
	StreamServerIP string `json:"stream_server_ip" gorm:"size:64;comment:流媒体服务器IP"` // 流媒体服务器IP

	// 打击模块
	StrikeIP   string `json:"strike_ip" gorm:"size:64;comment:打击模块IP"` // 打击模块IP
	StrikePort int    `json:"strike_port" gorm:"comment:打击模块端口"`       // 打击模块端口

	CommonModel
}

// TableName 设置表名
func (DeviceModel) TableName() string {
	return "devices"
}

// BeforeCreate GORM钩子，在创建记录之前调用
func (d *DeviceModel) BeforeCreate(tx *gorm.DB) (err error) {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	// 如果 Status 是零值（StatusDisabled = 0），但我们想要明确设置它
	// 可以根据业务需求决定是否需要默认值
	return
}
