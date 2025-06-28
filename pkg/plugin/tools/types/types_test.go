// Package types 类型定义测试
package types

import (
	"testing"
	"time"
)

// TestSyncMode 测试同步模式枚举
func TestSyncMode(t *testing.T) {
	tests := []struct {
		mode     SyncMode
		expected string
	}{
		{SyncModeUpload, "upload"},
		{SyncModeDownload, "download"},
		{SyncModeBidirectional, "bidirectional"},
		{SyncMode(999), "unknown"},
	}
	
	for _, test := range tests {
		if got := test.mode.String(); got != test.expected {
			t.Errorf("SyncMode.String() = %v, want %v", got, test.expected)
		}
	}
}

// TestTransferType 测试传输类型枚举
func TestTransferType(t *testing.T) {
	tests := []struct {
		transferType TransferType
		expected     string
	}{
		{TransferTypeUpload, "upload"},
		{TransferTypeDownload, "download"},
		{TransferType(999), "unknown"},
	}
	
	for _, test := range tests {
		if got := test.transferType.String(); got != test.expected {
			t.Errorf("TransferType.String() = %v, want %v", got, test.expected)
		}
	}
}

// TestTransferResult 测试传输结果结构
func TestTransferResult(t *testing.T) {
	result := &TransferResult{
		OperationID:      "test-op-001",
		Type:             TransferTypeUpload,
		LocalPath:        "/local/test.txt",
		RemotePath:       "/remote/test.txt",
		BytesTransferred: 1024,
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(time.Second),
		Duration:         time.Second,
		AverageSpeed:     1024.0,
		Success:          true,
	}
	
	if result.OperationID != "test-op-001" {
		t.Errorf("Expected OperationID to be 'test-op-001', got %v", result.OperationID)
	}
	
	if result.Type != TransferTypeUpload {
		t.Errorf("Expected Type to be TransferTypeUpload, got %v", result.Type)
	}
	
	if !result.Success {
		t.Errorf("Expected Success to be true, got %v", result.Success)
	}
}

// TestFileInfo 测试文件信息结构
func TestFileInfo(t *testing.T) {
	info := &FileInfo{
		Name:    "test.txt",
		Path:    "/path/to/test.txt",
		Size:    1024,
		IsDir:   false,
		ModTime: time.Now(),
	}
	
	if info.Name != "test.txt" {
		t.Errorf("Expected Name to be 'test.txt', got %v", info.Name)
	}
	
	if info.Size != 1024 {
		t.Errorf("Expected Size to be 1024, got %v", info.Size)
	}
	
	if info.IsDir {
		t.Errorf("Expected IsDir to be false, got %v", info.IsDir)
	}
} 