package store

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
	"xacms/internal/models"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2/log"
)

type CommonStore interface {
	GetDecryptToken() *string
	GetDeviceList() []models.DeviceModel
	SetDeviceList([]models.DeviceModel)
	GetDeviceParseIDByParseIP(ip string) (int, bool)
	GetDeviceDetectionIDByParseID(parseID int) (int, bool)
	GetDeviceByDetectionID(detectionID int) (*models.DeviceModel, bool)
}

type commonStore struct {
	ctx context.Context

	decryptTokenMu sync.RWMutex
	decryptToken   *string

	deviceList   []models.DeviceModel
	deviceListMu sync.RWMutex
}

func NewCommonStore(ctx context.Context) CommonStore {
	commonStore := &commonStore{
		ctx:            ctx,
		decryptTokenMu: sync.RWMutex{},
		decryptToken:   nil,
	}

	commonStore.refreshDecryptToken()
	go commonStore.startRefreshDecryptToken()

	return commonStore
}

// startRefreshDecryptToken 每24小时刷新一次解密token
func (s *commonStore) startRefreshDecryptToken() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.refreshDecryptToken()
		}
	}
}

type TokenData struct {
	Token  string   `json:"token"`
	Orders []string `json:"orders"`
}

type LoginResponse struct {
	Success bool      `json:"success"`
	Msg     string    `json:"msg"`
	Data    TokenData `json:"data"`
}

// refreshDecryptToken 刷新解密token
func (s *commonStore) refreshDecryptToken() {
	// 构造请求 URL
	url := fmt.Sprintf("http://101.227.171.238:5000/api/login?username=%s&password=%s", "zkzp", "askewrp23k2j")

	// 发起 GET 请求
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("请求登录接口失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("读取响应内容失败: %v", err)
		return
	}

	var result LoginResponse
	if err := sonic.Unmarshal(body, &result); err != nil {
		log.Errorf("解析响应内容失败: %v", err)
		return
	}

	if result.Success {
		s.setDecryptToken(result.Data.Token)
	} else {
		log.Errorf("登录失败: %s", result.Msg)
	}
}

// setDecryptToken 设置解密token
func (s *commonStore) setDecryptToken(token string) {
	s.decryptTokenMu.Lock()
	defer s.decryptTokenMu.Unlock()
	s.decryptToken = &token
}

// GetDecryptToken 获取解密token
func (s *commonStore) GetDecryptToken() *string {
	s.decryptTokenMu.RLock()
	defer s.decryptTokenMu.RUnlock()
	return s.decryptToken
}

// GetDeviceList 获取设备列表
func (s *commonStore) GetDeviceList() []models.DeviceModel {
	s.deviceListMu.RLock()
	defer s.deviceListMu.RUnlock()
	return s.deviceList
}

// SetDeviceList 设置设备列表
func (s *commonStore) SetDeviceList(deviceList []models.DeviceModel) {
	s.deviceListMu.Lock()
	defer s.deviceListMu.Unlock()
	s.deviceList = deviceList
}

// 根据解析IP获取解析ID
func (s *commonStore) GetDeviceParseIDByParseIP(ip string) (int, bool) {
	s.deviceListMu.RLock()
	defer s.deviceListMu.RUnlock()

	for _, device := range s.GetDeviceList() {
		if device.ParseIP == ip {
			return device.ParseID, true
		}
	}
	return 0, false
}

// 根据解析ID获取设备侦测ID
func (s *commonStore) GetDeviceDetectionIDByParseID(parseID int) (int, bool) {
	s.deviceListMu.RLock()
	defer s.deviceListMu.RUnlock()

	for _, device := range s.GetDeviceList() {
		if device.ParseID == parseID {
			return device.DetectionID, true
		}
	}
	return 0, false
}

// 根据侦测ID获取设备信息
func (s *commonStore) GetDeviceByDetectionID(detectionID int) (*models.DeviceModel, bool) {
	s.deviceListMu.RLock()
	defer s.deviceListMu.RUnlock()
	for _, device := range s.GetDeviceList() {
		if device.DetectionID == detectionID {
			return &device, true
		}
	}
	return nil, false
}
