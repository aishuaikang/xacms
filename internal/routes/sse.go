package routes

import (
	"context"
	"fmt"
	"time"
	"uav_defender/internal/cache"
	"uav_defender/internal/pkg/global"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
)

type SSEHandler struct {
	Ctx                 context.Context
	DevicesCache        cache.DevicesCache
	FPVWarningDataCache cache.FPVWarningDataCache
	ParseCache          cache.ParseCache
}

// RegisterRoutes 注册SSE相关路由
func (s *SSEHandler) RegisterRoutes(router *gin.RouterGroup) {
	sseGroup := router.Group("/sse")
	sseGroup.GET("/device", s.DeviceInfoListSSE)
	sseGroup.GET("/fpv", s.FPVWarningDataListSSE)
	sseGroup.GET("/parse", s.ParseDataListSSE)
}

// DeviceInfoListSSE 使用 SSE 实时获取设备信息
func (h *SSEHandler) DeviceInfoListSSE(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		// 客户端断开链接后退出
		select {
		case <-h.Ctx.Done():
			global.Logger.Info("服务器关闭，停止发送SSE")
			return
		case <-c.Request.Context().Done():
			global.Logger.Info("客户端断开连接，停止发送SSE")
			return
		case <-ticker.C:
			devices := h.DevicesCache.GetDisplayDevices()
			data, err := sonic.Marshal(devices)
			if err != nil {
				fmt.Fprintf(c.Writer, "data: {\"error\":\"marshal failed\"}\n\n")
			} else {
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			}
			c.Writer.Flush()
		}
	}

}

// FPVWarningDataListSSE 使用 SSE 实时获取 FPV 警告数据
func (s *SSEHandler) FPVWarningDataListSSE(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		// 客户端断开链接后退出
		select {
		case <-s.Ctx.Done():
			global.Logger.Info("服务器关闭，停止发送SSE")
			return
		case <-c.Request.Context().Done():
			global.Logger.Info("客户端断开连接，停止发送SSE")
			return
		case <-ticker.C:
			fpvWarningDataList := s.FPVWarningDataCache.GetFPVWarningDataList()
			data, err := sonic.Marshal(fpvWarningDataList)
			if err != nil {
				fmt.Fprintf(c.Writer, "data: {\"error\":\"marshal failed\"}\n\n")
			} else {
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			}
			c.Writer.Flush()
		}
	}
}

// ParseDataListSSE 使用 SSE 实时获取 Parse 数据
func (h *SSEHandler) ParseDataListSSE(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		// 客户端断开链接后退出
		select {
		case <-h.Ctx.Done():
			global.Logger.Info("服务器关闭，停止发送SSE")
			return
		case <-c.Request.Context().Done():
			global.Logger.Info("客户端断开连接，停止发送SSE")
			return
		case <-ticker.C:
			parseDataList := h.ParseCache.GetParseDataList()
			data, err := sonic.Marshal(parseDataList)
			if err != nil {
				fmt.Fprintf(c.Writer, "data: {\"error\":\"marshal failed\"}\n\n")
			} else {
				fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			}

			c.Writer.Flush()
		}
	}

}
