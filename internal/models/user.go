package models

type User struct {
	ID       uint64  `json:"id,string" gorm:"primaryKey;autoIncrement;comment:唯一ID"`                     // 唯一ID
	Nickname string  `json:"nickname" gorm:"size:64;not null;comment:用户昵称"`                              // 用户昵称
	Username string  `json:"username" gorm:"uniqueIndex:idx_user_username;size:64;not null;comment:用户名"` // 用户名
	Password string  `json:"-" gorm:"size:128;not null;comment:用户密码"`                                    // 用户密码
	Email    *string `json:"email" gorm:"uniqueIndex:idx_user_email;size:128;comment:用户邮箱"`              // 用户邮箱
	Phone    *string `json:"phone" gorm:"uniqueIndex:idx_user_phone;size:20;comment:用户电话"`               // 用户电话
	Avatar   *string `json:"avatar" gorm:"size:255;comment:用户头像"`                                        // 用户头像

	RoleID *uint64 `json:"role_id,string" gorm:"comment:角色ID"`         // 角色ID
	Role   *Role   `json:"role" gorm:"foreignKey:RoleID;comment:用户角色"` // 用户角色

	CommonModel
}

// TableName 设置表名
func (User) TableName() string {
	return "users"
}
