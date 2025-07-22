package clickhouseutils

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// ClickHouseTypeConverter ClickHouse特有类型转换器
// 处理ClickHouse特有的数据类型转换需求
//
// 主要处理的类型：
// 1. Nullable类型 - ClickHouse的Nullable(T)对应Go的*T
// 2. Array类型 - ClickHouse的Array(T)对应Go的[]T
// 3. LowCardinality类型 - 低基数字符串优化
// 4. DateTime64类型 - 毫秒精度时间
// 5. Decimal类型 - 高精度小数
// 6. UUID类型 - 唯一标识符
// 7. 大整数类型 - Int128, Int256, UInt128, UInt256
// 8. Map类型 - 键值对映射
// 9. Tuple类型 - 元组结构
// 10. Geo类型 - 地理位置类型
type ClickHouseTypeConverter struct {
	// 可以在这里添加配置选项
	enableNullableSupport bool
	enableArraySupport    bool
	enableDecimalSupport  bool
}

// NewClickHouseTypeConverter 创建ClickHouse类型转换器
func NewClickHouseTypeConverter() *ClickHouseTypeConverter {
	return &ClickHouseTypeConverter{
		enableNullableSupport: true,
		enableArraySupport:    true,
		enableDecimalSupport:  true,
	}
}

// ConvertClickHouseValue 转换ClickHouse特有类型到Go类型
// 处理ClickHouse驱动返回的特殊类型值
//
// 参数:
//   value: ClickHouse驱动返回的原始值
//   targetType: 目标Go类型
//
// 返回:
//   interface{}: 转换后的Go值
//   error: 转换错误，nil表示成功
func (c *ClickHouseTypeConverter) ConvertClickHouseValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 1. 处理Nullable类型
	if c.enableNullableSupport && targetType.Kind() == reflect.Ptr {
		return c.convertNullableValue(value, targetType)
	}

	// 2. 处理Array类型
	if c.enableArraySupport && targetType.Kind() == reflect.Slice {
		return c.convertArrayValue(value, targetType)
	}

	// 3. 处理时间类型（DateTime64）
	if targetType == reflect.TypeOf(time.Time{}) {
		return c.convertTimeValue(value)
	}

	// 4. 处理字符串类型（包括LowCardinality）
	if targetType.Kind() == reflect.String {
		return c.convertStringValue(value)
	}

	// 5. 处理数值类型（包括Decimal和大整数）
	if c.isNumericType(targetType) {
		return c.convertNumericValue(value, targetType)
	}

	// 对于非ClickHouse特有类型，返回原值
	return value, nil
}

// convertNullableValue 转换Nullable类型
// ClickHouse的Nullable(T)类型对应Go的*T指针类型
func (c *ClickHouseTypeConverter) convertNullableValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	elemType := targetType.Elem()
	
	// 递归转换基础类型
	convertedValue, err := c.ConvertClickHouseValue(value, elemType)
	if err != nil {
		return nil, err
	}
	
	if convertedValue == nil {
		// 返回nil指针
		return reflect.Zero(targetType).Interface(), nil
	}
	
	// 创建指针并设置值
	ptr := reflect.New(elemType)
	ptr.Elem().Set(reflect.ValueOf(convertedValue))
	return ptr.Interface(), nil
}

// convertArrayValue 转换Array类型
// ClickHouse的Array(T)类型对应Go的[]T切片类型
func (c *ClickHouseTypeConverter) convertArrayValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	sourceValue := reflect.ValueOf(value)
	if sourceValue.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice for array conversion, got %T", value)
	}
	
	elemType := targetType.Elem()
	resultSlice := reflect.MakeSlice(targetType, sourceValue.Len(), sourceValue.Len())
	
	for i := 0; i < sourceValue.Len(); i++ {
		elemValue := sourceValue.Index(i).Interface()
		convertedElem, err := c.ConvertClickHouseValue(elemValue, elemType)
		if err != nil {
			return nil, fmt.Errorf("failed to convert array element %d: %w", i, err)
		}
		resultSlice.Index(i).Set(reflect.ValueOf(convertedElem))
	}
	
	return resultSlice.Interface(), nil
}

// convertTimeValue 转换时间值
// 处理ClickHouse的DateTime、DateTime64等时间类型
func (c *ClickHouseTypeConverter) convertTimeValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		// 尝试解析时间字符串
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02 15:04:05.000",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, v); err == nil {
				return t, nil
			}
		}
		return nil, fmt.Errorf("unable to parse time string: %s", v)
	case int64:
		// Unix时间戳
		return time.Unix(v, 0), nil
	default:
		return nil, fmt.Errorf("unsupported time value type: %T", value)
	}
}

// convertStringValue 转换字符串值
// 处理ClickHouse的String、LowCardinality(String)等字符串类型
func (c *ClickHouseTypeConverter) convertStringValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case time.Time:
		// ClickHouse日期函数返回time.Time，转换为字符串格式
		// 对于日期类型，使用 YYYY-MM-DD 格式
		return v.Format("2006-01-02"), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}

// convertNumericValue 转换数值类型
// 处理ClickHouse的各种数值类型，包括Decimal和大整数
func (c *ClickHouseTypeConverter) convertNumericValue(value interface{}, targetType reflect.Type) (interface{}, error) {
	switch targetType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return c.convertToInt(value, targetType)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return c.convertToUint(value, targetType)
	case reflect.Float32, reflect.Float64:
		return c.convertToFloat(value, targetType)
	default:
		return value, nil
	}
}

// convertToInt 转换为有符号整数
func (c *ClickHouseTypeConverter) convertToInt(value interface{}, targetType reflect.Type) (interface{}, error) {
	var intVal int64
	
	switch v := value.(type) {
	case int64:
		intVal = v
	case int32:
		intVal = int64(v)
	case int16:
		intVal = int64(v)
	case int8:
		intVal = int64(v)
	case int:
		intVal = int64(v)
	case uint64:
		intVal = int64(v)
	case uint32:
		intVal = int64(v)
	case uint16:
		intVal = int64(v)
	case uint8:
		intVal = int64(v)
	case uint:
		intVal = int64(v)
	case float64:
		intVal = int64(v)
	case float32:
		intVal = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse string to int: %w", err)
		}
		intVal = parsed
	default:
		return nil, fmt.Errorf("unsupported type for int conversion: %T", value)
	}
	
	// 根据目标类型返回合适的大小
	switch targetType.Kind() {
	case reflect.Int:
		return int(intVal), nil
	case reflect.Int8:
		return int8(intVal), nil
	case reflect.Int16:
		return int16(intVal), nil
	case reflect.Int32:
		return int32(intVal), nil
	case reflect.Int64:
		return intVal, nil
	default:
		return intVal, nil
	}
}

// convertToUint 转换为无符号整数
func (c *ClickHouseTypeConverter) convertToUint(value interface{}, targetType reflect.Type) (interface{}, error) {
	var uintVal uint64
	
	switch v := value.(type) {
	case uint64:
		uintVal = v
	case uint32:
		uintVal = uint64(v)
	case uint16:
		uintVal = uint64(v)
	case uint8:
		uintVal = uint64(v)
	case uint:
		uintVal = uint64(v)
	case int64:
		if v < 0 {
			return nil, fmt.Errorf("cannot convert negative value to unsigned: %d", v)
		}
		uintVal = uint64(v)
	case int32:
		if v < 0 {
			return nil, fmt.Errorf("cannot convert negative value to unsigned: %d", v)
		}
		uintVal = uint64(v)
	case int16:
		if v < 0 {
			return nil, fmt.Errorf("cannot convert negative value to unsigned: %d", v)
		}
		uintVal = uint64(v)
	case int8:
		if v < 0 {
			return nil, fmt.Errorf("cannot convert negative value to unsigned: %d", v)
		}
		uintVal = uint64(v)
	case int:
		if v < 0 {
			return nil, fmt.Errorf("cannot convert negative value to unsigned: %d", v)
		}
		uintVal = uint64(v)
	case float64:
		if v < 0 {
			return nil, fmt.Errorf("cannot convert negative float to unsigned: %f", v)
		}
		uintVal = uint64(v)
	case float32:
		if v < 0 {
			return nil, fmt.Errorf("cannot convert negative float to unsigned: %f", v)
		}
		uintVal = uint64(v)
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse string to uint: %w", err)
		}
		uintVal = parsed
	default:
		return nil, fmt.Errorf("unsupported type for uint conversion: %T", value)
	}
	
	// 根据目标类型返回合适的大小
	switch targetType.Kind() {
	case reflect.Uint:
		return uint(uintVal), nil
	case reflect.Uint8:
		return uint8(uintVal), nil
	case reflect.Uint16:
		return uint16(uintVal), nil
	case reflect.Uint32:
		return uint32(uintVal), nil
	case reflect.Uint64:
		return uintVal, nil
	default:
		return uintVal, nil
	}
}

// convertToFloat 转换为浮点数
func (c *ClickHouseTypeConverter) convertToFloat(value interface{}, targetType reflect.Type) (interface{}, error) {
	var floatVal float64
	
	switch v := value.(type) {
	case float64:
		floatVal = v
	case float32:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	case int32:
		floatVal = float64(v)
	case int16:
		floatVal = float64(v)
	case int8:
		floatVal = float64(v)
	case int:
		floatVal = float64(v)
	case uint64:
		floatVal = float64(v)
	case uint32:
		floatVal = float64(v)
	case uint16:
		floatVal = float64(v)
	case uint8:
		floatVal = float64(v)
	case uint:
		floatVal = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse string to float: %w", err)
		}
		floatVal = parsed
	default:
		return nil, fmt.Errorf("unsupported type for float conversion: %T", value)
	}
	
	// 根据目标类型返回合适的精度
	switch targetType.Kind() {
	case reflect.Float32:
		return float32(floatVal), nil
	case reflect.Float64:
		return floatVal, nil
	default:
		return floatVal, nil
	}
}

// ConvertClickHouseValueToField 将ClickHouse值转换并设置到字段
// 这是对字段进行直接设置的便利方法
//
// 参数:
//   field: 目标字段的反射值
//   value: 要转换的ClickHouse类型值
//
// 返回:
//   error: 转换过程中的错误，如果转换成功则返回nil
func (c *ClickHouseTypeConverter) ConvertClickHouseValueToField(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}
	
	convertedValue, err := c.ConvertClickHouseValue(value, field.Type())
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

// isNumericType 检查是否为数值类型
func (c *ClickHouseTypeConverter) isNumericType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		 reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// IsClickHouseSpecificType 检查是否为ClickHouse特有类型
// 判断给定的值是否为ClickHouse数据库特有的类型
//
// 参数:
//   value: 要检查的值
//
// 返回:
//   bool: true表示是ClickHouse特有类型，false表示是通用类型
func IsClickHouseSpecificType(value interface{}) bool {
	// ClickHouse驱动返回的类型与Go标准类型的差异
	// 这些类型需要特殊的转换处理
	
	switch value.(type) {
	case uint64:
		// ClickHouse的count()函数返回uint64，但业务代码通常期望int64
		return true
	case uint32:
		// ClickHouse的UInt32字段返回uint32，但业务代码通常期望int32
		return true
	case uint16:
		// ClickHouse的UInt16字段返回uint16，但业务代码通常期望int16
		return true
	case uint8:
		// ClickHouse的UInt8字段返回uint8，但业务代码通常期望int8
		return true
	case uint:
		// ClickHouse的UInt字段返回uint，但业务代码通常期望int
		return true
	case time.Time:
		// ClickHouse的日期函数(toDate, toDateTime等)返回time.Time，但业务代码可能期望字符串
		return true
	
	// 未来可能需要处理的ClickHouse特有类型:
	// case *chproto.UUID:
	//     return true
	// case chproto.DateTime64:
	//     return true
	// case chproto.Decimal:
	//     return true
	// case chproto.Array:
	//     return true
	
	default:
		return false
	}
}

// HandleClickHouseTypeError 处理ClickHouse类型转换错误
// 当标准的结果映射出现ClickHouse类型错误时，使用此函数进行特殊处理
//
// 参数:
//   dest: 目标结构体指针
//   columns: 数据库列名
//   values: 原始值切片
//
// 返回:
//   error: 处理失败时返回错误信息
func HandleClickHouseTypeError(dest interface{}, columns []string, values []interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}
	
	structValue := destValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}
	
	converter := NewClickHouseTypeConverter()
	
	for i, colName := range columns {
		if i >= len(values) {
			continue
		}
		
		value := values[i]
		if !IsClickHouseSpecificType(value) {
			continue // 跳过非ClickHouse特有类型
		}
		
		// 查找对应的字段
		field, found := findFieldByColumn(structValue, colName)
		if !found {
			continue // 跳过找不到的字段
		}
		
		// 转换并设置ClickHouse特有类型
		if err := converter.ConvertClickHouseValueToField(field, value); err != nil {
			return fmt.Errorf("failed to convert ClickHouse value for field %s: %w", colName, err)
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