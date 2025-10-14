package models

type ScanType uint8

const (
	RegularScan ScanType = 1 + iota // 常规扫描
	KeyScan                         // 重点扫描
	FullScan                        // 全频扫描
	FPVScan                         // FPV扫描
)

type ScanEnable uint8

const (
	ScanEnabled  ScanEnable = 1 + iota // 启动
	ScanDisabled                       // 关闭
)

// ModuleGroupDevice 模块组设备信息表
type ModuleGroupDevice struct {
	ID            uint64     `json:"id,string" gorm:"primaryKey;autoIncrement;comment:唯一ID"` // 唯一ID
	Name          string     `json:"name" gorm:"column:name;uniqueIndex;type:varchar(31);not null" comment:"设备名称"`
	Model         string     `json:"model" gorm:"column:model;type:varchar(31);default:''" comment:"设备型号"`
	LocationLng   *float64   `json:"location_lng" gorm:"column:location_lng" comment:"部署位置的经度,精确到小数点后7位,车载/单兵类设备此项为空"`
	LocationLat   *float64   `json:"location_lat" gorm:"column:location_lat" comment:"部署位置的纬度,精确到小数点后7位,车载/单兵类设备此项为空"`
	DetectionID   string     `json:"detection_id" gorm:"column:detection_id;type:varchar(31);uniqueIndex;not null" comment:"侦测模块ID"`
	DetectionIP   string     `json:"detection_ip" gorm:"column:detection_ip;type:varchar(31);uniqueIndex:idx_detection;not null" comment:"侦测模块IP"`
	DetectionPort uint16     `json:"detection_port" gorm:"column:detection_port;type:int;uniqueIndex:idx_detection;not null" comment:"侦测模块端口"`
	ParseID       string     `json:"parse_id" gorm:"column:parse_id;type:varchar(31);uniqueIndex;not null" comment:"解析模块ID"`
	ParseIP       string     `json:"parse_ip" gorm:"column:parse_ip;type:varchar(31);uniqueIndex;not null" comment:"解析模块IP"`
	GpsIP         string     `json:"gps_ip" gorm:"column:gps_ip;type:varchar(31);uniqueIndex;not null" comment:"GPS模块IP"`
	FpvIP         string     `json:"fpv_ip" gorm:"column:fpv_ip;type:varchar(31);uniqueIndex;not null" comment:"FPV模块IP"`
	VideoIP       string     `json:"video_ip" gorm:"column:video_ip;type:varchar(31);uniqueIndex;not null" comment:"视频解码板IP"`
	StrikeIP      string     `json:"strike_ip" gorm:"column:strike_ip;type:varchar(31);uniqueIndex;not null" comment:"打击模块IP"`
	ScanType      ScanType   `json:"scan_type" gorm:"column:scan_type;default:2;type:tinyint" comment:"扫描类型 1常规扫描 2重点扫描 3全频扫描 4FPV扫描"`
	ScanEnable    ScanEnable `json:"scan_enable" gorm:"column:scan_enable;default:1;type:tinyint" comment:"扫描启动标识 1是启动 2是关闭"`

	CommonModel
}

func (ModuleGroupDevice) TableName() string {
	return "module_group_devices"
}
