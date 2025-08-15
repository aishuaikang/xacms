package models

import (
	"github.com/google/uuid"
)

type DepartmentModel struct {
	ID       uuid.UUID  `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"` // 唯一ID
	Name     string     `json:"name" gorm:"size:64;not null;comment:部门名称"`       // 部门名称
	ParentID *uuid.UUID `json:"parent_id" gorm:"type:char(36);comment:父级部门ID"`   // 父级部门ID

	Users []*UserModel `json:"users" gorm:"foreignKey:DepartmentID;comment:部门用户"` // 部门用户

	CommonModel
}

// TableName 设置表名
func (DepartmentModel) TableName() string {
	return "departments"
}
