package tasks

import (
	"context"
	"time"
	"xacms/internal/cache"

	"github.com/gofiber/fiber/v2/log"
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
				log.Info("设备存储停止刷新")
				return
			case <-ticker.C:

				t.parseCache.CleanupExpiredParseData(t.commonCache.GetTTL()) // 假设TTL为30秒
			}
		}
	}()
}
