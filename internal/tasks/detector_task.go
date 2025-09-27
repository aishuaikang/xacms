package tasks

import (
	"context"
	"time"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/global"
)

type DetectorTask struct {
	ctx           context.Context
	commonCache   cache.CommonCache
	detectorCache cache.DetectorCache
}

func NewDetectorTask(ctx context.Context, commonCache cache.CommonCache, detectorCache cache.DetectorCache) *DetectorTask {
	return &DetectorTask{
		ctx:           ctx,
		commonCache:   commonCache,
		detectorCache: detectorCache,
	}
}

func (t *DetectorTask) Execute() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-t.ctx.Done():
				global.Logger.Info("ParseTask 上下文已取消，正在退出 goroutine")
				return
			case <-ticker.C:

				t.detectorCache.CleanupExpiredDetectorData(t.commonCache.GetTTL()) // 假设TTL为30秒
			}
		}
	}()
}
