package filewrite

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"gohub/internal/gateway/logwrite/types"
	"gohub/pkg/logger"
)

// FileWriter 文件日志写入器
// 支持日志轮转、压缩、缓冲写入和并发安全
type FileWriter struct {
	// 配置
	config     *types.LogConfig
	fileConfig *types.FileOutputConfig
	
	// 文件相关
	currentFile *os.File
	writer      *bufio.Writer
	
	// 轮转相关
	currentSize int64
	lastRotate  time.Time
	
	// 并发控制
	mutex sync.RWMutex
	
	// 缓冲控制
	buffer      []string
	bufferSize  int
	flushTicker *time.Ticker
	
	// 关闭控制
	closeChan chan struct{}
	wg        sync.WaitGroup
}

// NewFileWriter 创建新的文件日志写入器
func NewFileWriter(config *types.LogConfig) (*FileWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	fileConfig, err := config.GetFileConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse file config: %w", err)
	}

	// 设置默认值
	if fileConfig.MaxSize == 0 {
		fileConfig.MaxSize = 100 // 100MB
	}
	if fileConfig.MaxAge == 0 {
		fileConfig.MaxAge = 30 // 30天
	}
	if fileConfig.MaxBackups == 0 {
		fileConfig.MaxBackups = 10
	}
	if fileConfig.Extension == "" {
		fileConfig.Extension = ".log"
	}

	writer := &FileWriter{
		config:      config,
		fileConfig:  fileConfig,
		bufferSize:  config.BufferSize,
		buffer:      make([]string, 0, config.FlushThreshold),
		closeChan:   make(chan struct{}),
		lastRotate:  time.Now(),
	}

	// 创建日志目录
	if err := os.MkdirAll(fileConfig.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// 打开初始文件
	if err := writer.openFile(); err != nil {
		return nil, fmt.Errorf("failed to open initial log file: %w", err)
	}

	// 启动定时刷新
	writer.startFlushTicker()

	// 启动清理协程
	writer.wg.Add(1)
	go writer.cleanupRoutine()

	return writer, nil
}

// Write 写入单条日志
func (w *FileWriter) Write(ctx context.Context, log *types.AccessLog) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 格式化日志
	formatted := w.formatLog(log)
	
	// 检查是否需要轮转
	if w.needRotate(len(formatted)) {
		if err := w.rotate(); err != nil {
			logger.Error("Failed to rotate log file", "error", err)
		}
	}

	// 添加到缓冲区
	w.buffer = append(w.buffer, formatted)
	
	// 检查是否需要立即刷新
	if len(w.buffer) >= w.config.FlushThreshold {
		return w.flushBuffer()
	}

	return nil
}

// BatchWrite 批量写入日志
func (w *FileWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 预估总大小
	var totalSize int
	formattedLogs := make([]string, len(logs))
	for i, log := range logs {
		formatted := w.formatLog(log)
		formattedLogs[i] = formatted
		totalSize += len(formatted)
	}

	// 检查是否需要轮转
	if w.needRotate(totalSize) {
		if err := w.rotate(); err != nil {
			logger.Error("Failed to rotate log file", "error", err)
		}
	}

	// 添加到缓冲区
	w.buffer = append(w.buffer, formattedLogs...)

	// 立即刷新
	return w.flushBuffer()
}

// Flush 刷新缓冲区
func (w *FileWriter) Flush(ctx context.Context) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.flushBuffer()
}

// Close 关闭写入器
func (w *FileWriter) Close() error {
	close(w.closeChan)
	w.wg.Wait()

	w.mutex.Lock()
	defer w.mutex.Unlock()

	// 刷新剩余缓冲区
	if err := w.flushBuffer(); err != nil {
		logger.Error("Failed to flush buffer on close", "error", err)
	}

	// 停止定时器
	if w.flushTicker != nil {
		w.flushTicker.Stop()
	}

	// 关闭文件
	if w.writer != nil {
		w.writer.Flush()
	}
	if w.currentFile != nil {
		return w.currentFile.Close()
	}

	return nil
}

// formatLog 格式化日志
func (w *FileWriter) formatLog(log *types.AccessLog) string {
	switch types.LogFormat(w.config.LogFormat) {
	case types.LogFormatJSON:
		if jsonStr, err := log.ToJSON(w.config); err == nil {
			return jsonStr
		}
		return fmt.Sprintf(`{"error": "failed to format log"}`)
	case types.LogFormatCSV:
		return log.ToCSV(w.config)
	default:
		return log.ToText(w.config)
	}
}

// needRotate 检查是否需要轮转
func (w *FileWriter) needRotate(additionalSize int) bool {
	if w.currentFile == nil {
		return false
	}

	// 基于大小的轮转
	if w.fileConfig.MaxSize > 0 {
		maxBytes := int64(w.fileConfig.MaxSize) * 1024 * 1024
		if w.currentSize+int64(additionalSize) > maxBytes {
			return true
		}
	}

	// 基于时间的轮转
	if w.fileConfig.RotationTime != "" {
		switch w.fileConfig.RotationTime {
		case "hourly":
			return time.Since(w.lastRotate) >= time.Hour
		case "daily":
			return time.Since(w.lastRotate) >= 24*time.Hour
		}
	}

	return false
}

// rotate 执行日志轮转
func (w *FileWriter) rotate() error {
	// 刷新当前缓冲区
	if err := w.flushBuffer(); err != nil {
		return fmt.Errorf("failed to flush before rotation: %w", err)
	}

	// 关闭当前文件
	if w.writer != nil {
		w.writer.Flush()
	}
	if w.currentFile != nil {
		w.currentFile.Close()
	}

	// 重命名当前文件
	if err := w.renameCurrentFile(); err != nil {
		return fmt.Errorf("failed to rename current file: %w", err)
	}

	// 打开新文件
	if err := w.openFile(); err != nil {
		return fmt.Errorf("failed to open new file: %w", err)
	}

	// 清理旧文件
	go w.cleanupOldFiles()

	w.lastRotate = time.Now()
	return nil
}

// openFile 打开日志文件
func (w *FileWriter) openFile() error {
	filename := w.getLogFileName()
	filePath := filepath.Join(w.fileConfig.Path, filename)

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// 获取文件大小
	if stat, err := file.Stat(); err == nil {
		w.currentSize = stat.Size()
	}

	w.currentFile = file
	w.writer = bufio.NewWriterSize(file, w.bufferSize)

	return nil
}

// getLogFileName 生成日志文件名
func (w *FileWriter) getLogFileName() string {
	prefix := w.fileConfig.Prefix
	if prefix == "" {
		prefix = "gateway"
	}
	
	extension := w.fileConfig.Extension
	if extension == "" {
		extension = ".log"
	}

	if w.fileConfig.LocalTime {
		return fmt.Sprintf("%s_%s%s", prefix, time.Now().Format("20060102_150405"), extension)
	}
	return fmt.Sprintf("%s_%s%s", prefix, time.Now().UTC().Format("20060102_150405"), extension)
}

// renameCurrentFile 重命名当前文件
func (w *FileWriter) renameCurrentFile() error {
	if w.currentFile == nil {
		return nil
	}

	currentPath := w.currentFile.Name()
	
	// 生成备份文件名
	dir := filepath.Dir(currentPath)
	base := filepath.Base(currentPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("%s_%s%s", name, timestamp, ext)
	backupPath := filepath.Join(dir, backupName)

	// 重命名文件
	if err := os.Rename(currentPath, backupPath); err != nil {
		return err
	}

	// 如果启用压缩，压缩文件
	if w.fileConfig.Compress {
		go w.compressFile(backupPath)
	}

	return nil
}

// compressFile 压缩文件
func (w *FileWriter) compressFile(filePath string) {
	gzipPath := filePath + ".gz"
	
	// 打开原文件
	src, err := os.Open(filePath)
	if err != nil {
		logger.Error("Failed to open file for compression", "file", filePath, "error", err)
		return
	}
	defer src.Close()

	// 创建gzip文件
	dst, err := os.Create(gzipPath)
	if err != nil {
		logger.Error("Failed to create gzip file", "file", gzipPath, "error", err)
		return
	}
	defer dst.Close()

	// 创建gzip写入器
	gzipWriter := gzip.NewWriter(dst)
	defer gzipWriter.Close()

	// 复制数据
	if _, err := io.Copy(gzipWriter, src); err != nil {
		logger.Error("Failed to compress file", "file", filePath, "error", err)
		return
	}

	// 删除原文件
	if err := os.Remove(filePath); err != nil {
		logger.Error("Failed to remove original file after compression", "file", filePath, "error", err)
	}
}

// flushBuffer 刷新缓冲区
func (w *FileWriter) flushBuffer() error {
	if len(w.buffer) == 0 || w.writer == nil {
		return nil
	}

	for _, line := range w.buffer {
		if _, err := w.writer.WriteString(line + "\n"); err != nil {
			return err
		}
		w.currentSize += int64(len(line) + 1)
	}

	// 清空缓冲区
	w.buffer = w.buffer[:0]

	// 刷新到磁盘
	return w.writer.Flush()
}

// startFlushTicker 启动定时刷新
func (w *FileWriter) startFlushTicker() {
	interval := time.Duration(w.config.AsyncFlushIntervalMs) * time.Millisecond
	if interval == 0 {
		interval = 5 * time.Second
	}

	w.flushTicker = time.NewTicker(interval)
	
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.flushTicker.C:
				w.mutex.Lock()
				w.flushBuffer()
				w.mutex.Unlock()
			case <-w.closeChan:
				return
			}
		}
	}()
}

// cleanupRoutine 清理协程
func (w *FileWriter) cleanupRoutine() {
	defer w.wg.Done()
	
	ticker := time.NewTicker(time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.cleanupOldFiles()
		case <-w.closeChan:
			return
		}
	}
}

// cleanupOldFiles 清理旧文件
func (w *FileWriter) cleanupOldFiles() {
	// 按数量清理
	if w.fileConfig.MaxBackups > 0 {
		w.cleanupByCount()
	}

	// 按时间清理
	if w.fileConfig.MaxAge > 0 {
		w.cleanupByAge()
	}
}

// cleanupByCount 按数量清理
func (w *FileWriter) cleanupByCount() {
	pattern := filepath.Join(w.fileConfig.Path, w.fileConfig.Prefix+"*"+w.fileConfig.Extension+"*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		logger.Error("Failed to list log files for cleanup", "error", err)
		return
	}

	if len(files) <= w.fileConfig.MaxBackups {
		return
	}

	// 按修改时间排序
	sort.Slice(files, func(i, j int) bool {
		statI, errI := os.Stat(files[i])
		statJ, errJ := os.Stat(files[j])
		if errI != nil || errJ != nil {
			return false
		}
		return statI.ModTime().Before(statJ.ModTime())
	})

	// 删除最旧的文件
	for i := 0; i < len(files)-w.fileConfig.MaxBackups; i++ {
		if err := os.Remove(files[i]); err != nil {
			logger.Error("Failed to remove old log file", "file", files[i], "error", err)
		}
	}
}

// cleanupByAge 按时间清理
func (w *FileWriter) cleanupByAge() {
	cutoff := time.Now().AddDate(0, 0, -w.fileConfig.MaxAge)
	
	pattern := filepath.Join(w.fileConfig.Path, w.fileConfig.Prefix+"*"+w.fileConfig.Extension+"*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		logger.Error("Failed to list log files for age cleanup", "error", err)
		return
	}

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}

		if stat.ModTime().Before(cutoff) {
			if err := os.Remove(file); err != nil {
				logger.Error("Failed to remove old log file", "file", file, "error", err)
			}
		}
	}
} 