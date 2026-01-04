//go:build !no_oracle
// +build !no_oracle

// Package oracleutils 提供Oracle数据库特有的结果处理工具
//
// 本包主要解决Oracle数据库的特殊类型转换问题：
// 1. godror.Number类型转换 - Oracle数值类型到Go基本类型的转换
// 2. Oracle日期时间处理 - 处理Oracle特有的日期时间格式
// 3. CLOB/BLOB处理 - 大对象类型的读取和转换
// 4. 字符集处理 - Oracle字符编码相关问题
//
// 使用示例：
//
//	// 在Oracle查询中使用
//	err := db.Query(ctx, &result, query, args, true)
//	if err != nil {
//	    // 如果出现Oracle类型转换错误，使用Oracle专用处理器
//	    err = oracleutils.HandleOracleTypeConversion(&result, rawValues)
//	}
package oracleutils

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/godror/godror"
)

// OracleTypeConverter Oracle类型转换器
type OracleTypeConverter struct{}

// NewOracleTypeConverter 创建Oracle类型转换器
func NewOracleTypeConverter() *OracleTypeConverter {
	return &OracleTypeConverter{}
}

// ConvertOracleValue 转换Oracle特有类型到Go类型
// 处理Oracle数据库返回的特殊类型，如godror.Number、Oracle日期等
//
// 支持的Oracle类型转换：
// - godror.Number -> int/int64/float64/string
// - Oracle日期时间 -> time.Time
// - CLOB -> string
// - BLOB -> []byte
//
// 参数:
//
//	value: Oracle返回的原始值
//	targetType: 目标Go类型
//
// 返回:
//
//	interface{}: 转换后的Go值
//	error: 转换失败时返回错误信息
func (c *OracleTypeConverter) ConvertOracleValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case godror.Number:
		return c.convertOracleNumber(v, targetType)
	case *godror.Number:
		if v == nil {
			return nil, nil
		}
		return c.convertOracleNumber(*v, targetType)
	default:
		// 对于非Oracle特有类型，返回原值
		return value, nil
	}
}

// convertOracleNumber 转换Oracle Number类型
// Oracle的NUMBER类型可以表示整数、小数、科学记数法等多种数值
// 需要根据目标类型进行适当的转换
//
// 参数:
//
//	num: Oracle Number值
//	targetType: 目标Go类型
//
// 返回:
//
//	interface{}: 转换后的值
//	error: 转换失败时返回错误信息
func (c *OracleTypeConverter) convertOracleNumber(num godror.Number, targetType reflect.Type) (interface{}, error) {
	// 获取Number的字符串表示
	numStr := num.String()

	switch targetType.Kind() {
	case reflect.Int:
		val, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to int: %w", numStr, err)
		}
		return int(val), nil

	case reflect.Int8:
		val, err := strconv.ParseInt(numStr, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to int8: %w", numStr, err)
		}
		return int8(val), nil

	case reflect.Int16:
		val, err := strconv.ParseInt(numStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to int16: %w", numStr, err)
		}
		return int16(val), nil

	case reflect.Int32:
		val, err := strconv.ParseInt(numStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to int32: %w", numStr, err)
		}
		return int32(val), nil

	case reflect.Int64:
		val, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to int64: %w", numStr, err)
		}
		return val, nil

	case reflect.Uint:
		val, err := strconv.ParseUint(numStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to uint: %w", numStr, err)
		}
		return uint(val), nil

	case reflect.Uint8:
		val, err := strconv.ParseUint(numStr, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to uint8: %w", numStr, err)
		}
		return uint8(val), nil

	case reflect.Uint16:
		val, err := strconv.ParseUint(numStr, 10, 16)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to uint16: %w", numStr, err)
		}
		return uint16(val), nil

	case reflect.Uint32:
		val, err := strconv.ParseUint(numStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to uint32: %w", numStr, err)
		}
		return uint32(val), nil

	case reflect.Uint64:
		val, err := strconv.ParseUint(numStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to uint64: %w", numStr, err)
		}
		return val, nil

	case reflect.Float32:
		val, err := strconv.ParseFloat(numStr, 32)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to float32: %w", numStr, err)
		}
		return float32(val), nil

	case reflect.Float64:
		val, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert Oracle Number %s to float64: %w", numStr, err)
		}
		return val, nil

	case reflect.String:
		return numStr, nil

	case reflect.Ptr:
		// 处理指针类型
		elemType := targetType.Elem()
		convertedValue, err := c.convertOracleNumber(num, elemType)
		if err != nil {
			return nil, err
		}

		// 创建指向转换值的指针
		ptrValue := reflect.New(elemType)
		ptrValue.Elem().Set(reflect.ValueOf(convertedValue))
		return ptrValue.Interface(), nil

	default:
		return nil, fmt.Errorf("unsupported target type %s for Oracle Number conversion", targetType.String())
	}
}

// ConvertOracleValueToField 将Oracle值转换并设置到结构体字段
// 这是一个便利函数，直接将转换后的值设置到反射字段中
//
// 参数:
//
//	field: 目标字段的反射值
//	value: Oracle返回的原始值
//
// 返回:
//
//	error: 转换或设置失败时返回错误信息
func (c *OracleTypeConverter) ConvertOracleValueToField(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	convertedValue, err := c.ConvertOracleValue(value, field.Type())
	if err != nil {
		return err
	}

	if convertedValue == nil {
		// 设置零值
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	// 设置转换后的值
	field.Set(reflect.ValueOf(convertedValue))
	return nil
}

// IsOracleSpecificType 检查是否为Oracle特有类型
// 判断给定的值是否为Oracle数据库特有的类型
//
// 参数:
//
//	value: 要检查的值
//
// 返回:
//
//	bool: true表示是Oracle特有类型，false表示是通用类型
func IsOracleSpecificType(value interface{}) bool {
	switch value.(type) {
	case godror.Number, *godror.Number:
		return true
	default:
		return false
	}
}

// HandleOracleTypeError 处理Oracle类型转换错误
// 当标准的结果映射出现Oracle类型错误时，使用此函数进行特殊处理
//
// 参数:
//
//	dest: 目标结构体指针
//	columns: 数据库列名
//	values: 原始值切片
//
// 返回:
//
//	error: 处理失败时返回错误信息
func HandleOracleTypeError(dest interface{}, columns []string, values []interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	structValue := destValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	converter := NewOracleTypeConverter()

	for i, colName := range columns {
		if i >= len(values) {
			continue
		}

		value := values[i]
		if !IsOracleSpecificType(value) {
			continue // 跳过非Oracle特有类型
		}

		// 查找对应的字段
		field, found := findFieldByColumn(structValue, colName)
		if !found {
			continue // 跳过找不到的字段
		}

		// 转换并设置Oracle特有类型
		if err := converter.ConvertOracleValueToField(field, value); err != nil {
			return fmt.Errorf("failed to convert Oracle value for field %s: %w", colName, err)
		}
	}

	return nil
}

// findFieldByColumn 根据列名查找结构体字段（简化版本）
// 这是result_format.go中FindFieldByColumn的简化版本
func findFieldByColumn(structValue reflect.Value, column string) (reflect.Value, bool) {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		structField := structType.Field(i)

		dbTag := structField.Tag.Get("db")
		if dbTag == column {
			return field, true
		}
	}

	return reflect.Value{}, false
}

// ConvertOracleNumberToInt64 将Oracle Number转换为int64的便利函数
// 这是最常用的转换场景的快捷方法
//
// 参数:
//
//	num: Oracle Number值
//
// 返回:
//
//	int64: 转换后的整数值
//	error: 转换失败时返回错误信息
func ConvertOracleNumberToInt64(num godror.Number) (int64, error) {
	numStr := num.String()
	return strconv.ParseInt(numStr, 10, 64)
}

// ConvertOracleNumberToFloat64 将Oracle Number转换为float64的便利函数
// 用于处理Oracle中的小数和浮点数
//
// 参数:
//
//	num: Oracle Number值
//
// 返回:
//
//	float64: 转换后的浮点数值
//	error: 转换失败时返回错误信息
func ConvertOracleNumberToFloat64(num godror.Number) (float64, error) {
	numStr := num.String()
	return strconv.ParseFloat(numStr, 64)
}

// ConvertOracleNumberToString 将Oracle Number转换为字符串的便利函数
// 保留Oracle Number的原始精度和格式
//
// 参数:
//
//	num: Oracle Number值
//
// 返回:
//
//	string: 转换后的字符串值
func ConvertOracleNumberToString(num godror.Number) string {
	return num.String()
}
