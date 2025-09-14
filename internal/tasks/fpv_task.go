package tasks

import (
	"context"
	"time"
	"xacms/internal/cache"

	"github.com/gofiber/fiber/v2/log"
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
				log.Debug("FPVTask 上下文已取消，正在退出 goroutine")
				return
			case <-ticker.C:

				t.fpVCache.ResetCacheIfExpired(30 * time.Second) // 假设TTL为30秒
			}
		}
	}()
}
