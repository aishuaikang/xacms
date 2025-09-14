package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	ID       uuid.UUID `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"`                            // 唯一ID
	Nickname string    `json:"nickname" gorm:"size:64;not null;comment:用户昵称"`                              // 用户昵称
	Username string    `json:"username" gorm:"uniqueIndex:idx_user_username;size:64;not null;comment:用户名"` // 用户名
	Password string    `json:"password" gorm:"size:128;not null;comment:用户密码"`                             // 用户密码
	Email    string    `json:"email" gorm:"uniqueIndex:idx_user_email;size:128;not null;comment:用户邮箱"`     // 用户邮箱
	Phone    string    `json:"phone" gorm:"uniqueIndex:idx_user_phone;size:20;not null;comment:用户电话"`      // 用户电话
	Avatar   *string   `json:"avatar" gorm:"size:255;comment:用户头像"`                                        // 用户头像
	Status   *Status   `json:"status" gorm:"type:tinyint;not null;default:1;comment:状态"`                   // 状态，1-启用，0-禁用

	RoleID *uuid.UUID `json:"role_id" gorm:"type:char(36);comment:角色ID"`  // 角色ID
	Role   *RoleModel `json:"role" gorm:"foreignKey:RoleID;comment:用户角色"` // 用户角色

	// TenantID *uuid.UUID   `json:"tenant_id" gorm:"type:char(36);comment:租户ID"`    // 租户ID
	// Tenant   *TenantModel `json:"tenant" gorm:"foreignKey:TenantID;comment:用户租户"` // 用户租户

	// DepartmentID *uuid.UUID       `json:"department_id" gorm:"type:char(36);comment:部门ID"`        // 部门ID
	// Department   *DepartmentModel `json:"department" gorm:"foreignKey:DepartmentID;comment:用户部门"` // 用户部门

	CommonModel
}

// TableName 设置表名
func (UserModel) TableName() string {
	return "users"
}

// BeforeCreate GORM钩子，在创建记录之前调用
func (u *UserModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	// 如果 Status 是零值（StatusDisabled = 0），但我们想要明确设置它
	// 可以根据业务需求决定是否需要默认值
	return
}
