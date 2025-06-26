package sqlutils

import (
	"database/sql"
	"fmt"
	"gohub/pkg/database"
	"reflect"
	"strings"
	"time"
)

// FieldMapper 字段映射器，用于处理数据库列到结构体字段的映射
type FieldMapper struct {
	columns     []string                    // 数据库列名列表，按查询结果顺序排列
	columnIndex map[string]int              // 列名到索引的快速查找映射表
	structInfo  *StructInfo                 // 目标结构体的详细信息
}

// StructInfo 结构体信息
type StructInfo struct {
	fields    []FieldInfo            // 可映射字段信息列表
	fieldMap  map[string]*FieldInfo  // 数据库字段名到字段信息的快速查找映射
	value     reflect.Value          // 结构体的反射值
}

// FieldInfo 字段信息
type FieldInfo struct {
	field     reflect.Value  // 字段的反射值
	dbName    string         // 对应的数据库字段名
	fieldType reflect.Type   // 字段的Go类型
	index     int            // 字段在结构体中的索引位置
}

// NewFieldMapper 创建字段映射器
func NewFieldMapper(columns []string, dest interface{}) (*FieldMapper, error) {
	structInfo, err := analyzeStruct(dest)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze struct: %w", err)
	}

	columnIndex := make(map[string]int, len(columns))
	for i, col := range columns {
		columnIndex[col] = i
	}

	return &FieldMapper{
		columns:     columns,
		columnIndex: columnIndex,
		structInfo:  structInfo,
	}, nil
}

// analyzeStruct 分析结构体
func analyzeStruct(dest interface{}) (*StructInfo, error) {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("dest must be a pointer, got %T", dest)
	}

	structValue := destValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("dest must be a pointer to struct, got pointer to %s", structValue.Kind())
	}

	structType := structValue.Type()
	info := &StructInfo{
		fields:   make([]FieldInfo, 0, structValue.NumField()),
		fieldMap: make(map[string]*FieldInfo),
		value:    structValue,
	}

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		structField := structType.Field(i)

		if !field.CanSet() {
			continue
		}

		dbTag := structField.Tag.Get("db")
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			dbTag = strings.ToLower(structField.Name)
		}

		fieldInfo := FieldInfo{
			field:     field,
			dbName:    dbTag,
			fieldType: field.Type(),
			index:     i,
		}

		info.fields = append(info.fields, fieldInfo)
		info.fieldMap[dbTag] = &info.fields[len(info.fields)-1]
	}

	return info, nil
}

// IsFieldCountMatched 检查字段数量是否匹配
func (fm *FieldMapper) IsFieldCountMatched() bool {
	return len(fm.columns) == len(fm.structInfo.fields)
}

// GetStructFieldCount 获取结构体字段数量
func (fm *FieldMapper) GetStructFieldCount() int {
	return len(fm.structInfo.fields)
}

// GetColumnCount 获取数据库列数量
func (fm *FieldMapper) GetColumnCount() int {
	return len(fm.columns)
}

// MapValues 将扫描到的值映射到结构体字段
func (fm *FieldMapper) MapValues(values []interface{}) error {
	if len(values) != len(fm.columns) {
		return fmt.Errorf("values count %d does not match columns count %d", len(values), len(fm.columns))
	}

	for i, colName := range fm.columns {
		fieldInfo, exists := fm.structInfo.fieldMap[colName]
		if !exists {
			continue
		}

		value := values[i]
		if err := fm.setFieldValue(fieldInfo, value); err != nil {
			return fmt.Errorf("failed to set field %s (column %d): %w", colName, i, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值
func (fm *FieldMapper) setFieldValue(fieldInfo *FieldInfo, value interface{}) error {
	if value == nil {
		if fieldInfo.field.Kind() == reflect.Ptr {
			fieldInfo.field.Set(reflect.Zero(fieldInfo.field.Type()))
		} else {
			fieldInfo.field.Set(reflect.Zero(fieldInfo.field.Type()))
		}
		return nil
	}

	return fm.convertAndSetValue(fieldInfo.field, value)
}

// convertAndSetValue 转换并设置值
func (fm *FieldMapper) convertAndSetValue(field reflect.Value, value interface{}) error {
	fieldType := field.Type()
	valueType := reflect.TypeOf(value)

	if valueType.AssignableTo(fieldType) {
		field.Set(reflect.ValueOf(value))
		return nil
	}

	if fieldType.Kind() == reflect.Ptr {
		elemType := fieldType.Elem()
		if valueType.AssignableTo(elemType) {
			newValue := reflect.New(elemType)
			newValue.Elem().Set(reflect.ValueOf(value))
			field.Set(newValue)
			return nil
		}
	}

	return fm.convertValue(field, value)
}

// convertValue 转换值类型
func (fm *FieldMapper) convertValue(field reflect.Value, value interface{}) error {
	fieldType := field.Type()
	
	switch v := value.(type) {
	case []byte:
		if fieldType.Kind() == reflect.String {
			field.SetString(string(v))
			return nil
		}
		// 处理指针类型的字符串字段
		if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.String {
			strValue := string(v)
			field.Set(reflect.ValueOf(&strValue))
			return nil
		}
	case int64:
		switch fieldType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field.SetInt(v)
			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v >= 0 {
				field.SetUint(uint64(v))
				return nil
			}
		case reflect.Ptr:
			// 处理指针类型的数值字段
			elemType := fieldType.Elem()
			switch elemType.Kind() {
			case reflect.Int:
				intValue := int(v)
				field.Set(reflect.ValueOf(&intValue))
				return nil
			case reflect.Int8:
				intValue := int8(v)
				field.Set(reflect.ValueOf(&intValue))
				return nil
			case reflect.Int16:
				intValue := int16(v)
				field.Set(reflect.ValueOf(&intValue))
				return nil
			case reflect.Int32:
				intValue := int32(v)
				field.Set(reflect.ValueOf(&intValue))
				return nil
			case reflect.Int64:
				field.Set(reflect.ValueOf(&v))
				return nil
			case reflect.Uint:
				if v >= 0 {
					uintValue := uint(v)
					field.Set(reflect.ValueOf(&uintValue))
					return nil
				}
			case reflect.Uint8:
				if v >= 0 {
					uintValue := uint8(v)
					field.Set(reflect.ValueOf(&uintValue))
					return nil
				}
			case reflect.Uint16:
				if v >= 0 {
					uintValue := uint16(v)
					field.Set(reflect.ValueOf(&uintValue))
					return nil
				}
			case reflect.Uint32:
				if v >= 0 {
					uintValue := uint32(v)
					field.Set(reflect.ValueOf(&uintValue))
					return nil
				}
			case reflect.Uint64:
				if v >= 0 {
					uintValue := uint64(v)
					field.Set(reflect.ValueOf(&uintValue))
					return nil
				}
			}
		}
	case float64:
		if fieldType.Kind() == reflect.Float32 || fieldType.Kind() == reflect.Float64 {
			field.SetFloat(v)
			return nil
		}
		// 处理指针类型的浮点字段
		if fieldType.Kind() == reflect.Ptr {
			elemType := fieldType.Elem()
			if elemType.Kind() == reflect.Float32 {
				floatValue := float32(v)
				field.Set(reflect.ValueOf(&floatValue))
				return nil
			} else if elemType.Kind() == reflect.Float64 {
				field.Set(reflect.ValueOf(&v))
				return nil
			}
		}
	case string:
		if fieldType.Kind() == reflect.String {
			field.SetString(v)
			return nil
		}
		// 处理指针类型的字符串字段
		if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.String {
			field.Set(reflect.ValueOf(&v))
			return nil
		}
	case time.Time:
		if fieldType == reflect.TypeOf(time.Time{}) {
			field.Set(reflect.ValueOf(v))
			return nil
		}
		// 处理指针类型的时间字段
		if fieldType == reflect.TypeOf(&time.Time{}) {
			field.Set(reflect.ValueOf(&v))
			return nil
		}
	case bool:
		if fieldType.Kind() == reflect.Bool {
			field.SetBool(v)
			return nil
		}
		// 处理指针类型的布尔字段
		if fieldType.Kind() == reflect.Ptr && fieldType.Elem().Kind() == reflect.Bool {
			field.Set(reflect.ValueOf(&v))
			return nil
		}
	}

	return fmt.Errorf("cannot convert %T to %s", value, fieldType.String())
}

// CreateInterfaceSlice 创建接口切片用于扫描
func CreateInterfaceSlice(columnCount int) []interface{} {
	values := make([]interface{}, columnCount)
	for i := range values {
		values[i] = new(interface{})
	}
	return values
}

// ExtractValues 从接口切片中提取实际值
func ExtractValues(values []interface{}) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		if ptr, ok := v.(*interface{}); ok {
			result[i] = *ptr
		} else {
			result[i] = v
		}
	}
	return result
}

// ScanRows 扫描多行结果到目标切片
// 将SQL查询返回的多行结果扫描到Go切片中
// 使用优雅的接口切片扫描方式，支持字段数量不匹配的情况
//
// 功能特性：
// - 智能字段匹配：支持数据库列数与结构体字段数不匹配的情况
// - 自动类型转换：安全处理NULL值和类型转换
// - 高性能扫描：字段匹配时使用传统高效方式，不匹配时使用灵活方式
// - db tag支持：通过tag映射自定义字段名
//
// 参数:
//   rows: SQL查询返回的行结果集
//   dest: 目标切片的指针，元素类型应为结构体或结构体指针
// 返回:
//   error: 扫描失败时返回错误信息
//
// 使用示例:
//   var users []User
//   err := ScanRows(rows, &users)
func ScanRows(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	sliceValue := destValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	elementType := sliceValue.Type().Elem()
	isPtr := elementType.Kind() == reflect.Ptr
	if isPtr {
		elementType = elementType.Elem()
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 创建一个临时结构体实例用于分析字段
	tempElement := reflect.New(elementType)
	fieldMapper, err := NewFieldMapper(columns, tempElement.Interface())
	if err != nil {
		return fmt.Errorf("failed to create field mapper: %v", err)
	}

	// 检查字段数量是否匹配，如果匹配则使用传统方式（更高效）
	if fieldMapper.IsFieldCountMatched() {
		return ScanRowsTraditional(rows, dest, columns, elementType, isPtr, sliceValue)
	}

	// 使用优雅的接口切片扫描方式
	return ScanRowsWithInterfaceSlice(rows, dest, columns, elementType, isPtr, sliceValue, fieldMapper)
}

// ScanOneRow 扫描单行结果到目标结构体（智能版本）
// 使用sql.Rows的智能字段映射功能处理单行查询结果
// 支持字段数量不匹配和动态字段映射
//
// 功能特性：
// - 智能字段匹配：支持数据库列数与结构体字段数不匹配
// - 自动类型转换：安全处理NULL值和类型转换
// - db tag支持：通过tag映射自定义字段名
// - 记录不存在检测：自动返回database.ErrRecordNotFound
//
// 参数:
//   rows: SQL查询返回的行结果集
//   dest: 目标结构体的指针
// 返回:
//   error: 扫描失败或记录不存在时返回错误信息
//
// 使用示例:
//   var user User
//   err := ScanOneRow(rows, &user)
//   if err == database.ErrRecordNotFound {
//       // 处理记录不存在
//   }
func ScanOneRow(rows *sql.Rows, dest interface{}) error {
	defer rows.Close()
	
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return database.ErrRecordNotFound
	}
	
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	structValue := destValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 创建字段映射器
	fieldMapper, err := NewFieldMapper(columns, dest)
	if err != nil {
		return fmt.Errorf("failed to create field mapper: %v", err)
	}

	// 检查字段数量是否匹配，如果匹配则使用传统方式（更高效）
	if fieldMapper.IsFieldCountMatched() {
		// 使用传统方式扫描
		scanTargets, fields := PrepareScanTargetsWithFields(structValue, columns)
		if len(scanTargets) == 0 {
			return fmt.Errorf("no valid scan targets prepared")
		}

		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}

		return ProcessScannedValues(scanTargets, fields)
	}

	// 使用智能接口切片扫描方式
	columnCount := len(columns)
	scanValues := CreateInterfaceSlice(columnCount)
	
	if err := rows.Scan(scanValues...); err != nil {
		return err
	}
	
	actualValues := ExtractValues(scanValues)
	
	return fieldMapper.MapValues(actualValues)
}

// ScanRow 扫描单行结果到目标结构体
// 将SQL查询返回的单行结果扫描到Go结构体中
// 注意：由于sql.Row没有Columns方法，这里使用简化的按字段顺序扫描
//
// 功能限制：
// - 要求数据库列顺序与结构体字段顺序一致
// - 无法处理列数不匹配的情况
// - 建议在可能的情况下使用Query+ScanRows替代
//
// 参数:
//   row: SQL查询返回的单行结果
//   dest: 目标结构体的指针
// 返回:
//   error: 扫描失败时返回错误信息
//
// 使用示例:
//   var user User
//   err := ScanRow(row, &user)
func ScanRow(row *sql.Row, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	structValue := destValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	// 由于sql.Row没有Columns方法，这里使用传统的按字段顺序扫描
	// 这是QueryOne方法的限制，建议在可能的情况下使用Query方法
	structType := structValue.Type()
	var scanTargets []interface{}
	var fields []reflect.Value

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		if !field.CanSet() {
			continue
		}
		
		structField := structType.Field(i)
		dbTag := structField.Tag.Get("db")
		if dbTag == "-" {
			continue
		}

		// 创建NULL值安全的扫描目标
		scanTarget := CreateNullSafeScanTarget(field)
		scanTargets = append(scanTargets, scanTarget)
		fields = append(fields, field)
	}

	if err := row.Scan(scanTargets...); err != nil {
		if err == sql.ErrNoRows {
			return database.ErrRecordNotFound
		}
		return err
	}

	// 处理扫描后的值转换
	return ProcessScannedValues(scanTargets, fields)
}

// PrepareScanTargetsWithFields 准备扫描目标并返回对应的字段
// 为scanRows函数提供的增强版本，同时返回扫描目标和对应字段
//
// 功能特性：
// - 自动字段匹配：根据db tag和字段名匹配数据库列
// - NULL值安全：为每个字段创建对应的NULL安全扫描目标
// - 智能跳过：对于找不到或无法设置的字段使用丢弃变量
//
// 参数:
//   structValue: 目标结构体的反射值
//   columns: 数据库列名切片
// 返回:
//   []interface{}: 扫描目标切片，每个元素对应一个数据库列
//   []reflect.Value: 对应的结构体字段切片
func PrepareScanTargetsWithFields(structValue reflect.Value, columns []string) ([]interface{}, []reflect.Value) {
	var scanTargets []interface{}
	var fields []reflect.Value

	for _, column := range columns {
		field, found := FindFieldByColumn(structValue, column)
		if !found {
			// 如果找不到对应字段，使用一个丢弃变量
			var discard interface{}
			scanTargets = append(scanTargets, &discard)
			fields = append(fields, reflect.Value{}) // 空值占位
			continue
		}

		if !field.CanSet() {
			// 字段不可设置，使用丢弃变量
			var discard interface{}
			scanTargets = append(scanTargets, &discard)
			fields = append(fields, reflect.Value{}) // 空值占位
			continue
		}

		// 创建NULL值安全的扫描目标
		scanTarget := CreateNullSafeScanTarget(field)
		scanTargets = append(scanTargets, scanTarget)
		fields = append(fields, field)
	}

	return scanTargets, fields
}

// FindFieldByColumn 根据列名查找对应的结构体字段
// 通过db tag或字段名（转小写）匹配数据库列名
// 支持db tag映射，优先使用tag定义的名称
//
// 匹配规则：
// - 优先匹配db tag定义的名称
// - 如果没有db tag，使用字段名转小写匹配
// - 支持忽略字段：db:"-"
//
// 参数:
//   structValue: 要搜索的结构体反射值
//   column: 要匹配的数据库列名
// 返回:
//   reflect.Value: 找到的字段反射值
//   bool: 是否找到匹配的字段
func FindFieldByColumn(structValue reflect.Value, column string) (reflect.Value, bool) {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		structField := structType.Field(i)

		// 获取数据库字段名
		dbTag := structField.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(structField.Name)
		}

		if dbTag == column {
			return field, true
		}
	}

	return reflect.Value{}, false
}

// CreateNullSafeScanTarget 创建NULL值安全的扫描目标
// 根据字段类型创建相应的sql.NullXXX类型，用于安全扫描可能为NULL的数据库值
//
// 支持的类型映射：
// - string -> sql.NullString
// - int/int64 -> sql.NullInt64
// - float64 -> sql.NullFloat64
// - bool -> sql.NullBool  
// - time.Time -> sql.NullTime
// - 指针类型 -> 对应基础类型的NULL版本
//
// 参数:
//   field: 目标字段的反射值
// 返回:
//   interface{}: 扫描目标，可以是sql.NullString、sql.NullInt64等
func CreateNullSafeScanTarget(field reflect.Value) interface{} {
	fieldType := field.Type()
	
	switch fieldType.Kind() {
	case reflect.String:
		return &sql.NullString{}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &sql.NullInt64{}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &sql.NullInt64{} // 使用Int64处理无符号整数
	case reflect.Float32, reflect.Float64:
		return &sql.NullFloat64{}
	case reflect.Bool:
		return &sql.NullBool{}
	case reflect.Ptr:
		// 如果是指针类型，创建对应基础类型的NULL扫描目标
		elemType := fieldType.Elem()
		switch elemType.Kind() {
		case reflect.String:
			return &sql.NullString{}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return &sql.NullInt64{}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return &sql.NullInt64{}
		case reflect.Float32, reflect.Float64:
			return &sql.NullFloat64{}
		case reflect.Bool:
			return &sql.NullBool{}
		default:
			if elemType == reflect.TypeOf(time.Time{}) {
				return &sql.NullTime{}
			}
		}
	case reflect.Struct:
		// 特殊处理时间类型
		if fieldType == reflect.TypeOf(time.Time{}) {
			return &sql.NullTime{}
		}
	}
	
	// 如果无法确定类型，返回通用接口
	var discard interface{}
	return &discard
}

// ProcessScannedValues 处理扫描后的值转换
// 将sql.NullXXX类型的值转换为目标字段类型，处理NULL值
//
// 转换规则：
// - sql.NullString -> string 或 *string
// - sql.NullInt64 -> int/int64/uint 或对应指针类型
// - sql.NullFloat64 -> float32/float64
// - sql.NullBool -> bool
// - sql.NullTime -> time.Time 或 *time.Time
// - NULL值处理：设置为零值或nil（指针类型）
//
// 参数:
//   scanTargets: 扫描目标切片
//   fields: 目标字段切片
// 返回:
//   error: 转换失败时返回错误信息
func ProcessScannedValues(scanTargets []interface{}, fields []reflect.Value) error {
	for i, scanTarget := range scanTargets {
		if i >= len(fields) {
			continue
		}
		
		field := fields[i]
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		
		// 根据扫描目标类型处理值转换
		switch v := scanTarget.(type) {
		case *sql.NullString:
			if field.Kind() == reflect.Ptr {
				// 处理指针类型字段
				if v.Valid {
					strValue := v.String
					field.Set(reflect.ValueOf(&strValue))
				} else {
					field.Set(reflect.Zero(field.Type()))
				}
			} else {
				// 处理非指针类型字段
				if v.Valid {
					field.SetString(v.String)
				} else {
					field.SetString("")
				}
			}
		case *sql.NullInt64:
			if field.Kind() == reflect.Ptr {
				// 处理指针类型字段
				if v.Valid {
					elemType := field.Type().Elem()
					switch elemType.Kind() {
					case reflect.Int:
						intValue := int(v.Int64)
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Int8:
						intValue := int8(v.Int64)
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Int16:
						intValue := int16(v.Int64)
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Int32:
						intValue := int32(v.Int64)
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Int64:
						intValue := v.Int64
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Uint:
						if v.Int64 >= 0 {
							uintValue := uint(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					case reflect.Uint8:
						if v.Int64 >= 0 {
							uintValue := uint8(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					case reflect.Uint16:
						if v.Int64 >= 0 {
							uintValue := uint16(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					case reflect.Uint32:
						if v.Int64 >= 0 {
							uintValue := uint32(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					case reflect.Uint64:
						if v.Int64 >= 0 {
							uintValue := uint64(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					}
				} else {
					field.Set(reflect.Zero(field.Type()))
				}
			} else {
				// 处理非指针类型字段
				if v.Valid {
					switch field.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.SetInt(v.Int64)
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						if v.Int64 >= 0 {
							field.SetUint(uint64(v.Int64))
						} else {
							field.SetUint(0)
						}
					}
				} else {
					switch field.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.SetInt(0)
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						field.SetUint(0)
					}
				}
			}
		case *sql.NullFloat64:
			if v.Valid {
				field.SetFloat(v.Float64)
			} else {
				field.SetFloat(0)
			}
		case *sql.NullBool:
			if v.Valid {
				field.SetBool(v.Bool)
			} else {
				field.SetBool(false)
			}
		case *sql.NullTime:
			if v.Valid {
				if field.Type() == reflect.TypeOf(time.Time{}) {
					field.Set(reflect.ValueOf(v.Time))
				} else if field.Type() == reflect.TypeOf(&time.Time{}) {
					field.Set(reflect.ValueOf(&v.Time))
				}
			} else {
				if field.Type() == reflect.TypeOf(time.Time{}) {
					field.Set(reflect.ValueOf(time.Time{}))
				} else if field.Type() == reflect.TypeOf(&time.Time{}) {
					field.Set(reflect.ValueOf((*time.Time)(nil)))
				}
			}
		default:
			// 对于其他类型，不做处理
		}
	}
	
	return nil
}

// ScanRowsTraditional 传统方式扫描多行结果（字段数量匹配时使用）
// 当数据库列数与结构体字段数匹配时，使用此方法以获得更好的性能
//
// 性能优势：
// - 直接字段映射，避免接口切片的开销
// - 减少反射操作，提高扫描速度
// - 适合标准的ORM场景
//
// 参数:
//   rows: SQL查询返回的行结果集
//   dest: 目标切片的指针
//   columns: 数据库列名切片
//   elementType: 切片元素类型
//   isPtr: 元素是否为指针类型
//   sliceValue: 切片的反射值
// 返回:
//   error: 扫描失败时返回错误信息
func ScanRowsTraditional(rows *sql.Rows, dest interface{}, columns []string, elementType reflect.Type, isPtr bool, sliceValue reflect.Value) error {
	for rows.Next() {
		// 创建新的结构体实例
		newElement := reflect.New(elementType)
		
		// 准备扫描目标（包含NULL值安全处理）
		scanTargets, fields := PrepareScanTargetsWithFields(newElement.Elem(), columns)
		if len(scanTargets) == 0 {
			return fmt.Errorf("no valid scan targets prepared")
		}

		// 扫描行数据
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}

		// 处理扫描后的值转换
		if err := ProcessScannedValues(scanTargets, fields); err != nil {
			return err
		}

		// 添加到切片
		if isPtr {
			sliceValue.Set(reflect.Append(sliceValue, newElement))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, newElement.Elem()))
		}
	}

	return rows.Err()
}

// ScanRowsWithInterfaceSlice 使用接口切片扫描多行结果（字段数量不匹配时使用）
// 当数据库列数与结构体字段数不匹配时，使用此优雅的方法
//
// 适用场景：
// - 数据库列多于结构体字段
// - 结构体字段多于数据库列
// - 列顺序与字段顺序不一致
// - 需要动态字段映射的场景
//
// 参数:
//   rows: SQL查询返回的行结果集
//   dest: 目标切片的指针
//   columns: 数据库列名切片
//   elementType: 切片元素类型
//   isPtr: 元素是否为指针类型
//   sliceValue: 切片的反射值
//   fieldMapper: 字段映射器
// 返回:
//   error: 扫描失败时返回错误信息
func ScanRowsWithInterfaceSlice(rows *sql.Rows, dest interface{}, columns []string, elementType reflect.Type, isPtr bool, sliceValue reflect.Value, fieldMapper *FieldMapper) error {
	columnCount := len(columns)
	
	for rows.Next() {
		// 创建新的结构体实例
		newElement := reflect.New(elementType)
		
		// 创建接口切片用于扫描所有列
		scanValues := CreateInterfaceSlice(columnCount)
		
		// 扫描所有列到接口切片
		if err := rows.Scan(scanValues...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}
		
		// 提取实际值
		actualValues := ExtractValues(scanValues)
		
		// 创建新的字段映射器用于当前行
		currentFieldMapper, err := NewFieldMapper(columns, newElement.Interface())
		if err != nil {
			return fmt.Errorf("failed to create field mapper for current row: %v", err)
		}
		
		// 将值映射到结构体字段
		if err := currentFieldMapper.MapValues(actualValues); err != nil {
			return fmt.Errorf("failed to map values to struct: %v", err)
		}

		// 添加到切片
		if isPtr {
			sliceValue.Set(reflect.Append(sliceValue, newElement))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, newElement.Elem()))
		}
	}

	return rows.Err()
} 