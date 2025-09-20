package services

import (
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DroneTargetService 无人机目标 服务接口
type DroneTargetService interface {
	GetDroneTargets(req dto.DroneTargetQueryRequest) (*dto.PaginatedResponse[models.DroneTargetModel], error)
	CreateDroneTarget(droneTarget *models.DroneTargetModel) (*models.DroneTargetModel, error)
	SyncParseDataListToDroneTargetDB(parseDataList []dto.ParseData) error
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
func (s *droneTargetService) GetDroneTargets(req dto.DroneTargetQueryRequest) (*dto.PaginatedResponse[models.DroneTargetModel], error) {
	query := s.db.Model(&models.DroneTargetModel{}).Preload(clause.Associations)

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

	var droneTargets []models.DroneTargetModel
	if err := paginate(query, req.Page, req.PageSize).Order("created_at DESC").Find(&droneTargets).Error; err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse[models.DroneTargetModel]{
		Total: total,
		Items: droneTargets,
	}, nil
}

// CreateDroneTarget 创建无人机目标
func (s *droneTargetService) CreateDroneTarget(droneTarget *models.DroneTargetModel) (*models.DroneTargetModel, error) {
	if err := s.db.Create(droneTarget).Error; err != nil {
		return nil, err
	}
	return droneTarget, nil
}

// SyncParseDataListToDroneTargetDB 将解析数据同步到无人机目标数据库
func (s *droneTargetService) SyncParseDataListToDroneTargetDB(palert []dto.ParseData) error {
	for _, pa := range palert {

		droneTarget := &models.DroneTargetModel{
			Serial:        pa.Serial,
			Model:         pa.Model,
			Distance:      pa.Distance,
			DroneLng:      pa.DroneGPS.Longitude,
			DroneLat:      pa.DroneGPS.Latitude,
			Height:        pa.Height,
			Device:        pa.Device,
			Frequency:     pa.Freq,
			DetectionType: models.DetectionTypeParse,
			Trajectories:  pa.TrajectoryList,
			PilotLat:      pa.PilotGPS.Latitude,
			PilotLng:      pa.PilotGPS.Longitude,
			VanishTime:    pa.Expires,
			CommonModel: models.CommonModel{
				CreatedAt: pa.IntrusionTime,
				UpdatedAt: pa.Expires,
			},
		}

		if err := s.db.Model(models.DroneTargetModel{}).Create(droneTarget).Error; err != nil {
			global.Logger.Error("同步解析告警到数据库失败", zap.Error(err), zap.String("serial", pa.Serial))
		}
	}
	return nil
}
