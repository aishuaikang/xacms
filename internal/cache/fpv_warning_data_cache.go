package cache

import (
	"sync"
	"xacms/internal/dto"
)

type FPVWarningDataCache interface {
	SetFPVWarningDataList(warnings []*dto.FPVWarningData)
	GetFPVWarningDataList() []*dto.FPVWarningData
	PushFPVWarning(warning *dto.FPVWarningData)
}

type fpvWarningDataCache struct {
	fpvWarnings      []*dto.FPVWarningData
	fpvWarningsMutex sync.RWMutex
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
