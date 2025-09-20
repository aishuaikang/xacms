package routes

import (
	"net/http"
	"uav_defender/internal/dto"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
)

// UserPublicRouter 用户处理器
type UserPublicRouter struct {
	CommonService services.CommonService
	UserService   services.UserService
}

// RegisterRoutes 注册用户相关路由
func (h *UserPublicRouter) RegisterRoutes(router *gin.RouterGroup) {
	// 用户登陆
	router.POST("/login", h.Login)
}

func (h *UserPublicRouter) Login(c *gin.Context) {
	// 解析请求体
	var req dto.LoginRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	userInfo, err := h.UserService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse(http.StatusUnauthorized, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(userInfo))
}
