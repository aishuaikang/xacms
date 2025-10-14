package dto

import (
	"encoding/json"
	"strconv"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// PaginatedResponse 分页响应结构
type PaginatedResponse[T any] struct {
	Total int64 `json:"total"`
	Items []T   `json:"items"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

type StringUint64Slice []uint64

func (s StringUint64Slice) MarshalJSON() ([]byte, error) {
	strs := make([]string, len(s))
	for i, v := range s {
		strs[i] = strconv.FormatUint(v, 10)
	}
	return json.Marshal(strs)
}

func (s *StringUint64Slice) UnmarshalJSON(data []byte) error {
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return err
	}
	*s = make(StringUint64Slice, len(strs))
	for i, str := range strs {
		v, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		(*s)[i] = v
	}
	return nil
}
