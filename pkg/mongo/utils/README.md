# MongoDB 文档转换工具 (utils)

## 概述

本包提供了MongoDB文档到Go结构体的转换工具，支持：

- **类型安全**：完整的类型检查和转换
- **标签支持**：bson/json标签映射
- **高性能**：反射缓存和对象池优化
- **并发安全**：使用池化实例避免并发冲突
- **内存管理**：自动清理缓存避免内存泄露
- **错误友好**：清晰的错误信息

## 功能特性

### 1. 基本转换功能
- 单个文档转换
- 批量文档转换
- 单个字段转换
- Map转换

### 2. 并发安全和性能优化
- 使用对象池管理转换器实例
- 自动清理缓存避免内存泄露
- 支持多协程并发调用
- 反射缓存提高性能

### 3. 简洁的API设计
- 静态方法开箱即用
- 支持创建专用实例
- 清晰的错误信息
- 一致的使用模式

## 使用方法

### 静态方法（推荐）

```go
// 单个文档转换
var user User
err := utils.ConvertDocument(doc, &user)
if err != nil {
    log.Printf("转换失败: %v", err)
    return
}

// 批量文档转换
var users []User
err := utils.ConvertDocuments(docs, &users)
if err != nil {
    log.Printf("批量转换失败: %v", err)
    return
}

// 单个字段转换
var userID string
err := utils.ConvertSingleField(doc["_id"], &userID)

// Map转换
result, err := utils.ConvertMap(doc)
```

### 创建专用实例（DAO层推荐）

```go
// 创建转换器实例
converter := utils.NewConverter()
defer converter.ClearCache() // 清理缓存

// 使用实例方法
err := converter.ConvertDocument(doc, &user)
if err != nil {
    log.Printf("转换失败: %v", err)
    return
}
```

### 在DAO层使用

```go
type UserDAO struct {
    converter *utils.DocumentConverter
}

func NewUserDAO() *UserDAO {
    return &UserDAO{
        converter: utils.NewConverter(),
    }
}

func (dao *UserDAO) FindUsers(ctx context.Context, filter interface{}) ([]User, error) {
    // 从MongoDB获取文档
    docs, err := dao.collection.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    
    // 转换文档
    var users []User
    err = dao.converter.ConvertDocuments(docs, &users)
    if err != nil {
        return nil, fmt.Errorf("转换用户文档失败: %v", err)
    }
    
    return users, nil
}

// 清理资源
func (dao *UserDAO) Close() {
    if dao.converter != nil {
        dao.converter.ClearCache()
    }
}
```

## 支持的类型

- **基本类型**：string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool
- **时间类型**：time.Time, *time.Time
- **指针类型**：*string, *int, *time.Time 等
- **结构体类型**：任意结构体

## 标签映射

支持以下标签优先级：
1. `bson` 标签
2. `json` 标签
3. 字段名

```go
type User struct {
    ID       string    `bson:"_id" json:"id"`
    Name     string    `bson:"name" json:"name"`
    Email    string    `bson:"email" json:"email"`
    Age      int       `bson:"age" json:"age"`
    IsActive bool      `bson:"is_active" json:"is_active"`
    Created  time.Time `bson:"created_at" json:"created_at"`
}
```

## 错误处理

```go
err := utils.ConvertDocument(doc, &user)
if err != nil {
    // 错误信息包含具体的失败原因
    log.Printf("转换失败: %v", err)
    return
}
```

## 性能优化

1. **对象池**：静态方法使用对象池，避免频繁创建实例
2. **反射缓存**：结构体字段映射会被缓存，减少反射开销
3. **批量转换**：批量操作比单个转换更高效
4. **自动清理**：池化实例自动清理缓存，避免内存泄露

## 并发安全说明

- **静态方法**：使用对象池，每次调用获取独立实例，完全线程安全
- **实例方法**：使用读写锁保护，支持多协程并发调用
- **自动管理**：静态方法使用后自动清理缓存并回收实例
- **无锁读取**：缓存命中时读操作无需加锁

## 最佳实践

1. **优先使用静态方法**：简单易用，自动管理内存
2. **DAO层使用实例**：创建专用实例，手动管理生命周期
3. **批量转换**：优先使用批量转换方法
4. **错误处理**：合理处理转换错误
5. **资源清理**：实例使用完后调用ClearCache()

```go
// 简单使用 - 推荐
func SimpleUsage() {
    var user User
    err := utils.ConvertDocument(doc, &user)
    // 无需额外处理，自动清理
}

// DAO层使用 - 长期复用
type UserDAO struct {
    converter *utils.DocumentConverter
}

func NewUserDAO() *UserDAO {
    return &UserDAO{
        converter: utils.NewConverter(),
    }
}

func (dao *UserDAO) Close() {
    dao.converter.ClearCache()
}
```

## 注意事项

1. **目标类型**：转换目标必须是指针类型
2. **结构体要求**：目标类型必须是结构体或结构体指针
3. **标签格式**：bson/json标签格式必须正确
4. **类型兼容**：源数据类型必须可以转换为目标类型
5. **nil值处理**：自动处理nil值和指针类型
6. **内存管理**：静态方法自动管理，实例方法需手动清理

## 常见问题

### Q: 转换失败怎么办？
A: 检查目标类型是否正确，标签是否匹配，数据类型是否兼容。

### Q: 如何处理未知字段？
A: 未匹配的字段会被自动忽略，这是正常行为。

### Q: 性能如何优化？
A: 优先使用批量转换和静态方法，避免频繁创建实例。

### Q: 并发安全吗？
A: 是的，静态方法使用对象池，实例方法使用读写锁。

### Q: 会有内存泄露吗？
A: 不会，静态方法自动清理缓存，实例方法需手动调用ClearCache()。

### Q: 什么时候使用实例方法？
A: 在DAO层或需要长期复用的场景下使用实例方法。 