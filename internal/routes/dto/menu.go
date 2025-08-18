package dto

import (
	"github.com/google/uuid"
)

// CreateMenuRequest 创建菜单请求结构
type CreateMenuRequest struct {
	Name       string     `json:"name" validate:"required,min=2,max=64"`
	Path       string     `json:"path" validate:"required,min=1,max=255"`
	Icon       string     `json:"icon" validate:"omitempty,max=128"`
	Component  string     `json:"component" validate:"omitempty,max=255"`
	ParentID   *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	Sort       int        `json:"sort" validate:"omitempty,min=0"`
	Type       int        `json:"type" validate:"required,oneof=1 2 3"` // 1:目录 2:菜单 3:按钮
	Permission string     `json:"permission" validate:"omitempty,max=128"`
	Hidden     bool       `json:"hidden" validate:"omitempty"`
	Status     int        `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateMenuRequest 更新菜单请求结构
type UpdateMenuRequest struct {
	Name       *string `json:"name" validate:"omitempty,min=2,max=64"`
	Path       *string `json:"path" validate:"omitempty,min=1,max=255"`
	Icon       *string `json:"icon" validate:"omitempty,max=128"`
	Component  *string `json:"component" validate:"omitempty,max=255"`
	Sort       *int    `json:"sort" validate:"omitempty,min=0"`
	Type       *int    `json:"type" validate:"omitempty,oneof=1 2 3"`
	Permission *string `json:"permission" validate:"omitempty,max=128"`
	Hidden     *bool   `json:"hidden" validate:"omitempty"`
	Status     *int    `json:"status" validate:"omitempty,oneof=0 1"`
}

// MenuQueryRequest 菜单查询请求结构
type MenuQueryRequest struct {
	BaseQueryRequest
	ParentID *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	Type     *int       `json:"type" validate:"omitempty,oneof=1 2 3"`
	Status   *int       `json:"status" validate:"omitempty,oneof=0 1"`
}

// MenuTreeRequest 菜单树查询请求结构
type MenuTreeRequest struct {
	Type   *int `json:"type" validate:"omitempty,oneof=1 2 3"`
	Status *int `json:"status" validate:"omitempty,oneof=0 1"`
}

// MoveMenuRequest 移动菜单请求结构
type MoveMenuRequest struct {
	NewParentID *uuid.UUID `json:"new_parent_id" validate:"omitempty,uuid"`
}
