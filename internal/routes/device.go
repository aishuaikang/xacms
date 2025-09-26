package routes

import (
	"context"
	"net/http"
	"strconv"
	"uav_defender/internal/cache"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// DeviceRouter 设备处理器
type DeviceRouter struct {
	Ctx           context.Context
	DeviceService services.DeviceService
	CommonService services.CommonService

	DevicesCache cache.DevicesCache
}

// RegisterRoutes 注册设备相关路由
func (h *DeviceRouter) RegisterRoutes(router *gin.RouterGroup) {
	deviceGroup := router.Group("/devices")

	deviceGroup.GET("", h.GetDevices)
	deviceGroup.POST("", h.CreateDevice)
	deviceGroup.GET("/:id", h.GetDevice)
	deviceGroup.PUT("/:id", h.UpdateDevice)
	deviceGroup.DELETE("/:id", h.DeleteDevice)
}

// GetDevices 获取设备列表
func (h *DeviceRouter) GetDevices(c *gin.Context) {
	// 获取设备列表
	var devices []models.Device
	if err := h.CommonService.GetItems(&devices); err != nil {
		global.Logger.Error("获取设备列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取设备列表失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(devices))
}

// CreateDevice 创建设备
func (h *DeviceRouter) CreateDevice(c *gin.Context) {
	// 解析请求体
	var req dto.CreateDeviceRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return

	}

	// 创建设备
	device, err := h.DeviceService.CreateDevice(req)
	if err != nil {
		global.Logger.Error("创建设备失败", zap.Error(err))

		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "设备已存在"))
				return
			}
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "创建设备失败"))
		return
	}

	h.DevicesCache.NotifyRefresh()
	h.DeviceService.RefreshMediaMtxConfig()
	c.JSON(http.StatusCreated, dto.SuccessResponse(device))
}

// GetDevice 获取设备详情
func (h *DeviceRouter) GetDevice(c *gin.Context) {
	id := c.Param("id")

	// 验证 ID 格式
	deviceID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "设备ID格式无效"))
		return
	}

	// 获取设备
	var device models.Device
	if err := h.CommonService.GetItemByID(uint(deviceID), &device); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse(http.StatusNotFound, "设备不存在"))
			return
		}
		global.Logger.Error("获取设备失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取设备失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(device))
}

// UpdateDevice 更新设备
func (h *DeviceRouter) UpdateDevice(c *gin.Context) {
	id := c.Param("id")

	// 验证 ID 格式
	deviceID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "设备ID格式无效"))
		return
	}

	// 解析请求体
	var req dto.UpdateDeviceRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 更新设备
	device, err := h.DeviceService.UpdateDevice(uint(deviceID), req)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "设备已存在"))
				return
			}
		}

		global.Logger.Error("更新设备失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "更新设备失败"))
		return
	}

	h.DevicesCache.NotifyRefresh()
	h.DeviceService.RefreshMediaMtxConfig()

	c.JSON(http.StatusOK, dto.SuccessResponse(device))
}

// DeleteDevice 删除设备
func (h *DeviceRouter) DeleteDevice(c *gin.Context) {
	id := c.Param("id")

	// 验证 ID 格式
	deviceID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "设备ID格式无效"))
		return
	}

	// 删除设备
	if err := h.CommonService.DeleteItemByID(&models.Device{}, uint(deviceID)); err != nil {
		global.Logger.Error("删除设备失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "删除设备失败"))
		return
	}

	h.DevicesCache.NotifyRefresh()
	h.DeviceService.RefreshMediaMtxConfig()

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}
