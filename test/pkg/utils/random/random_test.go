package random_test

import (
	"net"
	"testing"

	"gateway/pkg/utils/random"
)

// TestGetNodeIP_NonEmpty 校验 GetNodeIP 在进程 init 后必有返回值（失败时 random 包回退为 127.0.0.1）。
func TestGetNodeIP_NonEmpty(t *testing.T) {
	ip := random.GetNodeIP()
	if ip == "" {
		t.Fatal("GetNodeIP 返回空字符串")
	}
	t.Logf("GetNodeIP: %q", ip)
}

// TestGetNodeIP_ValidIPv4 校验返回值为合法 IPv4。
func TestGetNodeIP_ValidIPv4(t *testing.T) {
	ip := random.GetNodeIP()
	parsed := net.ParseIP(ip)
	if parsed == nil {
		t.Fatalf("GetNodeIP 返回值无法解析为 IP: %q", ip)
	}
	if parsed.To4() == nil {
		t.Fatalf("GetNodeIP 应为 IPv4: %q", ip)
	}
}

// TestGetNodeIP_Consistent 校验节点 IP 为包级缓存，多次读取一致。
func TestGetNodeIP_Consistent(t *testing.T) {
	a := random.GetNodeIP()
	b := random.GetNodeIP()
	if a != b {
		t.Errorf("多次调用应返回同一缓存值: %q vs %q", a, b)
	}
}

// TestGetNodeIP_NotLinkLocalUnicast 不应把 169.254.x.x 等 APIPA 当作节点 IP；存在私网/公网单播时应选其一，否则回退 127.0.0.1。
func TestGetNodeIP_NotLinkLocalUnicast(t *testing.T) {
	got := random.GetNodeIP()
	ip := net.ParseIP(got)
	if ip == nil {
		t.Fatalf("无法解析: %q", got)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		t.Fatalf("应为 IPv4: %q", got)
	}
	if ip4.IsLinkLocalUnicast() {
		t.Errorf("不应选取链路本地 IPv4(如 169.254.x.x) 作为节点 IP: %s", got)
	}
}
