// Package client 提供MongoDB集合实现
//
// 此文件包含Collection结构体和基本CRUD方法的实现
package client

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gohub/pkg/mongo/errors"
	"gohub/pkg/mongo/types"
)

// Collection MongoDB集合实现
// 实现types.MongoCollection接口，提供集合级别的CRUD操作
type Collection struct {
	coll     *mongo.Collection  // MongoDB驱动集合
	database *Database         // 父数据库引用
	name     string            // 集合名称
}

// Name 获取集合名称
func (c *Collection) Name() string {
	return c.name
}

// Database 获取所属的数据库实例
func (c *Collection) Database() types.MongoDatabase {
	return c.database
}

// === 查询操作 ===

// Find 查找多个文档
// 根据过滤条件查询文档，返回游标用于遍历结果
func (c *Collection) Find(ctx context.Context, filter types.Filter, opts *types.FindOptions) (types.MongoCursor, error) {
	// 构建查询选项
	findOptions := options.Find()

	if opts != nil {
		if opts.Sort != nil {
			findOptions.SetSort(opts.Sort)
		}
		if opts.Skip != nil {
			findOptions.SetSkip(*opts.Skip)
		}
		if opts.Limit != nil {
			findOptions.SetLimit(*opts.Limit)
		}
		if opts.Projection != nil {
			findOptions.SetProjection(opts.Projection)
		}
	}

	// 执行查询
	cursor, err := c.coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, errors.NewQueryError("failed to execute find query", err)
	}

	return &Cursor{cursor: cursor}, nil
}

// FindOne 查找单个文档
// 根据过滤条件查询单个文档，返回单个结果
func (c *Collection) FindOne(ctx context.Context, filter types.Filter, opts *types.FindOneOptions) types.SingleResult {
	// 构建查询选项
	findOneOptions := options.FindOne()

	if opts != nil {
		if opts.Sort != nil {
			findOneOptions.SetSort(opts.Sort)
		}
		if opts.Skip != nil {
			findOneOptions.SetSkip(*opts.Skip)
		}
		if opts.Projection != nil {
			findOneOptions.SetProjection(opts.Projection)
		}
	}

	// 执行查询
	result := c.coll.FindOne(ctx, filter, findOneOptions)
	return &Result{result: result}
}

// Count 计算文档数量
// 根据过滤条件统计匹配的文档数量
func (c *Collection) Count(ctx context.Context, filter types.Filter, opts *types.CountOptions) (int64, error) {
	// 构建计数选项
	countOptions := options.Count()

	if opts != nil {
		if opts.Limit != nil {
			countOptions.SetLimit(*opts.Limit)
		}
		if opts.Skip != nil {
			countOptions.SetSkip(*opts.Skip)
		}
		if opts.MaxTime != nil {
			countOptions.SetMaxTime(*opts.MaxTime)
		}
	}

	// 执行计数
	count, err := c.coll.CountDocuments(ctx, filter, countOptions)
	if err != nil {
		return 0, errors.NewQueryError("failed to count documents", err)
	}

	return count, nil
}

// === 写入操作 ===

// InsertOne 插入单个文档
// 向集合中插入一个文档
func (c *Collection) InsertOne(ctx context.Context, document types.Document, opts *types.InsertOneOptions) (*types.InsertOneResult, error) {
	// 构建插入选项
	insertOptions := options.InsertOne()

	if opts != nil && opts.BypassDocumentValidation != nil {
		insertOptions.SetBypassDocumentValidation(*opts.BypassDocumentValidation)
	}

	// 执行插入
	result, err := c.coll.InsertOne(ctx, document, insertOptions)
	if err != nil {
		return nil, errors.NewInsertError("failed to insert document", err)
	}

	return &types.InsertOneResult{
		InsertedID: result.InsertedID,
	}, nil
}

// InsertMany 插入多个文档
// 向集合中批量插入多个文档
func (c *Collection) InsertMany(ctx context.Context, documents []types.Document, opts *types.InsertManyOptions) (*types.InsertManyResult, error) {
	// 转换文档类型
	docs := make([]interface{}, len(documents))
	for i, doc := range documents {
		docs[i] = doc
	}

	// 构建插入选项
	insertOptions := options.InsertMany()

	if opts != nil {
		if opts.BypassDocumentValidation != nil {
			insertOptions.SetBypassDocumentValidation(*opts.BypassDocumentValidation)
		}
		if opts.Ordered != nil {
			insertOptions.SetOrdered(*opts.Ordered)
		}
	}

	// 执行批量插入
	result, err := c.coll.InsertMany(ctx, docs, insertOptions)
	if err != nil {
		return nil, errors.NewInsertError("failed to insert documents", err)
	}

	return &types.InsertManyResult{
		InsertedIDs: result.InsertedIDs,
	}, nil
}

// UpdateOne 更新单个文档
// 根据过滤条件更新第一个匹配的文档
func (c *Collection) UpdateOne(ctx context.Context, filter types.Filter, update types.Update, opts *types.UpdateOptions) (*types.UpdateResult, error) {
	// 构建更新选项
	updateOptions := options.Update()

	if opts != nil {
		if opts.BypassDocumentValidation != nil {
			updateOptions.SetBypassDocumentValidation(*opts.BypassDocumentValidation)
		}
		if opts.Upsert != nil {
			updateOptions.SetUpsert(*opts.Upsert)
		}
	}

	// 执行更新
	result, err := c.coll.UpdateOne(ctx, filter, update, updateOptions)
	if err != nil {
		return nil, errors.NewUpdateError("failed to update document", err)
	}

	return &types.UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// UpdateMany 更新多个文档
// 根据过滤条件更新所有匹配的文档
func (c *Collection) UpdateMany(ctx context.Context, filter types.Filter, update types.Update, opts *types.UpdateOptions) (*types.UpdateResult, error) {
	// 构建更新选项
	updateOptions := options.Update()

	if opts != nil {
		if opts.BypassDocumentValidation != nil {
			updateOptions.SetBypassDocumentValidation(*opts.BypassDocumentValidation)
		}
		if opts.Upsert != nil {
			updateOptions.SetUpsert(*opts.Upsert)
		}
	}

	// 执行批量更新
	result, err := c.coll.UpdateMany(ctx, filter, update, updateOptions)
	if err != nil {
		return nil, errors.NewUpdateError("failed to update documents", err)
	}

	return &types.UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// DeleteOne 删除单个文档
// 根据过滤条件删除第一个匹配的文档
func (c *Collection) DeleteOne(ctx context.Context, filter types.Filter, opts *types.DeleteOptions) (*types.DeleteResult, error) {
	// 构建删除选项
	deleteOptions := options.Delete()

	// 执行删除
	result, err := c.coll.DeleteOne(ctx, filter, deleteOptions)
	if err != nil {
		return nil, errors.NewDeleteError("failed to delete document", err)
	}

	return &types.DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// DeleteMany 删除多个文档
// 根据过滤条件删除所有匹配的文档
func (c *Collection) DeleteMany(ctx context.Context, filter types.Filter, opts *types.DeleteOptions) (*types.DeleteResult, error) {
	// 构建删除选项
	deleteOptions := options.Delete()

	// 执行批量删除
	result, err := c.coll.DeleteMany(ctx, filter, deleteOptions)
	if err != nil {
		return nil, errors.NewDeleteError("failed to delete documents", err)
	}

	return &types.DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// Drop 删除集合
// 删除整个集合及其所有数据和索引
func (c *Collection) Drop(ctx context.Context) error {
	if err := c.coll.Drop(ctx); err != nil {
		return errors.NewQueryError("failed to drop collection", err)
	}
	return nil
} 