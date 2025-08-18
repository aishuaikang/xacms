package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MenuModel struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"`                        // 唯一ID
	Name         string    `json:"name" gorm:"size:64;not null;comment:菜单名称"`                              // 菜单名称
	RouteName    string    `json:"route_name" gorm:"size:64;not null;comment:路由名称"`                        // 路由名称
	RoutePath    string    `json:"route_path" gorm:"size:255;not null;comment:路由路径"`                       // 路由路径
	IsHidden     bool      `json:"is_hidden" gorm:"type:boolean;not null;default:false;comment:是否隐藏"`      // 是否隐藏
	IsFullScreen bool      `json:"is_full_screen" gorm:"type:boolean;not null;default:false;comment:是否全屏"` // 是否全屏
	IsTabs       bool      `json:"is_tabs" gorm:"type:boolean;not null;default:false;comment:是否添加到tabs"`   // 是否添加到tabs
	Component    string    `json:"component" gorm:"size:255;not null;comment:组件路径"`                        // 组件路径
	Icon         string    `json:"icon" gorm:"size:64;comment:侧边栏图标"`                                      // 侧边栏图标
	Order        uint      `json:"order" gorm:"type:int;not null;default:0;comment:排序"`                    // 排序

	CommonModel
}

// TableName 设置表名
func (MenuModel) TableName() string {
	return "menus"
}

// BeforeCreate GORM钩子，在创建记录之前调用
func (u *MenuModel) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
