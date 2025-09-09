package services

import (
	"errors"
	"xacms/internal/models"
	"xacms/internal/routes/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceService 用户服务接口
type DeviceService interface {
	CreateDevice(req dto.CreateDeviceRequest) (*models.DeviceModel, error)
	UpdateDevice(userId uuid.UUID, req dto.UpdateDeviceRequest) (*models.DeviceModel, error)
	InitDevices() ([]models.DeviceModel, error)
}

// deviceService 设备服务实现
type deviceService struct {
	db            *gorm.DB
	commonService CommonService
}

// NewDeviceService 创建设备服务实例
func NewDeviceService(db *gorm.DB, commonService CommonService) DeviceService {
	return &deviceService{
		db:            db,
		commonService: commonService,
	}
}

// CreateDevice 创建用户
func (s *deviceService) CreateDevice(req dto.CreateDeviceRequest) (*models.DeviceModel, error) {
	deviceData := &models.DeviceModel{
		ID:        uuid.New(),
		Name:      req.Name,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,

		// 侦测模块
		DetectionID:   req.DetectionID,
		DetectionIP:   req.DetectionIP,
		DetectionPort: req.DetectionPort,

		// 解析模块
		ParseID: req.ParseID,
		ParseIP: req.ParseIP,

		// FPV模块
		FPVIP:          req.FPVIP,
		StreamServerIP: req.StreamServerIP,

		// 打击模块
		StrikeIP:   req.StrikeIP,
		StrikePort: req.StrikePort,
	}

	if err := s.db.Create(deviceData).Error; err != nil {
		return nil, err
	}
	return deviceData, nil
}

// UpdateDevice 修改设备
func (s *deviceService) UpdateDevice(userId uuid.UUID, req dto.UpdateDeviceRequest) (*models.DeviceModel, error) {
	var user models.DeviceModel
	if err := s.commonService.GetItemByID(userId, &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("设备不存在")
		}
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Longitude != nil {
		user.Longitude = *req.Longitude
	}

	if req.Latitude != nil {
		user.Latitude = *req.Latitude
	}

	if req.DetectionID != nil {
		user.DetectionID = *req.DetectionID
	}

	if req.DetectionIP != nil {
		user.DetectionIP = *req.DetectionIP
	}

	if req.DetectionPort != nil {
		user.DetectionPort = *req.DetectionPort

	}

	if req.ParseID != nil {
		user.ParseID = *req.ParseID
	}

	if req.ParseIP != nil {
		user.ParseIP = *req.ParseIP
	}

	if req.FPVIP != nil {
		user.FPVIP = *req.FPVIP
	}

	if req.StreamServerIP != nil {
		user.StreamServerIP = *req.StreamServerIP
	}

	if req.StrikeIP != nil {
		user.StrikeIP = *req.StrikeIP
	}

	if req.StrikePort != nil {
		user.StrikePort = *req.StrikePort
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// InitDevices 初始化设备列表
func (s *deviceService) InitDevices() ([]models.DeviceModel, error) {
	var devices []models.DeviceModel
	if err := s.commonService.GetItems(&devices); err != nil {
		return nil, err
	}

	return devices, nil
}
