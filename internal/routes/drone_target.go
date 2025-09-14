package routes

import (
	"context"
	"xacms/internal/dto"
	"xacms/internal/models"
	"xacms/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/datatypes"
)

// DroneTargetHandler 无人机目标处理器
type DroneTargetHandler struct {
	Ctx                context.Context
	CommonService      services.CommonService
	DroneTargetService services.DroneTargetService
}

// RegisterRoutes 注册无人机目标相关路由
func (h *DroneTargetHandler) RegisterRoutes(router fiber.Router) {
	droneTargetGroup := router.Group("/drone-targets").Name("无人机目标管理.")

	droneTargetGroup.Get("", h.GetDroneTargets).Name("获取无人机目标列表")
	droneTargetGroup.Delete("/:id<guid>", h.DeleteDroneTarget).Name("删除无人机目标")
	// 导出无人机目标CSV
	// droneTargetGroup.Get("/export", h.ExportDroneTargets).Name("导出无人机目标CSV")
	// 根据获取无人机目标轨迹

}

// GetDroneTargets 获取无人机目标列表
func (h *DroneTargetHandler) GetDroneTargets(c *fiber.Ctx) error {
	// 解析查询参数
	var req dto.DroneTargetQueryRequest
	err := h.CommonService.ValidateQuery(c, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, err.Error()))
	}

	// 获取无人机目标列表
	droneTargets, err := h.DroneTargetService.GetDroneTargets(req)
	if err != nil {
		log.Errorf("获取无人机目标列表失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "获取无人机目标列表失败"))
	}

	return c.JSON(dto.SuccessResponse(droneTargets))
}

// DeleteDroneTarget 删除无人机目标
func (h *DroneTargetHandler) DeleteDroneTarget(c *fiber.Ctx) error {
	id := c.Params("id")

	// 验证 UUID 格式
	userUUID := datatypes.UUID{}
	if err := userUUID.Scan(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse(fiber.StatusBadRequest, "用户ID格式无效"))
	}

	log.Infof("删除无人机目标 ID: %s", userUUID.String())

	// 删除无人机目标
	if err := h.CommonService.DeleteItemByID(&models.DroneTargetModel{}, userUUID); err != nil {
		log.Errorf("删除无人机目标失败: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse(fiber.StatusInternalServerError, "删除无人机目标失败"))
	}

	return c.JSON(dto.SuccessResponse(nil))
}
