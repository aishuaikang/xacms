package tasks

import (
	"context"
	"time"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"go.uber.org/zap"
)

type DroneTargetTask struct {
	ctx                context.Context
	droneTargetCache   cache.DroneTargetCache
	droneTargetService services.DroneTargetService
}

func NewDroneTargetTask(ctx context.Context, droneTargetCache cache.DroneTargetCache, droneTargetService services.DroneTargetService) *DroneTargetTask {
	return &DroneTargetTask{
		ctx:                ctx,
		droneTargetCache:   droneTargetCache,
		droneTargetService: droneTargetService,
	}
}

func (t *DroneTargetTask) Execute() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-t.ctx.Done():
				global.Logger.Info("DroneTargetTask 上下文已取消，正在退出 goroutine")
				return
			case <-ticker.C:
				staleTargets := t.droneTargetCache.MarkStaleTargetsAsVanishedAndRemove(10)
				if len(staleTargets) <= 0 {
					continue
				}

				global.Logger.Debug("找到 %d 个超过2分钟未更新且未消失的无人机目标", zap.Int("count", len(staleTargets)))
				for _, target := range staleTargets {
					if _, err := t.droneTargetService.CreateDroneTarget(target); err != nil {
						global.Logger.Error("存储无人机目标失败", zap.String("serial", target.Serial), zap.Error(err))
						continue
					}
				}

			}
		}
	}()
}
