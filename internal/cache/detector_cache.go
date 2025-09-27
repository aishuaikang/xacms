package cache

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"time"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
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
	GetDirectionalDataTargets() []*dto.DetectorData
	UpdateOrientationByFreq(freq float64) error
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

// 获取需要定向的DetectorData列表
func (c *detectorCache) GetDirectionalDataTargets() []*dto.DetectorData {
	c.detectorDataListMutex.RLock()

	alerts := make([]dto.DetectorData, len(c.detectorDataList))
	copy(alerts, c.detectorDataList)
	c.detectorDataListMutex.RUnlock()

	var targets []*dto.DetectorData
	for i := range alerts {
		data := &alerts[i]
		if c.shouldIncludeInDirectionalTargets(data) {
			// 创建副本避免外部修改
			dataCopy := *data
			targets = append(targets, &dataCopy)
		}
	}
	return targets
}

// shouldIncludeInDirectionalTargets 判断数据是否应该包含在定向目标中
func (c *detectorCache) shouldIncludeInDirectionalTargets(data *dto.DetectorData) bool {
	// if !utils.IsDJIDrone(data.Model) || data.IsCracked {
	// 	return false
	// }
	// _, ok := utils.GetRemoteControllerModelByModelSource(data.Model)
	// return !ok
	return true
}

// GetDetectorDataByFreq 获取指定频点对应的完整detectorData记录
func (c *detectorCache) GetDetectorDataIndexByFreq(freq float64) int {
	c.detectorDataListMutex.RLock()
	defer c.detectorDataListMutex.RUnlock()

	similarThreshold := config.AppConfig.Configuration.SimilarThreshold
	for i := range c.detectorDataList {
		if math.Abs(c.detectorDataList[i].Freq-freq) < similarThreshold {
			return i
		}
	}

	return -1
}

// 修改指定频点对应的 Orientation 字段
func (c *detectorCache) UpdateOrientationByFreq(freq float64) error {
	index := c.GetDetectorDataIndexByFreq(freq)
	if index == -1 {
		global.Logger.Warn("未找到指定频点的detectorData记录", zap.Float64("freq", freq))
		return fmt.Errorf("未找到指定频点的detectorData记录: %f", freq)
	}

	c.detectorDataListMutex.Lock()
	defer c.detectorDataListMutex.Unlock()
	data := &c.detectorDataList[index]

	// 构建天线数据
	antennas := c.buildAntennaDataFromGpios(data.GpiosData[:])

	// 计算候选方向角
	candidateAngle := utils.CalculateOrientationAngle(antennas)

	// 特殊情况：无有效数据
	if candidateAngle == 500 {
		data.Orientation = 500
		data.OrientationTS = models.CustomTime(time.Now())
		return nil
	}

	// 计算新的方向角
	currentAngle := data.Orientation
	newAngle := utils.CalculateNewOrientationAngle(candidateAngle, currentAngle)

	// 更新方向角和时间戳
	data.Orientation = newAngle
	data.OrientationTS = models.CustomTime(time.Now())

	// 记录两天线中间角度的日志
	if len(antennas) == 2 {
		sort.Slice(antennas, func(i, j int) bool {
			return antennas[i].RssiAvg > antennas[j].RssiAvg
		})
		rssiDiff := antennas[0].RssiAvg - antennas[1].RssiAvg
		if rssiDiff <= 2.0 {
			global.Logger.Info("两天线中间角度",
				zap.Float64("angle", candidateAngle),
				zap.Float64("rssiDiff", rssiDiff),
				zap.Any("antennas", antennas))
		}
	}

	return nil
}

// buildAntennaDataFromGpios 从GpiosData构建天线数据
func (c *detectorCache) buildAntennaDataFromGpios(gpiosData []dto.GpioData) []utils.AntennaData {
	antennas := make([]utils.AntennaData, 0, 8)
	for _, d := range gpiosData {
		if d.Count > 0 && d.Gpio >= 0 { // 确保gpio有效
			antennas = append(antennas, utils.AntennaData{
				Gpio:    d.Gpio,
				RssiAvg: d.RssiAverage,
			})
		}
	}
	return antennas
}
