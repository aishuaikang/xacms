package dto

import (
	"new-spbatc-drone-platform/internal/models"

	"github.com/google/uuid"
)

// UserQueryRequest 用户查询请求结构
type UserQueryRequest struct {
	BaseQueryRequest
	// Status       *models.Status `query:"status" validate:"omitempty,oneof=0 1"`
	// RoleID       *uuid.UUID     `query:"role_id" validate:"omitempty,uuid"`
	// TenantID     *uuid.UUID     `query:"tenant_id" validate:"omitempty,uuid"`
	// DepartmentID *uuid.UUID     `query:"department_id" validate:"omitempty,uuid"`
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	Nickname string         `json:"nickname" validate:"required,min=2,max=64"`
	Username string         `json:"username" validate:"required,min=3,max=64"`
	Password string         `json:"password" validate:"required,min=6,max=128"`
	Email    string         `json:"email" validate:"required,email,max=128"`
	Phone    string         `json:"phone" validate:"required,phone"`
	Avatar   *string        `json:"avatar" validate:"omitempty,max=255"`
	Status   *models.Status `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateUserRequest 更新用户请求结构
type UpdateUserRequest struct {
	Nickname *string        `json:"nickname" validate:"omitempty,min=2,max=64"`
	Username *string        `json:"username" validate:"omitempty,min=3,max=64"`
	Email    *string        `json:"email" validate:"omitempty,email,max=128"`
	Phone    *string        `json:"phone" validate:"omitempty,phone"`
	Avatar   *string        `json:"avatar" validate:"omitempty,max=255"`
	Status   *models.Status `json:"status" validate:"omitempty,oneof=0 1"`
}

// AssignRoleRequest 分配角色请求结构
type AssignRoleRequest struct {
	RoleID uuid.UUID `json:"role_id" validate:"required,uuid"`
}

// // ChangePasswordRequest 修改密码请求结构
// type ChangePasswordRequest struct {
// 	OldPassword string `json:"old_password" validate:"required,min=6,max=128"`
// 	NewPassword string `json:"new_password" validate:"required,min=6,max=128"`
// }

// // ResetPasswordRequest 重置密码请求结构
// type ResetPasswordRequest struct {
// 	UserID      uuid.UUID `json:"user_id" validate:"required,uuid"`
// 	NewPassword string    `json:"new_password" validate:"required,min=6,max=128"`
// }
