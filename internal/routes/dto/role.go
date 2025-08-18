package dto

import (
	"github.com/google/uuid"
)

// CreateRoleRequest 创建角色请求结构
type CreateRoleRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=64"`
	Code        string     `json:"code" validate:"required,min=2,max=64"`
	Description string     `json:"description" validate:"omitempty,max=255"`
	Sort        int        `json:"sort" validate:"omitempty,min=0"`
	Status      int        `json:"status" validate:"omitempty,oneof=0 1"`
	TenantID    *uuid.UUID `json:"tenant_id" validate:"omitempty,uuid"`
}

// UpdateRoleRequest 更新角色请求结构
type UpdateRoleRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=64"`
	Code        *string `json:"code" validate:"omitempty,min=2,max=64"`
	Description *string `json:"description" validate:"omitempty,max=255"`
	Sort        *int    `json:"sort" validate:"omitempty,min=0"`
	Status      *int    `json:"status" validate:"omitempty,oneof=0 1"`
}

// RoleQueryRequest 角色查询请求结构
type RoleQueryRequest struct {
	BaseQueryRequest
	Status   *int       `json:"status" validate:"omitempty,oneof=0 1"`
	TenantID *uuid.UUID `json:"tenant_id" validate:"omitempty,uuid"`
}

// AssignPermissionsRequest 分配权限请求结构
type AssignPermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids" validate:"required,min=1"`
}
