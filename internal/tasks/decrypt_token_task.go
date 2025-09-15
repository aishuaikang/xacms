package tasks

import (
	"context"
	"time"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/utils"

	"github.com/gofiber/fiber/v2/log"
)

type DecryptTokenTask struct {
	ctx               context.Context
	decryptTokenCache cache.DecryptTokenCache
}

func NewDecryptTokenTask(ctx context.Context, decryptTokenCache cache.DecryptTokenCache) *DecryptTokenTask {
	return &DecryptTokenTask{
		ctx:               ctx,
		decryptTokenCache: decryptTokenCache,
	}
}

func (t *DecryptTokenTask) Execute() {
	token, err := utils.RefreshDecryptToken()
	if err != nil {
		log.Errorf("获取解密Token失败: %v", err)
	} else {
		t.decryptTokenCache.SetDecryptToken(token)
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-t.ctx.Done():
				log.Debug("DecryptTokenTask 上下文已取消，正在退出 goroutine")
				return
			case <-ticker.C:
				token, err := utils.RefreshDecryptToken()
				if err != nil {
					log.Errorf("获取解密Token失败: %v", err)
					continue
				}
				t.decryptTokenCache.SetDecryptToken(token)
			}
		}
	}()
}
