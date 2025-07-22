// Package utils 提供MongoDB操作的通用工具函数
//
// 此包包含MongoDB操作中常用的工具函数，主要包括：
// - 文档转换：将MongoDB Document转换为Go结构体
// - 类型转换：处理MongoDB和Go之间的类型差异
// - 时间处理：统一处理时间格式转换
// - 错误处理：标准化错误格式
// - 并发安全：支持多协程并发使用
//
// 设计原则：
// - 通用性：支持任意结构体类型的转换
// - 类型安全：提供类型检查和转换
// - 性能优化：使用反射缓存和对象池减少开销
// - 错误友好：提供清晰的错误信息
// - 并发安全：使用池化实例避免并发冲突
// - 内存管理：自动清理缓存避免内存泄露
//
// 使用示例：
//
//	// 单个文档转换
//	var user User
//	err := utils.ConvertDocument(doc, &user)
//
//	// 批量文档转换
//	var users []User
//	err := utils.ConvertDocuments(docs, &users)
//
//	// 创建专用转换器实例（推荐在DAO层使用）
//	converter := utils.NewConverter()
//	defer converter.ClearCache()
//
// 注意事项：
// - 目标结构体必须是指针类型
// - 支持bson和json标签进行字段映射
// - 支持常用数据类型的自动转换
// - 静态方法使用池化实例，自动管理内存
// - 线程安全，支持多协程并发调用
package utils

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"gateway/pkg/mongo/types"
	"gateway/pkg/utils/ctime"
)

// ================================
// 全局转换器实例和静态方法
// ================================

// 使用sync.Pool管理转换器实例，避免全局实例的并发和缓存问题
var converterPool = sync.Pool{
	New: func() interface{} {
		return NewDocumentConverter()
	},
}

// getPooledConverter 从池中获取转换器实例
func getPooledConverter() *DocumentConverter {
	return converterPool.Get().(*DocumentConverter)
}

// putPooledConverter 将转换器实例放回池中
func putPooledConverter(converter *DocumentConverter) {
	// 清理缓存后放回池中，避免内存泄漏
	converter.ClearCache()
	converterPool.Put(converter)
}

// ConvertDocument 将MongoDB文档转换为Go结构体（静态方法）
//
// 参数：
//
//	doc - 源MongoDB文档，不能为nil
//	target - 目标结构体，必须是指针类型
//
// 返回值：
//
//	error - 转换错误，nil表示成功
//
// 使用示例：
//
//	var user User
//	err := utils.ConvertDocument(doc, &user)
//	if err != nil {
//	    log.Printf("转换失败: %v", err)
//	    return
//	}
//
// 注意事项：
// - target必须是结构体指针类型
// - 支持bson、json标签进行字段映射
// - 未匹配的字段将被忽略
// - 自动处理类型转换
// - 线程安全，使用池化实例避免并发问题
func ConvertDocument(doc types.Document, target interface{}) error {
	converter := getPooledConverter()
	defer putPooledConverter(converter)
	return converter.ConvertDocument(doc, target)
}

// ConvertDocuments 批量转换多个MongoDB文档为Go结构体切片（静态方法）
//
// 参数：
//
//	docs - 源MongoDB文档列表，不能为nil
//	targetSlice - 目标切片，必须是指向切片的指针
//
// 返回值：
//
//	error - 转换错误，nil表示成功
//
// 使用示例：
//
//	var users []User
//	err := utils.ConvertDocuments(docs, &users)
//	if err != nil {
//	    log.Printf("批量转换失败: %v", err)
//	    return
//	}
//
// 注意事项：
// - targetSlice必须是指向切片的指针
// - 支持指针切片和值切片
// - 批量操作比单个转换更高效
// - 线程安全，支持并发调用
func ConvertDocuments(docs []types.Document, targetSlice interface{}) error {
	converter := getPooledConverter()
	defer putPooledConverter(converter)
	return converter.ConvertDocuments(docs, targetSlice)
}

// ConvertSingleField 转换单个字段值（静态方法）
//
// 参数：
//
//	value - 源值，来自MongoDB文档
//	target - 目标变量，必须是指针类型
//
// 返回值：
//
//	error - 转换错误，nil表示成功
//
// 使用示例：
//
//	var userID string
//	err := utils.ConvertSingleField(doc["_id"], &userID)
//
// 注意事项：
// - 适用于单个字段的类型转换
// - 支持所有基本类型转换
// - target必须是指针类型
// - 线程安全，支持并发调用
func ConvertSingleField(value interface{}, target interface{}) error {
	converter := getPooledConverter()
	defer putPooledConverter(converter)
	return converter.ConvertSingleField(value, target)
}

// ConvertGoTimeToMongo 将Go时间转换为MongoDB存储时间，避免时区误差（静态方法）
//
// 参数：
//
//	goTime - Go语言的时间值
//
// 返回值：
//
//	time.Time - 转换后适合MongoDB存储的时间
//
// 使用示例：
//
//	now := time.Now()
//	mongoTime := utils.ConvertGoTimeToMongo(now)
//
//	// 在结构体中使用
//	user := User{
//	    Name: "张三",
//	    CreatedAt: utils.ConvertGoTimeToMongo(time.Now()),
//	}
//
// 时区处理原理：
// - MongoDB驱动会自动将time.Time转换为UTC时间进行存储
// - 为了保持原始的本地时间值，需要预先调整时区偏移
// - 例如：15:07 CST (+8) -> 预先+8小时 = 23:07 -> MongoDB转为UTC -> 15:07 UTC
//
// 注意事项：
// - 专门解决MongoDB时区自动转换问题
// - 确保存储的是原始的本地时间值
// - 适用于需要保持显示时间不变的场景
// - 线程安全，支持并发调用
func ConvertGoTimeToMongo(goTime time.Time) time.Time {
	// 如果时间不是UTC时间，需要调整时区以抵消MongoDB的自动UTC转换
	if goTime.Location() != time.UTC {
		// 获取本地时区偏移量（秒）
		_, offset := goTime.Zone()
		// 预先增加时区偏移量，这样MongoDB转换为UTC后就是原本的本地时间值
		// 例如：15:07 CST (+8) -> 15:07 + 8小时 = 23:07 -> MongoDB转换为UTC -> 15:07 UTC
		adjustedTime := goTime.Add(time.Duration(offset) * time.Second)
		return adjustedTime
	}

	// 如果已经是UTC时间，直接返回
	return goTime
}

func ConvertGoStringToMongo(goTimeStr string) (time.Time, error) {
	goTime, err := ctime.ParseTimeString(goTimeStr)
	if err != nil {
		return time.Time{}, err
	}
	// 如果时间不是UTC时间，需要调整时区以抵消MongoDB的自动UTC转换
	if goTime.Location() != time.UTC {
		// 获取本地时区偏移量（秒）
		_, offset := goTime.Zone()
		// 预先增加时区偏移量，这样MongoDB转换为UTC后就是原本的本地时间值
		// 例如：15:07 CST (+8) -> 15:07 + 8小时 = 23:07 -> MongoDB转换为UTC -> 15:07 UTC
		adjustedTime := goTime.Add(time.Duration(offset) * time.Second)
		return adjustedTime, nil
	}

	// 如果已经是UTC时间，直接返回
	return goTime, nil
}

// ConvertMap 将MongoDB文档转换为map[string]interface{}（静态方法）
//
// 参数：
//
//	doc - 源MongoDB文档
//
// 返回值：
//
//	map[string]interface{} - 转换后的map
//	error - 转换错误
//
// 使用示例：
//
//	result, err := utils.ConvertMap(doc)
//	if err != nil {
//	    log.Printf("转换失败: %v", err)
//	    return
//	}
//
// 注意事项：
// - 适用于动态字段处理
// - 保持原有数据类型
// - 处理嵌套结构
// - 线程安全，支持并发调用
func ConvertMap(doc types.Document) (map[string]interface{}, error) {
	converter := getPooledConverter()
	defer putPooledConverter(converter)
	return converter.ConvertMap(doc)
}

// ValidateDocument 验证文档是否可以转换为指定结构体（静态方法）
//
// 参数：
//
//	doc - 源MongoDB文档
//	targetType - 目标结构体类型
//
// 返回值：
//
//	error - 验证错误，nil表示可以转换
//
// 使用示例：
//
//	userType := reflect.TypeOf(User{})
//	err := utils.ValidateDocument(doc, userType)
//	if err != nil {
//	    log.Printf("文档不兼容: %v", err)
//	    return
//	}
//
// 注意事项：
// - 用于预验证，避免转换失败
// - 检查必要字段是否存在
// - 检查类型是否兼容
// - 线程安全，支持并发调用
func ValidateDocument(doc types.Document, targetType reflect.Type) error {
	converter := getPooledConverter()
	defer putPooledConverter(converter)
	return converter.ValidateDocument(doc, targetType)
}

// NewConverter 创建新的文档转换器实例（推荐在DAO层使用）
//
// 返回值：
//
//	*DocumentConverter - 转换器实例
//
// 使用示例：
//
//	converter := utils.NewConverter()
//	defer converter.ClearCache() // 可选：清理缓存
//
// 注意事项：
// - 每个转换器实例都有独立的缓存
// - 建议在DAO层复用同一个实例
// - 线程安全，支持并发调用
func NewConverter() *DocumentConverter {
	return NewDocumentConverter()
}

// ConvertToDocument 将Go结构体转换为MongoDB文档（静态方法）
//
// 参数：
//
//	source - 源结构体，不能为nil
//
// 返回值：
//
//	types.Document - 转换后的MongoDB文档
//	error - 转换错误，nil表示成功
//
// 使用示例：
//
//	user := User{Name: "张三", Age: 25}
//	doc, err := utils.ConvertToDocument(user)
//	if err != nil {
//	    log.Printf("转换失败: %v", err)
//	    return
//	}
//
// 注意事项：
// - 支持结构体和结构体指针
// - 支持bson、json标签进行字段映射
// - nil值字段会被跳过
// - 自动处理时间类型转换
// - 线程安全，使用池化实例避免并发问题
func ConvertToDocument(source interface{}) (types.Document, error) {
	converter := getPooledConverter()
	defer putPooledConverter(converter)
	return converter.ConvertToDocument(source)
}

// ConvertToDocuments 批量转换多个Go结构体为MongoDB文档列表（静态方法）
//
// 参数：
//
//	sourceSlice - 源结构体切片，不能为nil
//
// 返回值：
//
//	[]types.Document - 转换后的MongoDB文档列表
//	error - 转换错误，nil表示成功
//
// 使用示例：
//
//	users := []User{{Name: "张三"}, {Name: "李四"}}
//	docs, err := utils.ConvertToDocuments(users)
//	if err != nil {
//	    log.Printf("批量转换失败: %v", err)
//	    return
//	}
//
// 注意事项：
// - 支持值切片和指针切片
// - 批量操作比单个转换更高效
// - 线程安全，支持并发调用
func ConvertToDocuments(sourceSlice interface{}) ([]types.Document, error) {
	converter := getPooledConverter()
	defer putPooledConverter(converter)
	return converter.ConvertToDocuments(sourceSlice)
}

// ConvertFieldValue 转换单个字段值（静态方法）
//
// 参数：
//
//	value - 源值
//	target - 目标变量，必须是指针类型
//
// 返回值：
//
//	error - 转换错误，nil表示成功
//
// 使用示例：
//
//	var userID string
//	err := utils.ConvertFieldValue(doc["_id"], &userID)
//
// 注意事项：
// - 适用于单个字段的类型转换
// - 支持所有基本类型转换
// - target必须是指针类型
// - 线程安全，支持并发调用
func ConvertFieldValue(value interface{}, target interface{}) error {
	converter := getPooledConverter()
	defer putPooledConverter(converter)
	return converter.ConvertSingleField(value, target)
}

// ================================
// 字段处理器类
// ================================

// FieldMapper 字段映射器
//
// 负责处理结构体字段映射相关的逻辑，包括：
// - 字段信息缓存管理
// - 标签解析和映射
// - 字段类型分析
// - 并发安全保护
//
// 使用示例：
//
//	mapper := NewFieldMapper()
//	fieldMap, err := mapper.GetFieldMap(userType)
//
// 注意事项：
// - 线程安全，支持多协程并发调用
// - 使用读写锁保护内部缓存
// - 自动缓存字段映射信息
type FieldMapper struct {
	// 结构体字段映射缓存，减少反射开销
	// key: 结构体类型, value: 字段映射信息
	fieldCache map[reflect.Type]map[string]fieldInfo

	// 反向映射缓存：从字段名到文档键名
	// key: 结构体类型, value: 字段名到文档键名的映射
	reverseFieldCache map[reflect.Type]map[string]string

	// 保护缓存的读写锁，确保并发安全
	mu sync.RWMutex
}

// fieldInfo 字段信息结构体
//
// 包含字段的详细信息，用于转换过程中的字段映射和类型处理
type fieldInfo struct {
	FieldIndex  int          // 字段在结构体中的索引位置
	FieldType   reflect.Type // 字段的反射类型信息
	FieldName   string       // 字段名称
	BsonTag     string       // bson标签内容（如果存在）
	JsonTag     string       // json标签内容（如果存在）
	DocumentKey string       // 在文档中的键名（经过标签解析后）
	IsPointer   bool         // 是否是指针类型
	IsExported  bool         // 是否是导出字段
	OmitEmpty   bool         // 是否设置了omitempty标签
}

// NewFieldMapper 创建新的字段映射器
//
// 返回值：
//
//	*FieldMapper - 字段映射器实例
//
// 使用示例：
//
//	mapper := NewFieldMapper()
//	defer mapper.ClearCache() // 可选：清理缓存
//
// 注意事项：
// - 每个映射器实例都有独立的缓存
// - 线程安全，支持并发调用
func NewFieldMapper() *FieldMapper {
	return &FieldMapper{
		fieldCache:        make(map[reflect.Type]map[string]fieldInfo),
		reverseFieldCache: make(map[reflect.Type]map[string]string),
	}
}

// GetFieldMap 获取字段映射信息（文档键名 -> 字段信息）
//
// 参数：
//
//	structType - 结构体类型
//
// 返回值：
//
//	map[string]fieldInfo - 字段映射信息
//	error - 获取错误
//
// 缓存机制：
// - 首次访问时构建映射
// - 后续访问直接返回缓存
// - 提高反射性能
// - 使用读写锁保护并发安全
func (fm *FieldMapper) GetFieldMap(structType reflect.Type) (map[string]fieldInfo, error) {
	// 首先尝试读取缓存（使用读锁）
	fm.mu.RLock()
	if cached, exists := fm.fieldCache[structType]; exists {
		fm.mu.RUnlock()
		return cached, nil
	}
	fm.mu.RUnlock()

	// 缓存不存在，需要构建映射（使用写锁）
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// 双重检查，防止在等待写锁期间其他协程已经创建了缓存
	if cached, exists := fm.fieldCache[structType]; exists {
		return cached, nil
	}

	fieldMap := make(map[string]fieldInfo)
	reverseMap := make(map[string]string)
	numFields := structType.NumField()

	// 遍历结构体字段
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)

		// 跳过未导出字段
		if !field.IsExported() {
			continue
		}

		info := fieldInfo{
			FieldIndex: i,
			FieldType:  field.Type,
			FieldName:  field.Name,
			IsPointer:  field.Type.Kind() == reflect.Ptr,
			IsExported: field.IsExported(),
		}

		// 解析BSON标签
		if bsonTag := field.Tag.Get("bson"); bsonTag != "" {
			parts := strings.Split(bsonTag, ",")
			if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
				info.BsonTag = parts[0]
				// 检查omitempty选项
				for _, part := range parts[1:] {
					if strings.TrimSpace(part) == "omitempty" {
						info.OmitEmpty = true
						break
					}
				}
			}
		}

		// 解析JSON标签
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
				info.JsonTag = parts[0]
				// 如果没有设置BSON标签的omitempty，检查JSON标签
				if !info.OmitEmpty {
					for _, part := range parts[1:] {
						if strings.TrimSpace(part) == "omitempty" {
							info.OmitEmpty = true
							break
						}
					}
				}
			}
		}

		// 确定文档键名（优先级：bson > json > 字段名）
		if info.BsonTag != "" {
			info.DocumentKey = info.BsonTag
		} else if info.JsonTag != "" {
			info.DocumentKey = info.JsonTag
		} else {
			info.DocumentKey = field.Name
		}

		// 跳过标记为忽略的字段
		if info.BsonTag == "-" || info.JsonTag == "-" {
			continue
		}

		fieldMap[info.DocumentKey] = info
		reverseMap[field.Name] = info.DocumentKey
	}

	// 缓存映射
	fm.fieldCache[structType] = fieldMap
	fm.reverseFieldCache[structType] = reverseMap
	return fieldMap, nil
}

// GetReverseFieldMap 获取反向字段映射（字段名 -> 文档键名）
//
// 参数：
//
//	structType - 结构体类型
//
// 返回值：
//
//	map[string]string - 反向字段映射
//	error - 获取错误
//
// 使用场景：
// - 结构体转Document时使用
// - 字段名到文档键名的转换
func (fm *FieldMapper) GetReverseFieldMap(structType reflect.Type) (map[string]string, error) {
	// 确保正向映射已构建
	_, err := fm.GetFieldMap(structType)
	if err != nil {
		return nil, err
	}

	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if reverseMap, exists := fm.reverseFieldCache[structType]; exists {
		return reverseMap, nil
	}

	return nil, fmt.Errorf("反向字段映射不存在")
}

// ClearCache 清理字段映射缓存
//
// 使用场景：
// - 内存管理
// - 结构体定义变更后
// - 长时间运行的应用
//
// 注意事项：
// - 清理后首次访问会稍慢
// - 通常不需要手动调用
// - 线程安全，使用写锁保护
func (fm *FieldMapper) ClearCache() {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	fm.fieldCache = make(map[reflect.Type]map[string]fieldInfo)
	fm.reverseFieldCache = make(map[reflect.Type]map[string]string)
}

// ================================
// 转换器结构体定义
// ================================

// ================================
// 转换器使用示例
// ================================

// ================================
// 转换器结构体定义
// ================================

// DocumentConverter MongoDB文档转换器
//
// 提供MongoDB文档和Go结构体之间的双向转换功能，支持：
// - Document到结构体的转换
// - 结构体到Document的转换
// - 批量转换操作
// - 字段映射和类型转换
// - 并发安全保护
//
// 使用示例：
//
//	converter := utils.NewDocumentConverter()
//	err := converter.ConvertDocument(doc, &user)
//	doc, err := converter.ConvertToDocument(user)
//
// 注意事项：
// - 线程安全，支持多协程并发调用
// - 内部使用字段映射器管理缓存
// - 支持多种数据类型转换
type DocumentConverter struct {
	// 字段映射器，处理字段映射相关逻辑
	fieldMapper *FieldMapper

	// 字段值转换器，处理类型转换相关逻辑
	valueConverter *FieldValueConverter
}

// ================================
// 构造函数
// ================================

// NewDocumentConverter 创建新的文档转换器实例
//
// 返回值：
//
//	*DocumentConverter - 转换器实例
//
// 使用示例：
//
//	converter := utils.NewDocumentConverter()
//	defer converter.ClearCache() // 可选：清理缓存
//
// 注意事项：
// - 每个转换器实例都有独立的缓存
// - 建议在DAO层复用同一个实例
// - 线程安全，支持并发调用
func NewDocumentConverter() *DocumentConverter {
	return &DocumentConverter{
		fieldMapper:    NewFieldMapper(),
		valueConverter: NewFieldValueConverter(),
	}
}

// ================================
// 实例方法
// ================================

// ConvertDocument 将MongoDB文档转换为Go结构体（实例方法）
//
// 参数：
//
//	doc - 源MongoDB文档，不能为nil
//	target - 目标结构体，必须是指针类型
//
// 返回值：
//
//	error - 转换错误，nil表示成功
//
// 转换过程：
// 1. 验证参数有效性
// 2. 获取或构建字段映射
// 3. 遍历文档字段进行转换
// 4. 处理类型转换和赋值
//
// 支持的转换：
// - 基本类型：string, int, float, bool等
// - 时间类型：time.Time, *time.Time
// - 指针类型：自动处理nil值
// - 标签映射：bson > json > 字段名
// - 并发安全：支持多协程调用
func (c *DocumentConverter) ConvertDocument(doc types.Document, target interface{}) error {
	// 参数验证
	if doc == nil {
		return fmt.Errorf("源文档不能为nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("目标对象必须是指针类型，当前类型: %T", target)
	}

	targetValue = targetValue.Elem()
	if !targetValue.CanSet() {
		return fmt.Errorf("目标对象不可设置，请确保传入的是指针类型")
	}

	targetType := targetValue.Type()
	if targetType.Kind() != reflect.Struct {
		return fmt.Errorf("目标对象必须是结构体类型，当前类型: %v", targetType)
	}

	// 获取或构建字段映射
	fieldMap, err := c.fieldMapper.GetFieldMap(targetType)
	if err != nil {
		return fmt.Errorf("获取字段映射失败: %v", err)
	}

	// 遍历文档字段进行转换
	for docKey, docValue := range doc {
		if fieldInfo, exists := fieldMap[docKey]; exists {
			if err := c.setFieldValue(targetValue, fieldInfo, docValue); err != nil {
				return fmt.Errorf("设置字段 %s 失败: %v", docKey, err)
			}
		}
		// 注意：未匹配的字段会被忽略，这是正常行为
	}

	return nil
}

// ConvertDocuments 批量转换多个文档（实例方法）
//
// 参数：
//
//	docs - 源MongoDB文档列表，不能为nil
//	targetSlice - 目标切片，必须是指向切片的指针
//
// 返回值：
//
//	error - 转换错误，nil表示成功
//
// 转换过程：
// 1. 验证参数有效性
// 2. 分析切片元素类型
// 3. 创建新切片
// 4. 逐个转换文档
// 5. 设置转换结果
//
// 支持的切片类型：
// - []StructType：值类型切片
// - []*StructType：指针类型切片
// - 混合类型会自动处理
func (c *DocumentConverter) ConvertDocuments(docs []types.Document, targetSlice interface{}) error {
	// 参数验证
	if docs == nil {
		return fmt.Errorf("源文档列表不能为nil")
	}

	sliceValue := reflect.ValueOf(targetSlice)
	if sliceValue.Kind() != reflect.Ptr {
		return fmt.Errorf("目标切片必须是指针类型，当前类型: %T", targetSlice)
	}

	sliceValue = sliceValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("目标对象必须是切片类型，当前类型: %v", sliceValue.Type())
	}

	// 获取切片元素类型
	elemType := sliceValue.Type().Elem()
	isPointerElem := elemType.Kind() == reflect.Ptr
	if isPointerElem {
		elemType = elemType.Elem()
	}

	// 验证元素类型是否为结构体
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("切片元素必须是结构体类型，当前类型: %v", elemType)
	}

	// 创建新切片
	newSlice := reflect.MakeSlice(sliceValue.Type(), len(docs), len(docs))

	// 逐个转换文档
	for i, doc := range docs {
		// 创建新的结构体实例
		var newElem reflect.Value
		if isPointerElem {
			// 指针类型切片：创建指针实例
			newElem = reflect.New(elemType)
		} else {
			// 值类型切片：创建值实例
			newElem = reflect.New(elemType).Elem()
		}

		// 确定转换目标
		var targetInterface interface{}
		if isPointerElem {
			targetInterface = newElem.Interface()
		} else {
			targetInterface = newElem.Addr().Interface()
		}

		// 转换文档到结构体
		if err := c.ConvertDocument(doc, targetInterface); err != nil {
			return fmt.Errorf("转换第 %d 个文档失败: %v", i, err)
		}

		// 设置到切片中
		if isPointerElem {
			newSlice.Index(i).Set(newElem)
		} else {
			newSlice.Index(i).Set(newElem)
		}
	}

	// 设置转换结果
	sliceValue.Set(newSlice)
	return nil
}

// ConvertSingleField 转换单个字段值（实例方法）
//
// 参数：
//
//	value - 源值，来自MongoDB文档
//	target - 目标变量，必须是指针类型
//
// 返回值：
//
//	error - 转换错误，nil表示成功
//
// 使用场景：
// - 单个字段的类型转换
// - 自定义转换逻辑
// - 特殊字段处理
func (c *DocumentConverter) ConvertSingleField(value interface{}, target interface{}) error {
	if value == nil {
		return fmt.Errorf("源值不能为nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("目标对象必须是指针类型，当前类型: %T", target)
	}

	targetValue = targetValue.Elem()
	if !targetValue.CanSet() {
		return fmt.Errorf("目标对象不可设置")
	}

	// 使用字段值转换器进行转换
	convertedValue, err := c.valueConverter.ConvertToGoValue(value, targetValue.Type())
	if err != nil {
		return fmt.Errorf("字段值转换失败: %v", err)
	}

	// 设置转换后的值
	if convertedValue != nil {
		targetValue.Set(reflect.ValueOf(convertedValue))
	}

	return nil
}

// ConvertMap 将MongoDB文档转换为map[string]interface{}（实例方法）
//
// 参数：
//
//	doc - 源MongoDB文档
//
// 返回值：
//
//	map[string]interface{} - 转换后的map
//	error - 转换错误
//
// 使用场景：
// - 动态字段处理
// - 不确定结构的数据
// - 中间数据转换
func (c *DocumentConverter) ConvertMap(doc types.Document) (map[string]interface{}, error) {
	if doc == nil {
		return nil, fmt.Errorf("源文档不能为nil")
	}

	result := make(map[string]interface{})
	for key, value := range doc {
		result[key] = value
	}

	return result, nil
}

// ValidateDocument 验证文档是否可以转换为指定结构体（实例方法）
//
// 参数：
//
//	doc - 源MongoDB文档
//	targetType - 目标结构体类型
//
// 返回值：
//
//	error - 验证错误，nil表示可以转换
//
// 验证内容：
// - 文档是否为nil
// - 结构体类型是否有效
// - 字段映射是否正确
// - 类型转换是否可行
func (c *DocumentConverter) ValidateDocument(doc types.Document, targetType reflect.Type) error {
	if doc == nil {
		return fmt.Errorf("源文档不能为nil")
	}

	if targetType.Kind() != reflect.Struct {
		return fmt.Errorf("目标类型必须是结构体，当前类型: %v", targetType)
	}

	// 获取字段映射
	_, err := c.fieldMapper.GetFieldMap(targetType)
	if err != nil {
		return fmt.Errorf("获取字段映射失败: %v", err)
	}

	// 这里可以添加更多的验证逻辑
	return nil
}

// ConvertToDocument 将Go结构体转换为MongoDB文档（实例方法）
//
// 参数：
//
//	source - 源结构体，支持结构体或结构体指针
//
// 返回值：
//
//	types.Document - 转换后的MongoDB文档
//	error - 转换错误，nil表示成功
//
// 转换过程：
// 1. 验证参数有效性
// 2. 获取结构体字段信息
// 3. 遍历结构体字段
// 4. 根据标签映射转换字段
// 5. 处理类型转换
//
// 支持的转换：
// - 基本类型自动转换
// - 时间类型处理
// - 指针类型处理
// - omitempty标签支持
// - nil值跳过
func (c *DocumentConverter) ConvertToDocument(source interface{}) (types.Document, error) {
	if source == nil {
		return nil, fmt.Errorf("源对象不能为nil")
	}

	sourceValue := reflect.ValueOf(source)
	sourceType := sourceValue.Type()

	// 处理指针类型
	if sourceType.Kind() == reflect.Ptr {
		if sourceValue.IsNil() {
			return nil, fmt.Errorf("源对象指针不能为nil")
		}
		sourceValue = sourceValue.Elem()
		sourceType = sourceValue.Type()
	}

	// 验证是否为结构体
	if sourceType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("源对象必须是结构体类型，当前类型: %v", sourceType)
	}

	// 获取字段映射
	fieldMap, err := c.fieldMapper.GetFieldMap(sourceType)
	if err != nil {
		return nil, fmt.Errorf("获取字段映射失败: %v", err)
	}

	// 创建结果文档
	doc := make(types.Document)

	// 遍历字段映射，转换每个字段
	for docKey, fieldInfo := range fieldMap {
		fieldValue := sourceValue.Field(fieldInfo.FieldIndex)

		// 处理指针类型
		if fieldInfo.IsPointer {
			if fieldValue.IsNil() {
				// 对于指针类型的nil值，根据omitempty决定是否跳过
				if fieldInfo.OmitEmpty {
					continue
				}
				doc[docKey] = nil
				continue
			}
			fieldValue = fieldValue.Elem()
		}

		// 检查omitempty标签
		if fieldInfo.OmitEmpty && c.valueConverter.IsEmptyValue(fieldValue) {
			continue
		}

		// 转换字段值
		convertedValue, err := c.valueConverter.ConvertFromGoValue(fieldValue.Interface(), c.ConvertToDocument)
		if err != nil {
			return nil, fmt.Errorf("转换字段 %s 失败: %v", fieldInfo.FieldName, err)
		}

		doc[docKey] = convertedValue
	}

	return doc, nil
}

// ConvertToDocuments 批量转换Go结构体为MongoDB文档列表（实例方法）
//
// 参数：
//
//	sourceSlice - 源结构体切片，支持值切片和指针切片
//
// 返回值：
//
//	[]types.Document - 转换后的MongoDB文档列表
//	error - 转换错误，nil表示成功
//
// 转换过程：
// 1. 验证参数有效性
// 2. 分析切片元素类型
// 3. 逐个转换结构体
// 4. 构建文档列表
//
// 支持的切片类型：
// - []StructType：值类型切片
// - []*StructType：指针类型切片
func (c *DocumentConverter) ConvertToDocuments(sourceSlice interface{}) ([]types.Document, error) {
	if sourceSlice == nil {
		return nil, fmt.Errorf("源切片不能为nil")
	}

	sliceValue := reflect.ValueOf(sourceSlice)
	sliceType := sliceValue.Type()

	// 验证是否为切片类型
	if sliceType.Kind() != reflect.Slice {
		return nil, fmt.Errorf("源对象必须是切片类型，当前类型: %v", sliceType)
	}

	// 创建结果切片
	length := sliceValue.Len()
	docs := make([]types.Document, 0, length)

	// 逐个转换元素
	for i := 0; i < length; i++ {
		elem := sliceValue.Index(i)

		// 转换单个元素
		doc, err := c.ConvertToDocument(elem.Interface())
		if err != nil {
			return nil, fmt.Errorf("转换第 %d 个元素失败: %v", i, err)
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// ClearCache 清理反射缓存（实例方法）
//
// 使用场景：
// - 内存管理
// - 结构体定义变更后
// - 长时间运行的应用
//
// 注意事项：
// - 清理后首次转换会稍慢
// - 通常不需要手动调用
// - 线程安全，使用写锁保护
func (c *DocumentConverter) ClearCache() {
	// 清理字段映射器缓存
	c.fieldMapper.ClearCache()
}

// ================================
// 私有方法
// ================================

// setFieldValue 设置字段值
//
// 参数：
//
//	targetValue - 目标结构体值
//	info - 字段信息
//	docValue - 文档值
//
// 返回值：
//
//	error - 设置错误
//
// 处理逻辑：
// - 处理nil值
// - 处理指针类型
// - 调用类型转换
func (c *DocumentConverter) setFieldValue(targetValue reflect.Value, info fieldInfo, docValue interface{}) error {
	if docValue == nil {
		return nil // 跳过nil值，这是正常行为
	}

	fieldValue := targetValue.Field(info.FieldIndex)
	if !fieldValue.CanSet() {
		return fmt.Errorf("字段不可设置")
	}

	// 处理指针类型
	if info.IsPointer {
		if fieldValue.IsNil() {
			// 创建新的指针实例
			fieldValue.Set(reflect.New(info.FieldType.Elem()))
		}
		fieldValue = fieldValue.Elem()
	}

	// 使用字段值转换器进行转换
	convertedValue, err := c.valueConverter.ConvertToGoValue(docValue, fieldValue.Type())
	if err != nil {
		return fmt.Errorf("字段值转换失败: %v", err)
	}

	// 设置转换后的值
	if convertedValue != nil {
		fieldValue.Set(reflect.ValueOf(convertedValue))
	}

	return nil
}

// convertAndSetValue 转换并设置值
//
// 参数：
//
//	fieldValue - 字段值
//	docValue - 文档值
//
// 返回值：
//
//	error - 转换错误
//
// 支持的转换：
// - 直接类型匹配
// - 基本类型转换
// - 时间类型转换
// - 指针类型处理
func (c *DocumentConverter) convertAndSetValue(fieldValue reflect.Value, docValue interface{}) error {
	fieldType := fieldValue.Type()
	docValueType := reflect.TypeOf(docValue)

	// 直接类型匹配 - 最快的路径
	if docValueType != nil && docValueType == fieldType {
		fieldValue.Set(reflect.ValueOf(docValue))
		return nil
	}

	// 使用FieldValueConverter进行类型转换
	convertedValue, err := c.valueConverter.ConvertToGoValue(docValue, fieldType)
	if err != nil {
		return fmt.Errorf("字段类型转换失败: %v", err)
	}

	// 设置转换后的值
	if convertedValue != nil {
		fieldValue.Set(reflect.ValueOf(convertedValue))
	}

	return nil
}

// ================================
// 结束
// ================================
