package collector

import (
	"testing"
	"time"

	"gohub/pkg/metric/collector/disk"
	"gohub/pkg/metric/types"

	"github.com/stretchr/testify/assert"
)

func TestDiskCollector_Basic(t *testing.T) {
	collector := disk.NewDiskCollector()
	assert.NotNil(t, collector)
	assert.Equal(t, types.CollectorNameDisk, collector.GetName())
	assert.Equal(t, disk.DefaultCollectTimeout, collector.GetTimeout())
}

func TestDiskCollector_Collect(t *testing.T) {
	collector := disk.NewDiskCollector()
	collector.SetEnabled(true)

	// 执行采集
	result, err := collector.Collect()
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// 验证返回类型
	metrics, ok := result.(*types.DiskMetrics)
	assert.True(t, ok)
	assert.NotNil(t, metrics)

	// 打印分区信息
	t.Log("\n=== 磁盘分区信息 ===")
	for _, partition := range metrics.Partitions {
		t.Logf("设备: %s", partition.Device)
		t.Logf("挂载点: %s", partition.MountPoint)
		t.Logf("文件系统: %s", partition.FileSystem)
		t.Logf("总空间: %.2f GB", float64(partition.Total)/(1024*1024*1024))
		t.Logf("已用空间: %.2f GB", float64(partition.Used)/(1024*1024*1024))
		t.Logf("剩余空间: %.2f GB", float64(partition.Free)/(1024*1024*1024))
		t.Logf("使用率: %.2f%%", partition.UsagePercent)
		t.Logf("---")
	}

	// 打印IO统计信息
	t.Log("\n=== 磁盘IO统计(累计值) ===")
	for _, iostat := range metrics.IOStats {
		t.Logf("设备: %s", iostat.Device)
		t.Logf("读取次数: %d", iostat.ReadCount)
		t.Logf("写入次数: %d", iostat.WriteCount)
		t.Logf("读取字节: %.2f GB", float64(iostat.ReadBytes)/(1024*1024*1024))
		t.Logf("写入字节: %.2f GB", float64(iostat.WriteBytes)/(1024*1024*1024))
		t.Logf("读取时间: %.2f s", float64(iostat.ReadTime)/1000)
		t.Logf("写入时间: %.2f s", float64(iostat.WriteTime)/1000)
		t.Logf("正在进行的IO: %d", iostat.IOInProgress)
		t.Logf("IO总时间: %.2f s", float64(iostat.IOTime)/1000)
		t.Logf("---")
	}
}

func TestDiskCollector_IORate(t *testing.T) {
	collector := disk.NewDiskCollector()
	collector.SetEnabled(true)

	// 第一次采集
	result1, err := collector.GetDiskIOStats()
	assert.NoError(t, err)
	
	// 等待一段时间
	time.Sleep(2 * time.Second)
	
	// 第二次采集
	result2, err := collector.GetDiskIOStats()
	assert.NoError(t, err)

	// 计算这段时间内的IO速率
	t.Log("\n=== 磁盘IO速率(2秒内) ===")
	for i := range result1 {
		if i >= len(result2) {
			continue
		}
		device := result1[i].Device
		readBytesRate := float64(result2[i].ReadBytes-result1[i].ReadBytes) / (2 * 1024 * 1024) // MB/s
		writeBytesRate := float64(result2[i].WriteBytes-result1[i].WriteBytes) / (2 * 1024 * 1024) // MB/s
		
		t.Logf("设备: %s", device)
		t.Logf("读取速率: %.2f MB/s", readBytesRate)
		t.Logf("写入速率: %.2f MB/s", writeBytesRate)
		t.Logf("---")
	}
}

func TestDiskCollector_VirtualFS(t *testing.T) {
	// 测试虚拟文件系统检测
	testCases := []struct {
		device     string
		filesystem string
		expected   bool
	}{
		{"/dev/sda1", "ext4", false},
		{"/dev/loop0", "squashfs", true},
		{"none", "proc", true},
		{"/dev/nvme0n1p1", "ntfs", false},
		{"udev", "devtmpfs", true},
	}

	for _, tc := range testCases {
		actual := disk.IsVirtualFileSystem(tc.device, tc.filesystem)
		assert.Equal(t, tc.expected, actual, 
			"设备:%s, 文件系统:%s 应该返回 %v", 
			tc.device, tc.filesystem, tc.expected)
	}
}

func TestDiskCollector_Partition(t *testing.T) {
	// 测试分区检测
	testCases := []struct {
		device   string
		expected bool
	}{
		{"sda", false},
		{"sda1", true},
		{"nvme0n1", false},
		{"nvme0n1p1", true},
		{"", false},
	}

	for _, tc := range testCases {
		actual := disk.IsPartition(tc.device)
		assert.Equal(t, tc.expected, actual,
			"设备:%s 应该返回 %v",
			tc.device, tc.expected)
	}
}
