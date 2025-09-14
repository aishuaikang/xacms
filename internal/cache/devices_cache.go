package cache

import (
	"sync"
	"xacms/internal/models"
)

type DevicesCache interface {
	SetDevices(deviceList []models.DeviceModel)
	GetDevices() []models.DeviceModel
	NotifyRefresh()
	GetRefreshChan() <-chan struct{}
	GetDeviceByParseIP(parseIP string) (*models.DeviceModel, bool)
	GetDeviceByParseID(parseID int) (*models.DeviceModel, bool)
	GetDeviceByDetectionID(detectionID int) (*models.DeviceModel, bool)
	GetDeviceByFPVIP(fpvIP string) (*models.DeviceModel, bool)
}

type devicesCache struct {
	devices      []models.DeviceModel
	devicesMutex sync.RWMutex

	devicesRefreshSignal chan struct{}
}

func NewDevicesCache() DevicesCache {
	return &devicesCache{
		devices:              []models.DeviceModel{},
		devicesMutex:         sync.RWMutex{},
		devicesRefreshSignal: make(chan struct{}, 1),
	}
}

// SetDevices 设置设备列表
func (c *devicesCache) SetDevices(devices []models.DeviceModel) {
	c.devicesMutex.Lock()
	defer c.devicesMutex.Unlock()
	c.devices = devices
}

// GetDevices 获取设备列表
func (c *devicesCache) GetDevices() []models.DeviceModel {
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
func (c *devicesCache) GetDeviceByParseIP(parseIP string) (*models.DeviceModel, bool) {
	devices := c.GetDevices()
	for _, device := range devices {
		if device.ParseIP == parseIP {
			return &device, true
		}
	}
	return nil, false
}

// GetDeviceByParseID 根据解析ID获取设备侦测ID
func (c *devicesCache) GetDeviceByParseID(parseID int) (*models.DeviceModel, bool) {
	devices := c.GetDevices()

	for _, device := range devices {
		if device.ParseID == parseID {
			return &device, true
		}
	}
	return nil, false
}

// GetDeviceByDetectionID 根据侦测ID获取设备信息
func (c *devicesCache) GetDeviceByDetectionID(detectionID int) (*models.DeviceModel, bool) {
	devices := c.GetDevices()

	for _, device := range devices {
		if device.DetectionID == detectionID {
			return &device, true
		}
	}
	return nil, false
}

// GetDeviceByFPVIP 根据FPVIP获取设备信息
func (c *devicesCache) GetDeviceByFPVIP(fpvIP string) (*models.DeviceModel, bool) {
	devices := c.GetDevices()
	for _, device := range devices {
		if device.FPVIP == fpvIP {
			return &device, true
		}
	}
	return nil, false
}
