package store

import (
	"context"
	"sync"
	"xacms/internal/models"
	"xacms/internal/services"

	"github.com/gofiber/fiber/v2/log"
)

type DeviceStore interface {
	GetDeviceList() []models.DeviceModel
	RefreshDeviceList() error
	GetRefreshChan() chan<- struct{}
}

type deviceStore struct {
	ctx           context.Context
	deviceService services.DeviceService
	deviceList    []models.DeviceModel
	mu            sync.RWMutex
	refreshChan   chan struct{}
}

func NewDeviceStore(ctx context.Context, deviceService services.DeviceService) DeviceStore {
	// 初始化设备列表
	deviceList, err := deviceService.InitDevices()
	if err != nil {
		log.Errorf("初始化设备列表失败: %v", err)
	}

	log.Infof("初始化设备列表成功: %v", deviceList)

	ds := &deviceStore{
		ctx:           ctx,
		deviceService: deviceService,
		deviceList:    deviceList,
		mu:            sync.RWMutex{},
		refreshChan:   make(chan struct{}, 1),
	}

	go ds.triggerRefreshLoop()

	return ds
}

// triggerRefreshLoop 触发刷新设备列表
func (s *deviceStore) triggerRefreshLoop() {
	for {
		select {
		case <-s.refreshChan:
			if err := s.RefreshDeviceList(); err != nil {
				log.Errorf("触发刷新设备列表失败: %v", err)
			}
		case <-s.ctx.Done():
			log.Info("设备存储停止刷新")
			return
		}
	}
}

// GetDeviceList 获取设备列表
func (s *deviceStore) GetDeviceList() []models.DeviceModel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deviceList
}

// RefreshDeviceList 刷新设备列表
func (s *deviceStore) RefreshDeviceList() error {
	deviceList, err := s.deviceService.InitDevices()
	if err != nil {
		log.Errorf("刷新设备列表失败: %v", err)
		return err
	}
	s.mu.Lock()
	s.deviceList = deviceList
	s.mu.Unlock()
	return nil
}

// GetRefreshChan 获取刷新通道
func (s *deviceStore) GetRefreshChan() chan<- struct{} {
	return s.refreshChan
}
