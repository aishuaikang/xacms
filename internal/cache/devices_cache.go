package cache

import (
	"sync"
	fpv_fsm "uav_defender/internal/app/devices/fms/fpv"
	parse_fsm "uav_defender/internal/app/devices/fms/parse"
	"uav_defender/internal/dto"
	"uav_defender/internal/services"
)

type DevicesCache interface {
	SetDevices(deviceList []dto.DeviceInfo)
	GetDevices() []dto.DeviceInfo
	GetDisplayDevices() []dto.DeviceDisplayInfo
	NotifyRefresh()
	GetRefreshChan() <-chan struct{}
	GetDeviceByParseIP(parseIP string) (*dto.DeviceInfo, bool)
	GetDeviceByParseID(parseID int) (*dto.DeviceInfo, bool)
	GetDeviceByDetectionID(detectionID int) (*dto.DeviceInfo, bool)
	GetDeviceByFPVIP(fpvIP string) (*dto.DeviceInfo, bool)
	RefreshDevices() error
}

type devicesCache struct {
	devices              []dto.DeviceInfo       // 设备列表
	devicesMutex         sync.RWMutex           // 保护设备列表的读写锁
	devicesRefreshSignal chan struct{}          // 设备列表刷新信号通道
	deviceService        services.DeviceService // 设备服务
}

func NewDevicesCache(deviceService services.DeviceService) DevicesCache {
	return &devicesCache{
		devices:              []dto.DeviceInfo{},
		devicesMutex:         sync.RWMutex{},
		devicesRefreshSignal: make(chan struct{}, 1),
		deviceService:        deviceService,
	}
}

// SetDevices 设置设备列表
func (c *devicesCache) SetDevices(devices []dto.DeviceInfo) {
	c.devicesMutex.Lock()
	defer c.devicesMutex.Unlock()
	c.devices = devices
}

// GetDevices 获取设备列表
func (c *devicesCache) GetDevices() []dto.DeviceInfo {
	c.devicesMutex.RLock()
	defer c.devicesMutex.RUnlock()
	return c.devices
}

// 获取用于展示的设备列表
func (c *devicesCache) GetDisplayDevices() []dto.DeviceDisplayInfo {
	device := c.GetDevices()

	var displayDevices []dto.DeviceDisplayInfo
	for _, device := range device {
		displayDevices = append(displayDevices, dto.DeviceDisplayInfo{
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
func (c *devicesCache) GetDeviceByParseIP(parseIP string) (*dto.DeviceInfo, bool) {
	devices := c.GetDevices()
	for _, device := range devices {
		if device.ParseIP == parseIP {
			return &device, true
		}
	}
	return nil, false
}

// GetDeviceByParseID 根据解析ID获取设备信息
func (c *devicesCache) GetDeviceByParseID(parseID int) (*dto.DeviceInfo, bool) {
	devices := c.GetDevices()

	for _, device := range devices {
		if device.ParseID == parseID {
			return &device, true
		}
	}
	return nil, false
}

// GetDeviceByDetectionID 根据侦测ID获取设备信息
func (c *devicesCache) GetDeviceByDetectionID(detectionID int) (*dto.DeviceInfo, bool) {
	devices := c.GetDevices()

	for _, device := range devices {
		if device.DetectionID == detectionID {
			return &device, true
		}
	}
	return nil, false
}

// GetDeviceByFPVIP 根据FPVIP获取设备信息
func (c *devicesCache) GetDeviceByFPVIP(fpvIP string) (*dto.DeviceInfo, bool) {
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

	var deviceInfos []dto.DeviceInfo
	for _, device := range devices {
		deviceInfos = append(deviceInfos, dto.DeviceInfo{
			DeviceModel: device,
			FPVFsm:      fpv_fsm.NewFPVFsm(),
			ParseFsm:    parse_fsm.NewParseFsm(),
		})
	}

	c.SetDevices(deviceInfos)
	return nil
}
