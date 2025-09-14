package cache

import (
	"sort"
	"sync"
	"time"
	"xacms/internal/dto"
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
}

func NewParseCache() ParseCache {
	return &parseCache{
		parseDataList:      make([]dto.ParseData, 0),
		parseDataListMutex: sync.RWMutex{},
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
		return c.parseDataList[i].Expires < c.parseDataList[j].Expires
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
	now := time.Now().Unix()
	for _, data := range c.parseDataList {
		expireTime := data.Expires + ttl
		if expireTime > now {
			validData = append(validData, data)
		}
	}
	c.parseDataList = validData
}
