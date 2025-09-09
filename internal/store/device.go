package store

import (
	"context"
	"xacms/internal/services"

	"github.com/gofiber/fiber/v2/log"
)

type DeviceStore interface {
	RefreshDeviceList() error
	GetRefreshChan() chan<- struct{}
}

type deviceStore struct {
	ctx           context.Context
	commonStore   CommonStore
	deviceService services.DeviceService
	refreshChan   chan struct{}
}

func NewDeviceStore(ctx context.Context, deviceService services.DeviceService, commonStore CommonStore) DeviceStore {
	// 初始化设备列表
	deviceList, err := deviceService.InitDevices()
	if err != nil {
		log.Errorf("初始化设备列表失败: %v", err)
	}
	commonStore.SetDeviceList(deviceList)
	log.Infof("初始化设备列表成功: %v", deviceList)

	ds := &deviceStore{
		ctx:           ctx,
		commonStore:   commonStore,
		deviceService: deviceService,
		refreshChan:   make(chan struct{}, 1),
	}

	go ds.triggerRefreshLoop()

	return ds
}

// triggerRefreshLoop 触发刷新设备列表
func (s *deviceStore) triggerRefreshLoop() {
	for {
		select {
		case <-s.ctx.Done():
			log.Info("设备存储停止刷新")
			return
		case <-s.refreshChan:
			if err := s.RefreshDeviceList(); err != nil {
				log.Errorf("触发刷新设备列表失败: %v", err)
			}

			log.Info("设备列表已刷新")

		}
	}
}

// RefreshDeviceList 刷新设备列表
func (s *deviceStore) RefreshDeviceList() error {
	deviceList, err := s.deviceService.InitDevices()
	if err != nil {
		log.Errorf("刷新设备列表失败: %v", err)
		return err
	}

	s.commonStore.SetDeviceList(deviceList)

	return nil
}

// GetRefreshChan 获取刷新通道
func (s *deviceStore) GetRefreshChan() chan<- struct{} {
	return s.refreshChan
}
