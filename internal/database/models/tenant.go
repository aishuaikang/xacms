package models

import (
	"github.com/dromara/carbon/v2"
	"github.com/google/uuid"
)

// 租户类型枚举
type TenantType uint8

const (
	TenantTypeEnterprise  TenantType = iota // 企业
	TenantTypeInstitution                   // 机构
)

type TenantModel struct {
	ID          uuid.UUID       `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"`                         // 唯一ID
	Code        string          `json:"code" gorm:"uniqueIndex:idx_tenant_code;size:64;not null;comment:租户编码"`   // 租户编码，如 TENANT-001
	Name        string          `json:"name" gorm:"size:255;not null;comment:租户名称"`                              // 租户名称，如 某某科技有限公司
	StartAt     carbon.DateTime `json:"start_at" gorm:"index:idx_tenant_start;not null;comment:开始时间"`            // 开始时间，如 2027-01-01 00:00:00
	EndAt       carbon.DateTime `json:"end_at" gorm:"index:idx_tenant_end;not null;comment:结束时间"`                // 结束时间，如 2027-01-01 00:00:00
	Type        TenantType      `json:"type" gorm:"type:tinyint;not null;default:0;comment:租户类型"`                // 租户类型，如 企业/机构
	Phone       string          `json:"phone" gorm:"uniqueIndex:idx_tenant_phone;size:20;not null;comment:租户电话"` // 租户电话，如 13800000000
	Description string          `json:"description" gorm:"type:text;comment:租户说明"`                               // 租户说明，如 这是一个测试租户
	Status      Status          `json:"status" gorm:"type:tinyint;not null;default:1;comment:租户状态"`              // 租户状态，1-启用，0-禁用

	// TODO: 缺少与用户进行关联
	User *UserModel `json:"user" gorm:"foreignKey:TenantID;comment:租户用户"` // 租户用户

	CommonModel
}

// TableName 设置表名
func (TenantModel) TableName() string {
	return "tenants"
}
