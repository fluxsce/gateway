// Package client 提供MongoDB游标和结果实现
//
// 此文件包含Cursor和Result结构体及其相关方法的实现
package client

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"

	"gateway/pkg/mongo/errors"
)

// Cursor MongoDB游标实现
// 实现types.MongoCursor接口，提供结果遍历功能
type Cursor struct {
	cursor *mongo.Cursor // MongoDB驱动游标
}

// Result 单个结果实现
// 实现types.SingleResult接口，提供单个文档查询结果处理
type Result struct {
	result *mongo.SingleResult // MongoDB驱动单个结果
}

// === 游标方法实现 ===

// Next 移动到下一个文档
// 返回是否有下一个文档可用
func (c *Cursor) Next(ctx context.Context) bool {
	return c.cursor.Next(ctx)
}

// Decode 解码当前文档
// 将当前文档解码到指定的结构体中
func (c *Cursor) Decode(val interface{}) error {
	if err := c.cursor.Decode(val); err != nil {
		return errors.NewQueryError("failed to decode document", err)
	}
	return nil
}

// All 获取所有剩余文档
// 将所有剩余文档解码到指定的切片中
func (c *Cursor) All(ctx context.Context, results interface{}) error {
	if err := c.cursor.All(ctx, results); err != nil {
		return errors.NewQueryError("failed to decode all documents", err)
	}
	return nil
}

// Close 关闭游标
// 释放游标占用的资源
func (c *Cursor) Close(ctx context.Context) error {
	return c.cursor.Close(ctx)
}

// Err 获取游标错误
// 返回游标操作中遇到的错误
func (c *Cursor) Err() error {
	return c.cursor.Err()
}

// === 单个结果方法实现 ===

// Decode 解码结果文档
// 将结果文档解码到指定的结构体中
func (r *Result) Decode(val interface{}) error {
	err := r.result.Decode(val)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrDocumentNotFound
		}
		return errors.NewQueryError("failed to decode result", err)
	}
	return nil
}

// Err 获取查询错误
// 返回查询操作中遇到的错误
func (r *Result) Err() error {
	err := r.result.Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.ErrDocumentNotFound
		}
		return err
	}
	return nil
}
