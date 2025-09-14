package tasks

import (
	"context"
	"time"
	"xacms/internal/cache"
	"xacms/internal/services"

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
				log.Info("无人机目标存储停止刷新")
				return
			case <-ticker.C:
				log.Debug("检查无人机目标缓存，标记并存储超过2分钟未更新的目标")
				// 标记超过2分钟未更新的无人机目标为已消失并移除
				staleTargets := t.droneTargetCache.MarkStaleTargetsAsVanishedAndRemove(10)
				if len(staleTargets) <= 0 {
					continue
				}

				log.Infof("找到 %d 个超过2分钟未更新且未消失的无人机目标", len(staleTargets))
				for _, target := range staleTargets {
					// 打印消失时间、创建时间和更新时间字段
					log.Infof("存储无人机目标 %s，消失时间: %v, 创建时间: %v, 更新时间: %v", target.Serial, target.VanishTime, target.CreatedAt, target.UpdatedAt)
					if _, err := t.droneTargetService.CreateDroneTarget(target); err != nil {
						log.Errorf("存储无人机目标 %s 失败: %v", target.Serial, err)
						continue
					}
				}

			}
		}
	}()
}
