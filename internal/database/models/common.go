package models

import (
	"github.com/dromara/carbon/v2"
	"gorm.io/gorm"
)

type CommonModel struct {
	CreatedAt carbon.DateTime `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt carbon.DateTime `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index:idx_deleted_at;comment:删除时间"`
}

type CommonNotDeletedModel struct {
	CreatedAt carbon.DateTime `json:"created_at" gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt carbon.DateTime `json:"updated_at" gorm:"autoUpdateTime;comment:更新时间"`
}

// 状态
type Status uint8

const (
	StatusEnabled  Status = iota // 禁用
	StatusDisabled               // 启用
)

// type CustomTime time.Time

// func (ct *CustomTime) Scan(value interface{}) error {
// 	if t, ok := value.(time.Time); ok {
// 		*ct = CustomTime(t)
// 		return nil
// 	}
// 	return fmt.Errorf("failed to scan CustomTime: %v", value)
// }

// func (ct CustomTime) Value() (driver.Value, error) {
// 	t := time.Time(ct)
// 	if t.IsZero() {
// 		return nil, nil
// 	}
// 	return t, nil
// }

// func (ct CustomTime) MarshalJSON() ([]byte, error) {
// 	t := time.Time(ct)
// 	if t.IsZero() {
// 		return json.Marshal(nil)
// 	}
// 	return json.Marshal(t.Format("2006-01-02 15:04:05"))
// }
