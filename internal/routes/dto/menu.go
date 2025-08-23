package dto

import (
	"new-spbatc-drone-platform/internal/models"

	"github.com/google/uuid"
)

// CreateMenuRequest 创建菜单请求结构
type CreateMenuRequest struct {
	Name         string         `json:"name" validate:"required,min=2,max=64"`
	RouteName    string         `json:"route_name" validate:"required,min=2,max=64"`
	RoutePath    string         `json:"route_path" validate:"required,min=1,max=255"`
	ApiIds       *models.ApiIds `json:"api_paths" validate:"omitempty"`
	IsHidden     bool           `json:"is_hidden" validate:"omitempty"`
	IsFullScreen bool           `json:"is_full_screen" validate:"omitempty"`
	IsTabs       bool           `json:"is_tabs" validate:"omitempty"`
	Component    string         `json:"component" validate:"required,max=255"`
	Icon         *string        `json:"icon" validate:"omitempty,max=128"`
	Order        uint           `json:"order" validate:"omitempty,min=0"`
}

// UpdateMenuRequest 更新菜单请求结构
type UpdateMenuRequest struct {
	Name         *string        `json:"name" validate:"omitempty,min=2,max=64"`
	RouteName    *string        `json:"route_name" validate:"omitempty,min=2,max=64"`
	RoutePath    *string        `json:"route_path" validate:"omitempty,min=1,max=255"`
	ApiIds       *models.ApiIds `json:"api_paths" validate:"omitempty"`
	IsHidden     *bool          `json:"is_hidden" validate:"omitempty"`
	IsFullScreen *bool          `json:"is_full_screen" validate:"omitempty"`
	IsTabs       *bool          `json:"is_tabs" validate:"omitempty"`
	Component    *string        `json:"component" validate:"omitempty,max=255"`
	Icon         *string        `json:"icon" validate:"omitempty,max=128"`
	Order        *uint          `json:"order" validate:"omitempty,min=0"`
}

// MenuQueryRequest 菜单查询请求结构
// type MenuQueryRequest struct {
// 	BaseQueryRequest
// 	ParentID *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
// 	Type     *int       `json:"type" validate:"omitempty,oneof=1 2 3"`
// 	Status   *int       `json:"status" validate:"omitempty,oneof=0 1"`
// }

// MenuTreeRequest 菜单树查询请求结构
type MenuTreeRequest struct {
	Type   *int `json:"type" validate:"omitempty,oneof=1 2 3"`
	Status *int `json:"status" validate:"omitempty,oneof=0 1"`
}

// MoveMenuRequest 移动菜单请求结构
type MoveMenuRequest struct {
	NewParentID *uuid.UUID `json:"new_parent_id" validate:"omitempty,uuid"`
}

// ApiItem API列表项结构
type ApiItem struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name"`
}
