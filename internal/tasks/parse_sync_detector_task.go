package tasks

import (
	"context"
	"time"
	"uav_defender/internal/cache"
	"uav_defender/internal/dto"
)

type ParseSyncDetectorTask struct {
	ctx           context.Context
	commonCache   cache.CommonCache
	parseCache    cache.ParseCache
	detectorCache cache.DetectorCache
}

func NewParseSyncDetectorTask(ctx context.Context, commonCache cache.CommonCache, parseCache cache.ParseCache, detectorCache cache.DetectorCache) *ParseSyncDetectorTask {
	return &ParseSyncDetectorTask{
		ctx:           ctx,
		commonCache:   commonCache,
		parseCache:    parseCache,
		detectorCache: detectorCache,
	}
}

func (t *ParseSyncDetectorTask) Execute() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-t.ctx.Done():
				return
			case <-ticker.C:
				// 获取所有DID的解析数据，转换为DetectorData
				allDIDParseDataList := t.parseCache.GetAllDIDDataAsDetectorData()
				if len(allDIDParseDataList) == 0 {
					continue
				}

				// 获取所有未破解的非DJI设备数据
				nonDJIUncrackedData := t.detectorCache.GetNonDJIUncrackedData()

				// 预分配容量并合并数据
				newDetectorDataList := make([]dto.DetectorData, 0, len(allDIDParseDataList)+len(nonDJIUncrackedData))
				newDetectorDataList = append(newDetectorDataList, allDIDParseDataList...)
				newDetectorDataList = append(newDetectorDataList, nonDJIUncrackedData...)

				t.detectorCache.SetDetectorDataList(newDetectorDataList)
			}
		}
	}()
}
