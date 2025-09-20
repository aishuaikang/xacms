package routes

import (
	"net/http"
	"uav_defender/internal/dto"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// WhitelistRouter 白名单路由器
type WhitelistRouter struct {
	CommonService    services.CommonService
	WhitelistService services.WhitelistService
}

// RegisterRoutes 注册白名单相关路由
func (h *WhitelistRouter) RegisterRoutes(router *gin.RouterGroup) {
	whitelistGroup := router.Group("/whitelists")

	// 获取白名单列表
	whitelistGroup.GET("", h.GetWhitelists)
	// 添加白名单
	whitelistGroup.POST("", h.AddWhitelist)
	// 根据Serial删除白名单
	whitelistGroup.DELETE("/by-serial/:serial", h.DeleteWhitelistBySerial)
}

// GetWhitelists 获取白名单列表
func (h *WhitelistRouter) GetWhitelists(c *gin.Context) {
	var req dto.WhitelistQueryRequest
	if err := h.CommonService.ValidateQuery(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	whitelist, err := h.WhitelistService.GetWhitelists(req)
	if err != nil {
		global.Logger.Error("获取白名单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取白名单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(whitelist))
}

// AddWhitelist 添加白名单
func (h *WhitelistRouter) AddWhitelist(c *gin.Context) {
	var req dto.WhitelistCreateRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	whitelist, err := h.WhitelistService.AddWhitelist(req)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "白名单已存在"))
				return
			}
		}

		global.Logger.Error("添加白名单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "添加白名单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(whitelist))
}

// 根据Serial删除白名单
func (h *WhitelistRouter) DeleteWhitelistBySerial(c *gin.Context) {
	serial := c.Param("serial")

	if err := h.WhitelistService.DeleteWhitelistBySerial(serial); err != nil {
		global.Logger.Error("删除白名单失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "删除白名单失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}
