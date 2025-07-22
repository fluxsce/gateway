// Package client 提供MongoDB GridFS实现
//
// 此文件包含GridFS桶的实现，支持大文件的分块存储和管理
package client

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"gohub/pkg/mongo/errors"
	"gohub/pkg/mongo/types"
)

// GridFSBucket GridFS桶实现
// 实现types.GridFSBucket接口，提供文件存储操作
type GridFSBucket struct {
	bucket   *gridfs.Bucket    // MongoDB GridFS桶
	database *Database         // 父数据库引用
	name     string           // 桶名称
}

// NewGridFSBucket 创建新的GridFS桶
// database: 数据库实例，必须是有效的Database对象
// opts: 桶选项，如果为nil则使用默认选项
// 返回值: GridFS桶接口实例，如果创建失败会返回一个包含错误信息的桶实例
func NewGridFSBucket(database *Database, opts *types.GridFSBucketOptions) types.GridFSBucket {
	var bucketOpts *options.BucketOptions
	
	if opts != nil {
		bucketOpts = options.GridFSBucket()
		if opts.BucketName != nil {
			bucketOpts.SetName(*opts.BucketName)
		}
		if opts.ChunkSizeBytes != nil {
			bucketOpts.SetChunkSizeBytes(*opts.ChunkSizeBytes)
		}
		if opts.WriteConcern != nil {
			if wc, ok := opts.WriteConcern.(*writeconcern.WriteConcern); ok {
				bucketOpts.SetWriteConcern(wc)
			}
		}
		if opts.ReadConcern != nil {
			if rc, ok := opts.ReadConcern.(*readconcern.ReadConcern); ok {
				bucketOpts.SetReadConcern(rc)
			}
		}
		if opts.ReadPreference != nil {
			if rp, ok := opts.ReadPreference.(*readpref.ReadPref); ok {
				bucketOpts.SetReadPreference(rp)
			}
		}
	}
	
	bucket, err := gridfs.NewBucket(database.db, bucketOpts)
	if err != nil {
		// 如果创建失败，返回一个包含错误信息的桶实例
		return &GridFSBucket{
			bucket:   nil,
			database: database,
			name:     getBucketName(opts),
		}
	}
	
	return &GridFSBucket{
		bucket:   bucket,
		database: database,
		name:     getBucketName(opts),
	}
}

// getBucketName 获取桶名称
// opts: GridFS桶选项，可以为nil
// 返回值: 桶名称，如果选项中指定了名称则返回指定名称，否则返回默认名称"fs"
func getBucketName(opts *types.GridFSBucketOptions) string {
	if opts != nil && opts.BucketName != nil {
		return *opts.BucketName
	}
	return "fs" // 默认桶名称
}

// === 文件上传操作 ===

// UploadFromStream 从流上传文件
// ctx: 上下文对象，用于控制请求的生命周期
// filename: 存储在GridFS中的文件名
// source: 数据源，必须实现io.Reader接口
// opts: 上传选项，可以为nil使用默认选项
// 返回值: 上传文件的ID和错误信息
func (g *GridFSBucket) UploadFromStream(ctx context.Context, filename string, source interface{}, opts *types.GridFSUploadOptions) (interface{}, error) {
	if g.bucket == nil {
		return nil, errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	reader, ok := source.(io.Reader)
	if !ok {
		return nil, errors.NewQueryError("source must implement io.Reader", nil)
	}
	
	var uploadOpts *options.UploadOptions
	if opts != nil {
		uploadOpts = options.GridFSUpload()
		if opts.ChunkSizeBytes != nil {
			uploadOpts.SetChunkSizeBytes(*opts.ChunkSizeBytes)
		}
		if opts.Metadata != nil {
			uploadOpts.SetMetadata(opts.Metadata)
		}
	}
	
	fileID, err := g.bucket.UploadFromStream(filename, reader, uploadOpts)
	if err != nil {
		return nil, errors.NewQueryError("failed to upload from stream", err)
	}
	
	return fileID, nil
}

// UploadFromFile 从本地文件上传
// ctx: 上下文对象，用于控制请求的生命周期
// localFilePath: 本地文件的完整路径
// filename: 存储在GridFS中的文件名，如果为空则使用本地文件名
// opts: 上传选项，可以为nil使用默认选项
// 返回值: 上传文件的ID和错误信息
func (g *GridFSBucket) UploadFromFile(ctx context.Context, localFilePath string, filename string, opts *types.GridFSUploadOptions) (interface{}, error) {
	if g.bucket == nil {
		return nil, errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	file, err := os.Open(localFilePath)
	if err != nil {
		return nil, errors.NewQueryError("failed to open local file", err)
	}
	defer file.Close()
	
	// 如果filename为空，使用本地文件名
	if filename == "" {
		filename = filepath.Base(localFilePath)
	}
	
	return g.UploadFromStream(ctx, filename, file, opts)
}

// UploadFromBytes 从字节数组上传文件
// ctx: 上下文对象，用于控制请求的生命周期
// filename: 存储在GridFS中的文件名
// data: 要上传的字节数据
// opts: 上传选项，可以为nil使用默认选项
// 返回值: 上传文件的ID和错误信息
func (g *GridFSBucket) UploadFromBytes(ctx context.Context, filename string, data []byte, opts *types.GridFSUploadOptions) (interface{}, error) {
	if g.bucket == nil {
		return nil, errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	reader := strings.NewReader(string(data))
	return g.UploadFromStream(ctx, filename, reader, opts)
}

// === 文件下载操作 ===

// DownloadToStream 下载文件到流
// ctx: 上下文对象，用于控制请求的生命周期
// fileID: 要下载的文件ID
// destination: 目标流，必须实现io.Writer接口
// opts: 下载选项，可以为nil使用默认选项
// 返回值: 错误信息，成功时为nil
func (g *GridFSBucket) DownloadToStream(ctx context.Context, fileID interface{}, destination interface{}, opts *types.GridFSDownloadOptions) error {
	if g.bucket == nil {
		return errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	writer, ok := destination.(io.Writer)
	if !ok {
		return errors.NewQueryError("destination must implement io.Writer", nil)
	}
	
	_, err := g.bucket.DownloadToStream(fileID, writer)
	if err != nil {
		return errors.NewQueryError("failed to download to stream", err)
	}
	
	return nil
}

// DownloadToStreamByName 按文件名下载文件到流
// ctx: 上下文对象，用于控制请求的生命周期
// filename: 要下载的文件名
// destination: 目标流，必须实现io.Writer接口
// opts: 下载选项，可以为nil使用默认选项
// 返回值: 错误信息，成功时为nil
func (g *GridFSBucket) DownloadToStreamByName(ctx context.Context, filename string, destination interface{}, opts *types.GridFSDownloadOptions) error {
	if g.bucket == nil {
		return errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	writer, ok := destination.(io.Writer)
	if !ok {
		return errors.NewQueryError("destination must implement io.Writer", nil)
	}
	
	var downloadOpts *options.NameOptions
	if opts != nil && opts.Revision != nil {
		downloadOpts = options.GridFSName().SetRevision(*opts.Revision)
	}
	
	_, err := g.bucket.DownloadToStreamByName(filename, writer, downloadOpts)
	if err != nil {
		return errors.NewQueryError("failed to download to stream by name", err)
	}
	
	return nil
}

// DownloadToFile 下载文件到本地文件
// ctx: 上下文对象，用于控制请求的生命周期
// fileID: 要下载的文件ID
// localFilePath: 本地文件保存路径，如果目录不存在会自动创建
// opts: 下载选项，可以为nil使用默认选项
// 返回值: 错误信息，成功时为nil
func (g *GridFSBucket) DownloadToFile(ctx context.Context, fileID interface{}, localFilePath string, opts *types.GridFSDownloadOptions) error {
	if g.bucket == nil {
		return errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	// 创建目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(localFilePath), 0755); err != nil {
		return errors.NewQueryError("failed to create directory", err)
	}
	
	file, err := os.Create(localFilePath)
	if err != nil {
		return errors.NewQueryError("failed to create local file", err)
	}
	defer file.Close()
	
	return g.DownloadToStream(ctx, fileID, file, opts)
}

// DownloadToFileByName 按文件名下载文件到本地文件
// ctx: 上下文对象，用于控制请求的生命周期
// filename: 要下载的文件名
// localFilePath: 本地文件保存路径，如果目录不存在会自动创建
// opts: 下载选项，可以为nil使用默认选项
// 返回值: 错误信息，成功时为nil
func (g *GridFSBucket) DownloadToFileByName(ctx context.Context, filename string, localFilePath string, opts *types.GridFSDownloadOptions) error {
	if g.bucket == nil {
		return errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	// 创建目录（如果不存在）
	if err := os.MkdirAll(filepath.Dir(localFilePath), 0755); err != nil {
		return errors.NewQueryError("failed to create directory", err)
	}
	
	file, err := os.Create(localFilePath)
	if err != nil {
		return errors.NewQueryError("failed to create local file", err)
	}
	defer file.Close()
	
	return g.DownloadToStreamByName(ctx, filename, file, opts)
}

// DownloadToBytes 下载文件到字节数组
// ctx: 上下文对象，用于控制请求的生命周期
// fileID: 要下载的文件ID
// opts: 下载选项，可以为nil使用默认选项
// 返回值: 文件数据的字节数组和错误信息
func (g *GridFSBucket) DownloadToBytes(ctx context.Context, fileID interface{}, opts *types.GridFSDownloadOptions) ([]byte, error) {
	if g.bucket == nil {
		return nil, errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	var buffer strings.Builder
	err := g.DownloadToStream(ctx, fileID, &buffer, opts)
	if err != nil {
		return nil, err
	}
	
	return []byte(buffer.String()), nil
}

// DownloadToBytesByName 按文件名下载文件到字节数组
// ctx: 上下文对象，用于控制请求的生命周期
// filename: 要下载的文件名
// opts: 下载选项，可以为nil使用默认选项
// 返回值: 文件数据的字节数组和错误信息
func (g *GridFSBucket) DownloadToBytesByName(ctx context.Context, filename string, opts *types.GridFSDownloadOptions) ([]byte, error) {
	if g.bucket == nil {
		return nil, errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	var buffer strings.Builder
	err := g.DownloadToStreamByName(ctx, filename, &buffer, opts)
	if err != nil {
		return nil, err
	}
	
	return []byte(buffer.String()), nil
}

// === 文件删除操作 ===

// Delete 删除文件
// ctx: 上下文对象，用于控制请求的生命周期
// fileID: 要删除的文件ID
// 返回值: 错误信息，成功时为nil
func (g *GridFSBucket) Delete(ctx context.Context, fileID interface{}) error {
	if g.bucket == nil {
		return errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	if err := g.bucket.Delete(fileID); err != nil {
		return errors.NewQueryError("failed to delete file", err)
	}
	
	return nil
}

// DeleteByName 按文件名删除文件
// ctx: 上下文对象，用于控制请求的生命周期
// filename: 要删除的文件名
// 返回值: 错误信息，成功时为nil
func (g *GridFSBucket) DeleteByName(ctx context.Context, filename string) error {
	if g.bucket == nil {
		return errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	// 查找文件ID
	fileInfo, err := g.FindOne(ctx, types.Filter{"filename": filename}, nil)
	if err != nil {
		return errors.NewQueryError("failed to find file by name", err)
	}
	
	return g.Delete(ctx, fileInfo.ID)
}

// === 文件查询操作 ===

// Find 查找文件
// ctx: 上下文对象，用于控制请求的生命周期
// filter: 查询过滤条件，如{"filename": "test.txt"}
// opts: 查找选项，如分页、排序等，可以为nil使用默认选项
// 返回值: 游标对象和错误信息，用于遍历查找结果
func (g *GridFSBucket) Find(ctx context.Context, filter types.Filter, opts *types.GridFSFindOptions) (types.MongoCursor, error) {
	if g.bucket == nil {
		return nil, errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	var findOpts *options.GridFSFindOptions
	if opts != nil {
		findOpts = options.GridFSFind()
		if opts.BatchSize != nil {
			findOpts.SetBatchSize(*opts.BatchSize)
		}
		if opts.Limit != nil {
			findOpts.SetLimit(*opts.Limit)
		}
		if opts.MaxTime != nil {
			findOpts.SetMaxTime(*opts.MaxTime)
		}
		if opts.NoCursorTimeout != nil {
			findOpts.SetNoCursorTimeout(*opts.NoCursorTimeout)
		}
		if opts.Skip != nil {
			findOpts.SetSkip(*opts.Skip)
		}
		if opts.Sort != nil {
			findOpts.SetSort(opts.Sort)
		}
	}
	
	cursor, err := g.bucket.Find(filter, findOpts)
	if err != nil {
		return nil, errors.NewQueryError("failed to find files", err)
	}
	
	return &Cursor{cursor: cursor}, nil
}

// FindOne 查找单个文件
// ctx: 上下文对象，用于控制请求的生命周期
// filter: 查询过滤条件，如{"filename": "test.txt"}
// opts: 查找选项，如排序等，可以为nil使用默认选项
// 返回值: 文件信息对象和错误信息，如果没找到文件返回错误
func (g *GridFSBucket) FindOne(ctx context.Context, filter types.Filter, opts *types.GridFSFindOptions) (*types.GridFSFile, error) {
	if g.bucket == nil {
		return nil, errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	cursor, err := g.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	if !cursor.Next(ctx) {
		return nil, errors.NewQueryError("no file found", nil)
	}
	
	var file types.GridFSFile
	if err := cursor.Decode(&file); err != nil {
		return nil, errors.NewQueryError("failed to decode file", err)
	}
	
	return &file, nil
}

// GetFileInfo 获取文件信息
// ctx: 上下文对象，用于控制请求的生命周期
// fileID: 要获取信息的文件ID
// 返回值: 文件信息对象和错误信息
func (g *GridFSBucket) GetFileInfo(ctx context.Context, fileID interface{}) (*types.GridFSFile, error) {
	return g.FindOne(ctx, types.Filter{"_id": fileID}, nil)
}

// GetFileInfoByName 按文件名获取文件信息
// ctx: 上下文对象，用于控制请求的生命周期
// filename: 要获取信息的文件名
// 返回值: 文件信息对象和错误信息
func (g *GridFSBucket) GetFileInfoByName(ctx context.Context, filename string) (*types.GridFSFile, error) {
	return g.FindOne(ctx, types.Filter{"filename": filename}, nil)
}

// === 文件管理操作 ===

// ListFiles 列出所有文件
// ctx: 上下文对象，用于控制请求的生命周期
// filter: 查询过滤条件，如{"filename": {"$regex": "^test.*"}}
// opts: 查找选项，如分页、排序等，可以为nil使用默认选项
// 返回值: 文件信息列表和错误信息
func (g *GridFSBucket) ListFiles(ctx context.Context, filter types.Filter, opts *types.GridFSFindOptions) ([]*types.GridFSFile, error) {
	cursor, err := g.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var files []*types.GridFSFile
	for cursor.Next(ctx) {
		var file types.GridFSFile
		if err := cursor.Decode(&file); err != nil {
			return nil, errors.NewQueryError("failed to decode file", err)
		}
		files = append(files, &file)
	}
	
	if err := cursor.Err(); err != nil {
		return nil, errors.NewQueryError("cursor error", err)
	}
	
	return files, nil
}

// GetFileCount 获取文件数量
// ctx: 上下文对象，用于控制请求的生命周期
// filter: 查询过滤条件，用于统计特定条件的文件数量
// 返回值: 文件数量和错误信息
func (g *GridFSBucket) GetFileCount(ctx context.Context, filter types.Filter) (int64, error) {
	if g.bucket == nil {
		return 0, errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	// 使用聚合查询统计文件数量
	collection := g.database.db.Collection(g.name + ".files")
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.NewQueryError("failed to count files", err)
	}
	
	return count, nil
}

// Rename 重命名文件
// ctx: 上下文对象，用于控制请求的生命周期
// fileID: 要重命名的文件ID
// newFilename: 新的文件名
// 返回值: 错误信息，成功时为nil
func (g *GridFSBucket) Rename(ctx context.Context, fileID interface{}, newFilename string) error {
	if g.bucket == nil {
		return errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	if err := g.bucket.Rename(fileID, newFilename); err != nil {
		return errors.NewQueryError("failed to rename file", err)
	}
	
	return nil
}

// === 批量操作 ===

// BatchUpload 批量上传文件
// ctx: 上下文对象，用于控制请求的生命周期
// files: 文件映射，key为本地文件路径，value为GridFS中的文件名
// opts: 上传选项，应用于所有文件，可以为nil使用默认选项
// 返回值: 上传成功的文件ID列表和错误信息，如果某个文件上传失败会立即返回错误
func (g *GridFSBucket) BatchUpload(ctx context.Context, files map[string]string, opts *types.GridFSUploadOptions) ([]interface{}, error) {
	var uploadedIDs []interface{}
	
	for localPath, gridfsName := range files {
		fileID, err := g.UploadFromFile(ctx, localPath, gridfsName, opts)
		if err != nil {
			return uploadedIDs, errors.NewQueryError("failed to upload file: "+localPath, err)
		}
		uploadedIDs = append(uploadedIDs, fileID)
	}
	
	return uploadedIDs, nil
}

// BatchDownload 批量下载文件
// ctx: 上下文对象，用于控制请求的生命周期
// files: 文件映射，key为文件ID，value为本地保存路径
// opts: 下载选项，应用于所有文件，可以为nil使用默认选项
// 返回值: 下载成功的文件路径列表和错误信息，如果某个文件下载失败会立即返回错误
func (g *GridFSBucket) BatchDownload(ctx context.Context, files map[interface{}]string, opts *types.GridFSDownloadOptions) ([]string, error) {
	var downloadedFiles []string
	
	for fileID, localPath := range files {
		err := g.DownloadToFile(ctx, fileID, localPath, opts)
		if err != nil {
			return downloadedFiles, errors.NewQueryError("failed to download file", err)
		}
		downloadedFiles = append(downloadedFiles, localPath)
	}
	
	return downloadedFiles, nil
}

// BatchDelete 批量删除文件
// ctx: 上下文对象，用于控制请求的生命周期
// fileIDs: 要删除的文件ID列表
// 返回值: 成功删除的文件数量和错误信息，即使某些文件删除失败也会继续删除其他文件
func (g *GridFSBucket) BatchDelete(ctx context.Context, fileIDs []interface{}) (int64, error) {
	var deletedCount int64
	
	for _, fileID := range fileIDs {
		if err := g.Delete(ctx, fileID); err != nil {
			// 记录错误但继续删除其他文件
			continue
		}
		deletedCount++
	}
	
	return deletedCount, nil
}

// === 桶管理操作 ===

// Drop 删除整个GridFS桶
// ctx: 上下文对象，用于控制请求的生命周期
// 返回值: 错误信息，成功时为nil
// 警告: 此操作会删除桶中的所有文件和分块数据，不可恢复
func (g *GridFSBucket) Drop(ctx context.Context) error {
	if g.bucket == nil {
		return errors.NewQueryError("GridFS bucket not initialized", nil)
	}
	
	if err := g.bucket.Drop(); err != nil {
		return errors.NewQueryError("failed to drop GridFS bucket", err)
	}
	
	return nil
}

// GetBucketName 获取桶名称
// 返回值: 桶名称字符串
func (g *GridFSBucket) GetBucketName() string {
	return g.name
}

// convertToObjectID 将interface{}转换为ObjectID
// id: 要转换的ID值，支持primitive.ObjectID和string类型
// 返回值: 转换后的ObjectID和错误信息
func convertToObjectID(id interface{}) (primitive.ObjectID, error) {
	switch v := id.(type) {
	case primitive.ObjectID:
		return v, nil
	case string:
		return primitive.ObjectIDFromHex(v)
	default:
		return primitive.NilObjectID, errors.NewQueryError("invalid file ID type", nil)
	}
}

// convertToBSONDocument 将types.Document转换为bson.D
// doc: 要转换的Document对象
// 返回值: 转换后的bson.D对象，保持键值对的顺序
func convertToBSONDocument(doc types.Document) bson.D {
	var result bson.D
	for key, value := range doc {
		result = append(result, bson.E{Key: key, Value: value})
	}
	return result
} 