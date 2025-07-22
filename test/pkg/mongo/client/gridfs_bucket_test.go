package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/pkg/mongo/types"
)

var (
	testGridFSBucket types.GridFSBucket
	testTempDir      string
)

// setupGridFSTest 设置GridFS测试环境
func setupGridFSTest() error {
	if testDatabase == nil {
		return fmt.Errorf("database not available")
	}

	// 使用logs目录作为测试目录
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// 创建临时子目录
	tempDir, err := os.MkdirTemp(logsDir, "gridfs_test_*")
	if err != nil {
		return err
	}
	testTempDir = tempDir

	// 创建GridFS桶
	bucketOpts := &types.GridFSBucketOptions{
		BucketName:     &[]string{"test_bucket"}[0],
		ChunkSizeBytes: &[]int32{1024}[0], // 1KB chunks for testing
	}

	// 使用Database接口的GridFSBucketWithOptions方法
	testGridFSBucket = testDatabase.GridFSBucketWithOptions(bucketOpts)

	return nil
}

// teardownGridFSTest 清理GridFS测试环境
func teardownGridFSTest() {
	if testGridFSBucket != nil {
		ctx := context.Background()
		testGridFSBucket.Drop(ctx)
	}

	if testTempDir != "" {
		os.RemoveAll(testTempDir)
	}
}

// skipIfNoGridFS 如果没有GridFS环境则跳过测试
func skipIfNoGridFS(t *testing.T) {
	if testGridFSBucket == nil {
		if err := setupGridFSTest(); err != nil {
			t.Skip("跳过测试：GridFS环境不可用 - " + err.Error())
		}
	}
}

// TestGridFSUploadOperations 测试GridFS上传操作
func TestGridFSUploadOperations(t *testing.T) {
	skipIfNoGridFS(t)
	defer teardownGridFSTest()

	ctx := context.Background()

	t.Run("UploadFromBytes", func(t *testing.T) {
		testData := []byte("Hello, GridFS! This is a test file.")
		filename := "test_bytes.txt"

		fileID, err := testGridFSBucket.UploadFromBytes(ctx, filename, testData, nil)
		assert.NoError(t, err, "从字节上传应该成功")
		assert.NotNil(t, fileID, "文件ID不应为空")

		t.Logf("上传文件ID: %v", fileID)

		// 验证文件信息
		fileInfo, err := testGridFSBucket.GetFileInfo(ctx, fileID)
		assert.NoError(t, err, "获取文件信息应该成功")
		assert.Equal(t, filename, fileInfo.Filename, "文件名应该匹配")
		assert.Equal(t, int64(len(testData)), fileInfo.Length, "文件大小应该匹配")
	})

	t.Run("UploadFromStream", func(t *testing.T) {
		testData := "This is a stream upload test with more content to test chunking."
		reader := strings.NewReader(testData)
		filename := "test_stream.txt"

		metadata := types.Document{
			"description": "Stream upload test",
			"category":    "test",
		}

		opts := &types.GridFSUploadOptions{
			Metadata: metadata,
		}

		fileID, err := testGridFSBucket.UploadFromStream(ctx, filename, reader, opts)
		assert.NoError(t, err, "从流上传应该成功")
		assert.NotNil(t, fileID, "文件ID不应为空")

		// 验证文件信息和元数据
		fileInfo, err := testGridFSBucket.GetFileInfo(ctx, fileID)
		assert.NoError(t, err, "获取文件信息应该成功")
		assert.Equal(t, filename, fileInfo.Filename, "文件名应该匹配")
		assert.Equal(t, int64(len(testData)), fileInfo.Length, "文件大小应该匹配")
		assert.NotNil(t, fileInfo.Metadata, "元数据不应为空")

		t.Logf("文件信息: %+v", fileInfo)
	})

	t.Run("UploadFromFile", func(t *testing.T) {
		// 创建临时文件
		testContent := "This is a file upload test with some content."
		testFile := filepath.Join(testTempDir, "test_upload.txt")

		err := os.WriteFile(testFile, []byte(testContent), 0644)
		require.NoError(t, err, "创建测试文件应该成功")

		filename := "uploaded_file.txt"

		fileID, err := testGridFSBucket.UploadFromFile(ctx, testFile, filename, nil)
		assert.NoError(t, err, "从文件上传应该成功")
		assert.NotNil(t, fileID, "文件ID不应为空")

		// 验证文件信息
		fileInfo, err := testGridFSBucket.GetFileInfo(ctx, fileID)
		assert.NoError(t, err, "获取文件信息应该成功")
		assert.Equal(t, filename, fileInfo.Filename, "文件名应该匹配")
		assert.Equal(t, int64(len(testContent)), fileInfo.Length, "文件大小应该匹配")

		t.Logf("上传文件: %s -> %s (ID: %v)", testFile, filename, fileID)
	})

	t.Run("UploadLargeFile", func(t *testing.T) {
		// 创建大文件 (5KB) 来测试分块
		largeData := make([]byte, 5*1024)
		_, err := rand.Read(largeData)
		require.NoError(t, err, "生成随机数据应该成功")

		filename := "large_file.bin"

		fileID, err := testGridFSBucket.UploadFromBytes(ctx, filename, largeData, nil)
		assert.NoError(t, err, "上传大文件应该成功")
		assert.NotNil(t, fileID, "文件ID不应为空")

		// 验证文件信息
		fileInfo, err := testGridFSBucket.GetFileInfo(ctx, fileID)
		assert.NoError(t, err, "获取文件信息应该成功")
		assert.Equal(t, filename, fileInfo.Filename, "文件名应该匹配")
		assert.Equal(t, int64(len(largeData)), fileInfo.Length, "文件大小应该匹配")
		assert.Equal(t, int32(1024), fileInfo.ChunkSize, "分块大小应该匹配")

		t.Logf("大文件上传成功: %s (大小: %d bytes, 分块: %d)", filename, fileInfo.Length, fileInfo.ChunkSize)
	})
}

// TestGridFSDownloadOperations 测试GridFS下载操作
func TestGridFSDownloadOperations(t *testing.T) {
	skipIfNoGridFS(t)
	defer teardownGridFSTest()

	ctx := context.Background()

	// 准备测试数据
	testData := []byte("Hello, GridFS Download! This is test content for download operations.")
	filename := "download_test.txt"

	// 上传测试文件
	fileID, err := testGridFSBucket.UploadFromBytes(ctx, filename, testData, nil)
	require.NoError(t, err, "上传测试文件应该成功")

	t.Run("DownloadToBytes", func(t *testing.T) {
		downloadedData, err := testGridFSBucket.DownloadToBytes(ctx, fileID, nil)
		assert.NoError(t, err, "下载到字节数组应该成功")
		assert.Equal(t, testData, downloadedData, "下载的数据应该匹配")

		t.Logf("下载数据长度: %d", len(downloadedData))
	})

	t.Run("DownloadToBytesByName", func(t *testing.T) {
		downloadedData, err := testGridFSBucket.DownloadToBytesByName(ctx, filename, nil)
		assert.NoError(t, err, "按名称下载到字节数组应该成功")
		assert.Equal(t, testData, downloadedData, "下载的数据应该匹配")

		t.Logf("按名称下载数据长度: %d", len(downloadedData))
	})

	t.Run("DownloadToStream", func(t *testing.T) {
		var buffer bytes.Buffer

		err := testGridFSBucket.DownloadToStream(ctx, fileID, &buffer, nil)
		assert.NoError(t, err, "下载到流应该成功")
		assert.Equal(t, testData, buffer.Bytes(), "下载的数据应该匹配")

		t.Logf("下载到流的数据长度: %d", buffer.Len())
	})

	t.Run("DownloadToStreamByName", func(t *testing.T) {
		var buffer bytes.Buffer

		err := testGridFSBucket.DownloadToStreamByName(ctx, filename, &buffer, nil)
		assert.NoError(t, err, "按名称下载到流应该成功")
		assert.Equal(t, testData, buffer.Bytes(), "下载的数据应该匹配")

		t.Logf("按名称下载到流的数据长度: %d", buffer.Len())
	})

	t.Run("DownloadToFile", func(t *testing.T) {
		downloadPath := filepath.Join(testTempDir, "downloaded_file.txt")

		err := testGridFSBucket.DownloadToFile(ctx, fileID, downloadPath, nil)
		assert.NoError(t, err, "下载到文件应该成功")

		// 验证文件内容
		downloadedData, err := os.ReadFile(downloadPath)
		assert.NoError(t, err, "读取下载文件应该成功")
		assert.Equal(t, testData, downloadedData, "下载的文件内容应该匹配")

		t.Logf("下载到文件: %s", downloadPath)
	})

	t.Run("DownloadToFileByName", func(t *testing.T) {
		downloadPath := filepath.Join(testTempDir, "downloaded_by_name.txt")

		err := testGridFSBucket.DownloadToFileByName(ctx, filename, downloadPath, nil)
		assert.NoError(t, err, "按名称下载到文件应该成功")

		// 验证文件内容
		downloadedData, err := os.ReadFile(downloadPath)
		assert.NoError(t, err, "读取下载文件应该成功")
		assert.Equal(t, testData, downloadedData, "下载的文件内容应该匹配")

		t.Logf("按名称下载到文件: %s", downloadPath)
	})
}

// TestGridFSQueryOperations 测试GridFS查询操作
func TestGridFSQueryOperations(t *testing.T) {
	skipIfNoGridFS(t)
	defer teardownGridFSTest()

	ctx := context.Background()

	// 准备测试数据
	testFiles := []struct {
		filename string
		content  string
		metadata types.Document
	}{
		{
			filename: "file1.txt",
			content:  "Content of file 1",
			metadata: types.Document{"category": "text", "size": "small"},
		},
		{
			filename: "file2.txt",
			content:  "Content of file 2 with more data",
			metadata: types.Document{"category": "text", "size": "medium"},
		},
		{
			filename: "image.jpg",
			content:  "Fake image content",
			metadata: types.Document{"category": "image", "size": "large"},
		},
	}

	var uploadedIDs []interface{}

	// 上传测试文件
	for _, file := range testFiles {
		opts := &types.GridFSUploadOptions{
			Metadata: file.metadata,
		}

		fileID, err := testGridFSBucket.UploadFromBytes(ctx, file.filename, []byte(file.content), opts)
		require.NoError(t, err, "上传测试文件应该成功")
		uploadedIDs = append(uploadedIDs, fileID)
	}

	t.Run("Find", func(t *testing.T) {
		// 查找所有文本文件
		cursor, err := testGridFSBucket.Find(ctx, types.Filter{"metadata.category": "text"}, nil)
		assert.NoError(t, err, "查找文件应该成功")
		defer cursor.Close(ctx)

		var files []types.GridFSFile
		err = cursor.All(ctx, &files)
		assert.NoError(t, err, "解码文件应该成功")
		assert.Len(t, files, 2, "应该找到2个文本文件")

		t.Logf("找到 %d 个文本文件", len(files))
		for _, file := range files {
			t.Logf("文件: %s (大小: %d)", file.Filename, file.Length)
		}
	})

	t.Run("FindWithOptions", func(t *testing.T) {
		// 带选项的查找
		opts := &types.GridFSFindOptions{
			Sort:  types.Document{"length": -1}, // 按大小降序
			Limit: &[]int32{2}[0],
		}

		cursor, err := testGridFSBucket.Find(ctx, types.Filter{}, opts)
		assert.NoError(t, err, "带选项的查找应该成功")
		defer cursor.Close(ctx)

		var files []types.GridFSFile
		err = cursor.All(ctx, &files)
		assert.NoError(t, err, "解码文件应该成功")
		assert.Len(t, files, 2, "应该返回2个文件")

		// 验证排序
		if len(files) >= 2 {
			assert.GreaterOrEqual(t, files[0].Length, files[1].Length, "文件应该按大小降序排列")
		}

		t.Logf("排序后的文件: %v", files)
	})

	t.Run("FindOne", func(t *testing.T) {
		file, err := testGridFSBucket.FindOne(ctx, types.Filter{"filename": "file1.txt"}, nil)
		assert.NoError(t, err, "查找单个文件应该成功")
		assert.Equal(t, "file1.txt", file.Filename, "文件名应该匹配")
		assert.Equal(t, int64(len("Content of file 1")), file.Length, "文件大小应该匹配")

		t.Logf("找到文件: %+v", file)
	})

	t.Run("GetFileInfo", func(t *testing.T) {
		fileInfo, err := testGridFSBucket.GetFileInfo(ctx, uploadedIDs[0])
		assert.NoError(t, err, "获取文件信息应该成功")
		assert.Equal(t, "file1.txt", fileInfo.Filename, "文件名应该匹配")
		assert.NotNil(t, fileInfo.Metadata, "元数据不应为空")

		t.Logf("文件信息: %+v", fileInfo)
	})

	t.Run("GetFileInfoByName", func(t *testing.T) {
		fileInfo, err := testGridFSBucket.GetFileInfoByName(ctx, "image.jpg")
		assert.NoError(t, err, "按名称获取文件信息应该成功")
		assert.Equal(t, "image.jpg", fileInfo.Filename, "文件名应该匹配")
		assert.Equal(t, "image", fileInfo.Metadata["category"], "元数据应该匹配")

		t.Logf("按名称获取文件信息: %+v", fileInfo)
	})

	t.Run("ListFiles", func(t *testing.T) {
		files, err := testGridFSBucket.ListFiles(ctx, types.Filter{}, nil)
		assert.NoError(t, err, "列出文件应该成功")
		assert.Len(t, files, 3, "应该有3个文件")

		t.Logf("文件列表:")
		for _, file := range files {
			t.Logf("  - %s (大小: %d, 类型: %s)", file.Filename, file.Length, file.Metadata["category"])
		}
	})

	t.Run("GetFileCount", func(t *testing.T) {
		count, err := testGridFSBucket.GetFileCount(ctx, types.Filter{})
		assert.NoError(t, err, "获取文件数量应该成功")
		assert.Equal(t, int64(3), count, "应该有3个文件")

		// 按类型统计
		textCount, err := testGridFSBucket.GetFileCount(ctx, types.Filter{"metadata.category": "text"})
		assert.NoError(t, err, "获取文本文件数量应该成功")
		assert.Equal(t, int64(2), textCount, "应该有2个文本文件")

		t.Logf("总文件数: %d, 文本文件数: %d", count, textCount)
	})
}

// TestGridFSDeleteOperations 测试GridFS删除操作
func TestGridFSDeleteOperations(t *testing.T) {
	skipIfNoGridFS(t)
	defer teardownGridFSTest()

	ctx := context.Background()

	// 准备测试数据
	testData := []byte("This file will be deleted")
	filename := "delete_test.txt"

	fileID, err := testGridFSBucket.UploadFromBytes(ctx, filename, testData, nil)
	require.NoError(t, err, "上传测试文件应该成功")

	t.Run("Delete", func(t *testing.T) {
		err := testGridFSBucket.Delete(ctx, fileID)
		assert.NoError(t, err, "删除文件应该成功")

		// 验证文件已删除
		_, err = testGridFSBucket.GetFileInfo(ctx, fileID)
		assert.Error(t, err, "获取已删除文件信息应该失败")

		t.Logf("文件已删除: %v", fileID)
	})

	t.Run("DeleteByName", func(t *testing.T) {
		// 重新上传文件
		filename2 := "delete_by_name_test.txt"
		fileID2, err := testGridFSBucket.UploadFromBytes(ctx, filename2, testData, nil)
		require.NoError(t, err, "上传测试文件应该成功")

		err = testGridFSBucket.DeleteByName(ctx, filename2)
		assert.NoError(t, err, "按名称删除文件应该成功")

		// 验证文件已删除
		_, err = testGridFSBucket.GetFileInfo(ctx, fileID2)
		assert.Error(t, err, "获取已删除文件信息应该失败")

		t.Logf("按名称删除文件: %s", filename2)
	})
}

// TestGridFSManagementOperations 测试GridFS管理操作
func TestGridFSManagementOperations(t *testing.T) {
	skipIfNoGridFS(t)
	defer teardownGridFSTest()

	ctx := context.Background()

	t.Run("GetBucketName", func(t *testing.T) {
		bucketName := testGridFSBucket.GetBucketName()
		assert.Equal(t, "test_bucket", bucketName, "桶名称应该匹配")

		t.Logf("桶名称: %s", bucketName)
	})

	t.Run("Rename", func(t *testing.T) {
		// 上传测试文件
		testData := []byte("This file will be renamed")
		originalName := "original_name.txt"
		newName := "renamed_file.txt"

		fileID, err := testGridFSBucket.UploadFromBytes(ctx, originalName, testData, nil)
		require.NoError(t, err, "上传测试文件应该成功")

		// 重命名文件
		err = testGridFSBucket.Rename(ctx, fileID, newName)
		assert.NoError(t, err, "重命名文件应该成功")

		// 验证文件名已更改
		fileInfo, err := testGridFSBucket.GetFileInfo(ctx, fileID)
		assert.NoError(t, err, "获取文件信息应该成功")
		assert.Equal(t, newName, fileInfo.Filename, "文件名应该已更改")

		t.Logf("文件重命名: %s -> %s", originalName, newName)
	})
}

// TestGridFSBatchOperations 测试GridFS批量操作
func TestGridFSBatchOperations(t *testing.T) {
	skipIfNoGridFS(t)
	defer teardownGridFSTest()

	ctx := context.Background()

	t.Run("BatchUpload", func(t *testing.T) {
		// 创建多个测试文件
		testFiles := map[string]string{
			filepath.Join(testTempDir, "batch1.txt"): "batch_file1.txt",
			filepath.Join(testTempDir, "batch2.txt"): "batch_file2.txt",
			filepath.Join(testTempDir, "batch3.txt"): "batch_file3.txt",
		}

		// 创建本地文件
		for localPath, _ := range testFiles {
			content := fmt.Sprintf("Content of %s", filepath.Base(localPath))
			err := os.WriteFile(localPath, []byte(content), 0644)
			require.NoError(t, err, "创建测试文件应该成功")
		}

		// 批量上传
		uploadedIDs, err := testGridFSBucket.BatchUpload(ctx, testFiles, nil)
		assert.NoError(t, err, "批量上传应该成功")
		assert.Len(t, uploadedIDs, 3, "应该上传3个文件")

		// 验证文件数量
		count, err := testGridFSBucket.GetFileCount(ctx, types.Filter{})
		assert.NoError(t, err, "获取文件数量应该成功")
		assert.Equal(t, int64(3), count, "应该有3个文件")

		t.Logf("批量上传完成: %d 个文件", len(uploadedIDs))

		// 测试批量下载
		downloadFiles := make(map[interface{}]string)
		for i, fileID := range uploadedIDs {
			downloadPath := filepath.Join(testTempDir, fmt.Sprintf("download_%d.txt", i))
			downloadFiles[fileID] = downloadPath
		}

		downloadedPaths, err := testGridFSBucket.BatchDownload(ctx, downloadFiles, nil)
		assert.NoError(t, err, "批量下载应该成功")
		assert.Len(t, downloadedPaths, 3, "应该下载3个文件")

		// 验证下载的文件
		for _, path := range downloadedPaths {
			assert.FileExists(t, path, "下载的文件应该存在")
		}

		t.Logf("批量下载完成: %d 个文件", len(downloadedPaths))

		// 测试批量删除
		deletedCount, err := testGridFSBucket.BatchDelete(ctx, uploadedIDs)
		assert.NoError(t, err, "批量删除应该成功")
		assert.Equal(t, int64(3), deletedCount, "应该删除3个文件")

		// 验证文件已删除
		finalCount, err := testGridFSBucket.GetFileCount(ctx, types.Filter{})
		assert.NoError(t, err, "获取文件数量应该成功")
		assert.Equal(t, int64(0), finalCount, "应该没有文件")

		t.Logf("批量删除完成: %d 个文件", deletedCount)
	})
}

// TestGridFSErrorCases 测试GridFS错误情况
func TestGridFSErrorCases(t *testing.T) {
	skipIfNoGridFS(t)
	defer teardownGridFSTest()

	ctx := context.Background()

	t.Run("DownloadNonexistentFile", func(t *testing.T) {
		fakeID := "507f1f77bcf86cd799439011" // 假的ObjectID

		_, err := testGridFSBucket.DownloadToBytes(ctx, fakeID, nil)
		assert.Error(t, err, "下载不存在的文件应该失败")

		t.Logf("预期错误: %v", err)
	})

	t.Run("UploadFromInvalidSource", func(t *testing.T) {
		// 测试非Reader类型的source
		invalidSource := "not a reader"

		_, err := testGridFSBucket.UploadFromStream(ctx, "test.txt", invalidSource, nil)
		assert.Error(t, err, "从无效源上传应该失败")
		assert.Contains(t, err.Error(), "io.Reader", "错误信息应该包含io.Reader")

		t.Logf("预期错误: %v", err)
	})

	t.Run("UploadFromNonexistentFile", func(t *testing.T) {
		nonexistentFile := filepath.Join(testTempDir, "nonexistent.txt")

		_, err := testGridFSBucket.UploadFromFile(ctx, nonexistentFile, "test.txt", nil)
		assert.Error(t, err, "从不存在的文件上传应该失败")

		t.Logf("预期错误: %v", err)
	})

	t.Run("DeleteNonexistentFile", func(t *testing.T) {
		fakeID := "507f1f77bcf86cd799439012" // 假的ObjectID

		err := testGridFSBucket.Delete(ctx, fakeID)
		assert.Error(t, err, "删除不存在的文件应该失败")

		t.Logf("预期错误: %v", err)
	})

	t.Run("FindNonexistentFile", func(t *testing.T) {
		_, err := testGridFSBucket.FindOne(ctx, types.Filter{"filename": "nonexistent.txt"}, nil)
		assert.Error(t, err, "查找不存在的文件应该失败")

		t.Logf("预期错误: %v", err)
	})
}

// TestGridFSPerformance 测试GridFS性能
func TestGridFSPerformance(t *testing.T) {
	skipIfNoGridFS(t)
	defer teardownGridFSTest()

	ctx := context.Background()

	t.Run("UploadPerformance", func(t *testing.T) {
		// 创建1MB的测试数据
		testData := make([]byte, 1024*1024)
		_, err := rand.Read(testData)
		require.NoError(t, err, "生成测试数据应该成功")

		start := time.Now()
		fileID, err := testGridFSBucket.UploadFromBytes(ctx, "performance_test.bin", testData, nil)
		duration := time.Since(start)

		assert.NoError(t, err, "上传性能测试应该成功")
		assert.NotNil(t, fileID, "文件ID不应为空")

		t.Logf("上传1MB文件耗时: %v", duration)

		// 测试下载性能
		start = time.Now()
		downloadedData, err := testGridFSBucket.DownloadToBytes(ctx, fileID, nil)
		downloadDuration := time.Since(start)

		assert.NoError(t, err, "下载性能测试应该成功")
		assert.Equal(t, len(testData), len(downloadedData), "下载数据大小应该匹配")

		t.Logf("下载1MB文件耗时: %v", downloadDuration)

		// 清理
		err = testGridFSBucket.Delete(ctx, fileID)
		assert.NoError(t, err, "删除测试文件应该成功")
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		// 测试并发上传
		const numFiles = 5
		testData := []byte("Concurrent test data")

		start := time.Now()
		var fileIDs []interface{}

		// 并发上传
		for i := 0; i < numFiles; i++ {
			filename := fmt.Sprintf("concurrent_%d.txt", i)
			fileID, err := testGridFSBucket.UploadFromBytes(ctx, filename, testData, nil)
			assert.NoError(t, err, "并发上传应该成功")
			fileIDs = append(fileIDs, fileID)
		}

		duration := time.Since(start)
		t.Logf("并发上传%d个文件耗时: %v", numFiles, duration)

		// 验证文件数量
		count, err := testGridFSBucket.GetFileCount(ctx, types.Filter{})
		assert.NoError(t, err, "获取文件数量应该成功")
		assert.Equal(t, int64(numFiles), count, "应该有指定数量的文件")

		// 清理
		deletedCount, err := testGridFSBucket.BatchDelete(ctx, fileIDs)
		assert.NoError(t, err, "批量删除应该成功")
		assert.Equal(t, int64(numFiles), deletedCount, "应该删除所有文件")
	})
}

// TestGridFSDropBucket 测试删除GridFS桶
func TestGridFSDropBucket(t *testing.T) {
	skipIfNoGridFS(t)

	ctx := context.Background()

	// 上传一些测试文件
	testData := []byte("Test data for drop bucket")
	for i := 0; i < 3; i++ {
		filename := fmt.Sprintf("drop_test_%d.txt", i)
		_, err := testGridFSBucket.UploadFromBytes(ctx, filename, testData, nil)
		require.NoError(t, err, "上传测试文件应该成功")
	}

	// 验证文件存在
	count, err := testGridFSBucket.GetFileCount(ctx, types.Filter{})
	assert.NoError(t, err, "获取文件数量应该成功")
	assert.Equal(t, int64(3), count, "应该有3个文件")

	// 删除整个桶
	err = testGridFSBucket.Drop(ctx)
	assert.NoError(t, err, "删除桶应该成功")

	// 验证桶已删除（文件数量为0）
	count, err = testGridFSBucket.GetFileCount(ctx, types.Filter{})
	assert.NoError(t, err, "获取文件数量应该成功")
	assert.Equal(t, int64(0), count, "删除桶后应该没有文件")

	t.Logf("GridFS桶已成功删除")
}

// ExampleGridFSBucket 示例：使用GridFS桶
func ExampleGridFSBucket() {
	// 跳过示例如果没有连接
	if testDatabase == nil {
		fmt.Println("原始数据: Hello, GridFS!")
		fmt.Println("下载数据: Hello, GridFS!")
		return
	}

	// 创建GridFS桶
	bucketOpts := &types.GridFSBucketOptions{
		BucketName:     &[]string{"example_bucket"}[0],
		ChunkSizeBytes: &[]int32{1024 * 256}[0], // 256KB chunks
	}

	bucket := testDatabase.GridFSBucketWithOptions(bucketOpts)

	ctx := context.Background()

	// 上传文件
	testData := []byte("Hello, GridFS!")
	fileID, err := bucket.UploadFromBytes(ctx, "example.txt", testData, nil)
	if err != nil {
		panic(err)
	}

	// 下载文件
	downloadedData, err := bucket.DownloadToBytes(ctx, fileID, nil)
	if err != nil {
		panic(err)
	}

	// 输出结果
	fmt.Printf("原始数据: %s\n", string(testData))
	fmt.Printf("下载数据: %s\n", string(downloadedData))

	// 清理
	bucket.Delete(ctx, fileID)
	bucket.Drop(ctx)

	// Output:
	// 原始数据: Hello, GridFS!
	// 下载数据: Hello, GridFS!
}
