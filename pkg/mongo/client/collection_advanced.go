// Package client 提供MongoDB集合高级操作实现
//
// 此文件包含Collection的高级操作方法，包括原子操作、聚合、索引管理等
package client

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gohub/pkg/mongo/errors"
	"gohub/pkg/mongo/types"
)

// === 替换操作 ===

// ReplaceOne 替换单个文档
// 根据过滤条件替换第一个匹配的文档
func (c *Collection) ReplaceOne(ctx context.Context, filter types.Filter, replacement types.Document, opts *types.ReplaceOptions) (*types.UpdateResult, error) {
	// 构建替换选项
	replaceOptions := options.Replace()

	if opts != nil {
		if opts.BypassDocumentValidation != nil {
			replaceOptions.SetBypassDocumentValidation(*opts.BypassDocumentValidation)
		}
		if opts.Upsert != nil {
			replaceOptions.SetUpsert(*opts.Upsert)
		}
	}

	// 执行替换
	result, err := c.coll.ReplaceOne(ctx, filter, replacement, replaceOptions)
	if err != nil {
		return nil, errors.NewUpdateError("failed to replace document", err)
	}

	return &types.UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
		UpsertedID:    result.UpsertedID,
	}, nil
}

// === 原子操作 ===

// FindOneAndUpdate 查找并更新单个文档
// 原子地查找、更新并返回文档（更新前或更新后的版本）
func (c *Collection) FindOneAndUpdate(ctx context.Context, filter types.Filter, update types.Update, opts *types.FindOneAndUpdateOptions) types.SingleResult {
	// 构建查找并更新选项
	findOneAndUpdateOptions := options.FindOneAndUpdate()

	if opts != nil {
		if opts.BypassDocumentValidation != nil {
			findOneAndUpdateOptions.SetBypassDocumentValidation(*opts.BypassDocumentValidation)
		}
		if opts.Projection != nil {
			findOneAndUpdateOptions.SetProjection(opts.Projection)
		}
		if opts.ReturnDocument != nil {
			if *opts.ReturnDocument == types.ReturnDocumentAfter {
				findOneAndUpdateOptions.SetReturnDocument(options.After)
			} else {
				findOneAndUpdateOptions.SetReturnDocument(options.Before)
			}
		}
		if opts.Sort != nil {
			findOneAndUpdateOptions.SetSort(opts.Sort)
		}
		if opts.Upsert != nil {
			findOneAndUpdateOptions.SetUpsert(*opts.Upsert)
		}
		if opts.MaxTime != nil {
			findOneAndUpdateOptions.SetMaxTime(*opts.MaxTime)
		}
	}

	// 执行查找并更新
	result := c.coll.FindOneAndUpdate(ctx, filter, update, findOneAndUpdateOptions)
	return &Result{result: result}
}

// FindOneAndDelete 查找并删除单个文档
// 原子地查找、删除并返回被删除的文档
func (c *Collection) FindOneAndDelete(ctx context.Context, filter types.Filter, opts *types.FindOneAndDeleteOptions) types.SingleResult {
	// 构建查找并删除选项
	findOneAndDeleteOptions := options.FindOneAndDelete()

	if opts != nil {
		if opts.Projection != nil {
			findOneAndDeleteOptions.SetProjection(opts.Projection)
		}
		if opts.Sort != nil {
			findOneAndDeleteOptions.SetSort(opts.Sort)
		}
		if opts.MaxTime != nil {
			findOneAndDeleteOptions.SetMaxTime(*opts.MaxTime)
		}
	}

	// 执行查找并删除
	result := c.coll.FindOneAndDelete(ctx, filter, findOneAndDeleteOptions)
	return &Result{result: result}
}

// FindOneAndReplace 查找并替换单个文档
// 原子地查找、替换并返回文档（替换前或替换后的版本）
func (c *Collection) FindOneAndReplace(ctx context.Context, filter types.Filter, replacement types.Document, opts *types.FindOneAndReplaceOptions) types.SingleResult {
	// 构建查找并替换选项
	findOneAndReplaceOptions := options.FindOneAndReplace()

	if opts != nil {
		if opts.BypassDocumentValidation != nil {
			findOneAndReplaceOptions.SetBypassDocumentValidation(*opts.BypassDocumentValidation)
		}
		if opts.Projection != nil {
			findOneAndReplaceOptions.SetProjection(opts.Projection)
		}
		if opts.ReturnDocument != nil {
			if *opts.ReturnDocument == types.ReturnDocumentAfter {
				findOneAndReplaceOptions.SetReturnDocument(options.After)
			} else {
				findOneAndReplaceOptions.SetReturnDocument(options.Before)
			}
		}
		if opts.Sort != nil {
			findOneAndReplaceOptions.SetSort(opts.Sort)
		}
		if opts.Upsert != nil {
			findOneAndReplaceOptions.SetUpsert(*opts.Upsert)
		}
		if opts.MaxTime != nil {
			findOneAndReplaceOptions.SetMaxTime(*opts.MaxTime)
		}
	}

	// 执行查找并替换
	result := c.coll.FindOneAndReplace(ctx, filter, replacement, findOneAndReplaceOptions)
	return &Result{result: result}
}

// === 聚合操作 ===

// Aggregate 执行聚合查询
// 根据聚合管道执行复杂的数据处理和分析
func (c *Collection) Aggregate(ctx context.Context, pipeline types.Pipeline, opts *types.AggregateOptions) (types.MongoCursor, error) {
	// 构建聚合选项
	aggregateOptions := options.Aggregate()

	if opts != nil {
		if opts.AllowDiskUse != nil {
			aggregateOptions.SetAllowDiskUse(*opts.AllowDiskUse)
		}
		if opts.BatchSize != nil {
			aggregateOptions.SetBatchSize(*opts.BatchSize)
		}
		if opts.MaxTime != nil {
			aggregateOptions.SetMaxTime(*opts.MaxTime)
		}
	}

	// 执行聚合
	cursor, err := c.coll.Aggregate(ctx, pipeline, aggregateOptions)
	if err != nil {
		return nil, errors.NewQueryError("failed to execute aggregation", err)
	}

	return &Cursor{cursor: cursor}, nil
}

// === 索引操作 ===

// CreateIndex 创建索引
// 根据索引模型创建索引，提高查询性能
func (c *Collection) CreateIndex(ctx context.Context, model types.IndexModel) (string, error) {
	// 转换索引模型
	indexModel := mongo.IndexModel{
		Keys: model.Keys,
	}

	if model.Options != nil {
		indexOptions := options.Index()
		
		if model.Options.Background != nil {
			indexOptions.SetBackground(*model.Options.Background)
		}
		if model.Options.ExpireAfterSeconds != nil {
			indexOptions.SetExpireAfterSeconds(*model.Options.ExpireAfterSeconds)
		}
		if model.Options.Name != nil {
			indexOptions.SetName(*model.Options.Name)
		}
		if model.Options.Sparse != nil {
			indexOptions.SetSparse(*model.Options.Sparse)
		}
		if model.Options.Unique != nil {
			indexOptions.SetUnique(*model.Options.Unique)
		}
		
		indexModel.Options = indexOptions
	}

	// 创建索引
	name, err := c.coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return "", errors.NewIndexError("failed to create index", err)
	}

	return name, nil
}

// ListIndexes 列出所有索引
// 返回集合中所有索引的信息
func (c *Collection) ListIndexes(ctx context.Context) (types.MongoCursor, error) {
	cursor, err := c.coll.Indexes().List(ctx)
	if err != nil {
		return nil, errors.NewQueryError("failed to list indexes", err)
	}

	return &Cursor{cursor: cursor}, nil
}

// DropIndex 删除指定索引
// 根据索引名称删除索引
func (c *Collection) DropIndex(ctx context.Context, name string) error {
	if _, err := c.coll.Indexes().DropOne(ctx, name); err != nil {
		return errors.NewIndexError("failed to drop index", err)
	}
	return nil
}

// === 变更流操作 ===

// Watch 监视集合变更
// 创建变更流以监视集合中的文档变更
func (c *Collection) Watch(ctx context.Context, pipeline types.Pipeline, opts *types.ChangeStreamOptions) (types.ChangeStream, error) {
	// 目前返回一个简单的错误，因为实际实现需要完整的变更流支持
	return nil, errors.NewQueryError("ChangeStream not implemented yet", nil)
} 