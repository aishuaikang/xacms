package services

import (
	"uav_defender/internal/dto"
	"uav_defender/internal/models"

	"gorm.io/gorm"
)

type FPVService interface {
	GetFPVs(req dto.FPVQueryRequest) (*dto.PaginatedResponse[models.FPVVideo], error)
	AddFPV(req dto.AddFPVRequest) error
}

// fpvService FPV服务实现
type fpvService struct {
	db            *gorm.DB
	commonService CommonService
}

// NewFPVService 创建FPV服务实例
func NewFPVService(db *gorm.DB, commonService CommonService) FPVService {
	return &fpvService{
		db:            db,
		commonService: commonService,
	}
}

// GetFPVs 获取FPV列表
func (s *fpvService) GetFPVs(req dto.FPVQueryRequest) (*dto.PaginatedResponse[models.FPVVideo], error) {
	query := s.db.Model(&models.FPVVideo{})

	if req.Frequency != nil {
		query = query.Where("frequency = ?", *req.Frequency)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var fpvs []models.FPVVideo
	if err := paginate(query, req.Page, req.PageSize).Order("created_at DESC").Find(&fpvs).Error; err != nil {
		return nil, err
	}
	return &dto.PaginatedResponse[models.FPVVideo]{
		Total: total,
		Items: fpvs,
	}, nil
}

// AddFPV 添加FPV记录
func (s *fpvService) AddFPV(req dto.AddFPVRequest) error {
	fpv := &models.FPVVideo{
		DeviceID:  req.DeviceID,
		Frequency: req.Frequency,
		Filename:  req.Filename,
	}
	return s.db.Create(fpv).Error
}
