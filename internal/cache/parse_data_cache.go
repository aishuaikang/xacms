package cache

import (
	"sort"
	"sync"
	"xacms/internal/dto"
)

type ParseDataCache interface {
	SetParseDataList(dataList []dto.ParseData)
	GetParseDataList() []dto.ParseData
	UpdateParseDataAtIndex(index int, data dto.ParseData)
	AddParseData(data dto.ParseData)
	SortParseDataListByExpires()
	GetParseDataListLength() int
}

type parseDataCache struct {
	parseDataList      []dto.ParseData
	parseDataListMutex sync.RWMutex
}

func NewParseDataCache() ParseDataCache {
	return &parseDataCache{
		parseDataList:      make([]dto.ParseData, 0),
		parseDataListMutex: sync.RWMutex{},
	}
}

// SetParseDataList 设置解析数据列表
func (c *parseDataCache) SetParseDataList(dataList []dto.ParseData) {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()
	c.parseDataList = dataList
}

// GetParseDataList 获取解析数据列表
func (c *parseDataCache) GetParseDataList() []dto.ParseData {
	c.parseDataListMutex.RLock()
	defer c.parseDataListMutex.RUnlock()
	return c.parseDataList
}

// 根据索引更新解析数据
func (c *parseDataCache) UpdateParseDataAtIndex(index int, data dto.ParseData) {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()
	if index >= 0 && index < len(c.parseDataList) {
		c.parseDataList[index] = data
	}
}

// AddParseData 添加新的解析数据
func (c *parseDataCache) AddParseData(data dto.ParseData) {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()
	c.parseDataList = append(c.parseDataList, data)
}

// SortParseDataListByExpires 根据 Expires 字段对解析数据列表进行排序
func (c *parseDataCache) SortParseDataListByExpires() {
	c.parseDataListMutex.Lock()
	defer c.parseDataListMutex.Unlock()
	sort.Slice(c.parseDataList, func(i, j int) bool {
		return c.parseDataList[i].Expires < c.parseDataList[j].Expires
	})
}

// 获取解析数据列表长度
func (c *parseDataCache) GetParseDataListLength() int {
	c.parseDataListMutex.RLock()
	defer c.parseDataListMutex.RUnlock()
	return len(c.parseDataList)
}
