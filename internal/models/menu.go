package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/bytedance/sonic"
)

type ApiNames []string

// Value 实现 driver.Valuer 接口，用于将 ApiNames 转换为数据库存储格式
func (a *ApiNames) Value() (driver.Value, error) {

	jsonData, err := sonic.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("无法将 ApiNames 转换为数据库存储格式: %v", a)
	}

	return jsonData, nil
}

// Scan 实现 sql.Scanner 接口，用于将数据库中的值转换为 ApiIds
func (a *ApiNames) Scan(value any) error {
	// v, ok := value.([]byte)
	// if !ok {
	// 	global.Logger.Errorf("无法将数据库中的值转换为 ApiNames: %v", value)
	// 	return fmt.Errorf("无法将数据库中的值转换为 ApiNames: %v", value)
	// }

	switch v := value.(type) {
	case []byte:
		// global.Logger.Infof("数据库中的值是字节数组: %s", string(v))
		return sonic.Unmarshal(v, a)
	case string:
		// global.Logger.Infof("数据库中的值是字符串: %s", v)
		return sonic.Unmarshal([]byte(v), a)
	case nil:
		*a = nil
		return nil
	default:
		return fmt.Errorf("无法将数据库中的值转换为 ApiNames: %v", value)
	}
}

type Menu struct {
	ID           uint64    `json:"id,string" gorm:"primaryKey;autoIncrement;comment:唯一ID"`                 // 唯一ID
	ParentID     *uint64   `json:"parent_id,string" gorm:"comment:父级ID"`                                   // 父级ID
	Name         string    `json:"name" gorm:"size:64;not null;comment:菜单名称"`                              // 菜单名称
	RouteName    string    `json:"route_name" gorm:"size:64;not null;unique;comment:路由名称"`                 // 路由名称，唯一
	RoutePath    string    `json:"route_path" gorm:"size:255;not null;comment:路由路径"`                       // 路由路径
	ApiNames     *ApiNames `json:"api_names" gorm:"type:text;comment:API路径"`                               // API路径
	IsHidden     bool      `json:"is_hidden" gorm:"type:boolean;not null;default:false;comment:是否隐藏"`      // 是否隐藏
	IsFullScreen bool      `json:"is_full_screen" gorm:"type:boolean;not null;default:false;comment:是否全屏"` // 是否全屏
	IsTabs       bool      `json:"is_tabs" gorm:"type:boolean;not null;default:false;comment:是否添加到tabs"`   // 是否添加到tabs
	Component    string    `json:"component" gorm:"size:255;not null;comment:组件路径"`                        // 组件路径
	Icon         *string   `json:"icon" gorm:"size:64;comment:侧边栏图标"`                                      // 侧边栏图标
	Order        uint      `json:"order" gorm:"type:int;not null;default:0;comment:排序"`                    // 排序

	CommonModel
}

// TODO:缺少按钮表,需要和菜单表进行关联

// TableName 设置表名
func (Menu) TableName() string {
	return "menus"
}
