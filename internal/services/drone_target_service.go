package services

import (
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DroneTargetService 无人机目标 服务接口
type DroneTargetService interface {
	GetDroneTargets(req dto.DroneTargetQueryRequest) (*dto.PaginatedResponse[models.DroneTarget], error)
	GetAllDroneTargets(req dto.DroneTargetExportRequest) ([]models.DroneTarget, error)
	CreateDroneTarget(droneTarget *models.DroneTarget) (*models.DroneTarget, error)
	SyncParseDataListToDroneTargetDB(parseDataList []dto.ParseData) error
	SyncDetectorDataListToDroneTargetDB(detectorDataList []dto.DetectorData) error
}

type droneTargetService struct {
	db            *gorm.DB
	commonService CommonService
}

// NewDronTargetService 创建无人机目标服务实例
func NewDronTargetService(db *gorm.DB, commonService CommonService) DroneTargetService {
	return &droneTargetService{
		db:            db,
		commonService: commonService,
	}
}

// GetDroneTargets 获取无人机目标列表
func (s *droneTargetService) GetDroneTargets(req dto.DroneTargetQueryRequest) (*dto.PaginatedResponse[models.DroneTarget], error) {
	query := s.db.Model(&models.DroneTarget{})

	if req.Model != nil {
		query = query.Where("model LIKE ?", "%"+*req.Model+"%")
	}
	if req.DetectionType != nil {
		query = query.Where("detection_type = ?", *req.DetectionType)
	}
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", req.StartTime)
	}

	if req.EndTime != nil {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var droneTargets []models.DroneTarget
	if err := paginate(query, req.Page, req.PageSize).Order("created_at DESC").Find(&droneTargets).Error; err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse[models.DroneTarget]{
		Total: total,
		Items: droneTargets,
	}, nil
}

// GetAllDroneTargets 获取所有无人机目标列表（不分页）
func (s *droneTargetService) GetAllDroneTargets(req dto.DroneTargetExportRequest) ([]models.DroneTarget, error) {
	query := s.db.Model(&models.DroneTarget{})

	if req.Model != nil {
		query = query.Where("model LIKE ?", "%"+*req.Model+"%")
	}
	if req.DetectionType != nil {
		query = query.Where("detection_type = ?", *req.DetectionType)
	}
	if req.StartTime != nil {
		query = query.Where("created_at >= ?", req.StartTime)
	}

	if req.EndTime != nil {
		query = query.Where("created_at <= ?", req.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var droneTargets []models.DroneTarget
	if err := query.Order("created_at DESC").Find(&droneTargets).Error; err != nil {
		return nil, err
	}

	return droneTargets, nil
}

// CreateDroneTarget 创建无人机目标
func (s *droneTargetService) CreateDroneTarget(droneTarget *models.DroneTarget) (*models.DroneTarget, error) {
	if err := s.db.Create(droneTarget).Error; err != nil {
		return nil, err
	}
	return droneTarget, nil
}

// SyncParseDataListToDroneTargetDB 将解析数据同步到无人机目标数据库
func (s *droneTargetService) SyncParseDataListToDroneTargetDB(palert []dto.ParseData) error {
	for _, pa := range palert {

		model, exist := utils.GetDroneModelByModelSource(pa.Model)
		if !exist {
			model = pa.Model
		}

		droneTarget := &models.DroneTarget{
			Serial:         pa.Serial,
			Model:          model,    // 映射后的
			ModelSource:    pa.Model, // 原始的
			DeviceID:       pa.DeviceID,
			SensorID:       pa.ParseID,
			Distance:       pa.Distance,
			Longitude:      pa.DroneGPS.Longitude,
			Latitude:       pa.DroneGPS.Latitude,
			Height:         pa.Height,
			Frequency:      pa.Freq,
			Trajectories:   pa.Trajectories,
			PilotLongitude: pa.PilotGPS.Longitude,
			PilotLatitude:  pa.PilotGPS.Latitude,
			DetectionType:  models.DetectionTypeParse,
			VanishTime:     pa.Expires,
			CommonModel: models.CommonModel{
				CreatedAt: pa.IntrusionTime,
				UpdatedAt: pa.Expires,
			},
		}

		if err := s.db.Model(models.DroneTarget{}).Create(droneTarget).Error; err != nil {
			global.Logger.Error("同步解析告警到数据库失败", zap.Error(err), zap.String("serial", pa.Serial))
		}
	}
	return nil
}

// SyncDetectorDataListToDroneTargetDB 将侦测器数据同步到无人机目标数据库
func (s *droneTargetService) SyncDetectorDataListToDroneTargetDB(detectorDataList []dto.DetectorData) error {
	for _, alert := range detectorDataList {
		serial := utils.GenerateRandomID(8)

		droneTarget := &models.DroneTarget{
			Serial:         serial,
			Model:          alert.UAV,   // 映射后的
			ModelSource:    alert.Model, // 原始的
			DeviceID:       alert.DeviceID,
			SensorID:       alert.DetectionID,
			Distance:       0,
			Longitude:      0,
			Latitude:       0,
			Height:         0,
			Frequency:      0,
			Trajectories:   nil,
			PilotLongitude: 0,
			PilotLatitude:  0,
			DetectionType:  models.DetectionTypeParse,
			VanishTime:     alert.LastTime,
			CommonModel: models.CommonModel{
				CreatedAt: alert.FirstSeen,
				UpdatedAt: alert.LastTime,
			},
		}

		if err := s.db.Model(models.DroneTarget{}).Create(droneTarget).Error; err != nil {
			global.Logger.Error("同步解析告警到数据库失败", zap.Error(err), zap.String("serial", droneTarget.Serial))
		}
	}

	return nil
}
