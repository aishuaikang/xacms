package tasks

import (
	"context"
	"time"
	"uav_defender/internal/cache"
	"uav_defender/internal/services"

	"github.com/gofiber/fiber/v2/log"
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
				log.Debug("DroneTargetTask 上下文已取消，正在退出 goroutine")
				return
			case <-ticker.C:
				staleTargets := t.droneTargetCache.MarkStaleTargetsAsVanishedAndRemove(10)
				if len(staleTargets) <= 0 {
					continue
				}

				log.Infof("找到 %d 个超过2分钟未更新且未消失的无人机目标", len(staleTargets))
				for _, target := range staleTargets {
					if _, err := t.droneTargetService.CreateDroneTarget(target); err != nil {
						log.Errorf("存储无人机目标 %s 失败: %v", target.Serial, err)
						continue
					}
				}

			}
		}
	}()
}
