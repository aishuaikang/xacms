package dto

// BaseQueryRequest 基础查询请求结构
type BaseQueryRequest struct {
	Page     int `query:"page,string" validate:"min=1"`
	PageSize int `query:"page_size,string" validate:"min=1,max=100"`
	// Keyword  string `query:"keyword" validate:"omitempty,max=100"`
}

// // IDRequest 通用ID请求结构
// type IDRequest struct {
// 	ID string `json:"id" validate:"required,uuid"`
// }

// // StatusRequest 状态请求结构
// type StatusRequest struct {
// 	Status int `json:"status" validate:"required,oneof=0 1"`
// }
