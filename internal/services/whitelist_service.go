package services

import (
	"uav_defender/internal/dto"
	"uav_defender/internal/models"

	"gorm.io/gorm"
)

type WhitelistService interface {
	GetWhitelists(req dto.WhitelistQueryRequest) (*dto.PaginatedResponse[models.Whitelist], error)
	AddWhitelist(req dto.WhitelistCreateRequest) (*models.Whitelist, error)
	DeleteWhitelistBySerial(serial string) error
	IsSerialWhitelisted(serial string) (bool, error)
}

// whitelistService 白名单服务实现
type whitelistService struct {
	db            *gorm.DB
	commonService CommonService
}

// NewWhitelistService 创建白名单服务实例
func NewWhitelistService(db *gorm.DB, commonService CommonService) WhitelistService {
	return &whitelistService{
		db:            db,
		commonService: commonService,
	}
}

// GetWhitelists 获取白名单列表
func (s *whitelistService) GetWhitelists(req dto.WhitelistQueryRequest) (*dto.PaginatedResponse[models.Whitelist], error) {
	query := s.db.Model(&models.Whitelist{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var whitelist []models.Whitelist
	if err := paginate(query, req.Page, req.PageSize).Order("created_at DESC").Find(&whitelist).Error; err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse[models.Whitelist]{
		Total: total,
		Items: whitelist,
	}, nil
}

// AddWhitelist 添加白名单
func (s *whitelistService) AddWhitelist(req dto.WhitelistCreateRequest) (*models.Whitelist, error) {
	whitelist := &models.Whitelist{
		DeviceID: req.DeviceID,
		Model:    req.Model,
		Serial:   req.Serial,
	}

	if err := s.db.Create(whitelist).Error; err != nil {
		return nil, err
	}

	return whitelist, nil
}

// DeleteWhitelistBySerial 根据Serial删除白名单
func (s *whitelistService) DeleteWhitelistBySerial(serial string) error {
	return s.db.Unscoped().Where("serial = ?", serial).Delete(&models.Whitelist{}).Error
}

// IsSerialWhitelisted 根据Serial查询是否存在白名单
func (s *whitelistService) IsSerialWhitelisted(serial string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.Whitelist{}).Where("serial = ?", serial).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
