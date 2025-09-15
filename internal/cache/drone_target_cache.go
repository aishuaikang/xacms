package cache

import (
	"errors"
	"sync"
	"time"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/utils"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/datatypes"
)

type DroneTargetCache interface {
	HandleParseDataToDroneTarget(data dto.ParseData) error
	FindDroneTargetBySerial(serial string) (*models.DroneTargetModel, int)
	AppendDroneTarget(target *models.DroneTargetModel)
	GetDroneTargetCount() int
	MarkStaleTargetsAsVanishedAndRemove(thresholdSeconds int64) []*models.DroneTargetModel
}

type droneTargetCache struct {
	droneTargets      []*models.DroneTargetModel
	droneTargetsMutex sync.RWMutex
}

func NewDroneTargetCache() DroneTargetCache {
	return &droneTargetCache{
		droneTargets:      []*models.DroneTargetModel{},
		droneTargetsMutex: sync.RWMutex{},
	}
}

// HandleParseDataToDroneTarget 处理ParseData到无人机目标的转换和存储
func (c *droneTargetCache) HandleParseDataToDroneTarget(parseData dto.ParseData) error {
	const coordThreshold = 0.001

	// 检查坐标有效性
	coordValid := utils.IsValidCoord(parseData.DroneGPS.Longitude, parseData.DroneGPS.Latitude, coordThreshold)
	if !coordValid {
		log.Warnf("目标 %s 坐标无效: (%.6f, %.6f)",
			parseData.Serial, parseData.DroneGPS.Longitude, parseData.DroneGPS.Latitude)
	}

	droneTarget, droneTargetIndex := c.FindDroneTargetBySerial(parseData.Serial)
	if droneTarget != nil && droneTargetIndex != -1 {

		// TODO 由于存在定时任务是在扫描2分钟内未更新的目标设置消失时间并且入库此逻辑是否冗余
		// 计算出上次更新时间和当前时间的差值（秒）
		// timeDiff := carbon.Now().DiffAbsInSeconds(droneTarget.UpdatedAt.Carbon)
		// 判断是否10分钟没有更新
		// if timeDiff > 10*60 {
		// 	log.Infof("目标 %s 超过10分钟未更新，创建新目标", parseData.Serial)
		// } else {
		c.updateDroneTargetFromParseData(droneTargetIndex, parseData, coordValid)
		// }

	} else {
		c.addDroneTargetFromParseData(parseData, coordValid)
	}

	return nil
}

// FindDroneTargetBySerial 根据序列号查找未消失的无人机目标
func (c *droneTargetCache) FindDroneTargetBySerial(serial string) (*models.DroneTargetModel, int) {
	c.droneTargetsMutex.RLock()
	defer c.droneTargetsMutex.RUnlock()
	for i, target := range c.droneTargets {
		// log.Debugf("检查目标: %s, 消失时间: %v", target.Serial, target.VanishTime.IsZero())
		// 只返回未消失的目标
		if target.Serial == serial && target.VanishTime.IsZero() {
			return target, i
		}
	}
	return nil, -1
}

// addDroneTargetFromParseData 从解析数据创建并添加新的无人机目标
func (c *droneTargetCache) addDroneTargetFromParseData(parseData dto.ParseData, coordValid bool) {
	log.Debugf("创建无人机目标: %s", parseData.Serial)
	now := time.Now()
	var trajectory datatypes.JSONSlice[models.Trajectory]

	if coordValid {
		trajectory = datatypes.JSONSlice[models.Trajectory]{
			{
				Lat:    parseData.DroneGPS.Latitude,
				Lng:    parseData.DroneGPS.Longitude,
				Height: parseData.Height,
			},
		}

		log.Debugf("为目标 %s 添加初始轨迹点: (%.6f, %.6f, %.1f)",
			parseData.Serial, parseData.DroneGPS.Latitude, parseData.DroneGPS.Longitude, parseData.Height)
	}

	droneTarget := &models.DroneTargetModel{
		Serial:        parseData.Serial,
		Model:         parseData.Model,
		Device:        parseData.Device,
		Distance:      parseData.Distance,
		DroneLng:      parseData.DroneGPS.Longitude,
		DroneLat:      parseData.DroneGPS.Latitude,
		Height:        parseData.Height,
		Frequency:     parseData.Freq,
		DetectionType: models.DetectionTypeParse,
		Trajectory:    trajectory,
		PilotLng:      parseData.PilotGPS.Longitude,
		PilotLat:      parseData.PilotGPS.Latitude,
		CommonModel: models.CommonModel{
			CreatedAt: models.CustomTime(now),
			UpdatedAt: models.CustomTime(now),
		},
	}

	c.AppendDroneTarget(droneTarget)

	log.Debugf("无人机目标 %s 已添加到缓存", parseData.Serial)
}

// updateDroneTargetFromParseData 使用解析数据更新现有的无人机目标
func (c *droneTargetCache) updateDroneTargetFromParseData(droneTargetIndex int, parseData dto.ParseData, coordValid bool) error {
	log.Debugf("更新无人机目标: %s", parseData.Serial)
	if droneTargetCount := c.GetDroneTargetCount(); droneTargetIndex < 0 || droneTargetIndex >= droneTargetCount {
		return errors.New("无效的目标索引")
	}

	c.droneTargetsMutex.Lock()
	defer c.droneTargetsMutex.Unlock()

	target := c.droneTargets[droneTargetIndex]
	now := time.Now()

	// 更新基本信息
	target.Distance = parseData.Distance
	target.Height = parseData.Height
	target.UpdatedAt = models.CustomTime(now)
	target.Device = parseData.Device
	target.Frequency = parseData.Freq

	// 更新轨迹数据
	if !coordValid {
		return nil
	}

	target.DroneLng = parseData.DroneGPS.Longitude
	target.DroneLat = parseData.DroneGPS.Latitude

	newTrajectory := models.Trajectory{
		Lat:    parseData.DroneGPS.Latitude,
		Lng:    parseData.DroneGPS.Longitude,
		Height: parseData.Height,
	}

	if target.Trajectory == nil {
		target.Trajectory = datatypes.JSONSlice[models.Trajectory]{newTrajectory}
		return nil
	}

	// 判断最后一个轨迹点是否与新点相同，避免重复添加
	lastPoint := target.Trajectory[len(target.Trajectory)-1]
	if lastPoint.Lat != newTrajectory.Lat || lastPoint.Lng != newTrajectory.Lng {
		target.Trajectory = append(target.Trajectory, newTrajectory)
	}

	log.Debugf("更新目标 %s 轨迹点: (%.6f, %.6f, %.1f)",
		parseData.Serial, parseData.DroneGPS.Latitude, parseData.DroneGPS.Longitude, parseData.Height)
	return nil
}

// AppendDroneTarget 添加无人机目标到缓存
func (c *droneTargetCache) AppendDroneTarget(target *models.DroneTargetModel) {
	c.droneTargetsMutex.Lock()
	defer c.droneTargetsMutex.Unlock()
	c.droneTargets = append(c.droneTargets, target)
}

// GetDroneTargetCount 获取当前缓存中的无人机目标数量
func (c *droneTargetCache) GetDroneTargetCount() int {
	c.droneTargetsMutex.RLock()
	defer c.droneTargetsMutex.RUnlock()
	return len(c.droneTargets)
}

// MarkStaleTargetsAsVanishedAndRemove 标记超过阈值未更新的目标为已消失并从缓存中移除，返回这些目标列表
func (c *droneTargetCache) MarkStaleTargetsAsVanishedAndRemove(thresholdSeconds int64) []*models.DroneTargetModel {
	var staleTargets []*models.DroneTargetModel
	var activeTargets []*models.DroneTargetModel
	now := time.Now()

	// 先使用读锁读取数据
	c.droneTargetsMutex.RLock()
	for _, target := range c.droneTargets {
		timeDiff := now.Unix() - time.Time(target.UpdatedAt).Unix()

		// 只处理未消失的目标
		if target.VanishTime.IsZero() && timeDiff > thresholdSeconds {
			// 设置消失时间
			target.VanishTime = models.CustomTime(now)
			staleTargets = append(staleTargets, target)
			log.Infof("目标 %s 超过 %d 秒未更新，设置消失时间并准备移除", target.Serial, thresholdSeconds)
		} else {
			activeTargets = append(activeTargets, target)
		}
	}
	c.droneTargetsMutex.RUnlock()

	// 如果有需要更新的目标，使用写锁更新缓存
	if len(staleTargets) > 0 {
		c.droneTargetsMutex.Lock()
		c.droneTargets = activeTargets
		c.droneTargetsMutex.Unlock()
	}

	return staleTargets
}
