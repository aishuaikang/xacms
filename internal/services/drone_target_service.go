package services

import (
	"xacms/internal/dto"
	"xacms/internal/models"

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
	var droneTargets []models.DroneTargetModel
	query := s.db.Model(&models.DroneTargetModel{}).Preload(clause.Associations)

	// 分页参数
	page := req.Page
	pageSize := req.PageSize

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

	offset := (page - 1) * pageSize

	if err := query.Offset(offset).Limit(pageSize).Find(&droneTargets).Order("created_at DESC").Error; err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse[models.DroneTargetModel]{
		Total: len(droneTargets),
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
