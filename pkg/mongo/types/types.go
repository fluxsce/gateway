// Package types 提供MongoDB操作的类型定义
//
// 此包定义了MongoDB操作中使用的所有类型，包括：
// - 基础类型：Document、Filter、Update等
// - 操作选项：各种操作的配置选项
// - 结果类型：操作返回的结果
// - 接口定义：MongoDB操作的核心接口
//
// 设计原则：
// - 类型安全：提供强类型的操作接口
// - 灵活性：支持多种操作模式和配置
// - 兼容性：与MongoDB驱动兼容
// - 可扩展性：易于添加新功能
//
// 文件结构：
// - 基础类型定义：定义基本的数据结构
// - 操作选项类型：定义各种操作的配置参数
// - 原子操作选项：定义原子操作的特殊配置
// - 操作结果类型：定义操作返回的结果结构
// - 索引相关类型：定义索引操作的相关类型
// - 会话和事务类型：定义会话和事务的配置
// - 其他配置类型：定义其他辅助配置
// - 核心接口定义：定义MongoDB操作的核心接口
package types

import (
	"context"
	"time"
)

// === 基础类型定义 ===

// Document 表示MongoDB文档的通用类型
// 使用map[string]interface{}作为底层类型，提供灵活的文档结构支持
// 可以存储任意的键值对，值可以是任何类型
//
// 示例:
//   doc := Document{
//       "name": "John",
//       "age": 30,
//       "email": "john@example.com",
//   }
type Document map[string]interface{}

// Filter 表示MongoDB查询过滤器
// 用于指定查询条件，支持MongoDB的所有查询操作符
// 如$eq, $ne, $gt, $gte, $lt, $lte, $in, $nin等
//
// 示例:
//   filter := Filter{
//       "age": Document{"$gte": 18},
//       "status": "active",
//   }
type Filter map[string]interface{}

// Update 表示MongoDB更新操作
// 用于指定更新操作，支持MongoDB的所有更新操作符
// 如$set, $inc, $push, $pull, $unset等
//
// 示例:
//   update := Update{
//       "$set": Document{"status": "updated"},
//       "$inc": Document{"count": 1},
//   }
type Update map[string]interface{}

// Sort 表示MongoDB排序条件
// 键为字段名，值为排序方向（1为升序，-1为降序）
//
// 示例:
//   sort := Sort{
//       "age": 1,     // 按年龄升序
//       "name": -1,   // 按姓名降序
//   }
type Sort map[string]interface{}

// Pipeline 聚合管道类型
// 表示MongoDB聚合管道，由多个聚合阶段组成
// 每个阶段都是一个Document，包含具体的聚合操作
//
// 示例:
//   pipeline := Pipeline{
//       Document{"$match": Document{"status": "active"}},
//       Document{"$group": Document{"_id": "$department", "count": Document{"$sum": 1}}},
//       Document{"$sort": Document{"count": -1}},
//   }
type Pipeline []Document

// === 操作选项类型 ===

// FindOptions 查找操作选项
// 配置Find操作的各种参数，提供灵活的查询控制
type FindOptions struct {
	Sort       Document       `bson:"sort,omitempty"`       // 排序条件，键为字段名，值为排序方向
	Skip       *int64         `bson:"skip,omitempty"`       // 跳过的文档数量，用于分页
	Limit      *int64         `bson:"limit,omitempty"`      // 返回的最大文档数量，用于限制结果集大小
	Projection Document       `bson:"projection,omitempty"` // 投影字段，指定返回的字段（1包含，0排除）
	Timeout    *time.Duration `bson:"timeout,omitempty"`    // 操作超时时间，防止长时间阻塞
}

// FindOneOptions 查找单个文档操作选项
// 配置FindOne操作的各种参数，与FindOptions类似但不支持Limit
type FindOneOptions struct {
	Sort       Document       `bson:"sort,omitempty"`       // 排序条件，用于确定返回哪个文档
	Skip       *int64         `bson:"skip,omitempty"`       // 跳过的文档数量
	Projection Document       `bson:"projection,omitempty"` // 投影字段，指定返回的字段
	Timeout    *time.Duration `bson:"timeout,omitempty"`    // 操作超时时间
}

// CountOptions 计数操作选项
// 配置Count操作的各种参数，用于统计文档数量
type CountOptions struct {
	Limit   *int64         `bson:"limit,omitempty"`   // 计数的最大文档数量
	Skip    *int64         `bson:"skip,omitempty"`    // 跳过的文档数量
	MaxTime *time.Duration `bson:"maxTimeMS,omitempty"` // 最大执行时间
}

// InsertOneOptions 插入单个文档操作选项
// 配置InsertOne操作的参数
type InsertOneOptions struct {
	BypassDocumentValidation *bool `bson:"bypassDocumentValidation,omitempty"` // 是否跳过文档验证
}

// InsertManyOptions 插入多个文档操作选项
// 配置InsertMany操作的参数，支持批量插入控制
type InsertManyOptions struct {
	BypassDocumentValidation *bool `bson:"bypassDocumentValidation,omitempty"` // 是否跳过文档验证
	Ordered                  *bool `bson:"ordered,omitempty"`                  // 是否按顺序插入（遇到错误时是否停止）
}

// UpdateOptions 更新操作选项
// 配置Update操作的参数，支持upsert等高级功能
type UpdateOptions struct {
	BypassDocumentValidation *bool `bson:"bypassDocumentValidation,omitempty"` // 是否跳过文档验证
	Upsert                   *bool `bson:"upsert,omitempty"`                   // 如果文档不存在是否插入新文档
}

// ReplaceOptions 替换操作选项
// 配置ReplaceOne操作的参数，用于完整替换文档
type ReplaceOptions struct {
	BypassDocumentValidation *bool `bson:"bypassDocumentValidation,omitempty"` // 是否跳过文档验证
	Upsert                   *bool `bson:"upsert,omitempty"`                   // 如果文档不存在是否插入新文档
}

// DeleteOptions 删除操作选项
// 配置Delete操作的参数
type DeleteOptions struct {
	Comment *string `bson:"comment,omitempty"` // 操作注释，用于日志记录和调试
}

// AggregateOptions 聚合操作选项
// 配置Aggregate操作的参数，用于控制聚合管道的执行
type AggregateOptions struct {
	AllowDiskUse *bool          `bson:"allowDiskUse,omitempty"` // 是否允许使用磁盘存储临时数据
	BatchSize    *int32         `bson:"batchSize,omitempty"`    // 每批返回的文档数量
	MaxTime      *time.Duration `bson:"maxTimeMS,omitempty"`    // 最大执行时间
}

// === 原子操作选项 ===

// ReturnDocument 返回文档类型枚举
// 用于指定原子操作返回文档的时机（操作前或操作后）
type ReturnDocument int

const (
	// ReturnDocumentBefore 返回操作前的文档
	// 在FindOneAndUpdate等操作中，返回更新前的文档状态
	ReturnDocumentBefore ReturnDocument = iota
	
	// ReturnDocumentAfter 返回操作后的文档
	// 在FindOneAndUpdate等操作中，返回更新后的文档状态
	ReturnDocumentAfter
)

// FindOneAndUpdateOptions 查找并更新选项
// 配置FindOneAndUpdate原子操作的参数
// 这是一个原子操作，查找、更新、返回在一个操作中完成
type FindOneAndUpdateOptions struct {
	BypassDocumentValidation *bool             `bson:"bypassDocumentValidation,omitempty"` // 是否跳过文档验证
	Projection               Document          `bson:"projection,omitempty"`               // 投影字段
	ReturnDocument           *ReturnDocument   `bson:"returnDocument,omitempty"`           // 返回更新前还是更新后的文档
	Sort                     Document          `bson:"sort,omitempty"`                     // 排序条件
	Upsert                   *bool             `bson:"upsert,omitempty"`                   // 如果文档不存在是否插入
	MaxTime                  *time.Duration    `bson:"maxTimeMS,omitempty"`                // 最大执行时间
}

// FindOneAndDeleteOptions 查找并删除选项
// 配置FindOneAndDelete原子操作的参数
// 这是一个原子操作，查找、删除、返回在一个操作中完成
type FindOneAndDeleteOptions struct {
	Projection Document       `bson:"projection,omitempty"` // 投影字段
	Sort       Document       `bson:"sort,omitempty"`       // 排序条件，用于确定删除哪个文档
	MaxTime    *time.Duration `bson:"maxTimeMS,omitempty"`  // 最大执行时间
}

// FindOneAndReplaceOptions 查找并替换选项
// 配置FindOneAndReplace原子操作的参数
// 这是一个原子操作，查找、替换、返回在一个操作中完成
type FindOneAndReplaceOptions struct {
	BypassDocumentValidation *bool             `bson:"bypassDocumentValidation,omitempty"` // 是否跳过文档验证
	Projection               Document          `bson:"projection,omitempty"`               // 投影字段
	ReturnDocument           *ReturnDocument   `bson:"returnDocument,omitempty"`           // 返回替换前还是替换后的文档
	Sort                     Document          `bson:"sort,omitempty"`                     // 排序条件
	Upsert                   *bool             `bson:"upsert,omitempty"`                   // 如果文档不存在是否插入
	MaxTime                  *time.Duration    `bson:"maxTimeMS,omitempty"`                // 最大执行时间
}

// === 操作结果类型 ===

// InsertOneResult 插入单个文档的结果
// 包含插入操作的结果信息
type InsertOneResult struct {
	InsertedID interface{} `bson:"insertedId"` // 插入文档的ID，通常是MongoDB自动生成的ObjectID
}

// InsertManyResult 插入多个文档的结果
// 包含批量插入操作的结果信息
type InsertManyResult struct {
	InsertedIDs []interface{} `bson:"insertedIds"` // 所有插入文档的ID列表，按插入顺序排列
}

// UpdateResult 更新操作的结果
// 包含更新操作的统计信息
type UpdateResult struct {
	MatchedCount  int64       `bson:"matchedCount"`  // 匹配过滤条件的文档数量
	ModifiedCount int64       `bson:"modifiedCount"` // 实际被修改的文档数量
	UpsertedCount int64       `bson:"upsertedCount"` // 通过upsert插入的文档数量
	UpsertedID    interface{} `bson:"upsertedId"`    // upsert操作插入文档的ID
}

// DeleteResult 删除操作的结果
// 包含删除操作的统计信息
type DeleteResult struct {
	DeletedCount int64 `bson:"deletedCount"` // 被删除的文档数量
}

// === 索引相关类型 ===

// IndexModel 索引模型
// 定义索引的结构和选项，用于创建数据库索引
type IndexModel struct {
	Keys    Document     `bson:"keys"`             // 索引键定义，键为字段名，值为索引类型（1升序，-1降序）
	Options *IndexOptions `bson:"options,omitempty"` // 索引选项
}

// IndexOptions 索引选项
// 定义索引的各种配置参数
type IndexOptions struct {
	Background         *bool  `bson:"background,omitempty"`         // 是否在后台创建索引
	ExpireAfterSeconds *int32 `bson:"expireAfterSeconds,omitempty"` // TTL索引的过期时间（秒）
	Name               *string `bson:"name,omitempty"`               // 索引名称
	Sparse             *bool  `bson:"sparse,omitempty"`             // 是否为稀疏索引（跳过缺失字段的文档）
	Unique             *bool  `bson:"unique,omitempty"`             // 是否为唯一索引
}

// === 会话和事务相关类型 ===

// SessionOptions 会话选项
// 配置MongoDB会话的参数，会话用于事务和因果一致性
type SessionOptions struct {
	// 简化的会话选项，可以根据需要扩展
	// 因果一致性确保在会话中的操作具有因果关系
	CausalConsistency *bool `bson:"causalConsistency,omitempty"` // 是否启用因果一致性
}

// TransactionOptions 事务选项
// 配置MongoDB事务的参数
type TransactionOptions struct {
	// 简化的事务选项，可以根据需要扩展
	// 事务允许多个操作作为一个原子单元执行
	MaxCommitTimeMS *time.Duration `bson:"maxCommitTimeMS,omitempty"` // 事务提交的最大时间
}

// === 其他配置相关类型 ===

// CommandOptions 命令选项
// 配置数据库命令的执行参数
type CommandOptions struct {
	// 简化的命令选项，可以根据需要扩展
	// 用于执行原生MongoDB命令
	MaxTime *time.Duration `bson:"maxTimeMS,omitempty"` // 命令执行的最大时间
}

// ChangeStreamOptions 变更流选项
// 配置变更流的参数，变更流用于监听数据库变更
type ChangeStreamOptions struct {
	// 简化的变更流选项，可以根据需要扩展
	// 变更流允许应用程序实时监听数据库的变更
	MaxAwaitTime *time.Duration `bson:"maxAwaitTime,omitempty"` // 等待变更的最大时间
}

// GridFSBucketOptions GridFS桶选项
// 配置GridFS文件存储桶的参数
type GridFSBucketOptions struct {
	// 简化的GridFS选项，可以根据需要扩展
	// GridFS用于存储大型文件，将文件分块存储
	BucketName *string `bson:"bucketName,omitempty"` // 存储桶名称
	ChunkSizeBytes *int32 `bson:"chunkSizeBytes,omitempty"` // 文件分块大小（字节）
	WriteConcern interface{} `bson:"writeConcern,omitempty"` // 写关注级别
	ReadConcern interface{} `bson:"readConcern,omitempty"` // 读关注级别
	ReadPreference interface{} `bson:"readPreference,omitempty"` // 读偏好
}

// GridFSUploadOptions GridFS上传选项
// 配置GridFS文件上传的参数
type GridFSUploadOptions struct {
	ChunkSizeBytes *int32 `bson:"chunkSizeBytes,omitempty"` // 文件分块大小
	Metadata Document `bson:"metadata,omitempty"` // 文件元数据
}

// GridFSDownloadOptions GridFS下载选项
// 配置GridFS文件下载的参数
type GridFSDownloadOptions struct {
	Revision *int32 `bson:"revision,omitempty"` // 文件版本号（-1表示最新版本）
}

// GridFSFindOptions GridFS查找选项
// 配置GridFS文件查找的参数
type GridFSFindOptions struct {
	BatchSize *int32 `bson:"batchSize,omitempty"` // 批次大小
	Limit *int32 `bson:"limit,omitempty"` // 限制返回数量
	MaxTime *time.Duration `bson:"maxTime,omitempty"` // 最大执行时间
	NoCursorTimeout *bool `bson:"noCursorTimeout,omitempty"` // 游标是否超时
	Skip *int32 `bson:"skip,omitempty"` // 跳过的文档数量
	Sort Document `bson:"sort,omitempty"` // 排序条件
}

// GridFSFile GridFS文件信息
// 表示GridFS中存储的文件信息
type GridFSFile struct {
	ID interface{} `bson:"_id"` // 文件ID
	Length int64 `bson:"length"` // 文件大小（字节）
	ChunkSize int32 `bson:"chunkSize"` // 分块大小
	UploadDate time.Time `bson:"uploadDate"` // 上传时间
	Filename string `bson:"filename"` // 文件名
	Metadata Document `bson:"metadata,omitempty"` // 文件元数据
	MD5 string `bson:"md5,omitempty"` // MD5校验值
}

// === 核心接口定义 ===

// MongoClient MongoDB客户端接口
// 定义客户端级别的操作，包括连接管理和数据库访问
type MongoClient interface {
	// Connect 连接到MongoDB服务器
	// config: 连接配置，可以是配置结构体或连接字符串
	Connect(ctx context.Context, config interface{}) error
	
	// Disconnect 断开MongoDB连接
	// 关闭所有连接并释放资源
	Disconnect(ctx context.Context) error
	
	// Ping 测试连接状态
	// 发送ping命令验证连接是否正常
	Ping(ctx context.Context) error
	
	// Database 获取数据库实例
	// name: 数据库名称
	// 返回指定数据库的操作接口
	Database(name string) MongoDatabase
	
	// DefaultDatabase 获取配置文件中指定的默认数据库实例
	// 返回配置中指定的默认数据库的操作接口和可能的错误
	DefaultDatabase() (MongoDatabase, error)
	
	// ListDatabaseNames 列出所有数据库名称
	// filter: 过滤条件
	// 返回数据库名称列表
	ListDatabaseNames(ctx context.Context, filter Document) ([]string, error)
	
	// StartSession 开始新的会话
	// opts: 会话选项
	// 返回会话接口，用于事务和因果一致性
	StartSession(opts *SessionOptions) (MongoSession, error)
	
	// Watch 监视客户端级别的变更
	// pipeline: 聚合管道，用于过滤变更事件
	// opts: 变更流选项
	// 返回变更流接口
	Watch(ctx context.Context, pipeline Pipeline, opts *ChangeStreamOptions) (ChangeStream, error)
	
	// NumberSessionsInProgress 获取进行中的会话数量
	// 返回当前活跃的会话数量
	NumberSessionsInProgress() int
}

// MongoDatabase MongoDB数据库接口
// 定义数据库级别的操作，包括集合管理和数据库命令
type MongoDatabase interface {
	// Name 获取数据库名称
	Name() string
	
	// Collection 获取集合实例
	// name: 集合名称
	// 返回指定集合的操作接口
	Collection(name string) MongoCollection
	
	// ListCollectionNames 列出所有集合名称
	// filter: 过滤条件
	// 返回集合名称列表
	ListCollectionNames(ctx context.Context, filter Document) ([]string, error)
	
	// CreateCollection 创建集合
	// name: 集合名称
	CreateCollection(ctx context.Context, name string) error
	
	// DropCollection 删除集合
	// name: 集合名称
	DropCollection(ctx context.Context, name string) error
	
	// Drop 删除整个数据库
	// 删除数据库及其所有集合和数据
	Drop(ctx context.Context) error
	
	// GridFSBucket 获取默认GridFS存储桶
	// 返回GridFS存储桶接口，用于文件存储
	GridFSBucket() GridFSBucket
	
	// GridFSBucketWithOptions 获取带选项的GridFS存储桶
	// opts: GridFS桶选项
	// 返回自定义配置的GridFS存储桶
	GridFSBucketWithOptions(opts *GridFSBucketOptions) GridFSBucket
	
	// RunCommand 执行数据库命令
	// command: 要执行的命令
	// opts: 命令选项
	// 返回命令执行结果
	RunCommand(ctx context.Context, command Document, opts *CommandOptions) SingleResult
	
	// Watch 监视数据库级别的变更
	// pipeline: 聚合管道
	// opts: 变更流选项
	// 返回变更流接口
	Watch(ctx context.Context, pipeline Pipeline, opts *ChangeStreamOptions) (ChangeStream, error)
}

// MongoCollection MongoDB集合接口
// 定义集合级别的操作，包括CRUD操作、索引管理等
type MongoCollection interface {
	// === 基本信息 ===
	
	// Name 获取集合名称
	Name() string
	
	// Database 获取所属的数据库实例
	Database() MongoDatabase
	
	// === 查询操作 ===
	
	// Find 查找多个文档
	// filter: 查询过滤条件
	// opts: 查询选项
	// 返回游标，用于遍历查询结果
	Find(ctx context.Context, filter Filter, opts *FindOptions) (MongoCursor, error)
	
	// FindOne 查找单个文档
	// filter: 查询过滤条件
	// opts: 查询选项
	// 返回单个结果
	FindOne(ctx context.Context, filter Filter, opts *FindOneOptions) SingleResult
	
	// Count 计算文档数量
	// filter: 过滤条件
	// opts: 计数选项
	// 返回匹配的文档数量
	Count(ctx context.Context, filter Filter, opts *CountOptions) (int64, error)
	
	// === 插入操作 ===
	
	// InsertOne 插入单个文档
	// document: 要插入的文档
	// opts: 插入选项
	// 返回插入结果
	InsertOne(ctx context.Context, document Document, opts *InsertOneOptions) (*InsertOneResult, error)
	
	// InsertMany 插入多个文档
	// documents: 要插入的文档列表
	// opts: 插入选项
	// 返回批量插入结果
	InsertMany(ctx context.Context, documents []Document, opts *InsertManyOptions) (*InsertManyResult, error)
	
	// === 更新操作 ===
	
	// UpdateOne 更新单个文档
	// filter: 过滤条件
	// update: 更新操作
	// opts: 更新选项
	// 返回更新结果
	UpdateOne(ctx context.Context, filter Filter, update Update, opts *UpdateOptions) (*UpdateResult, error)
	
	// UpdateMany 更新多个文档
	// filter: 过滤条件
	// update: 更新操作
	// opts: 更新选项
	// 返回更新结果
	UpdateMany(ctx context.Context, filter Filter, update Update, opts *UpdateOptions) (*UpdateResult, error)
	
	// ReplaceOne 替换单个文档
	// filter: 过滤条件
	// replacement: 替换的文档
	// opts: 替换选项
	// 返回替换结果
	ReplaceOne(ctx context.Context, filter Filter, replacement Document, opts *ReplaceOptions) (*UpdateResult, error)
	
	// === 原子操作 ===
	
	// FindOneAndUpdate 查找并更新单个文档（原子操作）
	// filter: 过滤条件
	// update: 更新操作
	// opts: 查找并更新选项
	// 返回更新前或更新后的文档
	FindOneAndUpdate(ctx context.Context, filter Filter, update Update, opts *FindOneAndUpdateOptions) SingleResult
	
	// FindOneAndDelete 查找并删除单个文档（原子操作）
	// filter: 过滤条件
	// opts: 查找并删除选项
	// 返回被删除的文档
	FindOneAndDelete(ctx context.Context, filter Filter, opts *FindOneAndDeleteOptions) SingleResult
	
	// FindOneAndReplace 查找并替换单个文档（原子操作）
	// filter: 过滤条件
	// replacement: 替换的文档
	// opts: 查找并替换选项
	// 返回替换前或替换后的文档
	FindOneAndReplace(ctx context.Context, filter Filter, replacement Document, opts *FindOneAndReplaceOptions) SingleResult
	
	// === 删除操作 ===
	
	// DeleteOne 删除单个文档
	// filter: 过滤条件
	// opts: 删除选项
	// 返回删除结果
	DeleteOne(ctx context.Context, filter Filter, opts *DeleteOptions) (*DeleteResult, error)
	
	// DeleteMany 删除多个文档
	// filter: 过滤条件
	// opts: 删除选项
	// 返回删除结果
	DeleteMany(ctx context.Context, filter Filter, opts *DeleteOptions) (*DeleteResult, error)
	
	// === 聚合操作 ===
	
	// Aggregate 执行聚合查询
	// pipeline: 聚合管道
	// opts: 聚合选项
	// 返回聚合结果游标
	Aggregate(ctx context.Context, pipeline Pipeline, opts *AggregateOptions) (MongoCursor, error)
	
	// === 索引操作 ===
	
	// CreateIndex 创建索引
	// model: 索引模型
	// 返回创建的索引名称
	CreateIndex(ctx context.Context, model IndexModel) (string, error)
	
	// ListIndexes 列出所有索引
	// 返回索引信息游标
	ListIndexes(ctx context.Context) (MongoCursor, error)
	
	// DropIndex 删除指定索引
	// name: 索引名称
	DropIndex(ctx context.Context, name string) error
	
	// === 集合管理 ===
	
	// Drop 删除集合
	// 删除整个集合及其所有数据和索引
	Drop(ctx context.Context) error
	
	// Watch 监视集合级别的变更
	// pipeline: 聚合管道
	// opts: 变更流选项
	// 返回变更流接口
	Watch(ctx context.Context, pipeline Pipeline, opts *ChangeStreamOptions) (ChangeStream, error)
}

// MongoCursor MongoDB游标接口
// 定义游标操作，用于遍历查询结果
type MongoCursor interface {
	// Next 移动到下一个文档
	// 返回是否有下一个文档可用
	Next(ctx context.Context) bool
	
	// Decode 解码当前文档
	// val: 解码目标，通常是结构体指针
	// 将当前文档解码到指定的结构体中
	Decode(val interface{}) error
	
	// All 获取所有剩余文档
	// results: 结果切片，通常是结构体切片的指针
	// 将所有剩余文档解码到指定的切片中
	All(ctx context.Context, results interface{}) error
	
	// Close 关闭游标
	// 释放游标占用的资源，必须调用以避免资源泄漏
	Close(ctx context.Context) error
	
	// Err 获取游标错误
	// 返回游标操作中遇到的错误
	Err() error
}

// SingleResult 单个结果接口
// 定义单个文档查询结果的操作
type SingleResult interface {
	// Decode 解码结果文档
	// val: 解码目标，通常是结构体指针
	// 将结果文档解码到指定的结构体中
	Decode(val interface{}) error
	
	// Err 获取查询错误
	// 返回查询操作中遇到的错误，如果没有找到文档会返回ErrNoDocuments
	Err() error
}

// MongoSession MongoDB会话接口
// 定义会话操作，用于事务和因果一致性
type MongoSession interface {
	// 简化的会话接口，可以根据需要扩展
	// 会话提供了事务支持和因果一致性保证
	
	// EndSession 结束会话
	// 释放会话资源，结束会话生命周期
	EndSession(ctx context.Context)
}

// ChangeStream 变更流接口
// 定义变更流操作，用于监听数据库变更
type ChangeStream interface {
	// 简化的变更流接口，可以根据需要扩展
	// 变更流允许应用程序实时监听数据库的变更事件
	
	// Next 移动到下一个变更事件
	// 返回是否有下一个变更事件可用
	Next(ctx context.Context) bool
	
	// Close 关闭变更流
	// 释放变更流资源，停止监听变更事件
	Close(ctx context.Context) error
}

// GridFSBucket GridFS桶接口
// 定义GridFS文件存储操作，支持大文件的分块存储和管理
type GridFSBucket interface {
	// === 文件上传操作 ===
	
	// UploadFromStream 从流上传文件
	// filename: 文件名
	// source: 数据源，通常是io.Reader
	// opts: 上传选项
	// 返回文件ID和错误
	UploadFromStream(ctx context.Context, filename string, source interface{}, opts *GridFSUploadOptions) (interface{}, error)
	
	// UploadFromFile 从本地文件上传
	// localFilePath: 本地文件路径
	// filename: 存储在GridFS中的文件名（如果为空则使用本地文件名）
	// opts: 上传选项
	// 返回文件ID和错误
	UploadFromFile(ctx context.Context, localFilePath string, filename string, opts *GridFSUploadOptions) (interface{}, error)
	
	// UploadFromBytes 从字节数组上传文件
	// filename: 文件名
	// data: 文件数据
	// opts: 上传选项
	// 返回文件ID和错误
	UploadFromBytes(ctx context.Context, filename string, data []byte, opts *GridFSUploadOptions) (interface{}, error)
	
	// === 文件下载操作 ===
	
	// DownloadToStream 下载文件到流
	// fileID: 文件ID
	// destination: 目标流，通常是io.Writer
	// opts: 下载选项
	// 返回错误
	DownloadToStream(ctx context.Context, fileID interface{}, destination interface{}, opts *GridFSDownloadOptions) error
	
	// DownloadToStreamByName 按文件名下载文件到流
	// filename: 文件名
	// destination: 目标流，通常是io.Writer
	// opts: 下载选项
	// 返回错误
	DownloadToStreamByName(ctx context.Context, filename string, destination interface{}, opts *GridFSDownloadOptions) error
	
	// DownloadToFile 下载文件到本地文件
	// fileID: 文件ID
	// localFilePath: 本地文件路径
	// opts: 下载选项
	// 返回错误
	DownloadToFile(ctx context.Context, fileID interface{}, localFilePath string, opts *GridFSDownloadOptions) error
	
	// DownloadToFileByName 按文件名下载文件到本地文件
	// filename: 文件名
	// localFilePath: 本地文件路径
	// opts: 下载选项
	// 返回错误
	DownloadToFileByName(ctx context.Context, filename string, localFilePath string, opts *GridFSDownloadOptions) error
	
	// DownloadToBytes 下载文件到字节数组
	// fileID: 文件ID
	// opts: 下载选项
	// 返回文件数据和错误
	DownloadToBytes(ctx context.Context, fileID interface{}, opts *GridFSDownloadOptions) ([]byte, error)
	
	// DownloadToBytesByName 按文件名下载文件到字节数组
	// filename: 文件名
	// opts: 下载选项
	// 返回文件数据和错误
	DownloadToBytesByName(ctx context.Context, filename string, opts *GridFSDownloadOptions) ([]byte, error)
	
	// === 文件删除操作 ===
	
	// Delete 删除文件
	// fileID: 文件ID
	// 返回错误
	Delete(ctx context.Context, fileID interface{}) error
	
	// DeleteByName 按文件名删除文件
	// filename: 文件名
	// 返回错误
	DeleteByName(ctx context.Context, filename string) error
	
	// === 文件查询操作 ===
	
	// Find 查找文件
	// filter: 查询过滤条件
	// opts: 查找选项
	// 返回文件游标
	Find(ctx context.Context, filter Filter, opts *GridFSFindOptions) (MongoCursor, error)
	
	// FindOne 查找单个文件
	// filter: 查询过滤条件
	// opts: 查找选项
	// 返回文件信息
	FindOne(ctx context.Context, filter Filter, opts *GridFSFindOptions) (*GridFSFile, error)
	
	// GetFileInfo 获取文件信息
	// fileID: 文件ID
	// 返回文件信息和错误
	GetFileInfo(ctx context.Context, fileID interface{}) (*GridFSFile, error)
	
	// GetFileInfoByName 按文件名获取文件信息
	// filename: 文件名
	// 返回文件信息和错误
	GetFileInfoByName(ctx context.Context, filename string) (*GridFSFile, error)
	
	// === 文件管理操作 ===
	
	// ListFiles 列出所有文件
	// filter: 过滤条件
	// opts: 查找选项
	// 返回文件列表
	ListFiles(ctx context.Context, filter Filter, opts *GridFSFindOptions) ([]*GridFSFile, error)
	
	// GetFileCount 获取文件数量
	// filter: 过滤条件
	// 返回文件数量和错误
	GetFileCount(ctx context.Context, filter Filter) (int64, error)
	
	// Rename 重命名文件
	// fileID: 文件ID
	// newFilename: 新文件名
	// 返回错误
	Rename(ctx context.Context, fileID interface{}, newFilename string) error
	
	// === 批量操作 ===
	
	// BatchUpload 批量上传文件
	// files: 文件路径映射（本地路径 -> GridFS文件名）
	// opts: 上传选项
	// 返回成功上传的文件ID列表和错误
	BatchUpload(ctx context.Context, files map[string]string, opts *GridFSUploadOptions) ([]interface{}, error)
	
	// BatchDownload 批量下载文件
	// files: 文件映射（文件ID -> 本地路径）
	// opts: 下载选项
	// 返回成功下载的文件列表和错误
	BatchDownload(ctx context.Context, files map[interface{}]string, opts *GridFSDownloadOptions) ([]string, error)
	
	// BatchDelete 批量删除文件
	// fileIDs: 文件ID列表
	// 返回删除的文件数量和错误
	BatchDelete(ctx context.Context, fileIDs []interface{}) (int64, error)
	
	// === 桶管理操作 ===
	
	// Drop 删除整个GridFS桶
	// 删除桶中的所有文件和分块数据
	Drop(ctx context.Context) error
	
	// GetBucketName 获取桶名称
	// 返回桶名称
	GetBucketName() string
} 
