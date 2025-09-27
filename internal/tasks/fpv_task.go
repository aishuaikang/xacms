package tasks

import (
	"context"
	"time"
	"uav_defender/internal/cache"
)

type FPVTask struct {
	ctx         context.Context
	commonCache cache.CommonCache
	fpVCache    cache.FPVWarningDataCache
}

func NewFPVTask(ctx context.Context, commonCache cache.CommonCache, fpVCache cache.FPVWarningDataCache) *FPVTask {
	return &FPVTask{
		ctx:         ctx,
		commonCache: commonCache,
		fpVCache:    fpVCache,
	}
}

func (t *FPVTask) Execute() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-t.ctx.Done():
				return
			case <-ticker.C:

				t.fpVCache.ResetCacheIfExpired(time.Duration(t.commonCache.GetTTL())) // 假设TTL为30秒
			}
		}
	}()
}
