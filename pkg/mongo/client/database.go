// Package client 提供MongoDB数据库实现
//
// 此文件包含Database结构体和相关方法的实现
package client

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"gateway/pkg/mongo/errors"
	"gateway/pkg/mongo/types"
)

// Database MongoDB数据库实现
// 实现types.MongoDatabase接口，提供数据库级别的操作
type Database struct {
	db     *mongo.Database // MongoDB驱动数据库
	client *Client         // 父客户端引用
	name   string          // 数据库名称
}

// Name 获取数据库名称
func (d *Database) Name() string {
	return d.name
}

// Collection 获取集合实例
// 返回指定名称的集合操作接口
func (d *Database) Collection(name string) types.MongoCollection {
	return &Collection{
		coll:     d.db.Collection(name),
		database: d,
		name:     name,
	}
}

// ListCollectionNames 列出所有集合名称
// 根据过滤条件返回集合名称列表
func (d *Database) ListCollectionNames(ctx context.Context, filter types.Document) ([]string, error) {
	names, err := d.db.ListCollectionNames(ctx, filter)
	if err != nil {
		return nil, errors.NewQueryError("failed to list collection names", err)
	}
	return names, nil
}

// CreateCollection 创建集合
// 在数据库中创建新的集合
func (d *Database) CreateCollection(ctx context.Context, name string) error {
	if err := d.db.CreateCollection(ctx, name); err != nil {
		return errors.NewQueryError("failed to create collection", err)
	}
	return nil
}

// DropCollection 删除集合
// 从数据库中删除指定的集合
func (d *Database) DropCollection(ctx context.Context, name string) error {
	if err := d.db.Collection(name).Drop(ctx); err != nil {
		return errors.NewQueryError("failed to drop collection", err)
	}
	return nil
}

// Drop 删除整个数据库
// 删除数据库及其所有集合和数据
func (d *Database) Drop(ctx context.Context) error {
	if err := d.db.Drop(ctx); err != nil {
		return errors.NewQueryError("failed to drop database", err)
	}
	return nil
}

// GridFSBucket 获取GridFS存储桶
// 返回默认的GridFS存储桶实例，用于文件存储操作
func (d *Database) GridFSBucket() types.GridFSBucket {
	// 使用默认选项创建GridFS桶
	return NewGridFSBucket(d, nil)
}

// GridFSBucketWithOptions 获取带选项的GridFS存储桶
// 返回自定义配置的GridFS存储桶实例
func (d *Database) GridFSBucketWithOptions(opts *types.GridFSBucketOptions) types.GridFSBucket {
	// 使用指定选项创建GridFS桶
	return NewGridFSBucket(d, opts)
}

// RunCommand 执行数据库命令
// 执行原生的MongoDB数据库命令
func (d *Database) RunCommand(ctx context.Context, command types.Document, opts *types.CommandOptions) types.SingleResult {
	// 执行数据库命令
	result := d.db.RunCommand(ctx, command)
	return &Result{result: result}
}

// Watch 监视数据库变更
// 创建变更流以监视整个数据库中的文档变更
func (d *Database) Watch(ctx context.Context, pipeline types.Pipeline, opts *types.ChangeStreamOptions) (types.ChangeStream, error) {
	// 目前返回一个简单的错误，因为实际实现需要完整的变更流支持
	return nil, errors.NewQueryError("ChangeStream not implemented yet", nil)
}
