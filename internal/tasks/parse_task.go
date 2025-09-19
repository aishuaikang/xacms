package tasks

import (
	"context"
	"time"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/global"
)

type ParseTask struct {
	ctx         context.Context
	commonCache cache.CommonCache
	parseCache  cache.ParseCache
}

func NewParseTask(ctx context.Context, commonCache cache.CommonCache, parseCache cache.ParseCache) *ParseTask {
	return &ParseTask{
		ctx:         ctx,
		commonCache: commonCache,
		parseCache:  parseCache,
	}
}

func (t *ParseTask) Execute() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-t.ctx.Done():
				global.Logger.Info("ParseTask 上下文已取消，正在退出 goroutine")
				return
			case <-ticker.C:

				t.parseCache.CleanupExpiredParseData(t.commonCache.GetTTL()) // 假设TTL为30秒
			}
		}
	}()
}
