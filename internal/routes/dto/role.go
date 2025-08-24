package dto

import (
	"github.com/google/uuid"
)

// CreateRoleRequest 创建角色请求结构
type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=64"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Order       uint   `json:"order" validate:"omitempty,min=0"`
}

// UpdateRoleRequest 更新角色请求结构
type UpdateRoleRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=64"`
	Description *string `json:"description" validate:"omitempty,max=255"`
	Order       *uint   `json:"order" validate:"omitempty,min=0"`
}

// AssignMenusRequest 分配菜单请求结构
type AssignMenusRequest struct {
	MenuIDs []uuid.UUID `json:"menu_ids" validate:"required,min=1,dive,uuid"`
}
