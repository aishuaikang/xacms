package tasks

import (
	"context"
	"time"
	"uav_defender/internal/cache"
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
				return
			case <-ticker.C:

				t.detectorCache.CleanupExpiredDetectorData(t.commonCache.GetTTL()) // 假设TTL为30秒
			}
		}
	}()
}
