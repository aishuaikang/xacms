package models

import (
	"github.com/google/uuid"
)

type RoleModel struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"`     // 唯一ID
	Name        string    `json:"name" gorm:"size:64;not null;comment:角色名称"`           // 角色名称
	Description string    `json:"description" gorm:"size:255;comment:角色描述"`            // 角色描述
	Order       uint      `json:"order" gorm:"type:int;not null;default:0;comment:排序"` // 排序

	Menus []*MenuModel `json:"menus" gorm:"many2many:role_menus;comment:角色菜单"` // 角色菜单

	Users []*UserModel `json:"users" gorm:"foreignKey:RoleID;comment:角色用户"` // 角色用户

	CommonModel
}

// TableName 设置表名
func (RoleModel) TableName() string {
	return "roles"
}
