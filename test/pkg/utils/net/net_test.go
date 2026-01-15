package net_test

import (
	"os"
	"strings"
	"testing"

	netutil "gateway/pkg/utils/net"
)

func TestGetAllMACAddresses(t *testing.T) {
	result := netutil.GetAllMACAddresses()
	// MAC地址可能为空（某些环境无网卡），但不应该panic
	t.Logf("GetAllMACAddresses: %s", result)

	if result != "" {
		// 如果有结果，验证格式（MAC地址格式：xx:xx:xx:xx:xx:xx）
		macs := strings.Split(result, ",")
		for _, mac := range macs {
			if len(mac) < 17 {
				t.Errorf("MAC地址格式不正确: %s", mac)
			}
		}
	}
}

func TestGetFirstIPv4Address(t *testing.T) {
	result := netutil.GetFirstIPv4Address()
	t.Logf("GetFirstIPv4Address: %s", result)

	// 应该总是返回一个有效的IPv4地址（至少是127.0.0.1）
	if result == "" {
		t.Error("GetFirstIPv4Address 返回空字符串")
	}

	// 验证是IPv4格式
	parts := strings.Split(result, ".")
	if len(parts) != 4 {
		t.Errorf("不是有效的IPv4地址: %s", result)
	}
}

func TestGetFirstIPv6Address(t *testing.T) {
	result := netutil.GetFirstIPv6Address()
	t.Logf("GetFirstIPv6Address: %s", result)

	// 应该总是返回一个有效的IPv6地址（至少是::1）
	if result == "" {
		t.Error("GetFirstIPv6Address 返回空字符串")
	}

	// 验证包含冒号（IPv6格式）
	if !strings.Contains(result, ":") {
		t.Errorf("不是有效的IPv6地址: %s", result)
	}
}

func TestGetAllIPv4Addresses(t *testing.T) {
	result := netutil.GetAllIPv4Addresses()
	t.Logf("GetAllIPv4Addresses: %v", result)

	// 可能为空，但不应该panic
	for _, ip := range result {
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			t.Errorf("不是有效的IPv4地址: %s", ip)
		}
	}
}

func TestGetHostname(t *testing.T) {
	result := netutil.GetHostname()
	t.Logf("GetHostname: %s", result)

	// 验证与os.Hostname()结果一致
	expected, err := os.Hostname()
	if err == nil && result != expected {
		t.Errorf("GetHostname 结果不一致: got %s, want %s", result, expected)
	}

	// 主机名不应该为空（正常情况下）
	if result == "" {
		t.Log("警告: 主机名为空")
	}
}

func TestGetAllMACAddresses_NotPanic(t *testing.T) {
	// 确保在各种环境下不会panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetAllMACAddresses panic: %v", r)
		}
	}()
	_ = netutil.GetAllMACAddresses()
}

func TestGetFirstIPv4Address_NotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetFirstIPv4Address panic: %v", r)
		}
	}()
	_ = netutil.GetFirstIPv4Address()
}

func TestGetFirstIPv6Address_NotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetFirstIPv6Address panic: %v", r)
		}
	}()
	_ = netutil.GetFirstIPv6Address()
}

func TestGetHostname_NotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetHostname panic: %v", r)
		}
	}()
	_ = netutil.GetHostname()
}

// BenchmarkGetAllMACAddresses 性能测试
func BenchmarkGetAllMACAddresses(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = netutil.GetAllMACAddresses()
	}
}

func BenchmarkGetFirstIPv4Address(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = netutil.GetFirstIPv4Address()
	}
}

func BenchmarkGetHostname(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = netutil.GetHostname()
	}
}
