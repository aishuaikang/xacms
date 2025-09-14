package tasks

import (
	"context"
	"time"
	"xacms/internal/cache"
	"xacms/internal/pkg/utils"

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
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-t.ctx.Done():
				return
			case <-ticker.C:
				token, err := utils.GetDecryptToken()
				if err != nil {
					log.Errorf("获取解密Token失败: %v", err)
					continue
				}
				t.decryptTokenCache.SetDecryptToken(token)
			}
		}
	}()
}
