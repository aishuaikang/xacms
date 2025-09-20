package cache

import (
	"sort"
	"sync"
	"time"
	"uav_defender/internal/dto"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"go.uber.org/zap"
)

type ParseCache interface {
	SetParseDataList(dataList []dto.ParseData)
	GetParseDataList() []dto.ParseData
	UpdateParseDataAtIndex(index int, data dto.ParseData)
	AddParseData(data dto.ParseData)
	SortParseDataListByExpires()
	GetParseDataListLength() int
	CleanupExpiredParseData(ttl int64)
}

type parseCache struct {
	parseDataList      []dto.ParseData
	parseDataListMutex sync.RWMutex
	droneTargetService services.DroneTargetService
}

func NewParseCache(droneTargetService services.DroneTargetService) ParseCache {
	return &parseCache{
		parseDataList:      make([]dto.ParseData, 0),
		parseDataListMutex: sync.RWMutex{},
		droneTargetService: droneTargetService,
	}
}

// SetParseDataList 设置解析数据列表
func (c *parseCache) SetParseDataList(dataList []dto.ParseData) {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()
	c.parseDataList = dataList
}

// GetParseDataList 获取解析数据列表
func (c *parseCache) GetParseDataList() []dto.ParseData {
	c.parseDataListMutex.RLock()
	defer c.parseDataListMutex.RUnlock()
	return c.parseDataList
}

// 根据索引更新解析数据
func (c *parseCache) UpdateParseDataAtIndex(index int, data dto.ParseData) {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()
	if index >= 0 && index < len(c.parseDataList) {
		c.parseDataList[index] = data
	}
}

// AddParseData 添加新的解析数据
func (c *parseCache) AddParseData(data dto.ParseData) {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()
	c.parseDataList = append(c.parseDataList, data)
}

// SortParseDataListByExpires 根据 Expires 字段对解析数据列表进行排序
func (c *parseCache) SortParseDataListByExpires() {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()
	sort.Slice(c.parseDataList, func(i, j int) bool {
		return c.parseDataList[i].Expires.Time().Unix() < c.parseDataList[j].Expires.Time().Unix()
	})
}

// GetParseDataListLength 获取解析数据列表长度
func (c *parseCache) GetParseDataListLength() int {
	c.parseDataListMutex.RLock()
	defer c.parseDataListMutex.RUnlock()
	return len(c.parseDataList)
}

// CleanupExpiredParseData 清理过期的解析数据
func (c *parseCache) CleanupExpiredParseData(ttl int64) {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()

	validData := make([]dto.ParseData, 0)
	invalidData := make([]dto.ParseData, 0)
	now := time.Now().Unix()
	for _, data := range c.parseDataList {
		expireTime := data.Expires.Time().Unix() + ttl
		if expireTime > now {
			validData = append(validData, data)
		} else {
			invalidData = append(invalidData, data)
		}
	}
	c.parseDataList = validData

	if len(invalidData) == 0 {
		return
	}

	// 将过期的数据同步到数据库
	if err := c.droneTargetService.SyncParseDataListToDroneTargetDB(invalidData); err != nil {
		global.Logger.Error("同步解析数据列表到无人机目标数据库失败", zap.Error(err))
		return
	}
	global.Logger.Info("入库过期解析数据", zap.Int("入库前数量", len(validData)+len(invalidData)), zap.Int("入库后数量", len(validData)), zap.Int("清理数量", len(invalidData)))
}
