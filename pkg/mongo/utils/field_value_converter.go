package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gateway/pkg/mongo/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ================================
// 字段值转换器类
// ================================

// FieldValueConverter 字段值转换器
//
// 负责处理各种数据类型之间的转换，包括：
// - MongoDB文档值到Go类型的转换
// - Go类型到MongoDB文档值的转换
// - 基本类型转换（字符串、数字、布尔值）
// - 复杂类型转换（时间、结构体、切片、映射）
// - 空值检查和处理
//
// 设计特点：
// - 无状态设计，线程安全
// - 支持递归转换
// - 完整的错误处理
// - 高性能的类型匹配
//
// 使用示例：
//
//	converter := NewFieldValueConverter()
//	result, err := converter.ConvertToGoValue(mongoValue, targetType)
//	result, err := converter.ConvertFromGoValue(goValue)
//
// 注意事项：
// - 线程安全，可以并发调用
// - 不维护状态，每次调用都是独立的
// - 支持复杂的嵌套结构转换
type FieldValueConverter struct {
	// 可以添加配置选项，目前保持无状态
}

// NewFieldValueConverter 创建新的字段值转换器
//
// 返回值：
//
//	*FieldValueConverter - 字段值转换器实例
//
// 使用示例：
//
//	converter := NewFieldValueConverter()
//	// 无需清理，因为是无状态的
//
// 注意事项：
// - 无状态设计，可以全局复用
// - 线程安全，支持并发调用
func NewFieldValueConverter() *FieldValueConverter {
	return &FieldValueConverter{}
}

// ================================
// 核心转换方法
// ================================

// ConvertToGoValue 将MongoDB文档值转换为Go类型（从MongoDB到Go）
//
// 参数：
//
//	value - MongoDB文档中的值
//	targetType - 目标Go类型
//
// 返回值：
//
//	interface{} - 转换后的Go值
//	error - 转换错误，nil表示成功
//
// 转换支持：
// - 基本类型：string, int, float, bool等
// - 时间类型：time.Time, *time.Time
// - 指针类型：自动处理nil值
// - 复杂类型：结构体、切片、映射
func (fvc *FieldValueConverter) ConvertToGoValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	valueType := reflect.TypeOf(value)

	// 直接类型匹配 - 最快的路径
	if valueType != nil && valueType == targetType {
		return value, nil
	}

	// 处理指针类型
	if targetType.Kind() == reflect.Ptr {
		if value == nil {
			return nil, nil
		}
		// 递归转换指针指向的类型
		innerValue, err := fvc.ConvertToGoValue(value, targetType.Elem())
		if err != nil {
			return nil, err
		}
		// 创建指针
		if innerValue == nil {
			return nil, nil
		}
		ptrValue := reflect.New(targetType.Elem())
		ptrValue.Elem().Set(reflect.ValueOf(innerValue))
		return ptrValue.Interface(), nil
	}

	// 根据目标类型进行转换
	switch targetType.Kind() {
	case reflect.String:
		return fvc.convertToString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := fvc.convertToInt64(value)
		if err != nil {
			return nil, err
		}
		return fvc.convertIntToTargetType(intVal, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := fvc.convertToUint64(value)
		if err != nil {
			return nil, err
		}
		return fvc.convertUintToTargetType(uintVal, targetType)
	case reflect.Float32, reflect.Float64:
		floatVal, err := fvc.convertToFloat64(value)
		if err != nil {
			return nil, err
		}
		return fvc.convertFloatToTargetType(floatVal, targetType)
	case reflect.Bool:
		return fvc.convertToBool(value)
	case reflect.Slice, reflect.Array:
		// 处理切片和数组类型，包括 MongoDB 的 primitive.A
		return fvc.convertToSliceOrArray(value, targetType)
	case reflect.Struct:
		// 特殊处理time.Time
		if targetType == reflect.TypeOf(time.Time{}) {
			return fvc.convertToTime(value)
		}
		// 其他结构体类型需要特殊处理
		return nil, fmt.Errorf("不支持的结构体类型转换: %v", targetType)
	default:
		return nil, fmt.Errorf("不支持的目标类型: %v", targetType)
	}
}

// ConvertFromGoValue 将Go值转换为MongoDB文档兼容的值（从Go到MongoDB）
//
// 参数：
//
//	value - Go语言的值
//	structConverter - 结构体转换回调函数，用于处理结构体类型
//
// 返回值：
//
//	interface{} - 转换后的MongoDB兼容值
//	error - 转换错误，nil表示成功
//
// 转换规则：
// - 基本类型直接返回
// - 时间类型保持原样
// - 结构体通过回调函数转换为Document
// - 切片递归转换
// - 映射递归转换
func (fvc *FieldValueConverter) ConvertFromGoValue(value interface{}, structConverter func(interface{}) (types.Document, error)) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	v := reflect.ValueOf(value)
	t := v.Type()

	switch t.Kind() {
	case reflect.String, reflect.Bool:
		// 基本类型直接返回
		return value, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 整数类型直接返回
		return value, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 无符号整数类型直接返回
		return value, nil

	case reflect.Float32, reflect.Float64:
		// 浮点数类型直接返回
		return value, nil

	case reflect.Struct:
		// 特殊处理时间类型
		if t == reflect.TypeOf(time.Time{}) {
			// 使用统一的时间转换方法，保持代码一致性
			timeValue := value.(time.Time)
			return ConvertGoTimeToMongo(timeValue), nil
		}
		// 其他结构体类型通过回调函数转换
		if structConverter != nil {
			return structConverter(value)
		}
		return nil, fmt.Errorf("结构体转换需要提供转换器回调函数: %v", t)

	case reflect.Slice, reflect.Array:
		// 切片和数组类型
		length := v.Len()
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			elem := v.Index(i)
			convertedElem, err := fvc.ConvertFromGoValue(elem.Interface(), structConverter)
			if err != nil {
				return nil, fmt.Errorf("转换数组元素 %d 失败: %v", i, err)
			}
			result[i] = convertedElem
		}
		return result, nil

	case reflect.Map:
		// 映射类型
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("映射键必须是字符串类型，当前类型: %v", t.Key())
		}
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			mapValue := v.MapIndex(key)
			convertedValue, err := fvc.ConvertFromGoValue(mapValue.Interface(), structConverter)
			if err != nil {
				return nil, fmt.Errorf("转换映射值失败，键: %s, 错误: %v", key.String(), err)
			}
			result[key.String()] = convertedValue
		}
		return result, nil

	case reflect.Ptr:
		// 指针类型
		if v.IsNil() {
			return nil, nil
		}
		return fvc.ConvertFromGoValue(v.Elem().Interface(), structConverter)

	case reflect.Interface:
		// 接口类型
		if v.IsNil() {
			return nil, nil
		}
		return fvc.ConvertFromGoValue(v.Elem().Interface(), structConverter)

	default:
		return nil, fmt.Errorf("不支持的类型转换: %v", t)
	}
}

// ConvertFromGoValueSimple 简单版本的Go值转换（不支持结构体递归转换）
//
// 参数：
//
//	value - Go语言的值
//
// 返回值：
//
//	interface{} - 转换后的MongoDB兼容值
//	error - 转换错误，nil表示成功
//
// 注意事项：
// - 不支持结构体递归转换
// - 用于简单类型的转换
// - 遇到结构体会返回错误
func (fvc *FieldValueConverter) ConvertFromGoValueSimple(value interface{}) (interface{}, error) {
	return fvc.ConvertFromGoValue(value, nil)
}

// IsEmptyValue 检查值是否为空（用于omitempty标签处理）
//
// 参数：
//
//	v - 要检查的反射值
//
// 返回值：
//
//	bool - true表示为空值
//
// 空值定义：
// - nil指针
// - 零值数字（0, 0.0）
// - 空字符串
// - 空切片、数组、映射
// - 零值时间
// - false布尔值
func (fvc *FieldValueConverter) IsEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// 特殊处理时间类型
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		// 其他结构体类型不视为空
		return false
	}
	return false
}

// ================================
// 基本类型转换方法
// ================================

// convertToString 转换为字符串
//
// 支持的源类型：
// - string, []byte
// - 各种数值类型
// - bool类型
func (fvc *FieldValueConverter) convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	default:
		return "", fmt.Errorf("无法转换为字符串，源类型: %T", value)
	}
}

// convertToInt64 转换为int64
//
// 支持的源类型：
// - 各种整数类型
// - 各种浮点数类型
// - 字符串类型
func (fvc *FieldValueConverter) convertToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("无法转换为int64，源类型: %T", value)
	}
}

// convertToUint64 转换为uint64
//
// 支持的源类型：
// - 各种整数类型
// - 各种浮点数类型
// - 字符串类型
func (fvc *FieldValueConverter) convertToUint64(value interface{}) (uint64, error) {
	switch v := value.(type) {
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int:
		return uint64(v), nil
	case int8:
		return uint64(v), nil
	case int16:
		return uint64(v), nil
	case int32:
		return uint64(v), nil
	case int64:
		return uint64(v), nil
	case float32:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("无法转换为uint64，源类型: %T", value)
	}
}

// convertToFloat64 转换为float64
//
// 支持的源类型：
// - 各种浮点数类型
// - 各种整数类型
// - 字符串类型
func (fvc *FieldValueConverter) convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("无法转换为float64，源类型: %T", value)
	}
}

// convertToBool 转换为bool
//
// 支持的源类型：
// - bool类型
// - 各种数值类型（0为false，非0为true）
// - 字符串类型
func (fvc *FieldValueConverter) convertToBool(value interface{}) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("无法转换为bool，源类型: %T", value)
	}
}

// convertToTime 将各种类型的值转换为time.Time
//
// 支持的源类型：
// - time.Time, *time.Time
// - 字符串（多种格式）
// - Unix时间戳（int64, float64）
// - primitive.DateTime（MongoDB BSON DateTime类型）
//
// 支持的时间格式：
// - RFC3339: 2006-01-02T15:04:05Z07:00
// - RFC3339Nano: 2006-01-02T15:04:05.999999999Z07:00
// - 通用格式: 2006-01-02 15:04:05
// - ISO格式: 2006-01-02T15:04:05
// - 日期格式: 2006-01-02
//
// 时区处理：
// - 保持原有时区信息，不自动转换为UTC
// - 无时区信息的字符串默认按本地时区解析
// - 如需UTC时间，可手动调用EnsureUTC方法
func (fvc *FieldValueConverter) convertToTime(value interface{}) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		// 直接返回时间，保持原有时区信息
		return v, nil
	case *time.Time:
		if v != nil {
			return *v, nil
		}
		return time.Time{}, nil
	case string:
		// 尝试多种时间格式
		formats := []string{
			time.RFC3339,          // 2006-01-02T15:04:05Z07:00 (带时区)
			time.RFC3339Nano,      // 2006-01-02T15:04:05.999999999Z07:00 (带时区)
			"2006-01-02 15:04:05", // 无时区信息，按本地时区解析
			"2006-01-02T15:04:05", // 无时区信息，按本地时区解析
			"2006-01-02",          // 仅日期，按本地时区解析
		}

		var parsedTime time.Time
		var err error

		for _, format := range formats {
			parsedTime, err = time.Parse(format, v)
			if err == nil {
				// 成功解析，直接返回
				return parsedTime, nil
			}
		}

		// 如果标准格式都无法解析，尝试按本地时区解析
		for _, format := range formats {
			parsedTime, err = time.ParseInLocation(format, v, time.Local)
			if err == nil {
				// 成功解析，直接返回
				return parsedTime, nil
			}
		}

		return time.Time{}, fmt.Errorf("无法解析时间格式，支持的格式: %v，当前值: %s", formats, v)
	case int64:
		// Unix时间戳（秒）- 直接返回本地时间
		return time.Unix(v, 0), nil
	case float64:
		// Unix时间戳（秒，支持小数）- 直接返回本地时间
		return time.Unix(int64(v), int64((v-float64(int64(v)))*1e9)), nil
	case primitive.DateTime:
		// primitive.DateTime 转换为 time.Time
		// primitive.DateTime 存储的是从Unix纪元开始的毫秒数（UTC时间）
		// 需要除以1000转换为秒，然后创建time.Time对象

		// 创建UTC时间（MongoDB存储的是UTC时间戳）
		utcTime := time.Unix(int64(v)/1000, (int64(v)%1000)*1000000).UTC()

		// 重要：由于写入时通过 ConvertGoTimeToMongo 预先加了时区偏移量
		// 读取时需要减去时区偏移量，恢复原始的本地时间显示
		if utcTime.Location() == time.UTC {
			// 获取本地时区偏移量（秒）
			_, offset := time.Now().Zone()
			// 减去时区偏移量，恢复原始的本地时间值
			// 例如：MongoDB存储 15:07 UTC -> 15:07 - 8小时 = 07:07 -> 转为本地时间 -> 15:07 CST
			adjustedTime := utcTime.Add(-time.Duration(offset) * time.Second)
			return adjustedTime.In(time.Local), nil
		}

		return utcTime, nil
	default:
		return time.Time{}, fmt.Errorf("无法转换为时间，源类型: %T", value)
	}
}

// ================================
// 类型适配方法
// ================================

// convertIntToTargetType 将int64转换为目标整数类型
func (fvc *FieldValueConverter) convertIntToTargetType(value int64, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.Int:
		return int(value), nil
	case reflect.Int8:
		return int8(value), nil
	case reflect.Int16:
		return int16(value), nil
	case reflect.Int32:
		return int32(value), nil
	case reflect.Int64:
		return value, nil
	default:
		return nil, fmt.Errorf("不支持的整数类型: %v", targetType)
	}
}

// convertUintToTargetType 将uint64转换为目标无符号整数类型
func (fvc *FieldValueConverter) convertUintToTargetType(value uint64, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.Uint:
		return uint(value), nil
	case reflect.Uint8:
		return uint8(value), nil
	case reflect.Uint16:
		return uint16(value), nil
	case reflect.Uint32:
		return uint32(value), nil
	case reflect.Uint64:
		return value, nil
	default:
		return nil, fmt.Errorf("不支持的无符号整数类型: %v", targetType)
	}
}

// convertFloatToTargetType 将float64转换为目标浮点数类型
func (fvc *FieldValueConverter) convertFloatToTargetType(value float64, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.Float32:
		return float32(value), nil
	case reflect.Float64:
		return value, nil
	default:
		return nil, fmt.Errorf("不支持的浮点数类型: %v", targetType)
	}
}

// ================================
// 时区处理工具方法
// ================================

// EnsureUTC 确保时间值为UTC时区（工具方法）
//
// 参数：
//
//	t - 时间值
//
// 返回值：
//
//	time.Time - UTC时区的时间值
//
// 使用场景：
// - 手动转换时间为UTC时区
// - 特殊业务需求的时区统一
//
// 注意：
// - 默认情况下，时间转换器不会自动转换为UTC
// - 如果需要UTC时间，可以手动调用此方法
func (fvc *FieldValueConverter) EnsureUTC(t time.Time) time.Time {
	return t.UTC()
}

// ParseTimeWithTimezone 按指定时区解析时间字符串
//
// 参数：
//
//	timeStr - 时间字符串
//	format - 时间格式
//	location - 时区信息，nil表示使用本地时区
//
// 返回值：
//
//	time.Time - 解析后的时间（保持指定时区）
//	error - 解析错误
//
// 使用示例：
//
//	// 按中国时区解析
//	loc, _ := time.LoadLocation("Asia/Shanghai")
//	t, err := converter.ParseTimeWithTimezone("2023-01-01 12:00:00", "2006-01-02 15:04:05", loc)
//
//	// 按本地时区解析
//	t, err := converter.ParseTimeWithTimezone("2023-01-01 12:00:00", "2006-01-02 15:04:05", nil)
func (fvc *FieldValueConverter) ParseTimeWithTimezone(timeStr, format string, location *time.Location) (time.Time, error) {
	if location == nil {
		location = time.Local
	}

	parsedTime, err := time.ParseInLocation(format, timeStr, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("解析时间失败: %v", err)
	}

	// 直接返回解析后的时间，保持原有时区信息
	return parsedTime, nil
}

// ConvertToLocalTime 将UTC时间转换为本地时间（仅用于显示）
//
// 参数：
//
//	utcTime - UTC时间
//
// 返回值：
//
//	time.Time - 本地时区的时间
//
// 注意事项：
// - 此方法用于时区转换
// - 可将UTC时间转换为本地时间用于显示
func (fvc *FieldValueConverter) ConvertToLocalTime(utcTime time.Time) time.Time {
	return utcTime.Local()
}

// ConvertToTimezone 将时间转换为指定时区
//
// 参数：
//
//	t - 源时间
//	location - 目标时区
//
// 返回值：
//
//	time.Time - 指定时区的时间
//	error - 转换错误
//
// 使用示例：
//
//	// 转换为中国时区
//	loc, _ := time.LoadLocation("Asia/Shanghai")
//	chinaTime, err := converter.ConvertToTimezone(utcTime, loc)
func (fvc *FieldValueConverter) ConvertToTimezone(t time.Time, location *time.Location) (time.Time, error) {
	if location == nil {
		return time.Time{}, fmt.Errorf("时区信息不能为nil")
	}

	return t.In(location), nil
}

// ================================
// 数组和切片转换方法
// ================================

// convertToSliceOrArray 将值转换为切片或数组类型
// 专门处理 MongoDB 的 primitive.A 类型和其他数组类型
//
// 参数：
//
//	value - 源值（可能是 primitive.A、[]interface{} 或其他切片类型）
//	targetType - 目标类型（切片或数组）
//
// 返回值：
//
//	interface{} - 转换后的切片或数组
//	error - 转换错误
//
// 支持的转换：
// - primitive.A -> []int, []string, []float64 等
// - []interface{} -> []int, []string, []float64 等
// - []int -> []interface{} 等
// - 其他切片类型之间的转换
func (fvc *FieldValueConverter) convertToSliceOrArray(value interface{}, targetType reflect.Type) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	var sourceSlice []interface{}

	// 处理不同的源类型
	switch v := value.(type) {
	case primitive.A:
		// MongoDB 数组类型，直接转换为 []interface{}
		sourceSlice = []interface{}(v)
	case []interface{}:
		// 已经是 []interface{} 类型
		sourceSlice = v
	case []int:
		// 如果目标类型完全匹配，直接返回
		if targetType == reflect.TypeOf([]int{}) {
			return v, nil
		}
		// 转换为 []interface{} 以便统一处理
		sourceSlice = make([]interface{}, len(v))
		for i, item := range v {
			sourceSlice[i] = item
		}
	case []int32:
		// 转换为 []interface{} 以便统一处理
		sourceSlice = make([]interface{}, len(v))
		for i, item := range v {
			sourceSlice[i] = item
		}
	case []int64:
		// 转换为 []interface{} 以便统一处理
		sourceSlice = make([]interface{}, len(v))
		for i, item := range v {
			sourceSlice[i] = item
		}
	case []float32:
		// 转换为 []interface{} 以便统一处理
		sourceSlice = make([]interface{}, len(v))
		for i, item := range v {
			sourceSlice[i] = item
		}
	case []float64:
		// 转换为 []interface{} 以便统一处理
		sourceSlice = make([]interface{}, len(v))
		for i, item := range v {
			sourceSlice[i] = item
		}
	case []string:
		// 转换为 []interface{} 以便统一处理
		sourceSlice = make([]interface{}, len(v))
		for i, item := range v {
			sourceSlice[i] = item
		}
	case []bool:
		// 转换为 []interface{} 以便统一处理
		sourceSlice = make([]interface{}, len(v))
		for i, item := range v {
			sourceSlice[i] = item
		}
	default:
		// 尝试使用反射处理其他切片类型
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			length := rv.Len()
			sourceSlice = make([]interface{}, length)
			for i := 0; i < length; i++ {
				sourceSlice[i] = rv.Index(i).Interface()
			}
		} else {
			return nil, fmt.Errorf("无法将 %T 类型转换为切片或数组", value)
		}
	}

	// 获取目标元素类型
	elemType := targetType.Elem()

	// 创建目标切片
	resultSlice := reflect.MakeSlice(targetType, len(sourceSlice), len(sourceSlice))

	// 转换每个元素
	for i, item := range sourceSlice {
		convertedItem, err := fvc.ConvertToGoValue(item, elemType)
		if err != nil {
			return nil, fmt.Errorf("转换切片元素 %d 失败: %v", i, err)
		}

		if convertedItem != nil {
			resultSlice.Index(i).Set(reflect.ValueOf(convertedItem))
		}
	}

	// 处理数组类型
	if targetType.Kind() == reflect.Array {
		// 创建数组
		resultArray := reflect.New(targetType).Elem()
		if resultSlice.Len() > targetType.Len() {
			return nil, fmt.Errorf("源切片长度 %d 超过目标数组长度 %d", resultSlice.Len(), targetType.Len())
		}
		reflect.Copy(resultArray, resultSlice)
		return resultArray.Interface(), nil
	}

	return resultSlice.Interface(), nil
}
