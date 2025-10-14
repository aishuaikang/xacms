package dto

import "uav_defender/internal/models"

// UserQueryRequest 用户查询请求结构
type UserQueryRequest struct {
	BaseQueryRequest
	// Status       *models.Status `form:"status" validate:"omitempty,oneof=0 1"`
	// RoleID       *int `form:"role_id" validate:"omitempty"`
	// TenantID     *int `form:"tenant_id" validate:"omitempty"`
	// DepartmentID *int `form:"department_id" validate:"omitempty"`
}

// UserQueryResponse 用户查询响应结构
type UserQueryResponse struct {
	models.User
	DeviceCount int64 `json:"device_count"` // 设备数量
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Nickname string  `json:"nickname" validate:"required,min=2,max=64"`
	Username string  `json:"username" validate:"required,min=3,max=64"`
	Password string  `json:"password" validate:"required,min=6,max=128"`
	RoleID   uint64  `json:"role_id,string" validate:"required"`
	Email    *string `json:"email" validate:"omitempty,email,max=128"`
	Phone    *string `json:"phone" validate:"omitempty,phone"`
	Avatar   *string `json:"avatar" validate:"omitempty,max=255"`
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Nickname *string `json:"nickname" validate:"omitempty,min=2,max=64"`
	Username *string `json:"username" validate:"omitempty,min=3,max=64"`
	RoleID   *uint64 `json:"role_id,string" validate:"omitempty"`
	Email    *string `json:"email" validate:"omitempty,email,max=128"`
	Phone    *string `json:"phone" validate:"omitempty,phone"`
	Avatar   *string `json:"avatar" validate:"omitempty,max=255"`
}

// AssignRoleRequest 分配角色请求结构
type AssignRoleRequest struct {
	RoleID uint64 `json:"role_id,string" validate:"required"`
}

// ChangePasswordRequest 修改密码请求结构
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=6,max=128"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=128"`
}

// // ResetPasswordRequest 重置密码请求结构
// type ResetPasswordRequest struct {
// 	UserID      int `json:"user_id" validate:"required"`
// 	NewPassword string    `json:"new_password" validate:"required,min=6,max=128"`
// }

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

type LoginResponse struct {
	models.User
	Token string `json:"token"`
}
