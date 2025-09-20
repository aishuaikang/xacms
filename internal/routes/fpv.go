package routes

import (
	"fmt"
	"net/http"
	"time"
	fpv_fsm "uav_defender/internal/app/devices/fms/fpv"
	"uav_defender/internal/cache"
	"uav_defender/internal/dto"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// FPVRouter 白名单路由器
type FPVRouter struct {
	CommonService services.CommonService
	FPVService    services.FPVService
	DevicesCache  cache.DevicesCache
}

// RegisterRoutes 注册白名单相关路由
func (h *FPVRouter) RegisterRoutes(router *gin.RouterGroup) {
	whitelistGroup := router.Group("/fpvs")

	// 获取FPV列表
	whitelistGroup.GET("", h.GetFPVs)
	// 进入和退出凝视模式使用sse
	whitelistGroup.GET("/sse", h.SSE)

}

// GetFPVs 获取FPV列表
func (h *FPVRouter) GetFPVs(c *gin.Context) {
	var req dto.FPVQueryRequest
	if err := h.CommonService.ValidateQuery(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	fpvs, err := h.FPVService.GetFPVs(req)
	if err != nil {
		global.Logger.Error("获取FPV列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取FPV列表失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(fpvs))
}

// SSE 进入和退出凝视模式
func (h *FPVRouter) SSE(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	var req dto.FPVSSERequest
	if err := h.CommonService.ValidateQuery(c, &req); err != nil {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error())
		c.Writer.Flush()
		return
	}

	device, exists := h.DevicesCache.GetDeviceByDetectionID(req.DetectionID)
	if !exists {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "设备不存在")
		c.Writer.Flush()
		return
	}

	if err := device.FPVFsm.Event(c, string(fpv_fsm.EventToGazing), req.Addr, req.Frequency); err != nil {
		global.Logger.Error("进入凝视模式失败", zap.Error(err))
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", fmt.Sprintf("进入凝视模式失败: %v", err))
		c.Writer.Flush()
		return
	}

	global.Logger.Info("设备进入凝视模式", zap.Int("detection_id", req.DetectionID), zap.Int("freq", req.Frequency), zap.String("addr", req.Addr))

	defer func() {
		if err := device.FPVFsm.Event(c, string(fpv_fsm.EventToScanning)); err != nil {
			global.Logger.Error("进入扫频模式失败", zap.Error(err))
			fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", fmt.Sprintf("进入扫频模式失败: %v", err))
			c.Writer.Flush()
			return
		}
		global.Logger.Info("设备退出凝视模式，进入扫频模式", zap.Int("detection_id", req.DetectionID), zap.Int("freq", req.Frequency), zap.String("addr", req.Addr))
	}()

	url := fmt.Sprintf("%s/stream_%d", config.AppConfig.Configuration.StreamMediaUrl, req.DetectionID)

	// 成功连接后，立即发送一次数据
	fmt.Fprintf(c.Writer, "data: %s\n\n", url)
	c.Writer.Flush()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		// 客户端断开链接后退出
		select {
		case <-c.Done():
			return
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(c.Writer, "data: %s\n\n", url)
			c.Writer.Flush()
		}
	}
}
