package cache

import (
	"fmt"
	"strconv"
	"sync"
	"time"
	"uav_defender/internal/dto"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"
	"uav_defender/internal/services"

	"go.uber.org/zap"
)

type DetectorCache interface {
	GetDetectorDataList() []dto.DetectorData
	GetDetectorDataListLen() int
	AddDetectorData(data dto.DetectorData) error
	UpdateDetectorDataList(newData dto.DetectorData) error
	CleanupExpiredDetectorData(ttl int64)
	GetNonDJIUncrackedData() []dto.DetectorData
	SetDetectorDataList(dataList []dto.DetectorData)
}

type detectorCache struct {
	detectorDataList      []dto.DetectorData
	detectorDataListMutex sync.RWMutex
	droneTargetService    services.DroneTargetService
}

func NewDetectorCache(droneTargetService services.DroneTargetService) DetectorCache {
	return &detectorCache{
		detectorDataList:      make([]dto.DetectorData, 0),
		detectorDataListMutex: sync.RWMutex{},
		droneTargetService:    droneTargetService,
	}
}

// GetDetectorDataList 获取侦测器数据列表
func (c *detectorCache) GetDetectorDataList() []dto.DetectorData {
	c.detectorDataListMutex.RLock()
	defer c.detectorDataListMutex.RUnlock()
	return c.detectorDataList
}

// GetDetectorDataListLen 获取侦测器数据列表长度
func (c *detectorCache) GetDetectorDataListLen() int {
	c.detectorDataListMutex.RLock()
	defer c.detectorDataListMutex.RUnlock()
	return len(c.detectorDataList)
}

// AddDetectorData 添加侦测器数据
func (c *detectorCache) AddDetectorData(data dto.DetectorData) error {
	c.detectorDataListMutex.Lock()
	defer c.detectorDataListMutex.Unlock()
	c.detectorDataList = append(c.detectorDataList, data)
	return nil
}

// UpdateDetectorDataList 更新侦测器数据列表
func (c *detectorCache) UpdateDetectorDataList(newData dto.DetectorData) error {
	// 判断DetectorDataList是否为空
	len := c.GetDetectorDataListLen()
	if len == 0 {
		return c.AddDetectorData(newData)
	}

	c.detectorDataListMutex.Lock()
	defer c.detectorDataListMutex.Unlock()

	// 这里是找到所有相似的目标进行更新
	similarThreshold := config.AppConfig.Configuration.SimilarThreshold
	found := false
	for i, data := range c.detectorDataList {
		if utils.IsSameDetectorData(&data, &newData, similarThreshold) {
			c.detectorDataList[i] = newData
			found = true
		}
	}

	// 如果没有找到相似目标，添加新数据
	if !found {
		c.detectorDataList = append(c.detectorDataList, newData)
	}

	// 对相同的目标进行合并
	mergedList := make([]dto.DetectorData, 0)
	mergedMap := make(map[string]dto.DetectorData) // 使用map来避免重复合并

	for _, data := range c.detectorDataList {
		// key := strconv.Itoa(int(data.DetectionID)) + strconv.Itoa(int(data.DeviceID)) + "_" + data.Model + "_" + strconv.FormatFloat(data.Freq, 'f', 2, 64)
		key := fmt.Sprintf("%d_%d_%s_%s_%s", data.DetectionID, data.DeviceID, data.ID, data.Model, strconv.FormatFloat(data.Freq, 'f', 2, 64))

		if _, ok := mergedMap[key]; !ok {
			mergedMap[key] = data
		}
	}

	for _, data := range mergedMap {
		mergedList = append(mergedList, data)
	}

	c.detectorDataList = mergedList
	return nil
}

// CleanupExpiredDetectorData 清理过期的侦测器数据
func (c *detectorCache) CleanupExpiredDetectorData(ttl int64) {
	c.detectorDataListMutex.Lock()
	defer c.detectorDataListMutex.Unlock()

	validData := make([]dto.DetectorData, 0)
	invalidData := make([]dto.DetectorData, 0)
	now := time.Now().Unix()
	for _, data := range c.detectorDataList {
		expireTime := data.LastTime.Time().Unix() + ttl
		if expireTime > now {
			validData = append(validData, data)
		} else {
			invalidData = append(invalidData, data)
		}
	}
	c.detectorDataList = validData

	if len(invalidData) == 0 {
		return
	}

	// 将过期的数据同步到数据库
	if err := c.droneTargetService.SyncDetectorDataListToDroneTargetDB(invalidData); err != nil {
		global.Logger.Error("同步侦测器数据列表到无人机目标数据库失败", zap.Error(err))
		return
	}
	global.Logger.Info("入库过期侦测器数据", zap.Int("入库前数量", len(validData)+len(invalidData)), zap.Int("入库后数量", len(validData)), zap.Int("清理数量", len(invalidData)))
}

// GetNonDJIUncrackedData 获取所有非大疆且未破解的侦测器数据
func (c *detectorCache) GetNonDJIUncrackedData() []dto.DetectorData {
	c.detectorDataListMutex.RLock()
	defer c.detectorDataListMutex.RUnlock()
	nonDJIUncrackedData := make([]dto.DetectorData, 0)
	for _, data := range c.detectorDataList {
		if !utils.IsDJIDrone(data.Model) && !data.IsCracked {
			nonDJIUncrackedData = append(nonDJIUncrackedData, data)
		}
	}
	return nonDJIUncrackedData
}

// SetDetectorDataList 设置侦测器数据列表
func (c *detectorCache) SetDetectorDataList(dataList []dto.DetectorData) {
	c.detectorDataListMutex.Lock()
	defer c.detectorDataListMutex.Unlock()
	c.detectorDataList = dataList
}
