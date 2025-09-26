package dto

import (
	"uav_defender/internal/models"
)

// CreateMenuRequest 创建菜单请求结构
type CreateMenuRequest struct {
	ParentID     *uint            `json:"parent_id" validate:"omitempty"`
	Name         string           `json:"name" validate:"required,min=2,max=64"`
	RouteName    string           `json:"route_name" validate:"required,min=2,max=64"`
	RoutePath    string           `json:"route_path" validate:"required,min=1,max=255"`
	ApiNames     *models.ApiNames `json:"api_names" validate:"omitempty"`
	IsHidden     bool             `json:"is_hidden" validate:"omitempty"`
	IsFullScreen bool             `json:"is_full_screen" validate:"omitempty"`
	IsTabs       bool             `json:"is_tabs" validate:"omitempty"`
	Component    string           `json:"component" validate:"required,max=255"`
	Icon         *string          `json:"icon" validate:"omitempty,max=128"`
	Order        uint             `json:"order" validate:"omitempty,min=0"`
}

// UpdateMenuRequest 更新菜单请求结构
type UpdateMenuRequest struct {
	ParentID     *uint            `json:"parent_id" validate:"omitempty"`
	Name         *string          `json:"name" validate:"omitempty,min=2,max=64"`
	RouteName    *string          `json:"route_name" validate:"omitempty,min=2,max=64"`
	RoutePath    *string          `json:"route_path" validate:"omitempty,min=1,max=255"`
	ApiNames     *models.ApiNames `json:"api_names" validate:"omitempty"`
	IsHidden     *bool            `json:"is_hidden" validate:"omitempty"`
	IsFullScreen *bool            `json:"is_full_screen" validate:"omitempty"`
	IsTabs       *bool            `json:"is_tabs" validate:"omitempty"`
	Component    *string          `json:"component" validate:"omitempty,max=255"`
	Icon         *string          `json:"icon" validate:"omitempty,max=128"`
	Order        *uint            `json:"order" validate:"omitempty,min=0"`
}

type MenuWithChildren struct {
	models.Menu
	Children []MenuWithChildren `json:"children"`
}

type APIInfo struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler"`
}
