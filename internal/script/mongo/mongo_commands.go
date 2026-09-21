package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoIndexCommand MongoDB 索引创建命令
type MongoIndexCommand struct {
	// CollectionName 集合名称
	CollectionName string
	// IndexModel 索引模型
	IndexModel IndexModel
	// Description 索引描述
	Description string
}

// IndexModel 索引模型
type IndexModel struct {
	// Keys 索引键
	Keys bson.D
	// Options 索引选项
	Options *IndexOptions
}

// IndexOptions 索引选项
type IndexOptions struct {
	// Name 索引名称
	Name string
	// Unique 是否唯一索引
	Unique bool
	// Background 是否后台创建
	Background bool
	// Sparse 是否稀疏索引
	Sparse bool
	// ExpireAfterSeconds TTL过期时间（秒）
	ExpireAfterSeconds *int32
}

// ToMongoIndexModel 转换为 MongoDB 驱动的 IndexModel
func (im IndexModel) ToMongoIndexModel() mongo.IndexModel {
	opts := options.Index()

	if im.Options != nil {
		if im.Options.Name != "" {
			opts.SetName(im.Options.Name)
		}
		if im.Options.Unique {
			opts.SetUnique(true)
		}
		if im.Options.Background {
			opts.SetBackground(true)
		}
		if im.Options.Sparse {
			opts.SetSparse(true)
		}
		if im.Options.ExpireAfterSeconds != nil {
			opts.SetExpireAfterSeconds(*im.Options.ExpireAfterSeconds)
		}
	}

	return mongo.IndexModel{
		Keys:    im.Keys,
		Options: opts,
	}
}

// managedMongoLogCollections 仅对这些网关日志集合做索引对齐，禁止扫库或动其它业务集合。
var managedMongoLogCollections = map[string]struct{}{
	"HUB_GW_ACCESS_LOG":        {},
	"HUB_GW_BACKEND_TRACE_LOG": {},
}

// IsManagedMongoLogCollection 是否为本系统日志集合。其它集合一律不建、不删索引。
func IsManagedMongoLogCollection(name string) bool {
	_, ok := managedMongoLogCollections[name]
	return ok
}

// GetMongoInitCommands 返回启动时创建的最小索引集，与 scripts/db/mongo/mongo.js 对齐。
// 只覆盖详情、按实例列表/监控、TTL；不为每个筛选项各建一条，以降低百万级写入维护成本。
func GetMongoInitCommands() []MongoIndexCommand {
	ttl30Days := int32(2592000) // 30天

	return []MongoIndexCommand{
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "tenantId", Value: 1},
					{Key: "traceId", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_tenant_trace",
					Background: true,
				},
			},
			Description: "详情 GetGatewayLogByKey，非唯一以免历史重复键建失败",
		},
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayInstanceId", Value: 1},
					{Key: "gatewayStartProcessingTime", Value: -1},
				},
				Options: &IndexOptions{
					Name:       "idx_instance_time",
					Background: true,
				},
			},
			Description: "按实例列表、监控与应用层清理",
		},
		{
			CollectionName: "HUB_GW_ACCESS_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "gatewayStartProcessingTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:               "idx_ttl_cleanup",
					Background:         true,
					ExpireAfterSeconds: &ttl30Days,
				},
			},
			Description: "TTL 30天，必须单字段 Date",
		},
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "tenantId", Value: 1},
					{Key: "traceId", Value: 1},
					{Key: "backendTraceId", Value: 1},
				},
				Options: &IndexOptions{
					Name:       "idx_tenant_trace",
					Background: true,
				},
			},
			Description: "按 trace 拉列表（左前缀）或按 tenantId+traceId+backendTraceId 取单条，非唯一以免历史重复键建失败",
		},
		{
			CollectionName: "HUB_GW_BACKEND_TRACE_LOG",
			IndexModel: IndexModel{
				Keys: bson.D{
					{Key: "requestStartTime", Value: 1},
				},
				Options: &IndexOptions{
					Name:               "idx_ttl_cleanup",
					Background:         true,
					ExpireAfterSeconds: &ttl30Days,
				},
			},
			Description: "TTL 30天，兼时间扫描",
		},
	}
}

// GetIndexCommandsByCollection 按集合分组获取索引命令
func GetIndexCommandsByCollection() map[string][]MongoIndexCommand {
	commands := GetMongoInitCommands()
	result := make(map[string][]MongoIndexCommand)

	for _, cmd := range commands {
		result[cmd.CollectionName] = append(result[cmd.CollectionName], cmd)
	}

	return result
}

// GetCollectionNames 获取所有需要创建索引的集合名称
func GetCollectionNames() []string {
	commandsByCollection := GetIndexCommandsByCollection()
	collections := make([]string, 0, len(commandsByCollection))

	for collName := range commandsByCollection {
		collections = append(collections, collName)
	}

	return collections
}
