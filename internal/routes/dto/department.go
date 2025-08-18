package dto

import (
	"github.com/google/uuid"
)

// CreateDepartmentRequest 创建部门请求结构
type CreateDepartmentRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=64"`
	Code        string     `json:"code" validate:"required,min=2,max=64"`
	Description string     `json:"description" validate:"omitempty,max=255"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	TenantID    *uuid.UUID `json:"tenant_id" validate:"omitempty,uuid"`
	ManagerID   *uuid.UUID `json:"manager_id" validate:"omitempty,uuid"`
	Sort        int        `json:"sort" validate:"omitempty,min=0"`
	Status      int        `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateDepartmentRequest 更新部门请求结构
type UpdateDepartmentRequest struct {
	Name        *string    `json:"name" validate:"omitempty,min=2,max=64"`
	Code        *string    `json:"code" validate:"omitempty,min=2,max=64"`
	Description *string    `json:"description" validate:"omitempty,max=255"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	ManagerID   *uuid.UUID `json:"manager_id" validate:"omitempty,uuid"`
	Sort        *int       `json:"sort" validate:"omitempty,min=0"`
	Status      *int       `json:"status" validate:"omitempty,oneof=0 1"`
}

// DepartmentQueryRequest 部门查询请求结构
type DepartmentQueryRequest struct {
	BaseQueryRequest
	ParentID *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	TenantID *uuid.UUID `json:"tenant_id" validate:"omitempty,uuid"`
	Status   *int       `json:"status" validate:"omitempty,oneof=0 1"`
}

// DepartmentTreeRequest 部门树查询请求结构
type DepartmentTreeRequest struct {
	TenantID *uuid.UUID `json:"tenant_id" validate:"omitempty,uuid"`
	Status   *int       `json:"status" validate:"omitempty,oneof=0 1"`
}

// MoveDepartmentRequest 移动部门请求结构
type MoveDepartmentRequest struct {
	NewParentID *uuid.UUID `json:"new_parent_id" validate:"omitempty,uuid"`
}
