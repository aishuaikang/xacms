package routes

import (
	"bufio"
	"context"
	"fmt"
	"time"
	"xacms/internal/models"
	"xacms/internal/routes/dto"
	"xacms/internal/services"
	"xacms/internal/store"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

// DeviceHandler 设备处理器
type DeviceHandler struct {
	Ctx           context.Context
	DeviceService services.DeviceService
	CommonService services.CommonService
	DeviceStore   store.DeviceStore
	FPVStore      store.FPVStore
}

// RegisterRoutes 注册设备相关路由
func (h *DeviceHandler) RegisterRoutes(router fiber.Router) {
	deviceGroup := router.Group("/devices").Name("设备管理.")

	deviceGroup.Get("", h.GetDevices).Name("获取设备列表")
	deviceGroup.Post("", h.CreateDevice).Name("创建设备")
	deviceGroup.Get("/:id<guid>", h.GetDevice).Name("获取设备详情")
	deviceGroup.Put("/:id<guid>", h.UpdateDevice).Name("更新设备")
	deviceGroup.Delete("/:id<guid>", h.DeleteDevice).Name("删除设备")

	// 使用 sse 实时获取设备信息
	deviceGroup.Get("/sse", h.DeviceListSSE).Name("实时获取设备信息")

	// 实时获取 fpv 警告数据
	deviceGroup.Get("/fpv/sse", h.FPVWaringDataListSSE).Name("实时获取FPV警告数据")
}

// GetDevices 获取设备列表
func (h *DeviceHandler) GetDevices(c *fiber.Ctx) error {
	// 获取设备列表
	var devices []models.DeviceModel
	if err := h.CommonService.GetItems(&devices); err != nil {
		log.Errorf("获取设备列表失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取设备列表失败"))
	}

	return c.JSON(dto.SuccessResponse(devices))
}

// CreateDevice 创建设备
func (h *DeviceHandler) CreateDevice(c *fiber.Ctx) error {
	// 解析请求体
	var req dto.CreateDeviceRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, err.Error()))
	}

	// 创建设备
	device, err := h.DeviceService.CreateDevice(req)
	if err != nil {
		log.Errorf("创建设备失败: %v", err)

		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "设备已存在"))
			}
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "创建设备失败"))
	}

	h.DeviceStore.GetRefreshChan() <- struct{}{}

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse(device))
}

// GetDevice 获取设备详情
func (h *DeviceHandler) GetDevice(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	deviceUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "设备ID格式无效"))
	}

	// 获取设备
	var device models.DeviceModel
	if err := h.CommonService.GetItemByID(deviceUUID, &device); err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse(fiber.StatusNotFound, "设备不存在"))
		}
		log.Errorf("获取设备失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取设备失败"))
	}

	return c.JSON(dto.SuccessResponse(device))
}

// UpdateDevice 更新设备
func (h *DeviceHandler) UpdateDevice(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	deviceUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "设备ID格式无效"))
	}

	// 解析请求体
	var req dto.UpdateDeviceRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, err.Error()))
	}

	// 更新设备
	device, err := h.DeviceService.UpdateDevice(deviceUUID, req)
	if err != nil {
		log.Errorf("更新设备失败: %v", err)
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "设备已存在"))
			}
		}

		log.Errorf("更新设备失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "更新设备失败"))
	}

	h.DeviceStore.GetRefreshChan() <- struct{}{}

	return c.JSON(dto.SuccessResponse(device))
}

// DeleteDevice 删除设备
func (h *DeviceHandler) DeleteDevice(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	deviceUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "设备ID格式无效"))
	}

	// 删除设备
	if err := h.CommonService.DeleteItemByID(&models.DeviceModel{}, deviceUUID); err != nil {
		log.Errorf("删除设备失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "删除设备失败"))
	}

	h.DeviceStore.GetRefreshChan() <- struct{}{}

	return c.JSON(dto.SuccessResponse(nil))
}

// DeviceListSSE 使用 SSE 实时获取设备信息
func (h *DeviceHandler) DeviceListSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-h.Ctx.Done():
				return

			case <-ticker.C:
				// TODO: 仅发送变化的设备数据以优化性能
				devices := h.DeviceStore.GetDeviceList()
				data, err := sonic.Marshal(devices)
				if err != nil {
					fmt.Fprintf(w, "data: {\"error\":\"marshal failed\"}\n\n")
				} else {
					fmt.Fprintf(w, "data: %s\n\n", data)
				}

				err = w.Flush()
				if err != nil {
					log.Errorf("刷新连接时发生错误: %v. 关闭 SSE 连接", err)
					return
				}
			}
		}
	}))

	return nil
}

// FPVWaringDataListSSE 使用 SSE 实时获取 FPV 警告数据
func (h *DeviceHandler) FPVWaringDataListSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-h.Ctx.Done():
				return

			case <-ticker.C:
				// TODO: 仅发送变化的设备数据以优化性能
				// devices := h.DeviceStore.GetDeviceList()

				fpvWaringDataList := h.FPVStore.GetFPVWaringDataList()
				data, err := sonic.Marshal(fpvWaringDataList)
				if err != nil {
					fmt.Fprintf(w, "data: {\"error\":\"marshal failed\"}\n\n")
				} else {
					fmt.Fprintf(w, "data: %s\n\n", data)
				}

				err = w.Flush()
				if err != nil {
					log.Errorf("刷新连接时发生错误: %v. 关闭 SSE 连接", err)
					return
				}
			}
		}
	}))

	return nil
}
