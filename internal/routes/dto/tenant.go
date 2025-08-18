package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateTenantRequest 创建租户请求结构
type CreateTenantRequest struct {
	Name         string    `json:"name" validate:"required,min=2,max=64"`
	Code         string    `json:"code" validate:"required,min=2,max=64"`
	Domain       string    `json:"domain" validate:"required,fqdn,max=128"`
	Logo         string    `json:"logo" validate:"omitempty,max=255"`
	Description  string    `json:"description" validate:"omitempty,max=255"`
	ContactName  string    `json:"contact_name" validate:"required,min=2,max=64"`
	ContactPhone string    `json:"contact_phone" validate:"required,phone"`
	ContactEmail string    `json:"contact_email" validate:"required,email,max=128"`
	ExpireDate   time.Time `json:"expire_date" validate:"required"`
	MaxUsers     int       `json:"max_users" validate:"required,min=1"`
	Status       int       `json:"status" validate:"omitempty,oneof=0 1"`
}

// UpdateTenantRequest 更新租户请求结构
type UpdateTenantRequest struct {
	Name         *string    `json:"name" validate:"omitempty,min=2,max=64"`
	Code         *string    `json:"code" validate:"omitempty,min=2,max=64"`
	Domain       *string    `json:"domain" validate:"omitempty,fqdn,max=128"`
	Logo         *string    `json:"logo" validate:"omitempty,max=255"`
	Description  *string    `json:"description" validate:"omitempty,max=255"`
	ContactName  *string    `json:"contact_name" validate:"omitempty,min=2,max=64"`
	ContactPhone *string    `json:"contact_phone" validate:"omitempty,phone"`
	ContactEmail *string    `json:"contact_email" validate:"omitempty,email,max=128"`
	ExpireDate   *time.Time `json:"expire_date" validate:"omitempty"`
	MaxUsers     *int       `json:"max_users" validate:"omitempty,min=1"`
	Status       *int       `json:"status" validate:"omitempty,oneof=0 1"`
}

// TenantQueryRequest 租户查询请求结构
type TenantQueryRequest struct {
	BaseQueryRequest
	Status     *int    `json:"status" validate:"omitempty,oneof=0 1"`
	Domain     *string `json:"domain" validate:"omitempty,fqdn"`
	ExpireSoon *bool   `json:"expire_soon" validate:"omitempty"`
}

// TenantUsersRequest 租户用户查询请求结构
type TenantUsersRequest struct {
	BaseQueryRequest
	TenantID uuid.UUID `json:"tenant_id" validate:"required,uuid"`
}

// TenantDepartmentsRequest 租户部门查询请求结构
type TenantDepartmentsRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required,uuid"`
}
