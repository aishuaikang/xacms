package cache

import (
	"sync"
	"time"
	"uav_defender/internal/dto"
)

type FPVWarningDataCache interface {
	SetFPVWarningDataList(warnings []*dto.FPVWarningData)
	GetFPVWarningDataList() []*dto.FPVWarningData
	PushFPVWarning(warning *dto.FPVWarningData)
	SetLastUpdated(t time.Time)
	GetLastUpdated() time.Time
	ResetCacheIfExpired(expireDuration time.Duration)
}

type fpvWarningDataCache struct {
	fpvWarnings      []*dto.FPVWarningData
	fpvWarningsMutex sync.RWMutex

	// 最后更新时间
	lastUpdated      time.Time
	lastUpdatedMutex sync.RWMutex
}

func NewFPVWarningDataCache() FPVWarningDataCache {
	return &fpvWarningDataCache{
		fpvWarnings:      make([]*dto.FPVWarningData, 0),
		fpvWarningsMutex: sync.RWMutex{},
	}
}

// SetFPVWarnings 设置FPV警告数据
func (c *fpvWarningDataCache) SetFPVWarningDataList(warnings []*dto.FPVWarningData) {
	c.fpvWarningsMutex.Lock()
	defer c.fpvWarningsMutex.Unlock()
	c.fpvWarnings = warnings
}

// GetFPVWarnings 获取FPV警告数据列表
func (c *fpvWarningDataCache) GetFPVWarningDataList() []*dto.FPVWarningData {
	c.fpvWarningsMutex.RLock()
	defer c.fpvWarningsMutex.RUnlock()
	return c.fpvWarnings
}

// PushFPVWarning 将FPV警告数据推送到列表并且过滤掉 time 与 当前时间差超过 2 秒的数据
func (c *fpvWarningDataCache) PushFPVWarning(warning *dto.FPVWarningData) {
	var filteredWarnings []*dto.FPVWarningData

	// 过滤掉 time 与 当前时间差超过 2 秒的数据
	fpvWarnings := c.GetFPVWarningDataList()
	for _, item := range fpvWarnings {
		if warning.Time-item.Time <= 2 {
			filteredWarnings = append(filteredWarnings, item)
		}
	}

	filteredWarnings = append(filteredWarnings, warning)

	c.SetFPVWarningDataList(filteredWarnings)
}

// SetLastUpdated 设置最后更新时间
func (c *fpvWarningDataCache) SetLastUpdated(t time.Time) {
	c.lastUpdatedMutex.Lock()
	defer c.lastUpdatedMutex.Unlock()
	c.lastUpdated = t
}

// GetLastUpdated 获取最后更新时间
func (c *fpvWarningDataCache) GetLastUpdated() time.Time {
	c.lastUpdatedMutex.RLock()
	defer c.lastUpdatedMutex.RUnlock()
	return c.lastUpdated
}

// ResetCacheIfExpired 如果缓存过期则重置缓存
func (c *fpvWarningDataCache) ResetCacheIfExpired(expireDuration time.Duration) {
	c.lastUpdatedMutex.Lock()
	defer c.lastUpdatedMutex.Unlock()
	now := time.Now()
	if now.Sub(c.lastUpdated) > expireDuration {
		c.fpvWarnings = make([]*dto.FPVWarningData, 0)
		c.lastUpdated = now
	}
}
