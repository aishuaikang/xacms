package routes

import (
	"context"
	"net/http"
	"time"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"
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
	droneTargetGroup.DELETE("", h.DeleteMultipleDroneTargets)
	droneTargetGroup.GET("/export/csv", h.ExportDroneTargetsCSV)

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
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "无人机目标ID格式无效"))
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

// DeleteMultipleDroneTargets 批量删除无人机目标
func (h *DroneTargetHandler) DeleteMultipleDroneTargets(c *gin.Context) {
	var req dto.DeleteMultipleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}
	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "请提供至少一个无人机目标ID"))
		return
	}
	if err := h.CommonService.DeleteItemsByIDs(&models.DroneTargetModel{}, req); err != nil {
		global.Logger.Error("批量删除无人机目标失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "批量删除无人机目标失败"))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// ExportDroneTargetsCSV 导出无人机目标为 CSV 文件
func (h *DroneTargetHandler) ExportDroneTargetsCSV(c *gin.Context) {
	var req dto.DroneTargetExportRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 获取无人机目标列表
	droneTargets, err := h.DroneTargetService.GetAllDroneTargets(req)
	if err != nil {
		global.Logger.Error("获取无人机目标列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取无人机目标列表失败"))
		return
	}

	// 生成 CSV 文件
	csvData, err := utils.GenerateDroneTargetTableCSVByLang(droneTargets, req.Lang)
	if err != nil {
		global.Logger.Error("生成 CSV 文件失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "生成 CSV 文件失败"))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=drone_targets_"+time.Now().Format(time.DateTime)+".csv")
	c.String(http.StatusOK, string(csvData))
}
