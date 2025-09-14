package routes

import (
	"bufio"
	"context"
	"fmt"
	"time"
	"xacms/internal/cache"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/valyala/fasthttp"
)

type SSEHandler struct {
	Ctx                 context.Context
	DevicesCache        cache.DevicesCache
	FPVWarningDataCache cache.FPVWarningDataCache
	ParseCache          cache.ParseCache
}

// RegisterRoutes 注册SSE相关路由
func (s *SSEHandler) RegisterRoutes(router fiber.Router) {
	sseGroup := router.Group("/sse").Name("SSE管理.")
	sseGroup.Get("/device", s.DeviceInfoListSSE).Name("实时获取设备信息")
	sseGroup.Get("/fpv", s.FPVWarningDataListSSE).Name("实时获取FPV警告数据")
	sseGroup.Get("/parse", s.ParseDataListSSE).Name("实时获取Parse数据")
}

// DeviceInfoListSSE 使用 SSE 实时获取设备信息
func (h *SSEHandler) DeviceInfoListSSE(c *fiber.Ctx) error {
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
				devices := h.DevicesCache.GetDevices()
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

// FPVWarningDataListSSE 使用 SSE 实时获取 FPV 警告数据
func (s *SSEHandler) FPVWarningDataListSSE(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.Ctx.Done():
				return

			case <-ticker.C:
				// TODO: 仅发送变化的设备数据以优化性能
				// devices := h.DeviceStore.GetDeviceList()

				fpvWarningDataList := s.FPVWarningDataCache.GetFPVWarningDataList()
				data, err := sonic.Marshal(fpvWarningDataList)
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

// ParseDataListSSE 使用 SSE 实时获取 Parse 数据
func (h *SSEHandler) ParseDataListSSE(c *fiber.Ctx) error {

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
				parseDataList := h.ParseCache.GetParseDataList()
				data, err := sonic.Marshal(parseDataList)
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
