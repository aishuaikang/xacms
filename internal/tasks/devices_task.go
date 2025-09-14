package tasks

import (
	"context"
	"xacms/internal/cache"
	"xacms/internal/services"

	"github.com/gofiber/fiber/v2/log"
)

type DevicesTask struct {
	ctx           context.Context
	devicesCache  cache.DevicesCache
	deviceService services.DeviceService
}

func NewDevicesTask(ctx context.Context, devicesCache cache.DevicesCache, deviceService services.DeviceService) *DevicesTask {
	return &DevicesTask{
		ctx:           ctx,
		devicesCache:  devicesCache,
		deviceService: deviceService,
	}
}

func (t *DevicesTask) Execute() {
	// 初始加载设备列表
	deviceList, err := t.deviceService.InitDevices()
	if err != nil {
		log.Errorf("初始化设备列表失败: %v", err)
	}
	t.devicesCache.SetDevices(deviceList)
	go func() {
		for {
			select {
			case <-t.ctx.Done():
				log.Info("设备存储停止刷新")
				return
			case <-t.devicesCache.GetRefreshChan():
				deviceList, err := t.deviceService.InitDevices()
				if err != nil {
					log.Errorf("刷新设备列表失败: %v", err)
					continue
				}
				t.devicesCache.SetDevices(deviceList)
			}
		}
	}()
}
