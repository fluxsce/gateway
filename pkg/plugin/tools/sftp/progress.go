package sftp

import (
	"context"
	"fmt"
	"io"
	"time"

	"gateway/pkg/plugin/tools/types"
)

// progressMonitor 进度监控器
// 用于跟踪文件传输进度并定期报告进度信息
type progressMonitor struct {
	// 上下文，用于取消监控
	ctx context.Context

	// 操作ID，用于标识具体的传输操作
	operationID string

	// 文件路径，用于标识正在传输的文件
	filePath string

	// 总字节数
	totalBytes int64

	// 已传输字节数
	transferredBytes int64

	// 传输开始时间
	startTime time.Time

	// 上次报告进度的时间
	lastReportTime time.Time

	// 进度回调函数
	progressCallback types.ProgressCallback

	// 报告间隔
	reportInterval time.Duration
}

// update 更新传输进度
// 当有新的数据被传输时调用此方法更新进度
// 参数:
//
//	n: 本次传输的字节数
func (pm *progressMonitor) update(n int) {
	// 更新已传输字节数
	pm.transferredBytes += int64(n)

	// 检查是否需要报告进度
	now := time.Now()
	shouldReport := false

	// 如果达到报告间隔或传输完成，则报告进度
	if pm.progressCallback != nil {
		if now.Sub(pm.lastReportTime) >= pm.reportInterval || pm.transferredBytes == pm.totalBytes {
			shouldReport = true
		}
	}

	if !shouldReport {
		return
	}

	// 计算进度百分比
	percentage := float64(0)
	if pm.totalBytes > 0 {
		percentage = float64(pm.transferredBytes) * 100 / float64(pm.totalBytes)
	}

	// 计算传输速度（字节/秒）
	elapsedSeconds := now.Sub(pm.startTime).Seconds()
	speed := float64(0)
	if elapsedSeconds > 0 {
		speed = float64(pm.transferredBytes) / elapsedSeconds
	}

	// 计算剩余时间
	var eta time.Duration
	if speed > 0 && pm.transferredBytes < pm.totalBytes {
		remainingBytes := pm.totalBytes - pm.transferredBytes
		etaSeconds := float64(remainingBytes) / speed
		eta = time.Duration(etaSeconds * float64(time.Second))
	}

	// 创建进度信息
	progress := &types.TransferProgress{
		OperationID:      pm.operationID,
		FilePath:         pm.filePath,
		TotalBytes:       pm.totalBytes,
		TransferredBytes: pm.transferredBytes,
		Percentage:       percentage,
		Speed:            speed,
		ETA:              eta,
		StartTime:        pm.startTime,
		CurrentTime:      now,
	}

	// 调用进度回调函数
	pm.progressCallback(progress)

	// 更新上次报告时间
	pm.lastReportTime = now
}

// progressReader 带进度监控的读取器
// 实现io.Reader接口，在读取数据的同时更新传输进度
type progressReader struct {
	// 底层读取器
	reader io.Reader

	// 进度监控器
	progress *progressMonitor
}

// Read 实现io.Reader接口
// 从底层读取器读取数据，并更新传输进度
// 参数:
//
//	p: 数据缓冲区
//
// 返回:
//
//	int: 实际读取的字节数
//	error: 读取过程中的错误
func (r *progressReader) Read(p []byte) (int, error) {
	// 从底层读取器读取数据
	n, err := r.reader.Read(p)

	// 如果读取了数据，更新进度
	if n > 0 {
		r.progress.update(n)
	}

	return n, err
}

// progressWriter 带进度监控的写入器
// 实现io.Writer接口，在写入数据的同时更新传输进度
type progressWriter struct {
	// 底层写入器
	writer io.Writer

	// 进度监控器
	progress *progressMonitor
}

// Write 实现io.Writer接口
// 向底层写入器写入数据，并更新传输进度
// 参数:
//
//	p: 要写入的数据
//
// 返回:
//
//	int: 实际写入的字节数
//	error: 写入过程中的错误
func (w *progressWriter) Write(p []byte) (int, error) {
	// 向底层写入器写入数据
	n, err := w.writer.Write(p)

	// 如果写入了数据，更新进度
	if n > 0 {
		w.progress.update(n)
	}

	return n, err
}

// createProgressMonitor 创建进度监控器
// 为文件传输操作创建一个新的进度监控器
// 参数:
//
//	ctx: 上下文
//	operationID: 操作ID
//	filePath: 文件路径
//	totalBytes: 文件总大小
//	startTime: 传输开始时间
//	callback: 进度回调函数
//	interval: 报告间隔
//
// 返回:
//
//	*progressMonitor: 创建的进度监控器
func createProgressMonitor(ctx context.Context, operationID, filePath string, totalBytes int64, startTime time.Time, callback types.ProgressCallback, interval time.Duration) *progressMonitor {
	// 如果未指定报告间隔，使用默认值
	if interval <= 0 {
		interval = time.Second // 默认每秒报告一次
	}

	return &progressMonitor{
		ctx:              ctx,
		operationID:      operationID,
		filePath:         filePath,
		totalBytes:       totalBytes,
		transferredBytes: 0,
		startTime:        startTime,
		lastReportTime:   startTime,
		progressCallback: callback,
		reportInterval:   interval,
	}
}

// formatSpeed 格式化传输速度
// 将字节/秒的速度格式化为人类可读的字符串
// 参数:
//
//	bytesPerSecond: 字节/秒
//
// 返回:
//
//	string: 格式化后的速度字符串
func formatSpeed(bytesPerSecond float64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytesPerSecond >= GB:
		return fmt.Sprintf("%.2f GB/s", bytesPerSecond/GB)
	case bytesPerSecond >= MB:
		return fmt.Sprintf("%.2f MB/s", bytesPerSecond/MB)
	case bytesPerSecond >= KB:
		return fmt.Sprintf("%.2f KB/s", bytesPerSecond/KB)
	default:
		return fmt.Sprintf("%.0f B/s", bytesPerSecond)
	}
}

// formatETA 格式化剩余时间
// 将剩余时间格式化为人类可读的字符串
// 参数:
//
//	eta: 剩余时间
//
// 返回:
//
//	string: 格式化后的时间字符串
func formatETA(eta time.Duration) string {
	if eta <= 0 {
		return "完成"
	}

	hours := int(eta.Hours())
	minutes := int(eta.Minutes()) % 60
	seconds := int(eta.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d:%02d", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}
