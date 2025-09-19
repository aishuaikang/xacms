package tasks

import (
	"context"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/global"

	"go.uber.org/zap"
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
		global.Logger.Error("刷新设备列表失败", zap.Error(err))
	}

	go func() {
		for {
			select {
			case <-t.ctx.Done():
				global.Logger.Info("DevicesTask 上下文已取消，正在退出 goroutine")
				return
			case <-t.devicesCache.GetRefreshChan():
				if err := t.devicesCache.RefreshDevices(); err != nil {
					global.Logger.Error("刷新设备列表失败", zap.Error(err))
					continue
				}
			}
		}
	}()
}
