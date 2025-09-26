package models

type Device struct {
	ID uint `json:"id" gorm:"primaryKey;autoIncrement;comment:唯一ID"` // 唯一ID，自增主键

	Name      string   `json:"name" gorm:"uniqueIndex;size:64;not null;comment:设备名称"` // 设备名称
	Longitude *float64 `json:"longitude" gorm:"type:decimal(10,6);comment:设备经度"`      // 设备经度
	Latitude  *float64 `json:"latitude" gorm:"type:decimal(10,6);comment:设备纬度"`       // 设备纬度

	// 侦测模块
	DetectionID   uint   `json:"detection_id" gorm:"comment:侦测模块ID"`         // 侦测模块ID
	DetectionIP   string `json:"detection_ip" gorm:"size:64;comment:侦测模块IP"` // 侦测模块IP
	DetectionPort uint   `json:"detection_port" gorm:"comment:侦测模块端口"`       // 侦测模块端口

	// 解析模块
	ParseID uint   `json:"parse_id" gorm:"comment:解析模块ID"`         // 解析模块ID
	ParseIP string `json:"parse_ip" gorm:"size:64;comment:解析模块IP"` // 解析模块IP

	// FPV模块
	FPVIP  string `json:"fpv_ip" gorm:"size:64;comment:FPV模块IP"` // FPV模块IP
	RTSPIP string `json:"rtsp_ip" gorm:"size:64;comment:RTSPIP"` // RTSPIP

	// 打击模块
	StrikeIP string `json:"strike_ip" gorm:"size:64;comment:打击模块IP"` // 打击模块IP

	CommonModel
}

// TableName 设置表名
func (Device) TableName() string {
	return "devices"
}
