package common

import (
	"encoding/json"
	"net/http"
)

// Response 通用响应结构体
type Response struct {
	// Code 状态码
	Code int `json:"code"`
	// Message 响应消息
	Message string `json:"message"`
	// Data 响应数据
	Data interface{} `json:"data,omitempty"`
}

// NewResponse 创建新的响应实例
// code: 状态码
// message: 响应消息
// data: 响应数据
// 返回: 响应实例
func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// JSON 将响应写入 HTTP 响应
// w: HTTP 响应写入器
func (r *Response) JSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Code)
	json.NewEncoder(w).Encode(r)
}

// Error 创建错误响应
// code: 状态码
// message: 错误消息
// 返回: 响应实例
func Error(code int, message string) *Response {
	return NewResponse(code, message, nil)
}

// Success 创建成功响应
// data: 响应数据
// 返回: 响应实例
func Success(data interface{}) *Response {
	return NewResponse(http.StatusOK, "success", data)
}
