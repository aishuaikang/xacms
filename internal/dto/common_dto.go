package dto

import "github.com/google/uuid"

// BaseQueryRequest 基础查询请求结构
type BaseQueryRequest struct {
	Page     int `form:"page,string" validate:"min=1"`
	PageSize int `form:"page_size,string" validate:"min=1,max=100"`
	// Keyword  string `form:"keyword" validate:"omitempty,max=100"`
}

// DeleteMultipleRequest 批量删除请求结构
type DeleteMultipleRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,min=1,dive,uuid"` // UUID 列表
}

// // IDRequest 通用ID请求结构
// type IDRequest struct {
// 	ID string `json:"id" validate:"required,uuid"`
// }

// // StatusRequest 状态请求结构
// type StatusRequest struct {
// 	Status int `json:"status" validate:"required,oneof=0 1"`
// }
