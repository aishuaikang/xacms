package models

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApiIds []string

// Value 实现 driver.Valuer 接口，用于将 ApiIds 转换为数据库存储格式
func (a *ApiIds) Value() (driver.Value, error) {
	jsonData, err := a.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// Scan 实现 sql.Scanner 接口，用于将数据库中的值转换为 ApiIds
func (a *ApiIds) Scan(value any) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			*a = strings.Split(string(v), ",")
		}

	default:
		return fmt.Errorf("无法将 %T 转换为 ApiIds", value)
	}
	return nil
}

// 实现 json.Marshaler 接口，用于将 ApiIds 转换为 JSON 格式
func (a *ApiIds) MarshalJSON() ([]byte, error) {
	if a == nil {
		return nil, nil
	}
	return []byte(strings.Join(*a, ",")), nil
}

// 实现 json.Unmarshaler 接口，用于将 JSON 格式的值转换为 ApiIds
func (a *ApiIds) UnmarshalJSON(data []byte) error {
	if data == nil {
		return nil
	}
	*a = ApiIds(strings.Split(string(data), ","))
	return nil
}

type MenuModel struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey;type:char(36);comment:唯一ID"`                        // 唯一ID
	Name         string    `json:"name" gorm:"size:64;not null;comment:菜单名称"`                              // 菜单名称
	RouteName    string    `json:"route_name" gorm:"size:64;not null;unique;comment:路由名称"`                 // 路由名称，唯一
	RoutePath    string    `json:"route_path" gorm:"size:255;not null;comment:路由路径"`                       // 路由路径
	ApiIds       *ApiIds   `json:"api_ids" gorm:"type:text;comment:API路径"`                                 // API路径
	IsHidden     bool      `json:"is_hidden" gorm:"type:boolean;not null;default:false;comment:是否隐藏"`      // 是否隐藏
	IsFullScreen bool      `json:"is_full_screen" gorm:"type:boolean;not null;default:false;comment:是否全屏"` // 是否全屏
	IsTabs       bool      `json:"is_tabs" gorm:"type:boolean;not null;default:false;comment:是否添加到tabs"`   // 是否添加到tabs
	Component    string    `json:"component" gorm:"size:255;not null;comment:组件路径"`                        // 组件路径
	Icon         *string   `json:"icon" gorm:"size:64;comment:侧边栏图标"`                                      // 侧边栏图标
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
