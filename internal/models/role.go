package models

type Role struct {
	ID          uint   `json:"id" gorm:"primaryKey;autoIncrement;comment:唯一ID"`                     // 唯一ID
	Name        string `json:"name" gorm:"uniqueIndex:idx_role_name;size:64;not null;comment:角色名称"` // 角色名称
	Description string `json:"description" gorm:"size:255;comment:角色描述"`                            // 角色描述
	Order       uint   `json:"order" gorm:"not null;default:0;comment:角色排序，越小越靠前"`                  // 角色排序，越小越靠前

	Menus []*Menu `json:"menus" gorm:"many2many:role_menus;comment:角色菜单"` // 角色菜单

	// Users []*User `json:"users" gorm:"foreignKey:RoleID;comment:角色用户"` // 角色用户

	CommonModel
}

// TableName 设置表名
func (Role) TableName() string {
	return "roles"
}
