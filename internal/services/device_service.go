package services

import (
	"errors"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DeviceService 设备 服务接口
type DeviceService interface {
	CreateDevice(req dto.CreateDeviceRequest) (*models.DeviceModel, error)
	UpdateDevice(deviceID uuid.UUID, req dto.UpdateDeviceRequest) (*models.DeviceModel, error)
	GetAllDevices() ([]models.DeviceModel, error)
	RefreshMediaMtxConfig()
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

// CreateDevice 创建设备
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
		FPVIP:  req.FPVIP,
		RTSPIP: req.RTSPIP,

		// 打击模块
		StrikeIP: req.StrikeIP,
	}

	if err := s.db.Create(deviceData).Error; err != nil {
		return nil, err
	}
	return deviceData, nil
}

// UpdateDevice 修改设备
func (s *deviceService) UpdateDevice(deviceID uuid.UUID, req dto.UpdateDeviceRequest) (*models.DeviceModel, error) {
	var user models.DeviceModel
	if err := s.commonService.GetItemByID(deviceID, &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("设备不存在")
		}
		return nil, err
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Longitude != nil {
		user.Longitude = req.Longitude
	}

	if req.Latitude != nil {
		user.Latitude = req.Latitude
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

	if req.RTSPIP != nil {
		user.RTSPIP = *req.RTSPIP
	}

	if req.StrikeIP != nil {
		user.StrikeIP = *req.StrikeIP
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllDevices 初始化设备列表
func (s *deviceService) GetAllDevices() ([]models.DeviceModel, error) {
	var devices []models.DeviceModel
	if err := s.db.Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

// 刷新 MediaMtx 配置
func (s *deviceService) RefreshMediaMtxConfig() {
	var devices []models.DeviceModel
	if err := s.db.Find(&devices).Error; err != nil {
		global.Logger.Error("获取设备列表失败, 无法刷新 MediaMtx 配置", zap.Error(err))
		return
	}

	utils.UpdateMediaMtxConfigPaths(devices)

	global.Logger.Info("已刷新 MediaMtx 配置")
}
