package routes

import (
	"bufio"
	"fmt"
	"time"
	"xacms/internal/models"
	"xacms/internal/routes/dto"
	"xacms/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

// DeviceHandler 设备处理器
type DeviceHandler struct {
	DeviceService services.DeviceService
	CommonService services.CommonService

	deviceList []models.DeviceModel
}

func NewDeviceHandler(deviceService services.DeviceService, commonService services.CommonService) *DeviceHandler {
	// 初始化设备列表
	deviceList, err := deviceService.InitDevices()
	if err != nil {
		log.Errorf("初始化设备列表失败: %v", err)
	}

	// log.Infof("初始化设备列表成功: %v", deviceList)

	return &DeviceHandler{
		DeviceService: deviceService,
		CommonService: commonService,
		deviceList:    deviceList,
	}
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
	deviceGroup.Get("/sse", h.DeviceSSE).Name("实时获取设备信息")
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

// CreateDevice 创建用户
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

	return c.JSON(dto.SuccessResponse(nil))
}

// DeviceSSE 使用 SSE 实时获取设备信息
func (h *DeviceHandler) DeviceSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		fmt.Println("WRITER")
		var i int
		for {
			i++

			var msg string

			// if there are messages that have been sent to the `/publish` endpoint
			// then use these first, otherwise just send the current time
			// if len(sseMessageQueue) > 0 {
			// 	msg = fmt.Sprintf("%d - message recieved: %s", i, sseMessageQueue[0])
			// 	// remove the message from the buffer
			// 	sseMessageQueue = sseMessageQueue[1:]
			// } else {
			msg = fmt.Sprintf("%d - the time is %v", i, time.Now())
			// }

			fmt.Fprintf(w, "data: Message: %s\n\n", msg)
			fmt.Println(msg)

			err := w.Flush()
			if err != nil {
				// Refreshing page in web browser will establish a new
				// SSE connection, but only (the last) one is alive, so
				// dead connections must be closed here.
				fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

				break
			}
			time.Sleep(2 * time.Second)
		}
	}))

	return nil
}
