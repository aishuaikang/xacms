package routes

import (
	"context"
	"net/http"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/google/uuid"
)

// DroneTargetHandler 无人机目标处理器
type DroneTargetHandler struct {
	Ctx                context.Context
	CommonService      services.CommonService
	DroneTargetService services.DroneTargetService
}

// RegisterRoutes 注册无人机目标相关路由
func (h *DroneTargetHandler) RegisterRoutes(router *gin.RouterGroup) {
	droneTargetGroup := router.Group("/drone-targets")

	droneTargetGroup.GET("", h.GetDroneTargets)
	droneTargetGroup.DELETE("/:id", h.DeleteDroneTarget)
	// 导出无人机目标CSV
	// droneTargetGroup.Get("/export", h.ExportDroneTargets).Name("导出无人机目标CSV")
	// 根据获取无人机目标轨迹

}

// GetDroneTargets 获取无人机目标列表
func (h *DroneTargetHandler) GetDroneTargets(c *gin.Context) {
	// 解析查询参数
	var req dto.DroneTargetQueryRequest
	err := h.CommonService.ValidateQuery(c, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 获取无人机目标列表
	droneTargets, err := h.DroneTargetService.GetDroneTargets(req)
	if err != nil {
		// log.Errorf("获取无人机目标列表失败: %v", err)
		global.Logger.Error("获取无人机目标列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取无人机目标列表失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(droneTargets))
}

// DeleteDroneTarget 删除无人机目标
func (h *DroneTargetHandler) DeleteDroneTarget(c *gin.Context) {
	id := c.Param("id")

	// 验证 UUID 格式
	userUUID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "用户ID格式无效"))
		return
	}

	global.Logger.Info("删除无人机目标", zap.String("id", userUUID.String()))

	// 删除无人机目标
	if err := h.CommonService.DeleteItemByID(&models.DroneTargetModel{}, userUUID); err != nil {
		global.Logger.Error("删除无人机目标失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "删除无人机目标失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}
