package sftp_test

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"gohub/pkg/plugin/tools/common"
	"gohub/pkg/plugin/tools/configs"
	"gohub/pkg/plugin/tools/sftp"
)

// ExampleNewClient 演示如何创建和使用SFTP客户端
func ExampleNewClient() {
	// 创建SFTP配置
	config := &configs.SFTPConfig{
		Host:     "example.com",
		Port:     22,
		Username: "testuser",
		
		// 配置密码认证
		PasswordAuth: &configs.PasswordAuthConfig{
			Password:           "password123",
			AllowEmptyPassword: false,
			MaxRetries:         3,
		},
		
		// 配置连接参数
		ConnectTimeout:       30 * time.Second,
		KeepAliveInterval:    30 * time.Second,
		MaxReconnectAttempts: 3,
		AutoReconnect:        true,
		
		// 配置传输选项
		DefaultTransferOptions: &configs.SFTPTransferOptions{
			OverwriteExisting:      false,
			SkipExisting:           true,
			CreateTargetDir:        true,
			PreservePermissions:    true,
			PreserveTimestamps:     true,
			BufferSize:             64 * 1024, // 64KB
			ProgressReportInterval: time.Second,
		},
	}
	
	// 创建客户端
	client, err := sftp.NewClient(config)
	if err != nil {
		log.Fatalf("创建SFTP客户端失败: %v", err)
	}
	defer client.Close()
	
	// 设置进度回调
	client.SetProgressCallback(func(progress *common.TransferProgress) {
		fmt.Printf("传输进度: %.2f%% (%d/%d 字节) 速度: %.2f KB/s\n",
			progress.Percentage,
			progress.TransferredBytes,
			progress.TotalBytes,
			progress.Speed/1024)
	})
	
	// 设置错误回调
	client.SetErrorCallback(func(err *common.TransferError) {
		fmt.Printf("传输错误: %s - %s\n", err.Type, err.Message)
	})
	
	// 连接到服务器
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("连接到SFTP服务器失败: %v", err)
	}
	
	fmt.Println("成功连接到SFTP服务器")
	
	// 上传文件示例
	uploadExample(client, ctx)
	
	// 下载文件示例
	downloadExample(client, ctx)
	
	// 目录操作示例
	directoryExample(client, ctx)
	
	// 批量传输示例
	batchTransferExample(client, ctx)
}

// uploadExample 演示文件上传
func uploadExample(client sftp.Client, ctx context.Context) {
	fmt.Println("\n=== 文件上传示例 ===")
	
	// 上传单个文件
	result, err := client.UploadFile(ctx, "local_file.txt", "/remote/path/uploaded_file.txt", nil)
	if err != nil {
		fmt.Printf("上传文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("上传成功: %d 字节, 用时: %v, 平均速度: %.2f KB/s\n",
		result.BytesTransferred,
		result.Duration,
		result.AverageSpeed/1024)
}

// downloadExample 演示文件下载
func downloadExample(client sftp.Client, ctx context.Context) {
	fmt.Println("\n=== 文件下载示例 ===")
	
	// 下载单个文件
	result, err := client.DownloadFile(ctx, "/remote/path/file.txt", "local_downloaded_file.txt", nil)
	if err != nil {
		fmt.Printf("下载文件失败: %v\n", err)
		return
	}
	
	fmt.Printf("下载成功: %d 字节, 用时: %v, 平均速度: %.2f KB/s\n",
		result.BytesTransferred,
		result.Duration,
		result.AverageSpeed/1024)
}

// directoryExample 演示目录操作
func directoryExample(client sftp.Client, ctx context.Context) {
	fmt.Println("\n=== 目录操作示例 ===")
	
	// 创建目录
	if err := client.CreateDirectory(ctx, "/remote/new_directory", true); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		return
	}
	fmt.Println("创建目录成功")
	
	// 列出目录内容
	files, err := client.ListDirectory(ctx, "/remote")
	if err != nil {
		fmt.Printf("列出目录失败: %v\n", err)
		return
	}
	
	fmt.Printf("目录内容 (%d 项):\n", len(files))
	for _, file := range files {
		fileType := "文件"
		if file.IsDir {
			fileType = "目录"
		}
		fmt.Printf("  %s: %s (%d 字节)\n", fileType, file.Name, file.Size)
	}
	
	// 获取文件信息
	info, err := client.GetFileInfo(ctx, "/remote/file.txt")
	if err != nil {
		fmt.Printf("获取文件信息失败: %v\n", err)
		return
	}
	
	fmt.Printf("文件信息: %s, 大小: %d 字节, 修改时间: %v\n",
		info.Name, info.Size, info.ModTime)
}

// batchTransferExample 演示批量传输
func batchTransferExample(client sftp.Client, ctx context.Context) {
	fmt.Println("\n=== 批量传输示例 ===")
	
	// 准备批量传输操作
	operations := []*common.TransferOperation{
		{
			Type:       common.TransferTypeUpload,
			LocalPath:  "file1.txt",
			RemotePath: "/remote/file1.txt",
		},
		{
			Type:       common.TransferTypeUpload,
			LocalPath:  "file2.txt",
			RemotePath: "/remote/file2.txt",
		},
		{
			Type:       common.TransferTypeDownload,
			LocalPath:  "downloaded_file.txt",
			RemotePath: "/remote/existing_file.txt",
		},
	}
	
	// 执行批量传输
	result, err := client.BatchTransfer(ctx, operations, nil)
	if err != nil {
		fmt.Printf("批量传输失败: %v\n", err)
		return
	}
	
	fmt.Printf("批量传输完成:\n")
	fmt.Printf("  总操作数: %d\n", result.TotalOperations)
	fmt.Printf("  成功: %d, 失败: %d\n", result.SuccessfulOperations, result.FailedOperations)
	fmt.Printf("  总传输字节: %d\n", result.TotalBytesTransferred)
	fmt.Printf("  总用时: %v\n", result.TotalDuration)
	fmt.Printf("  平均速度: %.2f KB/s\n", result.AverageSpeed/1024)
	
	if len(result.Errors) > 0 {
		fmt.Println("错误信息:")
		for _, errMsg := range result.Errors {
			fmt.Printf("  - %s\n", errMsg)
		}
	}
}

// ExampleSFTPConfig 演示不同的配置选项
func ExampleSFTPConfig() {
	// 使用公钥认证的配置
	publicKeyConfig := &configs.SFTPConfig{
		Host:     "secure-server.com",
		Port:     2222,
		Username: "user",
		
		// 配置公钥认证
		PublicKeyAuth: &configs.PublicKeyAuthConfig{
			PrivateKeyPath: "/path/to/private/key",
			Passphrase:     "key_passphrase",
		},
		
		// 配置主机密钥验证
		HostKeyVerification: &configs.HostKeyVerificationConfig{
			Enabled:               true,
			StrictHostKeyChecking: true,
			KnownHostsFile:        "/path/to/known_hosts",
		},
		
		// 配置性能参数
		ConcurrentTransfers: 5,
		BufferSize:          128 * 1024, // 128KB
		
		// 配置自动重连
		AutoReconnect:        true,
		MaxReconnectAttempts: 5,
		ReconnectInterval:    10 * time.Second,
	}
	
	client, err := sftp.NewClient(publicKeyConfig)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()
	
	fmt.Println("使用公钥认证的SFTP客户端创建成功")
}

// ExampleAdvancedTransferOptions 演示高级传输选项
func ExampleAdvancedTransferOptions() {
	config := configs.DefaultSFTPConfig()
	config.Host = "example.com"
	config.Username = "user"
	config.PasswordAuth = &configs.PasswordAuthConfig{Password: "password"}
	
	client, err := sftp.NewClient(config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}
	defer client.Close()
	
	// 高级传输选项
	transferOptions := &configs.SFTPTransferOptions{
		OverwriteExisting:   true,  // 覆盖已存在的文件
		CreateTargetDir:     true,  // 自动创建目标目录
		PreservePermissions: true,  // 保持文件权限
		PreserveTimestamps:  true,  // 保持时间戳
		VerifyIntegrity:     true,  // 验证传输完整性
		
		// 性能优化
		BufferSize:          256 * 1024, // 256KB缓冲区
		ConcurrentTransfers: 3,          // 3个并发传输
		
		// 重试设置
		RetryCount:    5,
		RetryInterval: 2 * time.Second,
		
		// 文件过滤
		MaxFileSize:       100 * 1024 * 1024, // 最大100MB
		IncludePatterns:   []string{"*.txt", "*.log"},
		ExcludePatterns:   []string{"*.tmp", "*.bak"},
		IncludeExtensions: []string{".txt", ".log", ".json"},
		ExcludeExtensions: []string{".tmp", ".bak", ".cache"},
		
		// 进度报告
		ProgressReportInterval: 500 * time.Millisecond,
		
		// 错误处理
		ContinueOnError: true, // 出错时继续执行
	}
	
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	
	// 使用高级选项上传文件
	result, err := client.UploadFile(ctx, "large_file.txt", "/remote/large_file.txt", transferOptions)
	if err != nil {
		fmt.Printf("上传失败: %v\n", err)
		return
	}
	
	fmt.Printf("高级选项上传完成: %d 字节\n", result.BytesTransferred)
} 