package sqlutils

import (
	"gateway/pkg/database/sqlutils/clickhouseutils"
	"reflect"
)

// IsClickHouseSpecificType 检查是否为ClickHouse特有类型
func (fm *FieldMapper) IsClickHouseSpecificType(value interface{}) bool {
	return clickhouseutils.IsClickHouseSpecificType(value)
}

// convertClickHouseValue 处理ClickHouse特有类型到Go类型的转换
// 这是FieldMapper内部使用的ClickHouse类型转换方法
//
// 参数：
//   - field: 目标字段的反射值
//   - value: 要转换的ClickHouse类型值
//
// 返回：
//   - error: 转换过程中的错误，如果转换成功则返回nil
//
// 说明：
//   - 该方法会在常规类型转换之前被调用
//   - 只处理ClickHouse特有类型，其他类型返回nil错误
//   - 使用clickhouseutils包的转换器确保类型转换的准确性
//   - 支持nil值的安全处理
func (fm *FieldMapper) convertClickHouseValue(field reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	if clickhouseutils.IsClickHouseSpecificType(value) {
		converter := clickhouseutils.NewClickHouseTypeConverter()
		return converter.ConvertClickHouseValueToField(field, value)
	}
	return nil
}

// HandleSpecialTypeConversion 处理特殊类型转换（ClickHouse扩展）
// 扩展原有的Oracle类型转换，增加ClickHouse支持
func HandleSpecialTypeConversionWithClickHouse(dest reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	// 先尝试ClickHouse类型转换
	if clickhouseutils.IsClickHouseSpecificType(value) {
		converter := clickhouseutils.NewClickHouseTypeConverter()
		return converter.ConvertClickHouseValueToField(dest, value)
	}

	// 如果不是ClickHouse类型，使用原有的Oracle类型转换逻辑
	return HandleSpecialTypeConversion(dest, value)
}
