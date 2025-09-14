package tasks

import (
	"context"
	"xacms/internal/cache"

	"github.com/gofiber/fiber/v2/log"
)

type DevicesTask struct {
	ctx          context.Context
	devicesCache cache.DevicesCache
}

func NewDevicesTask(ctx context.Context, devicesCache cache.DevicesCache) *DevicesTask {
	return &DevicesTask{
		ctx:          ctx,
		devicesCache: devicesCache,
	}
}

func (t *DevicesTask) Execute() {
	if err := t.devicesCache.RefreshDevices(); err != nil {
		log.Errorf("刷新设备列表失败: %v", err)
	}

	go func() {
		for {
			select {
			case <-t.ctx.Done():
				log.Debug("DevicesTask 上下文已取消，正在退出 goroutine")
				return
			case <-t.devicesCache.GetRefreshChan():
				if err := t.devicesCache.RefreshDevices(); err != nil {
					log.Errorf("刷新设备列表失败: %v", err)
					continue
				}
			}
		}
	}()
}
