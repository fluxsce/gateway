// Package serialize 提供通用的序列化/反序列化工具函数
//
// 支持的格式：
//   - JSON: Marshal/Unmarshal
//   - XML: Unmarshal
//
// 使用示例：
//
//	// JSON序列化
//	data, err := serialize.JSONMarshal(obj)
//
//	// JSON反序列化
//	var result MyStruct
//	err := serialize.JSONUnmarshal(data, &result)
//
//	// XML反序列化
//	var result MyStruct
//	err := serialize.XMLUnmarshal(data, &result)
package serialize

import (
	"encoding/json"
	"encoding/xml"
)

// JSONMarshal JSON序列化
//
// 将任意类型序列化为JSON字节数组
//
// 参数：
//   - v: 要序列化的对象
//
// 返回：
//   - []byte: JSON字节数组
//   - error: 序列化失败时返回错误
func JSONMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// JSONMarshalIndent JSON序列化（格式化输出）
//
// 将任意类型序列化为格式化的JSON字节数组，便于阅读
//
// 参数：
//   - v: 要序列化的对象
//   - prefix: 每行前缀
//   - indent: 缩进字符串（如 "  " 或 "\t"）
//
// 返回：
//   - []byte: 格式化的JSON字节数组
//   - error: 序列化失败时返回错误
func JSONMarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// JSONUnmarshal JSON反序列化
//
// 将JSON字节数组反序列化为指定类型
//
// 参数：
//   - data: JSON字节数组
//   - v: 目标对象指针
//
// 返回：
//   - error: 反序列化失败时返回错误
func JSONUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// XMLMarshal XML序列化
//
// 将任意类型序列化为XML字节数组
//
// 参数：
//   - v: 要序列化的对象
//
// 返回：
//   - []byte: XML字节数组
//   - error: 序列化失败时返回错误
func XMLMarshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

// XMLMarshalIndent XML序列化（格式化输出）
//
// 将任意类型序列化为格式化的XML字节数组
//
// 参数：
//   - v: 要序列化的对象
//   - prefix: 每行前缀
//   - indent: 缩进字符串
//
// 返回：
//   - []byte: 格式化的XML字节数组
//   - error: 序列化失败时返回错误
func XMLMarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return xml.MarshalIndent(v, prefix, indent)
}

// XMLUnmarshal XML反序列化
//
// 将XML字节数组反序列化为指定类型
//
// 参数：
//   - data: XML字节数组
//   - v: 目标对象指针
//
// 返回：
//   - error: 反序列化失败时返回错误
func XMLUnmarshal(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}
