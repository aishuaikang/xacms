package services

import (
	"uav_defender/internal/dto"
	"uav_defender/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DroneTargetService 无人机目标 服务接口
type DroneTargetService interface {
	GetDroneTargets(req dto.DroneTargetQueryRequest) (*dto.PaginatedResponse[models.DroneTargetModel], error)
	CreateDroneTarget(droneTarget *models.DroneTargetModel) (*models.DroneTargetModel, error)
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
