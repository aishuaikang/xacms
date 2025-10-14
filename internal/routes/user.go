package routes

import (
	"net/http"
	"strconv"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

// UserRouter 用户路由器
type UserRouter struct {
	CommonService services.CommonService
	UserService   services.UserService
}

// RegisterRoutes 注册用户相关路由
func (h *UserRouter) RegisterRoutes(router *gin.RouterGroup) {
	userGroup := router.Group("/users")

	userGroup.GET("", h.GetUsers)
	userGroup.POST("", h.CreateUser)
	userGroup.GET("/:id", h.GetUser)
	userGroup.PUT("/:id", h.UpdateUser)
	userGroup.DELETE("/:id", h.DeleteUser)
	userGroup.POST("/:id/role", h.AssignRole)
	userGroup.POST("/:id/password", h.ChangePassword)

	// 获取所有用户列表
	userGroup.GET("/all", h.GetUsersAll)
}

// GetUsers 获取用户列表
func (h *UserRouter) GetUsers(c *gin.Context) {
	// 解析查询参数
	var req dto.UserQueryRequest
	if err := h.CommonService.ValidateQuery(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 获取用户列表
	users, err := h.UserService.GetUsers(req)
	if err != nil {
		global.Logger.Error("获取用户列表失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(users))
}

// GetUsersAll 获取所有用户列表
func (h *UserRouter) GetUsersAll(c *gin.Context) {
	var users []models.User
	// 获取所有用户
	err := h.CommonService.GetItems(&users)
	if err != nil {
		global.Logger.Error("获取所有用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取所有用户失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(users))
}

// CreateUser 创建用户
func (h *UserRouter) CreateUser(c *gin.Context) {
	// 解析请求体
	var req dto.CreateUserRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 创建用户
	user, err := h.UserService.CreateUser(req)
	if err != nil {
		global.Logger.Error("创建用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "创建用户失败"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(user))
}

// GetUser 获取用户详情
func (h *UserRouter) GetUser(c *gin.Context) {
	id := c.Param("id")

	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "用户ID格式无效"))
		return
	}

	// 获取用户
	var user models.User
	if err := h.CommonService.GetItemByID(userID, &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponse(http.StatusNotFound, "用户不存在"))
			return
		}
		global.Logger.Error("获取用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "获取用户失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(user))
}

// UpdateUser 更新用户
func (h *UserRouter) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "用户ID格式无效"))
		return
	}

	// 解析请求体
	var req dto.UpdateUserRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 更新用户
	user, err := h.UserService.UpdateUser(userID, req)
	if err != nil {
		global.Logger.Error("更新用户失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "更新用户失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(user))
}

// DeleteUser 删除用户
func (h *UserRouter) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "用户ID格式无效"))
		return
	}

	// 删除用户
	if err := h.UserService.DeleteUser(userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

// AssignRole 分配角色
func (h *UserRouter) AssignRole(c *gin.Context) {
	id := c.Param("id")

	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "用户ID格式无效"))
		return
	}

	// 解析请求体
	var req dto.AssignRoleRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 分配角色
	user, err := h.UserService.AssignRole(userID, req)
	if err != nil {
		global.Logger.Error("分配角色失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "分配角色失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(user))
}

// ChangePassword 修改用户密码
func (h *UserRouter) ChangePassword(c *gin.Context) {
	id := c.Param("id")

	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, "用户ID格式无效"))
		return
	}

	// 解析请求体
	var req dto.ChangePasswordRequest
	if err := h.CommonService.ValidateBody(c, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(http.StatusBadRequest, err.Error()))
		return
	}

	// 修改密码
	if err := h.UserService.ChangePassword(userID, req); err != nil {
		global.Logger.Error("修改密码失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(http.StatusInternalServerError, "修改密码失败"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}
