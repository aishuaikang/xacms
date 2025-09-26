package dto

import (
	"uav_defender/internal/models"
)

// UserQueryRequest 用户查询请求结构
type UserQueryRequest struct {
	BaseQueryRequest
	// Status       *models.Status `form:"status" validate:"omitempty,oneof=0 1"`
	// RoleID       *uint `form:"role_id" validate:"omitempty,uuid"`
	// TenantID     *uint `form:"tenant_id" validate:"omitempty,uuid"`
	// DepartmentID *uint `form:"department_id" validate:"omitempty,uuid"`
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Nickname string  `json:"nickname" validate:"required,min=2,max=64"`
	Username string  `json:"username" validate:"required,min=3,max=64"`
	Password string  `json:"password" validate:"required,min=6,max=128"`
	Email    string  `json:"email" validate:"required,email,max=128"`
	Phone    string  `json:"phone" validate:"required,phone"`
	Avatar   *string `json:"avatar" validate:"omitempty,max=255"`
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Nickname *string `json:"nickname" validate:"omitempty,min=2,max=64"`
	Username *string `json:"username" validate:"omitempty,min=3,max=64"`
	Email    *string `json:"email" validate:"omitempty,email,max=128"`
	Phone    *string `json:"phone" validate:"omitempty,phone"`
	Avatar   *string `json:"avatar" validate:"omitempty,max=255"`
}

// AssignRoleRequest 分配角色请求结构
type AssignRoleRequest struct {
	RoleID uint `json:"role_id" validate:"required"`
}

// // ChangePasswordRequest 修改密码请求结构
// type ChangePasswordRequest struct {
// 	OldPassword string `json:"old_password" validate:"required,min=6,max=128"`
// 	NewPassword string `json:"new_password" validate:"required,min=6,max=128"`
// }

// // ResetPasswordRequest 重置密码请求结构
// type ResetPasswordRequest struct {
// 	UserID      uint `json:"user_id" validate:"required,uuid"`
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
