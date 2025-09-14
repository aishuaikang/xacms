package cache

import (
	"sync"
	"xacms/internal/dto"
	"xacms/internal/services"
)

type DevicesCache interface {
	SetDevices(deviceList []dto.DeviceInfo)
	GetDevices() []dto.DeviceInfo
	NotifyRefresh()
	GetRefreshChan() <-chan struct{}
	GetDeviceByParseIP(parseIP string) (*dto.DeviceInfo, bool)
	GetDeviceByParseID(parseID int) (*dto.DeviceInfo, bool)
	GetDeviceByDetectionID(detectionID int) (*dto.DeviceInfo, bool)
	GetDeviceByFPVIP(fpvIP string) (*dto.DeviceInfo, bool)
	RefreshDevices() error
}

type devicesCache struct {
	devices      []dto.DeviceInfo
	devicesMutex sync.RWMutex

	devicesRefreshSignal chan struct{}
	deviceService        services.DeviceService
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

// GetDeviceByParseID 根据解析ID获取设备侦测ID
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

// 将设备信息初始化到缓存中
func (c *devicesCache) RefreshDevices() error {
	devices, err := c.deviceService.GetAllDevices()
	if err != nil {
		return err
	}

	var deviceInfos []dto.DeviceInfo
	for _, device := range devices {
		deviceInfos = append(deviceInfos, dto.DeviceInfo{
			DeviceModel:    device,
			HeartbeatCount: 0,
			Expires:        0,
			Status:         dto.DeviceInfoStatusOffline,
			StrikeInfo: dto.StrikeInfo{
				Mode:      dto.StrikeModeIdle,
				Status:    dto.StrikeStatusNotStriked,
				Frequency: nil,
			},
		})
	}

	c.SetDevices(deviceInfos)
	return nil
}
