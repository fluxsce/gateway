// 构建标签说明：
// !no_oracle 表示启用Oracle数据库支持
// 当项目需要Oracle数据库支持时，使用此文件中的实现
// 使用方法：
//   1. 默认情况下（无特殊构建标签），此文件会被编译使用
//   2. 如果设置了 -tags no_oracle，此文件将被排除在编译之外
//   3. 适用于需要Oracle支持的生产环境
//
//go:build !no_oracle
// +build !no_oracle

package sqlutils

import (
	"gohub/pkg/database/sqlutils/oracleutils"
	"reflect"
)

// IsOracleSpecificType 检查值是否为Oracle特有类型
// 用于在类型转换前快速判断是否需要特殊处理
//
// 参数：
//   - value: 要检查的值
//
// 返回：
//   - bool: 如果是Oracle特有类型返回true，否则返回false
//
// 说明：
//   - 支持所有Oracle特有类型（如CLOB、BLOB、NCLOB等）
//   - 对nil值返回false，避免空指针异常
//   - 通过oracleutils包进行具体的类型判断
func (fm *FieldMapper) IsOracleSpecificType(value interface{}) bool {
    if value == nil {
        return false
    }
    return oracleutils.IsOracleSpecificType(value)
}

// HandleSpecialTypeConversion 处理特殊数据库类型的转换
// 目前主要用于处理Oracle特有类型到Go类型的转换
//
// 参数：
//   - dest: 目标字段的反射值
//   - value: 要转换的原始值
//
// 返回：
//   - error: 转换过程中的错误，如果转换成功则返回nil
//
// 说明：
//   - 优先处理nil值，避免后续转换出错
//   - 使用oracleutils包中的转换器进行具体转换
//   - 支持Oracle所有特有类型到Go类型的映射
func HandleSpecialTypeConversion(dest reflect.Value, value interface{}) error {
    if value == nil {
        return nil
    }
    
    if oracleutils.IsOracleSpecificType(value) {
        converter := oracleutils.NewOracleTypeConverter()
        return converter.ConvertOracleValueToField(dest, value)
    }
    return nil
}

// convertOracleValue 处理Oracle特有类型到Go类型的转换
// 这是FieldMapper内部使用的Oracle类型转换方法
//
// 参数：
//   - field: 目标字段的反射值
//   - value: 要转换的Oracle类型值
//
// 返回：
//   - error: 转换过程中的错误，如果转换成功则返回nil
//
// 说明：
//   - 该方法会在常规类型转换之前被调用
//   - 只处理Oracle特有类型，其他类型返回nil错误
//   - 使用oracleutils包的转换器确保类型转换的准确性
//   - 支持nil值的安全处理
func (fm *FieldMapper) convertOracleValue(field reflect.Value, value interface{}) error {
    if value == nil {
        return nil
    }
    
    if oracleutils.IsOracleSpecificType(value) {
        converter := oracleutils.NewOracleTypeConverter()
        return converter.ConvertOracleValueToField(field, value)
    }
    return nil
}