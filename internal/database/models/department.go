package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
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

// BeforeCreate GORM钩子，在创建记录之前调用
func (u *DepartmentModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
