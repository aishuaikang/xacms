package cache

import (
	"sync"
	"uav_defender/internal/app/devices/conn"
	fpv_fsm "uav_defender/internal/app/devices/fms/fpv"
	parse_fsm "uav_defender/internal/app/devices/fms/parse"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/utils"
	"uav_defender/internal/services"

	"github.com/google/uuid"
)

// DeviceInfo 设备信息
type DeviceInfo struct {
	models.DeviceModel
	FPVFsm   *fpv_fsm.FPVFsm
	ParseFsm *parse_fsm.ParseFsm
}

// DeviceDisplayInfo 设备展示信息
type DeviceDisplayInfo struct {
	models.DeviceModel
	FPVState   fpv_fsm.FPVState     `json:"fpv_state"`   // FPV状态
	ParseState parse_fsm.ParseState `json:"parse_state"` // 解析状态
}

type DevicesCache interface {
	SetDevices(deviceList []DeviceInfo)
	GetDevices() []DeviceInfo
	GetDisplayDevices() []DeviceDisplayInfo
	NotifyRefresh()
	GetRefreshChan() <-chan struct{}
	GetDeviceByParseIP(parseIP string) (*DeviceInfo, bool)
	// GetDeviceByParseID(parseID int) (*DeviceInfo, bool)
	// GetDeviceByDetectionID(detectionID int) (*DeviceInfo, bool)
	GetDeviceByFPVIP(fpvIP string) (*DeviceInfo, bool)
	GetDeviceByID(id uuid.UUID) (*DeviceInfo, bool)
	RefreshDevices() error
}

type devicesCache struct {
	devices              []DeviceInfo           // 设备列表
	devicesMutex         sync.RWMutex           // 保护设备列表的读写锁
	devicesRefreshSignal chan struct{}          // 设备列表刷新信号通道
	deviceService        services.DeviceService // 设备服务
	fpvConnection        *conn.FpvConnection
}

func NewDevicesCache(deviceService services.DeviceService, fpvConnection *conn.FpvConnection) DevicesCache {
	return &devicesCache{
		devices:              []DeviceInfo{},
		devicesMutex:         sync.RWMutex{},
		devicesRefreshSignal: make(chan struct{}, 1),
		deviceService:        deviceService,
		fpvConnection:        fpvConnection,
	}
}

// SetDevices 设置设备列表
func (c *devicesCache) SetDevices(devices []DeviceInfo) {
	c.devicesMutex.Lock()
	defer c.devicesMutex.Unlock()
	c.devices = devices
}

// GetDevices 获取设备列表
func (c *devicesCache) GetDevices() []DeviceInfo {
	c.devicesMutex.RLock()
	defer c.devicesMutex.RUnlock()
	return c.devices
}

// 获取用于展示的设备列表
func (c *devicesCache) GetDisplayDevices() []DeviceDisplayInfo {
	device := c.GetDevices()

	var displayDevices []DeviceDisplayInfo
	for _, device := range device {
		displayDevices = append(displayDevices, DeviceDisplayInfo{
			DeviceModel: device.DeviceModel,
			FPVState:    fpv_fsm.FPVState(device.FPVFsm.Current()),
			ParseState:  parse_fsm.ParseState(device.ParseFsm.Current()),
		})
	}
	return displayDevices
}

// NotifyRefresh 触发设备列表刷新
func (c *devicesCache) NotifyRefresh() {
	c.devicesRefreshSignal <- struct{}{}
}

// GetRefreshChan 获取设备列表刷新通道
func (c *devicesCache) GetRefreshChan() <-chan struct{} {
	return c.devicesRefreshSignal
}

// GetDeviceByParseIP 根据解析IP获取设备
func (c *devicesCache) GetDeviceByParseIP(parseIP string) (*DeviceInfo, bool) {
	devices := c.GetDevices()
	for _, device := range devices {
		if device.ParseIP == parseIP {
			return &device, true
		}
	}
	return nil, false
}

// // GetDeviceByParseID 根据解析ID获取设备信息
// func (c *devicesCache) GetDeviceByParseID(parseID int) (*DeviceInfo, bool) {
// 	devices := c.GetDevices()

// 	for _, device := range devices {
// 		if device.ParseID == parseID {
// 			return &device, true
// 		}
// 	}
// 	return nil, false
// }

// // GetDeviceByDetectionID 根据侦测ID获取设备信息
// func (c *devicesCache) GetDeviceByDetectionID(detectionID int) (*DeviceInfo, bool) {
// 	devices := c.GetDevices()

// 	for _, device := range devices {
// 		if device.DetectionID == detectionID {
// 			return &device, true
// 		}
// 	}
// 	return nil, false
// }

// GetDeviceByFPVIP 根据FPVIP获取设备信息
func (c *devicesCache) GetDeviceByFPVIP(fpvIP string) (*DeviceInfo, bool) {
	devices := c.GetDevices()
	for _, device := range devices {
		if device.FPVIP == fpvIP {
			return &device, true
		}
	}
	return nil, false
}

// RefreshDevices 刷新设备信息
func (c *devicesCache) RefreshDevices() error {
	devices, err := c.deviceService.GetAllDevices()
	if err != nil {
		return err
	}

	var deviceInfos []DeviceInfo
	for _, device := range devices {
		deviceInfos = append(deviceInfos, DeviceInfo{
			DeviceModel: device,
			FPVFsm:      fpv_fsm.NewFPVFsm(c.fpvConnection),
			ParseFsm:    parse_fsm.NewParseFsm(),
		})
	}

	c.SetDevices(deviceInfos)
	return nil
}

// GetDeviceByID 根据设备ID获取设备信息
func (c *devicesCache) GetDeviceByID(id uuid.UUID) (*DeviceInfo, bool) {
	devices := c.GetDevices()
	for _, device := range devices {
		if utils.EqualUUID(&device.ID, &id) {
			return &device, true
		}

	}
	return nil, false
}
