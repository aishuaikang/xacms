package routes

import (
	"context"
	"net/http"
	"uav_defender/internal/cache"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"gorm.io/gorm"
)

// DeviceHandler 设备处理器
type DeviceHandler struct {
	Ctx           context.Context
	DeviceService services.DeviceService
	CommonService services.CommonService

	DevicesCache cache.DevicesCache
}

// RegisterRoutes 注册设备相关路由
func (h *DeviceHandler) RegisterRoutes(router *gin.RouterGroup) {
	deviceGroup := router.Group("/devices")

	deviceGroup.GET("", h.GetDevices)
	deviceGroup.POST("", h.CreateDevice)
	deviceGroup.GET("/:id", h.GetDevice)
	deviceGroup.PUT("/:id", h.UpdateDevice)
	deviceGroup.DELETE("/:id", h.DeleteDevice)
}

// GetDevices 获取设备列表
func (h *DeviceHandler) GetDevices(c *gin.Context) {
	// 获取设备列表
	var devices []models.DeviceModel
	if err := h.CommonService.GetItems(&devices); err != nil {
		// log.Errorf("获取设备列表失败: %v", err)
		global.Logger.Error("获取设备列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取设备列表失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(devices))
}

// CreateDevice 创建设备
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
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

	c.JSON(http.StatusCreated, dto.SuccessResponse(device))
}

// GetDevice 获取设备详情
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	deviceUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "设备ID格式无效"))
		return
	}

	// 获取设备
	var device models.DeviceModel
	if err := h.CommonService.GetItemByID(deviceUUID, &device); err != nil {
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
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	deviceUUID, err := uuid.Parse(id)
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
	device, err := h.DeviceService.UpdateDevice(deviceUUID, req)
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

	c.JSON(http.StatusOK, dto.SuccessResponse(device))
}

// DeleteDevice 删除设备
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	deviceUUID, err := uuid.Parse(id)
	if err != nil {
		// return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "设备ID格式无效"))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "设备ID格式无效"))
		return
	}

	// 删除设备
	if err := h.CommonService.DeleteItemByID(&models.DeviceModel{}, deviceUUID); err != nil {
		global.Logger.Error("删除设备失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "删除设备失败"))
		return
	}

	h.DevicesCache.NotifyRefresh()

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}
