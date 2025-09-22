package routes

import (
	"fmt"
	"net/http"
	"time"
	"uav_defender/internal/app/devices/conn"
	fpv_fsm "uav_defender/internal/app/devices/fms/fpv"
	"uav_defender/internal/cache"
	"uav_defender/internal/dto"
	"uav_defender/internal/pkg/config"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// FPVRouter 白名单路由器
type FPVRouter struct {
	CommonService services.CommonService
	FPVService    services.FPVService
	DevicesCache  cache.DevicesCache
	FpvConnection *conn.FpvConnection
}

// RegisterRoutes 注册白名单相关路由
func (h *FPVRouter) RegisterRoutes(router *gin.RouterGroup) {
	whitelistGroup := router.Group("/fpvs")

	// 获取FPV列表
	whitelistGroup.GET("", h.GetFPVs)
	// 进入和退出凝视模式使用sse
	whitelistGroup.GET("/sse", h.SSE)
	// 设置频点
	whitelistGroup.POST("/frequency", h.SetFrequency)

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

	device, exists := h.DevicesCache.GetDeviceByID(uuid.MustParse(req.DeviceID))
	if !exists {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", "设备不存在")
		c.Writer.Flush()
		return
	}

	filename := fmt.Sprintf("%d_%d.mp4", req.Frequency, time.Now().Unix())

	if err := device.FPVFsm.Event(c, string(fpv_fsm.EventToGazing), req.DeviceID, req.Frequency, filename); err != nil {
		global.Logger.Error("进入凝视模式失败", zap.Error(err))
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", fmt.Sprintf("进入凝视模式失败: %v", err))
		c.Writer.Flush()
		return
	}

	defer func() {
		// 切换回扫频模式
		if err := device.FPVFsm.Event(c, string(fpv_fsm.EventToScanning)); err != nil {
			global.Logger.Error("进入扫频模式失败", zap.Error(err))
			fmt.Fprintf(c.Writer, "event: error\ndata: 进入扫频模式失败: %v\n\n", err)
			c.Writer.Flush()
			return
		}
		global.Logger.Info("设备退出凝视模式，进入扫频模式",
			zap.String("device_id", req.DeviceID),
			zap.Int("freq", req.Frequency),
			zap.String("filename", filename),
		)

		// 保存FPV记录到数据库
		addReq := dto.AddFPVRequest{
			DeviceID:  uuid.MustParse(req.DeviceID),
			Frequency: req.Frequency,
			FileName:  filename,
		}
		if err := h.FPVService.AddFPV(addReq); err != nil {
			global.Logger.Error("保存FPV记录失败", zap.Error(err))
			fmt.Fprintf(c.Writer, "event: error\ndata: 保存FPV记录失败\n\n")
			c.Writer.Flush()
			return
		}
		global.Logger.Info("保存FPV记录成功",
			zap.String("device_id", req.DeviceID),
			zap.Int("freq", req.Frequency),
			zap.String("filename", filename),
		)
	}()

	global.Logger.Info("设备进入凝视模式", zap.String("device_id", req.DeviceID), zap.Int("freq", req.Frequency), zap.String("filename", filename))

	// 成功连接后，立即发送一次数据

	streamKey := fmt.Sprintf("stream_%s", req.DeviceID)
	url := fmt.Sprintf("%s/%s", config.AppConfig.Configuration.StreamMediaUrl, streamKey)

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
			global.Logger.Info("发送心跳", zap.String("device_id", req.DeviceID))
			fmt.Fprintf(c.Writer, "data: %s\n\n", url)
			c.Writer.Flush()
		}
	}
}

func (h *FPVRouter) SetFrequency(c *gin.Context) {
	var req dto.SetFrequencyRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	device, exists := h.DevicesCache.GetDeviceByID(req.DeviceID)
	if !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse(http.StatusNotFound, "设备不存在"))
		return
	}

	// 必须是凝视状态才能设置点频
	if !device.FPVFsm.FSM.Is(string(fpv_fsm.StateGazing)) {
		global.Logger.Warn("设备不在凝视状态，不能设置频点", zap.String("device_id", req.DeviceID.String()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "设备不在凝视状态，不能设置频点"))
		return
	}

	conn, exists := h.FpvConnection.GetConnection(req.DeviceID)
	if !exists {
		c.JSON(http.StatusNotFound, dto.ErrorResponse(http.StatusNotFound, "FPV连接不存在"))
		return
	}

	command := 9
	fpvCommand, expectedResponse := utils.BuildFPVCommand(req.Frequency, command)

	// 发送命令
	if err := conn.SendCommand(fpvCommand); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("发送命令失败: %v", err)))
		return
	}

	// 等待响应
	response, err := conn.WaitResponse()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("等待响应失败: %v", err)))
		return
	}

	global.Logger.Info("收到响应", zap.String("address", req.DeviceID.String()), zap.String("response", response))

	// 验证响应
	if !utils.IsExpectedResponse(response, expectedResponse) {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, fmt.Sprintf("设备响应不符合预期: %q != %q", response, expectedResponse)))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse("设置频点成功"))
}
